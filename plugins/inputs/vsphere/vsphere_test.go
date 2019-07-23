package vsphere

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/influxdata/telegraf/internal"
	itls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var configHeader = `
[agent]
  interval = "10s"
  round_interval = true
  metric_batch_size = 1000
  metric_buffer_limit = 10000
  collection_jitter = "0s"
  flush_interval = "10s"
  flush_jitter = "0s"
  precision = ""
  debug = false
  quiet = false
  logfile = ""
  hostname = ""
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
		ClusterInclude:       []string{"/**"},
		HostMetricInclude: []string{
			"cpu.coreUtilization.average",
			"cpu.costop.summation",
			"cpu.demand.average",
			"cpu.idle.summation",
			"cpu.latency.average",
			"cpu.readiness.average",
			"cpu.ready.summation",
			"cpu.swapwait.summation",
			"cpu.usage.average",
			"cpu.usagemhz.average",
			"cpu.used.summation",
			"cpu.utilization.average",
			"cpu.wait.summation",
			"disk.deviceReadLatency.average",
			"disk.deviceWriteLatency.average",
			"disk.kernelReadLatency.average",
			"disk.kernelWriteLatency.average",
			"disk.numberReadAveraged.average",
			"disk.numberWriteAveraged.average",
			"disk.read.average",
			"disk.totalReadLatency.average",
			"disk.totalWriteLatency.average",
			"disk.write.average",
			"mem.active.average",
			"mem.latency.average",
			"mem.state.latest",
			"mem.swapin.average",
			"mem.swapinRate.average",
			"mem.swapout.average",
			"mem.swapoutRate.average",
			"mem.totalCapacity.average",
			"mem.usage.average",
			"mem.vmmemctl.average",
			"net.bytesRx.average",
			"net.bytesTx.average",
			"net.droppedRx.summation",
			"net.droppedTx.summation",
			"net.errorsRx.summation",
			"net.errorsTx.summation",
			"net.usage.average",
			"power.power.average",
			"storageAdapter.numberReadAveraged.average",
			"storageAdapter.numberWriteAveraged.average",
			"storageAdapter.read.average",
			"storageAdapter.write.average",
			"sys.uptime.latest"},
		HostMetricExclude: nil,
		HostInclude:       []string{"/**"},
		VMMetricInclude: []string{
			"cpu.demand.average",
			"cpu.idle.summation",
			"cpu.latency.average",
			"cpu.readiness.average",
			"cpu.ready.summation",
			"cpu.run.summation",
			"cpu.usagemhz.average",
			"cpu.used.summation",
			"cpu.wait.summation",
			"mem.active.average",
			"mem.granted.average",
			"mem.latency.average",
			"mem.swapin.average",
			"mem.swapinRate.average",
			"mem.swapout.average",
			"mem.swapoutRate.average",
			"mem.usage.average",
			"mem.vmmemctl.average",
			"net.bytesRx.average",
			"net.bytesTx.average",
			"net.droppedRx.summation",
			"net.droppedTx.summation",
			"net.usage.average",
			"power.power.average",
			"virtualDisk.numberReadAveraged.average",
			"virtualDisk.numberWriteAveraged.average",
			"virtualDisk.read.average",
			"virtualDisk.readOIO.latest",
			"virtualDisk.throughput.usage.average",
			"virtualDisk.totalReadLatency.average",
			"virtualDisk.totalWriteLatency.average",
			"virtualDisk.write.average",
			"virtualDisk.writeOIO.latest",
			"sys.uptime.latest"},
		VMMetricExclude: nil,
		VMInclude:       []string{"/**"},
		DatastoreMetricInclude: []string{
			"disk.used.*",
			"disk.provsioned.*"},
		DatastoreMetricExclude:  nil,
		DatastoreInclude:        []string{"/**"},
		DatacenterMetricInclude: nil,
		DatacenterMetricExclude: nil,
		DatacenterInclude:       []string{"/**"},
		ClientConfig:            itls.ClientConfig{InsecureSkipVerify: true},

		MaxQueryObjects:         256,
		MaxQueryMetrics:         256,
		ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
		Timeout:                 internal.Duration{Duration: time.Second * 20},
		ForceDiscoverOnInit:     true,
		DiscoverConcurrency:     1,
		CollectConcurrency:      1,
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
	return model, s, nil
}

func testAlignUniform(t *testing.T, n int) {
	now := time.Now().Truncate(60 * time.Second)
	info := make([]types.PerfSampleInfo, n)
	values := make([]int64, n)
	for i := 0; i < n; i++ {
		info[i] = types.PerfSampleInfo{
			Timestamp: now.Add(time.Duration(20*i) * time.Second),
			Interval:  20,
		}
		values[i] = 1
	}
	newInfo, newValues := alignSamples(info, values, 60*time.Second)
	require.Equal(t, n/3, len(newInfo), "Aligned infos have wrong size")
	require.Equal(t, n/3, len(newValues), "Aligned values have wrong size")
	for _, v := range newValues {
		require.Equal(t, 1.0, v, "Aligned value should be 1")
	}
}

func TestAlignMetrics(t *testing.T) {
	testAlignUniform(t, 3)
	testAlignUniform(t, 30)
	testAlignUniform(t, 333)

	// 20s to 60s of 1,2,3,1,2,3... (should average to 2)
	n := 30
	now := time.Now().Truncate(60 * time.Second)
	info := make([]types.PerfSampleInfo, n)
	values := make([]int64, n)
	for i := 0; i < n; i++ {
		info[i] = types.PerfSampleInfo{
			Timestamp: now.Add(time.Duration(20*i) * time.Second),
			Interval:  20,
		}
		values[i] = int64(i%3 + 1)
	}
	newInfo, newValues := alignSamples(info, values, 60*time.Second)
	require.Equal(t, n/3, len(newInfo), "Aligned infos have wrong size")
	require.Equal(t, n/3, len(newValues), "Aligned values have wrong size")
	for _, v := range newValues {
		require.Equal(t, 2.0, v, "Aligned value should be 2")
	}
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

func TestThrottledExecutor(t *testing.T) {
	max := int64(0)
	ngr := int64(0)
	n := 10000
	var mux sync.Mutex
	results := make([]int, 0, n)
	te := NewThrottledExecutor(5)
	for i := 0; i < n; i++ {
		func(i int) {
			te.Run(context.Background(), func() {
				atomic.AddInt64(&ngr, 1)
				mux.Lock()
				defer mux.Unlock()
				results = append(results, i*2)
				if ngr > max {
					max = ngr
				}
				time.Sleep(100 * time.Microsecond)
				atomic.AddInt64(&ngr, -1)
			})
		}(i)
	}
	te.Wait()
	sort.Ints(results)
	for i := 0; i < n; i++ {
		require.Equal(t, results[i], i*2, "Some jobs didn't run")
	}
	require.Equal(t, int64(5), max, "Wrong number of goroutines spawned")
}

func TestTimeout(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}

	m, s, err := createSim()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	v := defaultVSphere()
	var acc testutil.Accumulator
	v.Vcenters = []string{s.URL.String()}
	v.Timeout = internal.Duration{Duration: 1 * time.Nanosecond}
	require.NoError(t, v.Start(nil)) // We're not using the Accumulator, so it can be nil.
	defer v.Stop()
	err = v.Gather(&acc)

	// The accumulator must contain exactly one error and it must be a deadline exceeded.
	require.Equal(t, 1, len(acc.Errors))
	require.True(t, strings.Contains(acc.Errors[0].Error(), "context deadline exceeded"))
}

func TestMaxQuery(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}
	m, s, err := createSim()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	v := defaultVSphere()
	v.MaxQueryMetrics = 256
	ctx := context.Background()
	c, err := NewClient(ctx, s.URL, v)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 256, v.MaxQueryMetrics)

	om := object.NewOptionManager(c.Client.Client, *c.Client.Client.ServiceContent.Setting)
	err = om.Update(ctx, []types.BaseOptionValue{&types.OptionValue{
		Key:   "config.vpxd.stats.maxQueryMetrics",
		Value: "42",
	}})
	if err != nil {
		t.Fatal(err)
	}

	v.MaxQueryMetrics = 256
	ctx = context.Background()
	c2, err := NewClient(ctx, s.URL, v)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 42, v.MaxQueryMetrics)
	c.close()
	c2.close()
}

func TestFinder(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}

	m, s, err := createSim()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	v := defaultVSphere()
	ctx := context.Background()

	c, err := NewClient(ctx, s.URL, v)

	f := Finder{c}

	dc := []mo.Datacenter{}
	err = f.Find(ctx, "Datacenter", "/DC0", &dc)
	require.NoError(t, err)
	require.Equal(t, 1, len(dc))
	require.Equal(t, "DC0", dc[0].Name)

	host := []mo.HostSystem{}
	err = f.Find(ctx, "HostSystem", "/DC0/host/DC0_H0/DC0_H0", &host)
	require.NoError(t, err)
	require.Equal(t, 1, len(host))
	require.Equal(t, "DC0_H0", host[0].Name)

	host = []mo.HostSystem{}
	err = f.Find(ctx, "HostSystem", "/DC0/host/DC0_C0/DC0_C0_H0", &host)
	require.NoError(t, err)
	require.Equal(t, 1, len(host))
	require.Equal(t, "DC0_C0_H0", host[0].Name)

	host = []mo.HostSystem{}
	err = f.Find(ctx, "HostSystem", "/DC0/host/DC0_C0/*", &host)
	require.NoError(t, err)
	require.Equal(t, 3, len(host))

	vm := []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/DC0/vm/DC0_H0_VM0", &vm)
	require.NoError(t, err)
	require.Equal(t, 1, len(dc))
	require.Equal(t, "DC0_H0_VM0", vm[0].Name)

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/DC0/vm/DC0_C0*", &vm)
	require.NoError(t, err)
	require.Equal(t, 1, len(dc))

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/DC0/*/DC0_H0_VM0", &vm)
	require.NoError(t, err)
	require.Equal(t, 1, len(dc))
	require.Equal(t, "DC0_H0_VM0", vm[0].Name)

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/DC0/*/DC0_H0_*", &vm)
	require.NoError(t, err)
	require.Equal(t, 2, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/DC0/**/DC0_H0_VM*", &vm)
	require.NoError(t, err)
	require.Equal(t, 2, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/DC0/**", &vm)
	require.NoError(t, err)
	require.Equal(t, 4, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/**", &vm)
	require.NoError(t, err)
	require.Equal(t, 4, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/**/DC0_H0_VM*", &vm)
	require.NoError(t, err)
	require.Equal(t, 2, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.Find(ctx, "VirtualMachine", "/**/vm/**", &vm)
	require.NoError(t, err)
	require.Equal(t, 4, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.FindAll(ctx, "VirtualMachine", []string{"/DC0/vm/DC0_H0*", "/DC0/vm/DC0_C0*"}, &vm)
	require.NoError(t, err)
	require.Equal(t, 4, len(vm))

	vm = []mo.VirtualMachine{}
	err = f.FindAll(ctx, "VirtualMachine", []string{"/**"}, &vm)
	require.NoError(t, err)
	require.Equal(t, 4, len(vm))
}

func TestAll(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}

	m, s, err := createSim()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	var acc testutil.Accumulator
	v := defaultVSphere()
	v.Vcenters = []string{s.URL.String()}
	v.Start(&acc)
	defer v.Stop()
	require.NoError(t, v.Gather(&acc))
	require.Equal(t, 0, len(acc.Errors), fmt.Sprintf("Errors found: %s", acc.Errors))
	require.True(t, len(acc.Metrics) > 0, "No metrics were collected")
}
