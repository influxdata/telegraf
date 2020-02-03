package vsphere

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
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
		Log: testutil.Logger{},
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

func createSim(folders int) (*simulator.Model, *simulator.Server, error) {
	model := simulator.VPX()

	model.Folder = folders
	model.Datacenter = 2
	//model.App = 1

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

func TestMaxQuery(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}
	m, s, err := createSim(0)
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

func testLookupVM(ctx context.Context, t *testing.T, f *Finder, path string, expected int, expectedName string) {
	poweredOn := types.VirtualMachinePowerState("poweredOn")
	var vm []mo.VirtualMachine
	err := f.Find(ctx, "VirtualMachine", path, &vm)
	require.NoError(t, err)
	require.Equal(t, expected, len(vm))
	if expectedName != "" {
		require.Equal(t, expectedName, vm[0].Name)
	}
	for _, v := range vm {
		require.Equal(t, poweredOn, v.Runtime.PowerState)
	}
}

func TestFinder(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}

	m, s, err := createSim(0)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	v := defaultVSphere()
	ctx := context.Background()

	c, err := NewClient(ctx, s.URL, v)

	f := Finder{c}

	var dc []mo.Datacenter
	err = f.Find(ctx, "Datacenter", "/DC0", &dc)
	require.NoError(t, err)
	require.Equal(t, 1, len(dc))
	require.Equal(t, "DC0", dc[0].Name)

	var host []mo.HostSystem
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

	var vm []mo.VirtualMachine
	testLookupVM(ctx, t, &f, "/DC0/vm/DC0_H0_VM0", 1, "")
	testLookupVM(ctx, t, &f, "/DC0/vm/DC0_C0*", 2, "")
	testLookupVM(ctx, t, &f, "/DC0/*/DC0_H0_VM0", 1, "DC0_H0_VM0")
	testLookupVM(ctx, t, &f, "/DC0/*/DC0_H0_*", 2, "")
	testLookupVM(ctx, t, &f, "/DC0/**/DC0_H0_VM*", 2, "")
	testLookupVM(ctx, t, &f, "/DC0/**", 4, "")
	testLookupVM(ctx, t, &f, "/DC1/**", 4, "")
	testLookupVM(ctx, t, &f, "/**", 8, "")
	testLookupVM(ctx, t, &f, "/**/vm/**", 8, "")
	testLookupVM(ctx, t, &f, "/*/host/**/*DC*", 8, "")
	testLookupVM(ctx, t, &f, "/*/host/**/*DC*VM*", 8, "")
	testLookupVM(ctx, t, &f, "/*/host/**/*DC*/*/*DC*", 4, "")

	vm = []mo.VirtualMachine{}
	err = f.FindAll(ctx, "VirtualMachine", []string{"/DC0/vm/DC0_H0*", "/DC0/vm/DC0_C0*"}, []string{}, &vm)
	require.NoError(t, err)
	require.Equal(t, 4, len(vm))

	rf := ResourceFilter{
		finder:       &f,
		paths:        []string{"/DC0/vm/DC0_H0*", "/DC0/vm/DC0_C0*"},
		excludePaths: []string{"/DC0/vm/DC0_H0_VM0"},
		resType:      "VirtualMachine",
	}
	vm = []mo.VirtualMachine{}
	require.NoError(t, rf.FindAll(ctx, &vm))
	require.Equal(t, 3, len(vm))

	rf = ResourceFilter{
		finder:       &f,
		paths:        []string{"/DC0/vm/DC0_H0*", "/DC0/vm/DC0_C0*"},
		excludePaths: []string{"/**"},
		resType:      "VirtualMachine",
	}
	vm = []mo.VirtualMachine{}
	require.NoError(t, rf.FindAll(ctx, &vm))
	require.Equal(t, 0, len(vm))

	rf = ResourceFilter{
		finder:       &f,
		paths:        []string{"/**"},
		excludePaths: []string{"/**"},
		resType:      "VirtualMachine",
	}
	vm = []mo.VirtualMachine{}
	require.NoError(t, rf.FindAll(ctx, &vm))
	require.Equal(t, 0, len(vm))

	rf = ResourceFilter{
		finder:       &f,
		paths:        []string{"/**"},
		excludePaths: []string{"/this won't match anything"},
		resType:      "VirtualMachine",
	}
	vm = []mo.VirtualMachine{}
	require.NoError(t, rf.FindAll(ctx, &vm))
	require.Equal(t, 8, len(vm))

	rf = ResourceFilter{
		finder:       &f,
		paths:        []string{"/**"},
		excludePaths: []string{"/**/*VM0"},
		resType:      "VirtualMachine",
	}
	vm = []mo.VirtualMachine{}
	require.NoError(t, rf.FindAll(ctx, &vm))
	require.Equal(t, 4, len(vm))
}

func TestFolders(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}

	m, s, err := createSim(1)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove()
	defer s.Close()

	v := defaultVSphere()
	ctx := context.Background()

	c, err := NewClient(ctx, s.URL, v)

	f := Finder{c}

	var folder []mo.Folder
	err = f.Find(ctx, "Folder", "/F0", &folder)
	require.NoError(t, err)
	require.Equal(t, 1, len(folder))
	require.Equal(t, "F0", folder[0].Name)

	var dc []mo.Datacenter
	err = f.Find(ctx, "Datacenter", "/F0/DC1", &dc)
	require.NoError(t, err)
	require.Equal(t, 1, len(dc))
	require.Equal(t, "DC1", dc[0].Name)

	testLookupVM(ctx, t, &f, "/F0/DC0/vm/**/F*", 0, "")
	testLookupVM(ctx, t, &f, "/F0/DC1/vm/**/F*/*VM*", 4, "")
	testLookupVM(ctx, t, &f, "/F0/DC1/vm/**/F*/**", 4, "")
}

func TestAll(t *testing.T) {
	// Don't run test on 32-bit machines due to bug in simulator.
	// https://github.com/vmware/govmomi/issues/1330
	var i int
	if unsafe.Sizeof(i) < 8 {
		return
	}

	m, s, err := createSim(0)
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
