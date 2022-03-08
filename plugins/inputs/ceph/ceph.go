package ceph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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
	typeMds     = "mds"
	typeRgw     = "rgw"
	osdPrefix   = "ceph-osd"
	monPrefix   = "ceph-mon"
	mdsPrefix   = "ceph-mds"
	rgwPrefix   = "ceph-client"
	sockSuffix  = "asok"
)

type Ceph struct {
	CephBinary             string `toml:"ceph_binary"`
	OsdPrefix              string `toml:"osd_prefix"`
	MonPrefix              string `toml:"mon_prefix"`
	MdsPrefix              string `toml:"mds_prefix"`
	RgwPrefix              string `toml:"rgw_prefix"`
	SocketDir              string `toml:"socket_dir"`
	SocketSuffix           string `toml:"socket_suffix"`
	CephUser               string `toml:"ceph_user"`
	CephConfig             string `toml:"ceph_config"`
	GatherAdminSocketStats bool   `toml:"gather_admin_socket_stats"`
	GatherClusterStats     bool   `toml:"gather_cluster_stats"`

	Log telegraf.Logger `toml:"-"`
}

func (c *Ceph) Description() string {
	return "Collects performance metrics from the MON, OSD, MDS and RGW nodes in a Ceph storage cluster."
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
  mds_prefix = "ceph-mds"
  rgw_prefix = "ceph-client"

  ## suffix used to identify socket files
  socket_suffix = "asok"

  ## Ceph user to authenticate as, ceph will search for the corresponding keyring
  ## e.g. client.admin.keyring in /etc/ceph, or the explicit path defined in the
  ## client section of ceph.conf for example:
  ##
  ##     [client.telegraf]
  ##         keyring = /etc/ceph/client.telegraf.keyring
  ##
  ## Consult the ceph documentation for more detail on keyring generation.
  ceph_user = "client.admin"

  ## Ceph configuration to use to locate the cluster
  ceph_config = "/etc/ceph/ceph.conf"

  ## Whether to gather statistics via the admin socket
  gather_admin_socket_stats = true

  ## Whether to gather statistics via ceph commands, requires ceph_user and ceph_config
  ## to be specified
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
			acc.AddError(fmt.Errorf("error reading from socket '%s': %v", s.socket, err))
			continue
		}
		data, err := c.parseDump(dump)
		if err != nil {
			acc.AddError(fmt.Errorf("error parsing dump from socket '%s': %v", s.socket, err))
			continue
		}
		for tag, metrics := range data {
			acc.AddFields(measurement,
				metrics,
				map[string]string{"type": s.sockType, "id": s.sockID, "collection": tag})
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
		output, err := c.execute(job.command)
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
		MdsPrefix:              mdsPrefix,
		RgwPrefix:              rgwPrefix,
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

	switch socket.sockType {
	case typeOsd:
		cmdArgs = append(cmdArgs, "perf", "dump")
	case typeMon:
		cmdArgs = append(cmdArgs, "perfcounters_dump")
	case typeMds:
		cmdArgs = append(cmdArgs, "perf", "dump")
	case typeRgw:
		cmdArgs = append(cmdArgs, "perf", "dump")
	default:
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
	listing, err := os.ReadDir(c.SocketDir)
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
		if strings.HasPrefix(f, c.MdsPrefix) {
			sockType = typeMds
			sockPrefix = mdsPrefix
		}
		if strings.HasPrefix(f, c.RgwPrefix) {
			sockType = typeRgw
			sockPrefix = rgwPrefix
		}

		if sockType == typeOsd || sockType == typeMon || sockType == typeMds || sockType == typeRgw {
			path := filepath.Join(c.SocketDir, f)
			sockets = append(sockets, &socket{parseSockID(f, sockPrefix, c.SocketSuffix), sockType, path})
		}
	}
	return sockets, nil
}

func parseSockID(fname, prefix, suffix string) string {
	s := fname
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, suffix)
	s = strings.Trim(s, ".-_")
	return s
}

type socket struct {
	sockID   string
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
			//nolint:errcheck,revive // should never return an error
			buf.WriteString(".")
		}
		//nolint:errcheck,revive // should never return an error
		buf.WriteString(m.pathStack[i])
	}
	return buf.String()
}

