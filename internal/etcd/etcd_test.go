package etcd

import (
	"golang.org/x/net/context"
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/plugins/outputs/influxdb"

	eclient "github.com/coreos/etcd/client"
)

func Test1Write(t *testing.T) {
	// Delete hostname conf file
	hostname, _ := os.Hostname()
	os.Remove("./testdata/test1/hosts/" + hostname + ".conf")
	// Get etcd client
	e := NewEtcdClient("http://localhost:2379", "/telegraf")
	// Delete old conf from etcd
	delOptions := &eclient.DeleteOptions{
		Recursive: true,
		Dir:       true,
	}
	e.Kapi.Delete(context.Background(), "/telegraf", delOptions)

	// Test write dir
	err := e.WriteConfigDir("./testdata/test1")
	require.NoError(t, err)
	resp, _ := e.Kapi.Get(context.Background(), "/telegraf/main", nil)
	assert.Equal(t,
		"[tags]\n  dc = \"us-east-1\"\n\n[agent]\n  interval = \"2s\"\n  round_interval = true\n  flush_interval = \"10s\"\n  flush_jitter = \"0s\"\n  debug = false\n  hostname = \"\"\n",
		resp.Node.Value)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/hosts/localhost", nil)
	assert.Equal(t,
		"\n[agent]\n  interval = \"2s\"\n  labels = [\"influx\"]\n\n[[inputs.cpu]]\n  percpu = true\n  totalcpu = true\n  drop = [\"cpu_time*\"]\n",
		resp.Node.Value)

	// Test write conf
	err = e.WriteLabelConfig("mylabel", "./testdata/test1/labels/network.conf")
	require.NoError(t, err)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/labels/mylabel", nil)
	assert.Equal(t,
		"[[inputs.net]]\n\n",
		resp.Node.Value)

	// Test read
	c := config.NewConfig()
	var inputFilters []string
	var outputFilters []string
	c.OutputFilters = outputFilters
	c.InputFilters = inputFilters

	net := inputs.Inputs["net"]().(*system.NetIOStats)
	influx := outputs.Outputs["influxdb"]().(*influxdb.InfluxDB)
	influx.URLs = []string{"http://localhost:8086"}
	influx.Database = "telegraf"
	influx.Precision = "s"

	c, err = e.ReadConfig(c, "mylabel,influx")
	require.NoError(t, err)
	assert.Equal(t, net, c.Inputs[0].Input,
		"Testdata did not produce a correct net struct.")
	assert.Equal(t, influx, c.Outputs[0].Output,
		"Testdata did not produce a correct influxdb struct.")

	// Test reload
	shutdown := make(chan struct{})
	signals := make(chan os.Signal)
	go e.LaunchWatcher(shutdown, signals)
	// Test write conf
	err = e.WriteLabelConfig("mylabel", "./testdata/test1/labels/network2.conf")
	require.NoError(t, err)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/labels/mylabel", nil)
	assert.Equal(t,
		"[[inputs.net]]\n\n  interfaces = [\"eth0\"]\n\n",
		resp.Node.Value)
	// TODO found a way to test reload ....
	sig := <-signals
	assert.Equal(t, syscall.SIGHUP, sig)

}

func Test2Error(t *testing.T) {
	e := NewEtcdClient("http://localhost:2379", "/telegraf")

	// Test write dir
	err := e.WriteConfigDir("./testdata/test2")
	require.Error(t, err)
}

func Test3Write(t *testing.T) {
	// Delete old hostname conf file
	hostname, _ := os.Hostname()
	os.Remove("./testdata/test1/hosts/" + hostname + ".conf")
	// Write host file
	if hostname != "" {
		r, err := os.Open("./testdata/test1/hosts/localhost.conf")
		if err != nil {
			panic(err)
		}
		defer r.Close()

		w, err := os.Create("./testdata/test1/hosts/" + hostname + ".conf")
		if err != nil {
			panic(err)
		}
		defer w.Close()

		// do the actual work
		_, err = io.Copy(w, r)
		if err != nil {
			panic(err)
		}
	}
	// Get tcd client
	e := NewEtcdClient("http://localhost:2379", "/telegraf")
	// Delete old conf from etcd
	delOptions := &eclient.DeleteOptions{
		Recursive: true,
		Dir:       true,
	}
	e.Kapi.Delete(context.Background(), "/telegraf", delOptions)

	// Test write dir
	err := e.WriteConfigDir("./testdata/test1")
	require.NoError(t, err)
	resp, _ := e.Kapi.Get(context.Background(), "/telegraf/main", nil)
	assert.Equal(t,
		"[tags]\n  dc = \"us-east-1\"\n\n[agent]\n  interval = \"2s\"\n  round_interval = true\n  flush_interval = \"10s\"\n  flush_jitter = \"0s\"\n  debug = false\n  hostname = \"\"\n",
		resp.Node.Value)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/hosts/localhost", nil)
	assert.Equal(t,
		"\n[agent]\n  interval = \"2s\"\n  labels = [\"influx\"]\n\n[[inputs.cpu]]\n  percpu = true\n  totalcpu = true\n  drop = [\"cpu_time*\"]\n",
		resp.Node.Value)

	// Test write conf
	err = e.WriteLabelConfig("mylabel", "./testdata/test1/labels/network.conf")
	require.NoError(t, err)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/labels/mylabel", nil)
	assert.Equal(t,
		"[[inputs.net]]\n\n",
		resp.Node.Value)

	// Test read
	c := config.NewConfig()
	var inputFilters []string
	var outputFilters []string
	c.OutputFilters = outputFilters
	c.InputFilters = inputFilters

	cpu := inputs.Inputs["cpu"]().(*system.CPUStats)
	cpu.PerCPU = true
	cpu.TotalCPU = true
	net := inputs.Inputs["net"]().(*system.NetIOStats)
	influx := outputs.Outputs["influxdb"]().(*influxdb.InfluxDB)
	influx.URLs = []string{"http://localhost:8086"}
	influx.Database = "telegraf"
	influx.Precision = "s"

	c, err = e.ReadConfig(c, "mylabel,influx")
	require.NoError(t, err)
	assert.Equal(t, cpu, c.Inputs[0].Input,
		"Testdata did not produce a correct net struct.")
	assert.Equal(t, net, c.Inputs[1].Input,
		"Testdata did not produce a correct net struct.")
	assert.Equal(t, influx, c.Outputs[0].Output,
		"Testdata did not produce a correct influxdb struct.")

	// Test reload
	shutdown := make(chan struct{})
	signals := make(chan os.Signal)
	go e.LaunchWatcher(shutdown, signals)
	// Test write conf
	err = e.WriteLabelConfig("mylabel", "./testdata/test1/labels/network2.conf")
	require.NoError(t, err)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/labels/mylabel", nil)
	assert.Equal(t,
		"[[inputs.net]]\n\n  interfaces = [\"eth0\"]\n\n",
		resp.Node.Value)
	// TODO found a way to test reload ....
	sig := <-signals
	assert.Equal(t, syscall.SIGHUP, sig)
	// Delete hostname conf file
	os.Remove("./testdata/test1/hosts/" + hostname + ".conf")
}
