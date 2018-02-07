package ceph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
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
	CephBinary             string
	OsdPrefix              string
	MonPrefix              string
	SocketDir              string
	SocketSuffix           string
	CephUser               string
	CephConfig             string
	GatherAdminSocketStats bool
	GatherClusterStats     bool
}

func (c *Ceph) Description() string {
	return "Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster."
}

var sampleConfig = `
  ## This is the recommended interval to poll.  Too frequent and you will lose
  ## data points due to timeouts during rebalancing and recovery
  interval = '1m'

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

  ## Ceph user to authenticate as
  ceph_user = "client.admin"

  ## Ceph configuration to use to locate the cluster
  ceph_config = "/etc/ceph/ceph.conf"

  ## Whether to gather statistics via the admin socket
  gather_admin_socket_stats = true

  ## Whether to gather statistics via ceph commands
  gather_cluster_stats = false
`

func (c *Ceph) SampleConfig() string {
	return sampleConfig
}

func (c *Ceph) Gather(acc telegraf.Accumulator) error {
	if c.GatherAdminSocketStats {
		if err := c.gatherAdminSocketStats(acc); err != nil {
			return err
		}
	}

	if c.GatherClusterStats {
		if err := c.gatherClusterStats(acc); err != nil {
			return err
		}
	}

	return nil
}

func (c *Ceph) gatherAdminSocketStats(acc telegraf.Accumulator) error {
	sockets, err := findSockets(c)
	if err != nil {
		return fmt.Errorf("failed to find sockets at path '%s': %v", c.SocketDir, err)
	}

	for _, s := range sockets {
		dump, err := perfDump(c.CephBinary, s)
		if err != nil {
			acc.AddError(fmt.Errorf("E! error reading from socket '%s': %v", s.socket, err))
			continue
		}
		data, err := parseDump(dump)
		if err != nil {
			acc.AddError(fmt.Errorf("E! error parsing dump from socket '%s': %v", s.socket, err))
			continue
		}
		for tag, metrics := range data {
			acc.AddFields(measurement,
				map[string]interface{}(metrics),
				map[string]string{"type": s.sockType, "id": s.sockId, "collection": tag})
		}
	}
	return nil
}

func (c *Ceph) gatherClusterStats(acc telegraf.Accumulator) error {
	jobs := []struct {
		command string
		parser  func(telegraf.Accumulator, string) error
	}{
		{"status", decodeStatus},
		{"df", decodeDf},
		{"osd pool stats", decodeOsdPoolStats},
	}

	// For each job, execute against the cluster, parse and accumulate the data points
	for _, job := range jobs {
		output, err := c.exec(job.command)
		if err != nil {
			return fmt.Errorf("error executing command: %v", err)
		}
		err = job.parser(acc, output)
		if err != nil {
			return fmt.Errorf("error parsing output: %v", err)
		}
	}

	return nil
}

func init() {
	c := Ceph{
		CephBinary:             "/usr/bin/ceph",
		OsdPrefix:              osdPrefix,
		MonPrefix:              monPrefix,
		SocketDir:              "/var/run/ceph",
		SocketSuffix:           sockSuffix,
		CephUser:               "client.admin",
		CephConfig:             "/etc/ceph/ceph.conf",
		GatherAdminSocketStats: true,
		GatherClusterStats:     false,
	}

	inputs.Add(measurement, func() telegraf.Input { return &c })

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
func parseDump(dump string) (taggedMetricMap, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(dump), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json: '%s': %v", dump, err)
	}

	return newTaggedMetricMap(data), nil
}

// Builds a TaggedMetricMap out of a generic string map.
// The top-level key is used as a tag and all sub-keys are flattened into metrics
func newTaggedMetricMap(data map[string]interface{}) taggedMetricMap {
	tmm := make(taggedMetricMap)
	for tag, datapoints := range data {
		mm := make(metricMap)
		for _, m := range flatten(datapoints) {
			mm[m.name()] = m.value
		}
		tmm[tag] = mm
	}
	return tmm
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
		log.Printf("I! Ignoring unexpected type '%T' for value %v", val, val)
	}

	return metrics
}

func (c *Ceph) exec(command string) (string, error) {
	cmdArgs := []string{"--conf", c.CephConfig, "--name", c.CephUser, "--format", "json"}
	cmdArgs = append(cmdArgs, strings.Split(command, " ")...)

	cmd := exec.Command(c.CephBinary, cmdArgs...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running ceph %v: %s", command, err)
	}

	output := out.String()

	// Ceph doesn't sanitize its output, and may return invalid JSON.  Patch this
	// up for them, as having some inaccurate data is better than none.
	output = strings.Replace(output, "-inf", "0", -1)
	output = strings.Replace(output, "inf", "0", -1)

	return output, nil
}