type metricMap map[string]interface{}

type taggedMetricMap map[string]metricMap

// Parses a raw JSON string into a taggedMetricMap
// Delegates the actual parsing to newTaggedMetricMap(..)
func (c *Ceph) parseDump(dump string) (taggedMetricMap, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(dump), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json: '%s': %v", dump, err)
	}

	return c.newTaggedMetricMap(data), nil
}

// Builds a TaggedMetricMap out of a generic string map.
// The top-level key is used as a tag and all sub-keys are flattened into metrics
func (c *Ceph) newTaggedMetricMap(data map[string]interface{}) taggedMetricMap {
	tmm := make(taggedMetricMap)
	for tag, datapoints := range data {
		mm := make(metricMap)
		for _, m := range c.flatten(datapoints) {
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
func (c *Ceph) flatten(data interface{}) []*metric {
	var metrics []*metric

	switch val := data.(type) {
	case float64:
		metrics = []*metric{{make([]string, 0, 1), val}}
	case map[string]interface{}:
		metrics = make([]*metric, 0, len(val))
		for k, v := range val {
			for _, m := range c.flatten(v) {
				m.pathStack = append(m.pathStack, k)
				metrics = append(metrics, m)
			}
		}
	default:
		c.Log.Infof("ignoring unexpected type '%T' for value %v", val, val)
	}

	return metrics
}

// execute executes the 'ceph' command with the supplied arguments, returning JSON formatted output
func (c *Ceph) execute(command string) (string, error) {
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
	FSMap struct {
		NumUp        float64 `json:"up"`
		NumIn        float64 `json:"in"`
		NumMax       float64 `json:"max"`
		NumUpStandby float64 `json:"up:standby"`
	} `json:"fsmap"`
	Health struct {
		Status string `json:"status"`
	} `json:"health"`
	MonMap struct {
		NumMons float64 `json:"num_mons"`
	} `json:"monmap"`
	OSDMap struct {
		Epoch          float64 `json:"epoch"`
		NumOSDs        float64 `json:"num_osds"`
		NumUpOSDs      float64 `json:"num_up_osds"`
		NumInOSDs      float64 `json:"num_in_osds"`
		NumRemappedPGs float64 `json:"num_remapped_pgs"`
	} `json:"osdmap"`
	PGMap struct {
		PGsByState []struct {
			StateName string  `json:"state_name"`
			Count     float64 `json:"count"`
		} `json:"pgs_by_state"`
		Version                 float64 `json:"version"`
		NumPGs                  float64 `json:"num_pgs"`
		NumPools                float64 `json:"num_pools"`
		NumObjects              float64 `json:"num_objects"`
		DataBytes               float64 `json:"data_bytes"`
		BytesUsed               float64 `json:"bytes_used"`
		BytesAvail              float64 `json:"bytes_avail"`
		BytesTotal              float64 `json:"bytes_total"`
		ReadBytesSec            float64 `json:"read_bytes_sec"`
		WriteBytesSec           float64 `json:"write_bytes_sec"`
		ReadOpPerSec            float64 `json:"read_op_per_sec"`
		WriteOpPerSec           float64 `json:"write_op_per_sec"`
		InactivePGsRatio        float64 `json:"inactive_pgs_ratio"`
		DegradedObjects         float64 `json:"degraded_objects"`
		DegraedTotal            float64 `json:"degraded_total"`
		DegradedRatio           float64 `json:"degraded_ratio"`
		RecoveringObjectsPerSec float64 `json:"recovering_objects_per_sec"`
		RecoveringBytesPerSec   float64 `json:"recovering_bytes_per_sec"`
		RecoveringKeysPerSec    float64 `json:"recovering_keys_per_sec"`
		NumOjbectRecovered      float64 `json:"num_objects_recovered"`
		NumBytesRecovered       float64 `json:"num_bytes_recovered"`
		NumKeysRecovered        float64 `json:"num_keys_recovered"`
	} `json:"pgmap"`
}

// decodeStatus decodes the output of 'ceph -s'
func decodeStatus(acc telegraf.Accumulator, input string) error {
	data := &CephStatus{}
	if err := json.Unmarshal([]byte(input), data); err != nil {
		return fmt.Errorf("failed to parse json: '%s': %v", input, err)
	}

	decoders := []func(telegraf.Accumulator, *CephStatus) error{
		decodeStatusFsmap,
		decodeStatusHealth,
		decodeStatusMonmap,
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

// decodeStatusFsmap decodes the FS map portion of the output of 'ceph -s'
func decodeStatusFsmap(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"up":         data.FSMap.NumUp,
		"in":         data.FSMap.NumIn,
		"max":        data.FSMap.NumMax,
		"up_standby": data.FSMap.NumUpStandby,
	}
	acc.AddFields("ceph_fsmap", fields, map[string]string{})
	return nil
}

// decodeStatusHealth decodes the health portion of the output of 'ceph status'
func decodeStatusHealth(acc telegraf.Accumulator, data *CephStatus) error {
	statusCodes := map[string]float64{
		"HEALTH_ERR":  0,
		"HEALTH_WARN": 1,
		"HEALTH_OK":   2,
	}
	fields := map[string]interface{}{
		"status":      data.Health.Status,
		"status_code": statusCodes[data.Health.Status],
	}
	acc.AddFields("ceph_health", fields, map[string]string{})
	return nil
}

// decodeStatusMonmap decodes the Mon map portion of the output of 'ceph -s'
func decodeStatusMonmap(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"num_mons": data.MonMap.NumMons,
	}
	acc.AddFields("ceph_monmap", fields, map[string]string{})
	return nil
}

// decodeStatusOsdmap decodes the OSD map portion of the output of 'ceph -s'
func decodeStatusOsdmap(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"epoch":            data.OSDMap.Epoch,
		"num_osds":         data.OSDMap.NumOSDs,
		"num_up_osds":      data.OSDMap.NumUpOSDs,
		"num_in_osds":      data.OSDMap.NumInOSDs,
		"num_remapped_pgs": data.OSDMap.NumRemappedPGs,
	}
	acc.AddFields("ceph_osdmap", fields, map[string]string{})
	return nil
}

// decodeStatusPgmap decodes the PG map portion of the output of 'ceph -s'
func decodeStatusPgmap(acc telegraf.Accumulator, data *CephStatus) error {
	fields := map[string]interface{}{
		"version":                    data.PGMap.Version,
		"num_pgs":                    data.PGMap.NumPGs,
		"num_pools":                  data.PGMap.NumPools,
		"num_objects":                data.PGMap.NumObjects,
		"data_bytes":                 data.PGMap.DataBytes,
		"bytes_used":                 data.PGMap.BytesUsed,
		"bytes_avail":                data.PGMap.BytesAvail,
		"bytes_total":                data.PGMap.BytesTotal,
		"read_bytes_sec":             data.PGMap.ReadBytesSec,
		"write_bytes_sec":            data.PGMap.WriteBytesSec,
		"read_op_per_sec":            data.PGMap.ReadOpPerSec,
		"write_op_per_sec":           data.PGMap.WriteOpPerSec,
		"inactive_pgs_ratio":         data.PGMap.InactivePGsRatio,
		"degraded_objects":           data.PGMap.DegradedObjects,
		"degraded_total":             data.PGMap.DegraedTotal,
		"degraded_ratio":             data.PGMap.DegradedRatio,
		"recovering_objects_per_sec": data.PGMap.RecoveringObjectsPerSec,
		"recovering_bytes_per_sec":   data.PGMap.RecoveringBytesPerSec,
		"recovering_keys_per_sec":    data.PGMap.RecoveringKeysPerSec,
		"num_objects_recovered":      data.PGMap.NumOjbectRecovered,
		"num_bytes_recovered":        data.PGMap.NumBytesRecovered,
		"num_keys_recovered":         data.PGMap.NumKeysRecovered,
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
		NumOSDs            float64 `json:"num_osds"`
		NumPerPoolOSDs     float64 `json:"num_per_pool_osds"`
		NumPerPoolOmapOSDs float64 `json:"num_per_pool_omap_osds"`
		TotalAvailBytes    float64 `json:"total_avail_bytes"`
		TotalBytes         float64 `json:"total_bytes"`
		TotalUsedBytes     float64 `json:"total_used_bytes"`
		TotalUsedRawBytes  float64 `json:"total_used_raw_bytes"`
		TotalUsedRawRatio  float64 `json:"total_used_raw_ratio"`
	} `json:"stats"`
	StatsbyClass map[string]map[string]float64 `json:"stats_by_class"`
	Pools        []struct {
		Name  string `json:"name"`
		Stats struct {
			BytesUsed   float64 `json:"bytes_used"`
			KBUsed      float64 `json:"kb_used"`
			MaxAvail    float64 `json:"max_avail"`
			Objects     float64 `json:"objects"`
			PercentUsed float64 `json:"percent_used"`
			Stored      float64 `json:"stored"`
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
		"num_osds":               data.Stats.NumOSDs,
		"num_per_pool_osds":      data.Stats.NumPerPoolOSDs,
		"num_per_pool_omap_osds": data.Stats.NumPerPoolOmapOSDs,
		"total_avail_bytes":      data.Stats.TotalAvailBytes,
		"total_bytes":            data.Stats.TotalBytes,
		"total_used_bytes":       data.Stats.TotalUsedBytes,
		"total_used_raw_bytes":   data.Stats.TotalUsedRawBytes,
		"total_used_raw_ratio":   data.Stats.TotalUsedRawRatio,
	}
	acc.AddFields("ceph_usage", fields, map[string]string{})

	// ceph.stats_by_class: records per device-class usage
	for class, stats := range data.StatsbyClass {
		tags := map[string]string{
			"class": class,
		}
		fields := map[string]interface{}{}
		for key, value := range stats {
			fields[key] = value
		}
		acc.AddFields("ceph_deviceclass_usage", fields, tags)
	}

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
			"stored":       pool.Stats.Stored,
		}
		acc.AddFields("ceph_pool_usage", fields, tags)
	}

	return nil
}

// CephOSDPoolStats is used to unmarshal 'ceph osd pool stats' output
type CephOSDPoolStats []struct {
	PoolName     string `json:"pool_name"`
	ClientIORate struct {
		ReadBytesSec  float64 `json:"read_bytes_sec"`
		WriteBytesSec float64 `json:"write_bytes_sec"`
		ReadOpPerSec  float64 `json:"read_op_per_sec"`
		WriteOpPerSec float64 `json:"write_op_per_sec"`
	} `json:"client_io_rate"`
	RecoveryRate struct {
		RecoveringObjectsPerSec float64 `json:"recovering_objects_per_sec"`
		RecoveringBytesPerSec   float64 `json:"recovering_bytes_per_sec"`
		RecoveringKeysPerSec    float64 `json:"recovering_keys_per_sec"`
		NumObjectRecovered      float64 `json:"num_objects_recovered"`
		NumBytesRecovered       float64 `json:"num_bytes_recovered"`
		NumKeysRecovered        float64 `json:"num_keys_recovered"`
	} `json:"recovery_rate"`
	Recovery struct {
		DegradedObjects float64 `json:"degraded_objects"`
		DegradedTotal   float64 `json:"degraded_total"`
		DegradedRatio   float64 `json:"degraded_ratio"`
	} `json:"recovery"`
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
			"read_op_per_sec":            pool.ClientIORate.ReadOpPerSec,
			"write_op_per_sec":           pool.ClientIORate.WriteOpPerSec,
			"recovering_objects_per_sec": pool.RecoveryRate.RecoveringObjectsPerSec,
			"recovering_bytes_per_sec":   pool.RecoveryRate.RecoveringBytesPerSec,
			"recovering_keys_per_sec":    pool.RecoveryRate.RecoveringKeysPerSec,
			"num_objects_recovered":      pool.RecoveryRate.NumObjectRecovered,
			"num_bytes_recovered":        pool.RecoveryRate.NumBytesRecovered,
			"num_keys_recovered":         pool.RecoveryRate.NumKeysRecovered,
			"degraded_objects":           pool.Recovery.DegradedObjects,
			"degraded_total":             pool.Recovery.DegradedTotal,
			"degraded_ratio":             pool.Recovery.DegradedRatio,
		}
		acc.AddFields("ceph_pool_stats", fields, tags)
	}

	return nil
}
