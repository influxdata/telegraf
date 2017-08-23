package libvirt

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_disk(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt integration tests.")
	}

	var acc testutil.Accumulator

	l := &Libvirt{Libvirt_uri: "test+tcp://127.0.0.1/root/test.xml"}

	err := acc.GatherError(l.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("disk"))
	assert.True(t, acc.HasTag("disk", "deploy_id"))
	assert.True(t, acc.HasTag("disk", "device"))

	assert.True(t, acc.HasInt64Field("disk", "read.request"))
	assert.True(t, acc.HasInt64Field("disk", "read.bytes"))
	assert.True(t, acc.HasInt64Field("disk", "write.request"))
	assert.True(t, acc.HasInt64Field("disk", "write.bytes"))
	assert.True(t, acc.HasInt64Field("disk", "total.requests"))
	assert.True(t, acc.HasInt64Field("disk", "total.bytes"))
	assert.True(t, acc.HasFloatField("disk", "current.read.request"))
	assert.True(t, acc.HasFloatField("disk", "current.write.bytes"))
	assert.True(t, acc.HasFloatField("disk", "current.total.requests"))
	assert.True(t, acc.HasFloatField("disk", "current.total.bytes"))
}

func Test_cpu(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt integration tests.")
	}

	var acc testutil.Accumulator

	l := &Libvirt{Libvirt_uri: "test+tcp://127.0.0.1/root/test.xml"}

	err := acc.GatherError(l.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("cpu"))
	assert.True(t, acc.HasTag("cpu", "deploy_id"))

	assert.True(t, acc.HasFloatField("cpu", "load"))
	assert.True(t, acc.HasUIntField("cpu", "time"))
}

func Test_net(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt integration tests.")
	}

	var acc testutil.Accumulator

	l := &Libvirt{Libvirt_uri: "test+tcp://127.0.0.1/root/test.xml"}

	err := acc.GatherError(l.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("network"))
	assert.True(t, acc.HasTag("network", "deploy_id"))

	assert.True(t, acc.HasInt64Field("network", "rx"))
	assert.True(t, acc.HasInt64Field("network", "tx"))
	assert.True(t, acc.HasFloatField("network", "current.rx"))
	assert.True(t, acc.HasFloatField("network", "current.tx"))
}

func Test_cpustat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt integration tests.")
	}

	var acc testutil.Accumulator

	l := &Libvirt{Libvirt_uri: "test+tcp://127.0.0.1/root/test.xml"}

	err := acc.GatherError(l.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasIntField("cpustat", "count"))

	assert.True(t, acc.HasMeasurement("cpustat"))

	assert.True(t, acc.HasInt32Field("cpustat", "cpu.cores"))
	assert.True(t, acc.HasFloatField("cpustat", "cpu.mhz"))
}

func Test_memory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping libvirt integration tests.")
	}

	var acc testutil.Accumulator

	l := &Libvirt{Libvirt_uri: "test+tcp://127.0.0.1/root/test.xml"}

	err := acc.GatherError(l.Gather)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("max"))
	assert.True(t, acc.HasTag("max", "deploy_id"))

	assert.True(t, acc.HasUIntField("max", "memory"))
	//assert.True(t, acc.HasUIntField("max", "vcpus"))
}
