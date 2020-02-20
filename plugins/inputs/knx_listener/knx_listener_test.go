package knx_listener

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

const epsilon = 1e-6

// Check if we can receive data
func TestRegularReceives_DPT1(t *testing.T) {
	acc := &testutil.Accumulator{}

	// Setup the unit-under-test
	listener := KNXListener{
		ServiceType:    "dummy",
		ServiceAddress: "manual",
		Measurements: []Measurement{
			{"test_1001", "1.001", []string{"1/0/1"}},
			{"test_1002", "1.002", []string{"1/0/2"}},
			{"test_1003", "1.003", []string{"1/0/3"}},
		},
	}

	err := listener.Start(acc)
	require.NoError(t, err)

	// Get the dummy interface to send data to the unit-under-test
	client := listener.client.(*KNXDummyInterface)

	tstart := time.Now()

	// Send some test data we expect in the config
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(1, 0, 1),
		Data:        dpt.DPT_1001(true).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(1, 0, 2),
		Data:        dpt.DPT_1002(false).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(1, 0, 3),
		Data:        dpt.DPT_1003(true).Pack(),
	})

	// Give the accumulator some time to collect the data
	time.Sleep(10 * time.Millisecond)

	// Stop the listener
	listener.Stop()
	tstop := time.Now()

	// Check if we got what we expected
	require.Len(t, acc.Metrics, 3)

	m := acc.Metrics[0]
	assert.Equal(t, "test_1001", m.Measurement)
	assert.Equal(t, "1/0/1", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.Equal(t, true, m.Fields["value"])
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[1]
	assert.Equal(t, "test_1002", m.Measurement)
	assert.Equal(t, "1/0/2", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.Equal(t, false, m.Fields["value"])
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[2]
	assert.Equal(t, "test_1003", m.Measurement)
	assert.Equal(t, "1/0/3", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.Equal(t, true, m.Fields["value"])
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))
}

// Check if we can receive data
func TestRegularReceives_DPT5(t *testing.T) {
	acc := &testutil.Accumulator{}

	// Setup the unit-under-test
	listener := KNXListener{
		ServiceType:    "dummy",
		ServiceAddress: "manual",
		Measurements: []Measurement{
			{"test_5001", "5.001", []string{"5/0/1"}},
			{"test_5003", "5.003", []string{"5/0/3"}},
			{"test_5004", "5.004", []string{"5/0/4"}},
		},
	}

	err := listener.Start(acc)
	require.NoError(t, err)

	// Get the dummy interface to send data to the unit-under-test
	client := listener.client.(*KNXDummyInterface)

	tstart := time.Now()

	// Send some test data we expect in the config
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(5, 0, 1),
		Data:        dpt.DPT_5001(12.12).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(5, 0, 3),
		Data:        dpt.DPT_5003(120.1).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(5, 0, 4),
		Data:        dpt.DPT_5004(25).Pack(),
	})

	// Give the accumulator some time to collect the data
	time.Sleep(10 * time.Millisecond)

	// Stop the listener
	listener.Stop()
	tstop := time.Now()

	// Check if we got what we expected
	require.Len(t, acc.Metrics, 3)

	m := acc.Metrics[0]
	assert.Equal(t, "test_5001", m.Measurement)
	assert.Equal(t, "5/0/1", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.InDelta(t, 12.12, m.Fields["value"], 100.0/255.0+epsilon)
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[1]
	assert.Equal(t, "test_5003", m.Measurement)
	assert.Equal(t, "5/0/3", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.InDelta(t, 120.1, m.Fields["value"], 360.0/255.0+epsilon)
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[2]
	assert.Equal(t, "test_5004", m.Measurement)
	assert.Equal(t, "5/0/4", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.Equal(t, uint64(25), m.Fields["value"])
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))
}

// Check if we can receive data
func TestRegularReceives_DPT9(t *testing.T) {
	acc := &testutil.Accumulator{}

	// Setup the unit-under-test
	listener := KNXListener{
		ServiceType:    "dummy",
		ServiceAddress: "manual",
		Measurements: []Measurement{
			{"test_9001", "9.001", []string{"9/0/1"}},
			{"test_9004", "9.004", []string{"9/0/4"}},
			{"test_9005", "9.005", []string{"9/0/5"}},
			{"test_9007", "9.007", []string{"9/0/7"}},
		},
	}

	err := listener.Start(acc)
	require.NoError(t, err)

	// Get the dummy interface to send data to the unit-under-test
	client := listener.client.(*KNXDummyInterface)

	tstart := time.Now()

	// Send some test data we expect in the config
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(9, 0, 1),
		Data:        dpt.DPT_9001(18.56).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(9, 0, 4),
		Data:        dpt.DPT_9004(243.9).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(9, 0, 5),
		Data:        dpt.DPT_9005(12.01).Pack(),
	})
	client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(9, 0, 7),
		Data:        dpt.DPT_9007(59.32).Pack(),
	})

	// Give the accumulator some time to collect the data
	time.Sleep(10 * time.Millisecond)

	// Stop the listener
	listener.Stop()
	tstop := time.Now()

	// Check if we got what we expected
	require.Len(t, acc.Metrics, 4)

	m := acc.Metrics[0]
	assert.Equal(t, "test_9001", m.Measurement)
	assert.Equal(t, "9/0/1", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.InDelta(t, 18.56, m.Fields["value"], 1e-3)
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[1]
	assert.Equal(t, "test_9004", m.Measurement)
	assert.Equal(t, "9/0/4", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.InDelta(t, 243.9, m.Fields["value"], 0.1)
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[2]
	assert.Equal(t, "test_9005", m.Measurement)
	assert.Equal(t, "9/0/5", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.InDelta(t, 12.01, m.Fields["value"], 1e-3)
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m = acc.Metrics[3]
	assert.Equal(t, "test_9007", m.Measurement)
	assert.Equal(t, "9/0/7", m.Tags["groupaddress"])
	assert.Len(t, m.Fields, 1)
	assert.InDelta(t, 59.32, m.Fields["value"], 1e-3)
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))
}
