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
		metrics = []*metric{{make([]string, 0, 1), val}}
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

// exec executes the 'ceph' command with the supplied arguments, returning JSON formatted output
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

// CephStatus is used to unmarshal "ceph -s" output
type CephStatus struct {
	Health struct {
		Status        string `json:"status"`
		OverallStatus string `json:"overall_status"`
	} `json:"health"`
	OSDMap struct {
		OSDMap struct {
			Epoch          float64 `json:"epoch"`
			NumOSDs        float64 `json:"num_osds"`
			NumUpOSDs      float64 `json:"num_up_osds"`
			NumInOSDs      float64 `json:"num_in_osds"`
			Full           bool    `json:"full"`
			NearFull       bool    `json:"nearfull"`
			NumRemappedPGs float64 `json:"num_remapped_pgs"`
		} `json:"osdmap"`
	} `json:"osdmap"`
	PGMap struct {
		PGsByState []struct {
			StateName string  `json:"state_name"`
			Count     float64 `json:"count"`
		} `json:"pgs_by_state"`
		Version       float64  `json:"version"`
		NumPGs        float64  `json:"num_pgs"`
		DataBytes     float64  `json:"data_bytes"`
		BytesUsed     float64  `json:"bytes_used"`
		BytesAvail    float64  `json:"bytes_avail"`
		BytesTotal    float64  `json:"bytes_total"`
		ReadBytesSec  float64  `json:"read_bytes_sec"`
		WriteBytesSec float64  `json:"write_bytes_sec"`
		OpPerSec      *float64 `json:"op_per_sec"` // This field is no longer reported in ceph 10 and later
		ReadOpPerSec  float64  `json:"read_op_per_sec"`
		WriteOpPerSec float64  `json:"write_op_per_sec"`
	} `json:"pgmap"`
}

// decodeStatus decodes the output of 'ceph -s'
func decodeStatus(acc telegraf.Accumulator, input string) error {
	data := &CephStatus{}
	if err := json.Unmarshal([]byte(input), data); err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	decoders := []func(telegraf.Accumulator, *CephStatus) error{
		decodeStatusHealth,
		decodeStatusOsdmap,
		decodeStatusPgmap,
		decodeStatusPgmapState,
	}

	for _, decoder := range decoders {
		if err := decoder(acc, data); err != nil {
			return err
		}
	}

	return nil
}

// decodeStatusHealth decodes the health portion of the output of 'ceph status'
func decodeStatusHealth(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"status":         data.Health.Status,
		"overall_status": data.Health.OverallStatus,
	}
	acc.AddFields("ceph_health", fields, map[string]string{})
	return nil
}

// decodeStatusOsdmap decodes the OSD map portion of the output of 'ceph -s'
func decodeStatusOsdmap(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"epoch":            data.OSDMap.OSDMap.Epoch,
		"num_osds":         data.OSDMap.OSDMap.NumOSDs,
		"num_up_osds":      data.OSDMap.OSDMap.NumUpOSDs,
		"num_in_osds":      data.OSDMap.OSDMap.NumInOSDs,
		"full":             data.OSDMap.OSDMap.Full,
		"nearfull":         data.OSDMap.OSDMap.NearFull,
		"num_remapped_pgs": data.OSDMap.OSDMap.NumRemappedPGs,
	}
	acc.AddFields("ceph_osdmap", fields, map[string]string{})
	return nil
}

// decodeStatusPgmap decodes the PG map portion of the output of 'ceph -s'
func decodeStatusPgmap(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"version":          data.PGMap.Version,
		"num_pgs":          data.PGMap.NumPGs,
		"data_bytes":       data.PGMap.DataBytes,
		"bytes_used":       data.PGMap.BytesUsed,
		"bytes_avail":      data.PGMap.BytesAvail,
		"bytes_total":      data.PGMap.BytesTotal,
		"read_bytes_sec":   data.PGMap.ReadBytesSec,
		"write_bytes_sec":  data.PGMap.WriteBytesSec,
		"op_per_sec":       data.PGMap.OpPerSec, // This field is no longer reported in ceph 10 and later
		"read_op_per_sec":  data.PGMap.ReadOpPerSec,
		"write_op_per_sec": data.PGMap.WriteOpPerSec,
	}
	acc.AddFields("ceph_pgmap", fields, map[string]string{})
	return nil
}

// decodeStatusPgmapState decodes the PG map state portion of the output of 'ceph -s'
func decodeStatusPgmapState(acc telegraf.Accumulator, data *CephStatus) error {
	for _, pgState := range data.PGMap.PGsByState {
		tags := map[string]string{
			"state": pgState.StateName,
		}
		fields := map[string]interface{}{
			"count": pgState.Count,
		}
		acc.AddFields("ceph_pgmap_state", fields, tags)
	}
	return nil
}

