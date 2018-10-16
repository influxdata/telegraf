package vsphere

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	itls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/simulator"
)

var configHeader = `
# Telegraf Configuration
#
# Telegraf is entirely plugin driven. All metrics are gathered from the
# declared inputs, and sent to the declared outputs.
#
# Plugins must be declared in here to be active.
# To deactivate a plugin, comment out the name and any variables.
#
# Use 'telegraf -config telegraf.conf -test' to see what metrics a config
# file would generate.
#
# Environment variables can be used anywhere in this config file, simply prepend
# them with $. For strings the variable must be within quotes (ie, "$STR_VAR"),
# for numbers and booleans they should be plain (ie, $INT_VAR, $BOOL_VAR)


# Global tags can be specified here in key="value" format.
[global_tags]
  # dc = "us-east-1" # will tag all metrics with dc=us-east-1
  # rack = "1a"
  ## Environment variables can be used as tags, and throughout the config file
  # user = "$USER"


# Configuration for telegraf agent
[agent]
  ## Default data collection interval for all inputs
  interval = "10s"
  ## Rounds collection interval to 'interval'
  ## ie, if interval="10s" then always collect on :00, :10, :20, etc.
  round_interval = true

  ## Telegraf will send metrics to outputs in batches of at most
  ## metric_batch_size metrics.
  ## This controls the size of writes that Telegraf sends to output plugins.
  metric_batch_size = 1000

  ## For failed writes, telegraf will cache metric_buffer_limit metrics for each
  ## output, and will flush this buffer on a successful write. Oldest metrics
  ## are dropped first when this buffer fills.
  ## This buffer only fills when writes fail to output plugin(s).
  metric_buffer_limit = 10000

  ## Collection jitter is used to jitter the collection by a random amount.
  ## Each plugin will sleep for a random time within jitter before collecting.
  ## This can be used to avoid many plugins querying things like sysfs at the
  ## same time, which can have a measurable effect on the system.
  collection_jitter = "0s"

  ## Default flushing interval for all outputs. You shouldn't set this below
  ## interval. Maximum flush_interval will be flush_interval + flush_jitter
  flush_interval = "10s"
  ## Jitter the flush interval by a random amount. This is primarily to avoid
  ## large write spikes for users running a large number of telegraf instances.
  ## ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
  flush_jitter = "0s"

  ## By default or when set to "0s", precision will be set to the same
  ## timestamp order as the collection interval, with the maximum being 1s.
  ##   ie, when interval = "10s", precision will be "1s"
  ##       when interval = "250ms", precision will be "1ms"
  ## Precision will NOT be used for service inputs. It is up to each individual
  ## service input to set the timestamp at the appropriate precision.
  ## Valid time units are "ns", "us" (or "µs"), "ms", "s".
  precision = ""

  ## Logging configuration:
  ## Run telegraf with debug log messages.
  debug = false
  ## Run telegraf in quiet mode (error log messages only).
  quiet = false
  ## Specify the log file name. The empty string means to log to stderr.
  logfile = ""

  ## Override default hostname, if empty use os.Hostname()
  hostname = ""
  ## If set to true, do no set the "host" tag in the telegraf agent.
  omit_hostname = false
`