func decodeStatus(acc telegraf.Accumulator, input string) error {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	err = decodeStatusOsdmap(acc, data)
	if err != nil {
		return err
	}

	err = decodeStatusPgmap(acc, data)
	if err != nil {
		return err
	}

	err = decodeStatusPgmapState(acc, data)
	if err != nil {
		return err
	}

	return nil
}

func decodeStatusOsdmap(acc telegraf.Accumulator, data map[string]interface{}) error {
	osdmap, ok := data["osdmap"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("WARNING %s - unable to decode osdmap", measurement)
	}
	fields, ok := osdmap["osdmap"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("WARNING %s - unable to decode osdmap", measurement)
	}
	acc.AddFields("ceph_osdmap", fields, map[string]string{})
	return nil
}

func decodeStatusPgmap(acc telegraf.Accumulator, data map[string]interface{}) error {
	pgmap, ok := data["pgmap"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("WARNING %s - unable to decode pgmap", measurement)
	}
	fields := make(map[string]interface{})
	for key, value := range pgmap {
		switch value.(type) {
		case float64:
			fields[key] = value
		}
	}
	acc.AddFields("ceph_pgmap", fields, map[string]string{})
	return nil
}

func extractPgmapStates(data map[string]interface{}) ([]interface{}, error) {
	const key = "pgs_by_state"

	pgmap, ok := data["pgmap"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("WARNING %s - unable to decode pgmap", measurement)
	}

	s, ok := pgmap[key]
	if !ok {
		return nil, fmt.Errorf("WARNING %s - pgmap is missing the %s field", measurement, key)
	}

	states, ok := s.([]interface{})
	if !ok {
		return nil, fmt.Errorf("WARNING %s - pgmap[%s] is not a list", measurement, key)
	}
	return states, nil
}

func decodeStatusPgmapState(acc telegraf.Accumulator, data map[string]interface{}) error {
	states, err := extractPgmapStates(data)
	if err != nil {
		return err
	}
	for _, state := range states {
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode pg state", measurement)
		}
		stateName, ok := stateMap["state_name"].(string)
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode pg state name", measurement)
		}
		stateCount, ok := stateMap["count"].(float64)
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode pg state count", measurement)
		}

		tags := map[string]string{
			"state": stateName,
		}
		fields := map[string]interface{}{
			"count": stateCount,
		}
		acc.AddFields("ceph_pgmap_state", fields, tags)
	}
	return nil
}

func decodeDf(acc telegraf.Accumulator, input string) error {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	// ceph.usage: records global utilization and number of objects
	stats_fields, ok := data["stats"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("WARNING %s - unable to decode df stats", measurement)
	}
	acc.AddFields("ceph_usage", stats_fields, map[string]string{})

	// ceph.pool.usage: records per pool utilization and number of objects
	pools, ok := data["pools"].([]interface{})
	if !ok {
		return fmt.Errorf("WARNING %s - unable to decode df pools", measurement)
	}

	for _, pool := range pools {
		pool_map, ok := pool.(map[string]interface{})
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode df pool", measurement)
		}
		pool_name, ok := pool_map["name"].(string)
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode df pool name", measurement)
		}
		fields, ok := pool_map["stats"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode df pool stats", measurement)
		}
		tags := map[string]string{
			"name": pool_name,
		}
		acc.AddFields("ceph_pool_usage", fields, tags)
	}

	return nil
}

func decodeOsdPoolStats(acc telegraf.Accumulator, input string) error {
	data := make([]map[string]interface{}, 0)
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	// ceph.pool.stats: records pre pool IO and recovery throughput
	for _, pool := range data {
		pool_name, ok := pool["pool_name"].(string)
		if !ok {
			return fmt.Errorf("WARNING %s - unable to decode osd pool stats name", measurement)
		}
		// Note: the 'recovery' object looks broken (in hammer), so it's omitted
		objects := []string{
			"client_io_rate",
			"recovery_rate",
		}
		fields := make(map[string]interface{})
		for _, object := range objects {
			perfdata, ok := pool[object].(map[string]interface{})
			if !ok {
				return fmt.Errorf("WARNING %s - unable to decode osd pool stats", measurement)
			}
			for key, value := range perfdata {
				fields[key] = value
			}
		}
		tags := map[string]string{
			"name": pool_name,
		}
		acc.AddFields("ceph_pool_stats", fields, tags)
	}

	return nil
}