// CephDF is used to unmarshal 'ceph df' output
type CephDf struct {
	Stats struct {
		TotalSpace      *float64 `json:"total_space"` // pre ceph 0.84
		TotalUsed       *float64 `json:"total_used"`  // pre ceph 0.84
		TotalAvail      *float64 `json:"total_avail"` // pre ceph 0.84
		TotalBytes      *float64 `json:"total_bytes"`
		TotalUsedBytes  *float64 `json:"total_used_bytes"`
		TotalAvailBytes *float64 `json:"total_avail_bytes"`
	} `json:"stats"`
	Pools []struct {
		Name  string `json:"name"`
		Stats struct {
			KBUsed      float64  `json:"kb_used"`
			BytesUsed   float64  `json:"bytes_used"`
			Objects     float64  `json:"objects"`
			PercentUsed *float64 `json:"percent_used"`
			MaxAvail    *float64 `json:"max_avail"`
		} `json:"stats"`
	} `json:"pools"`
}

// decodeDf decodes the output of 'ceph df'
func decodeDf(acc telegraf.Accumulator, input string) error {
	data := &CephDf{}
	if err := json.Unmarshal([]byte(input), data); err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	// ceph.usage: records global utilization and number of objects
	fields := map[string]interface{}{
		"total_space":       data.Stats.TotalSpace,
		"total_used":        data.Stats.TotalUsed,
		"total_avail":       data.Stats.TotalAvail,
		"total_bytes":       data.Stats.TotalBytes,
		"total_used_bytes":  data.Stats.TotalUsedBytes,
		"total_avail_bytes": data.Stats.TotalAvailBytes,
	}
	acc.AddFields("ceph_usage", fields, map[string]string{})

	// ceph.pool.usage: records per pool utilization and number of objects
	for _, pool := range data.Pools {
		tags := map[string]string{
			"name": pool.Name,
		}
		fields := map[string]interface{}{
			"kb_used":      pool.Stats.KBUsed,
			"bytes_used":   pool.Stats.BytesUsed,
			"objects":      pool.Stats.Objects,
			"percent_used": pool.Stats.PercentUsed,
			"max_avail":    pool.Stats.MaxAvail,
		}
		acc.AddFields("ceph_pool_usage", fields, tags)
	}

	return nil
}

// CephOSDPoolStats is used to unmarshal 'ceph osd pool stats' output
type CephOSDPoolStats []struct {
	PoolName     string `json:"pool_name"`
	ClientIORate struct {
		ReadBytesSec  float64  `json:"read_bytes_sec"`
		WriteBytesSec float64  `json:"write_bytes_sec"`
		OpPerSec      *float64 `json:"op_per_sec"` // This field is no longer reported in ceph 10 and later
		ReadOpPerSec  float64  `json:"read_op_per_sec"`
		WriteOpPerSec float64  `json:"write_op_per_sec"`
	} `json:"client_io_rate"`
	RecoveryRate struct {
		RecoveringObjectsPerSec float64 `json:"recovering_objects_per_sec"`
		RecoveringBytesPerSec   float64 `json:"recovering_bytes_per_sec"`
		RecoveringKeysPerSec    float64 `json:"recovering_keys_per_sec"`
	} `json:"recovery_rate"`
}

// decodeOsdPoolStats decodes the output of 'ceph osd pool stats'
func decodeOsdPoolStats(acc telegraf.Accumulator, input string) error {
	data := CephOSDPoolStats{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	// ceph.pool.stats: records pre pool IO and recovery throughput
	for _, pool := range data {
		tags := map[string]string{
			"name": pool.PoolName,
		}
		fields := map[string]interface{}{
			"read_bytes_sec":             pool.ClientIORate.ReadBytesSec,
			"write_bytes_sec":            pool.ClientIORate.WriteBytesSec,
			"op_per_sec":                 pool.ClientIORate.OpPerSec, // This field is no longer reported in ceph 10 and later
			"read_op_per_sec":            pool.ClientIORate.ReadOpPerSec,
			"write_op_per_sec":           pool.ClientIORate.WriteOpPerSec,
			"recovering_objects_per_sec": pool.RecoveryRate.RecoveringObjectsPerSec,
			"recovering_bytes_per_sec":   pool.RecoveryRate.RecoveringBytesPerSec,
			"recovering_keys_per_sec":    pool.RecoveryRate.RecoveringKeysPerSec,
		}
		acc.AddFields("ceph_pool_stats", fields, tags)
	}

	return nil
}
