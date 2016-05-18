package ceph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	measurement = "ceph"
	typeMon     = "monitor"
	typeOsd     = "osd"
	osdPrefix   = "ceph-osd"
	monPrefix   = "ceph-mon"
	sockSuffix  = "asok"
)

type Ceph struct {
	CephBinary   string
	OsdPrefix    string
	MonPrefix    string
	SocketDir    string
	SocketSuffix string
}

func (c *Ceph) setDefaults() {
	if c.CephBinary == "" {
		c.CephBinary = "/usr/bin/ceph"
	}

	if c.OsdPrefix == "" {
		c.OsdPrefix = osdPrefix
	}

	if c.MonPrefix == "" {
		c.MonPrefix = monPrefix
	}

	if c.SocketDir == "" {
		c.SocketDir = "/var/run/ceph"
	}

	if c.SocketSuffix == "" {
		c.SocketSuffix = sockSuffix
	}
}

func (c *Ceph) Description() string {
	return "Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster."
}

var sampleConfig = `
  ## All configuration values are optional, defaults are shown below

  ## location of ceph binary
  ceph_binary = "/usr/bin/ceph"

  ## directory in which to look for socket files
  socket_dir = "/var/run/ceph"

  ## prefix of MON and OSD socket files, used to determine socket type
  mon_prefix = "ceph-mon"
  osd_prefix = "ceph-osd"

  ## suffix used to identify socket files
  socket_suffix = "asok"
`

func (c *Ceph) SampleConfig() string {
	return sampleConfig
}

func (c *Ceph) Gather(acc telegraf.Accumulator) error {
	c.setDefaults()
	sockets, err := findSockets(c)
	if err != nil {
		return fmt.Errorf("failed to find sockets at path '%s': %v", c.SocketDir, err)
	}

	for _, s := range sockets {
		dump, err := perfDump(c.CephBinary, s)
		if err != nil {
			log.Printf("error reading from socket '%s': %v", s.socket, err)
			continue
		}
		data, err := parseDump(dump)
		if err != nil {
			log.Printf("error parsing dump from socket '%s': %v", s.socket, err)
			continue
		}
		for tag, metrics := range *data {
			acc.AddFields(measurement,
				map[string]interface{}(metrics),
				map[string]string{"type": s.sockType, "id": s.sockId, "collection": tag})
		}
	}
	return nil
}

func init() {
	inputs.Add(measurement, func() telegraf.Input { return &Ceph{} })
}

var perfDump = func(binary string, socket *socket) (string, error) {
	cmdArgs := []string{"--admin-daemon", socket.socket}
	if socket.sockType == typeOsd {
		cmdArgs = append(cmdArgs, "perf", "dump")
	} else if socket.sockType == typeMon {
		cmdArgs = append(cmdArgs, "perfcounters_dump")
	} else {
		return "", fmt.Errorf("ignoring unknown socket type: %s", socket.sockType)
	}

	cmd := exec.Command(binary, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running ceph dump: %s", err)
	}

	return out.String(), nil
}

var findSockets = func(c *Ceph) ([]*socket, error) {
	listing, err := ioutil.ReadDir(c.SocketDir)
	if err != nil {
		return []*socket{}, fmt.Errorf("Failed to read socket directory '%s': %v", c.SocketDir, err)
	}
	sockets := make([]*socket, 0, len(listing))
	for _, info := range listing {
		f := info.Name()
		var sockType string
		var sockPrefix string
		if strings.HasPrefix(f, c.MonPrefix) {
			sockType = typeMon
			sockPrefix = monPrefix
		}
		if strings.HasPrefix(f, c.OsdPrefix) {
			sockType = typeOsd
			sockPrefix = osdPrefix

		}
		if sockType == typeOsd || sockType == typeMon {
			path := filepath.Join(c.SocketDir, f)
			sockets = append(sockets, &socket{parseSockId(f, sockPrefix, c.SocketSuffix), sockType, path})
		}
	}
	return sockets, nil
}

func parseSockId(fname, prefix, suffix string) string {
	s := fname
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, suffix)
	s = strings.Trim(s, ".-_")
	return s
}

type socket struct {
	sockId   string
	sockType string
	socket   string
}

type metric struct {
	pathStack []string // lifo stack of name components
	value     float64
}

// Pops names of pathStack to build the flattened name for a metric
func (m *metric) name() string {
	buf := bytes.Buffer{}
	for i := len(m.pathStack) - 1; i >= 0; i-- {
		if buf.Len() > 0 {
			buf.WriteString(".")
		}
		buf.WriteString(m.pathStack[i])
	}
	return buf.String()
}

type metricMap map[string]interface{}

type taggedMetricMap map[string]metricMap

// Parses a raw JSON string into a taggedMetricMap
// Delegates the actual parsing to newTaggedMetricMap(..)
func parseDump(dump string) (*taggedMetricMap, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(dump), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json: '%s': %v", dump, err)
	}

	tmm := newTaggedMetricMap(data)

	if err != nil {
		return nil, fmt.Errorf("failed to tag dataset: '%v': %v", tmm, err)
	}

	return tmm, nil
}

// Builds a TaggedMetricMap out of a generic string map.
// The top-level key is used as a tag and all sub-keys are flattened into metrics
func newTaggedMetricMap(data map[string]interface{}) *taggedMetricMap {
	tmm := make(taggedMetricMap)
	for tag, datapoints := range data {
		mm := make(metricMap)
		for _, m := range flatten(datapoints) {
			mm[m.name()] = m.value
		}
		tmm[tag] = mm
	}
	return &tmm
}

// Recursively flattens any k-v hierarchy present in data.
// Nested keys are flattened into ordered slices associated with a metric value.
// The key slices are treated as stacks, and are expected to be reversed and concatenated
// when passed as metrics to the accumulator. (see (*metric).name())
func flatten(data interface{}) []*metric {
	var metrics []*metric

	switch val := data.(type) {
	case float64:
		metrics = []*metric{&metric{make([]string, 0, 1), val}}
	case map[string]interface{}:
		metrics = make([]*metric, 0, len(val))
		for k, v := range val {
			for _, m := range flatten(v) {
				m.pathStack = append(m.pathStack, k)
				metrics = append(metrics, m)
			}
		}
	default:
		log.Printf("Ignoring unexpected type '%T' for value %v", val, val)
	}

	return metrics
}