func defaultVSphere() *VSphere {
	return &VSphere{
		ClusterMetricInclude: []string{
			"cpu.usage.*",
			"cpu.usagemhz.*",
			"mem.usage.*",
			"mem.active.*"},
		ClusterMetricExclude: nil,
		HostMetricInclude: []string{
			"cpu.ready.summation.delta.millisecond",
			"cpu.latency.average.rate.percent",
			"cpu.coreUtilization.average.rate.percent",
			"mem.usage.average.absolute.percent",
			"mem.swapinRate.average.rate.kiloBytesPerSecond",
			"mem.state.latest.absolute.number",
			"mem.latency.average.absolute.percent",
			"mem.vmmemctl.average.absolute.kiloBytes",
			"disk.read.average.rate.kiloBytesPerSecond",
			"disk.write.average.rate.kiloBytesPerSecond",
			"disk.numberReadAveraged.average.rate.number",
			"disk.numberWriteAveraged.average.rate.number",
			"disk.deviceReadLatency.average.absolute.millisecond",
			"disk.deviceWriteLatency.average.absolute.millisecond",
			"disk.totalReadLatency.average.absolute.millisecond",
			"disk.totalWriteLatency.average.absolute.millisecond",
			"storageAdapter.read.average.rate.kiloBytesPerSecond",
			"storageAdapter.write.average.rate.kiloBytesPerSecond",
			"storageAdapter.numberReadAveraged.average.rate.number",
			"storageAdapter.numberWriteAveraged.average.rate.number",
			"net.errorsRx.summation.delta.number",
			"net.errorsTx.summation.delta.number",
			"net.bytesRx.average.rate.kiloBytesPerSecond",
			"net.bytesTx.average.rate.kiloBytesPerSecond",
			"cpu.used.summation.delta.millisecond",
			"cpu.usage.average.rate.percent",
			"cpu.utilization.average.rate.percent",
			"cpu.wait.summation.delta.millisecond",
			"cpu.idle.summation.delta.millisecond",
			"cpu.readiness.average.rate.percent",
			"cpu.costop.summation.delta.millisecond",
			"cpu.swapwait.summation.delta.millisecond",
			"mem.swapoutRate.average.rate.kiloBytesPerSecond",
			"disk.kernelReadLatency.average.absolute.millisecond",
			"disk.kernelWriteLatency.average.absolute.millisecond"},
		HostMetricExclude: nil,
		VMMetricInclude: []string{
			"cpu.ready.summation.delta.millisecond",
			"mem.swapinRate.average.rate.kiloBytesPerSecond",
			"virtualDisk.numberReadAveraged.average.rate.number",
			"virtualDisk.numberWriteAveraged.average.rate.number",
			"virtualDisk.totalReadLatency.average.absolute.millisecond",
			"virtualDisk.totalWriteLatency.average.absolute.millisecond",
			"virtualDisk.readOIO.latest.absolute.number",
			"virtualDisk.writeOIO.latest.absolute.number",
			"net.bytesRx.average.rate.kiloBytesPerSecond",
			"net.bytesTx.average.rate.kiloBytesPerSecond",
			"net.droppedRx.summation.delta.number",
			"net.droppedTx.summation.delta.number",
			"cpu.run.summation.delta.millisecond",
			"cpu.used.summation.delta.millisecond",
			"mem.swapoutRate.average.rate.kiloBytesPerSecond",
			"virtualDisk.read.average.rate.kiloBytesPerSecond",
			"virtualDisk.write.average.rate.kiloBytesPerSecond"},
		VMMetricExclude: nil,
		DatastoreMetricInclude: []string{
			"disk.used.*",
			"disk.provsioned.*"},
		DatastoreMetricExclude: nil,
		ClientConfig:           itls.ClientConfig{InsecureSkipVerify: true},

		MaxQueryObjects:         256,
		ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
		Timeout:                 internal.Duration{Duration: time.Second * 20},
		ForceDiscoverOnInit:     true,
	}
}

func createSim() (*simulator.Model, *simulator.Server, error) {
	model := simulator.VPX()

	err := model.Create()
	if err != nil {
		return nil, nil, err
	}

	model.Service.TLS = new(tls.Config)

	s := model.Service.NewServer()
	//fmt.Printf("Server created at: %s\n", s.URL)

	return model, s, nil
}

func TestParseConfig(t *testing.T) {
	v := VSphere{}
	c := v.SampleConfig()
	p := regexp.MustCompile("\n#")
	fmt.Printf("Source=%s", p.ReplaceAllLiteralString(c, "\n"))
	c = configHeader + "\n[[inputs.vsphere]]\n" + p.ReplaceAllLiteralString(c, "\n")
	fmt.Printf("Source=%s", c)
	tab, err := toml.Parse([]byte(c))
	require.NoError(t, err)
	require.NotNil(t, tab)
}

func TestWorkerPool(t *testing.T) {
	wp := NewWorkerPool(100)
	ctx := context.Background()
	wp.Run(ctx, func(ctx context.Context, p interface{}) interface{} {
		return p.(int) * 2
	}, 10)

	n := 100000
	wp.Fill(ctx, func(ctx context.Context, f PushFunc) {
		for i := 0; i < n; i++ {
			f(ctx, i)
		}
	})
	results := make([]int, n)
	i := 0
	wp.Drain(ctx, func(ctx context.Context, p interface{}) bool {
		results[i] = p.(int)
		i++
		return true
	})
	sort.Ints(results)
	for i := 0; i < n; i++ {
		require.Equal(t, results[i], i*2)
	}
}

func TestTimeout(t *testing.T) {
	m, s, err := createSim()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	var acc testutil.Accumulator
	v := defaultVSphere()
	v.Vcenters = []string{s.URL.String()}
	v.Timeout = internal.Duration{Duration: 1 * time.Nanosecond}
	require.NoError(t, v.Start(nil)) // We're not using the Accumulator, so it can be nil.
	defer v.Stop()
	require.NoError(t, v.Gather(&acc))

	// The accumulator must contain exactly one error and it must be a deadline exceeded.
	require.Equal(t, 1, len(acc.Errors))
	require.True(t, strings.Contains(acc.Errors[0].Error(), "context deadline exceeded"))
}

func TestAll(t *testing.T) {
	m, s, err := createSim()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	var acc testutil.Accumulator
	v := defaultVSphere()
	v.Vcenters = []string{s.URL.String()}
	v.Start(nil) // We're not using the Accumulator, so it can be nil.
	defer v.Stop()
	require.NoError(t, v.Gather(&acc))
}
