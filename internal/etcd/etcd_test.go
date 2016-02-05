package etcd

import (
	"golang.org/x/net/context"
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
)

func TestWrite(t *testing.T) {
	e := NewEtcdClient("http://localhost:2379", "/telegraf")

	// Test write dir
	err := e.WriteConfigDir("./testdata/")
	require.NoError(t, err)
	resp, _ := e.Kapi.Get(context.Background(), "/telegraf/main", nil)
	assert.Equal(t,
		`{"agent":{"debug":false,"flush_interval":"10s","flush_jitter":"0s","hostname":"","interval":"2s","round_interval":true},"tags":{"dc":"us-east-1"}}`,
		resp.Node.Value)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/hosts/localhost", nil)
	assert.Equal(t,
		`{"agent":{"interval":"2s","labels":["influx"]},"inputs":{"cpu":[{"drop":["cpu_time*"],"percpu":true,"totalcpu":true}]}}`,
		resp.Node.Value)

	// Test write conf
	err = e.WriteLabelConfig("mylabel", "./testdata/labels/network.conf")
	require.NoError(t, err)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/labels/mylabel", nil)
	assert.Equal(t, `{"inputs":{"net":[{}]}}`, resp.Node.Value)

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
	err = e.WriteLabelConfig("mylabel", "./testdata/labels/network2.conf")
	require.NoError(t, err)
	resp, _ = e.Kapi.Get(context.Background(), "/telegraf/labels/mylabel", nil)
	assert.Equal(t, `{"inputs":{"net":[{"interfaces":["eth0"]}]}}`, resp.Node.Value)
	// TODO found a way to test reload ....
	sig := <-signals
	assert.Equal(t, syscall.SIGHUP, sig)

}
