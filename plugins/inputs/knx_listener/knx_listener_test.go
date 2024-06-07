package knx_listener

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"

	"github.com/influxdata/telegraf/testutil"
)

const epsilon = 1e-3

func setValue(data dpt.DatapointValue, value interface{}) error {
	d := reflect.Indirect(reflect.ValueOf(data))
	if !d.CanSet() {
		return fmt.Errorf("cannot set datapoint %v", data)
	}
	switch v := value.(type) {
	case bool:
		d.SetBool(v)
	case float64:
		d.SetFloat(v)
	case int64:
		d.SetInt(v)
	case uint64:
		d.SetUint(v)
	case string:
		d.SetString(v)
	default:
		return fmt.Errorf("unknown type '%T' when setting value for DPT", value)
	}
	return nil
}

type message struct {
	address string
	dpt     string
	value   interface{}
}

func produceKnxEvent(t *testing.T, address string, datapoint string, value interface{}) *knx.GroupEvent {
	addr, err := cemi.NewGroupAddrString(address)
	require.NoError(t, err)

	data, ok := dpt.Produce(datapoint)
	require.True(t, ok)
	err = setValue(data, value)
	require.NoError(t, err)

	return &knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: addr,
		Data:        data.Pack(),
	}
}

func TestRegularReceives_DPT(t *testing.T) {
	// Define the test-cases
	var testcases = []struct {
		address  string
		dpt      string
		asstring bool
		value    interface{}
		expected interface{}
	}{
		{"1/0/1", "1.001", false, true, true},
		{"1/0/2", "1.002", false, false, false},
		{"1/0/3", "1.003", false, true, true},
		{"1/0/9", "1.009", false, false, false},
		{"1/1/0", "1.010", false, true, true},
		{"1/1/1", "1.001", true, true, "On"},
		{"1/1/2", "1.001", true, false, "Off"},
		{"1/1/3", "1.002", true, true, "True"},
		{"1/1/4", "1.002", true, false, "False"},
		{"1/1/5", "1.003", true, true, "Enable"},
		{"1/1/6", "1.003", true, false, "Disable"},
		{"1/1/7", "1.009", true, true, "Close"},
		{"1/1/8", "1.009", true, false, "Open"},
		{"1/2/0", "1.010", true, true, "Start"},
		{"1/2/1", "1.010", true, false, "Stop"},
		{"5/0/1", "5.001", false, 12.157, 12.157},
		{"5/0/3", "5.003", false, 121.412, 121.412},
		{"5/0/4", "5.004", false, uint64(25), uint64(25)},
		{"9/0/1", "9.001", false, 18.56, 18.56},
		{"9/0/4", "9.004", false, 243.84, 243.84},
		{"9/0/5", "9.005", false, 12.01, 12.01},
		{"9/0/7", "9.007", false, 59.32, 59.32},
		{"13/0/1", "13.001", false, int64(-15), int64(-15)},
		{"13/0/2", "13.002", false, int64(183), int64(183)},
		{"13/1/0", "13.010", false, int64(-141), int64(-141)},
		{"13/1/1", "13.011", false, int64(277), int64(277)},
		{"13/1/2", "13.012", false, int64(-4096), int64(-4096)},
		{"13/1/3", "13.013", false, int64(8192), int64(8192)},
		{"13/1/4", "13.014", false, int64(-65536), int64(-65536)},
		{"13/1/5", "13.015", false, int64(2147483647), int64(2147483647)},
		{"14/0/0", "14.000", false, -1.31, -1.31},
		{"14/0/1", "14.001", false, 0.44, 0.44},
		{"14/0/2", "14.002", false, 32.08, 32.08},
		{"14/0/3", "14.003", false, 92.69, 92.69},
		{"14/0/4", "14.004", false, 1.00794, 1.00794},
		{"14/1/0", "14.010", false, 5963.78, 5963.78},
		{"14/1/1", "14.011", false, 150.95, 150.95},
		{"16/0/0", "16.000", false, "hello world", "hello world"},
	}
	acc := &testutil.Accumulator{}

	// Setup the unit-under-test
	measurements := make([]Measurement, 0, len(testcases))
	for _, testcase := range testcases {
		measurements = append(measurements, Measurement{
			Name:      "test",
			Dpt:       testcase.dpt,
			AsString:  testcase.asstring,
			Addresses: []string{testcase.address},
		})
	}
	listener := KNXListener{
		ServiceType:  "dummy",
		Measurements: measurements,
		Log:          testutil.Logger{Name: "knx_listener"},
	}
	require.NoError(t, listener.Init())

	// Setup the listener to test
	err := listener.Start(acc)
	require.NoError(t, err)
	client := listener.client.(*KNXDummyInterface)

	tstart := time.Now()

	// Send the defined test data
	for _, testcase := range testcases {
		event := produceKnxEvent(t, testcase.address, testcase.dpt, testcase.value)
		client.Send(*event)
	}

	// Give the accumulator some time to collect the data
	acc.Wait(len(testcases))

	// Stop the listener
	listener.Stop()
	tstop := time.Now()

	// Check if we got what we expected
	require.Len(t, acc.Metrics, len(testcases))
	for i, m := range acc.Metrics {
		require.Equal(t, "test", m.Measurement)
		require.Equal(t, testcases[i].address, m.Tags["groupaddress"])
		require.Len(t, m.Fields, 1)
		switch v := testcases[i].expected.(type) {
		case string, bool, int64, uint64:
			require.Equal(t, v, m.Fields["value"])
		case float64:
			require.InDelta(t, v, m.Fields["value"], epsilon)
		}
		require.False(t, tstop.Before(m.Time))
		require.False(t, tstart.After(m.Time))
	}
}

func TestRegularReceives_MultipleMessages(t *testing.T) {
	listener := KNXListener{
		ServiceType: "dummy",
		Measurements: []Measurement{
			{Name: "temperature", Dpt: "1.001", Addresses: []string{"1/1/1"}},
		},
		Log: testutil.Logger{Name: "knx_listener"},
	}
	require.NoError(t, listener.Init())

	acc := &testutil.Accumulator{}

	// Setup the listener to test
	err := listener.Start(acc)
	require.NoError(t, err)
	client := listener.client.(*KNXDummyInterface)

	testMessages := []message{
		{"1/1/1", "1.001", true},
		{"1/1/1", "1.001", false},
		{"1/1/2", "1.001", false},
		{"1/1/2", "1.001", true},
	}

	for _, testcase := range testMessages {
		event := produceKnxEvent(t, testcase.address, testcase.dpt, testcase.value)
		client.Send(*event)
	}

	// Give the accumulator some time to collect the data
	acc.Wait(2)

	// Stop the listener
	listener.Stop()

	// Check if we got what we expected
	require.Len(t, acc.Metrics, 2)

	require.Equal(t, "temperature", acc.Metrics[0].Measurement)
	require.Equal(t, "1/1/1", acc.Metrics[0].Tags["groupaddress"])
	require.Len(t, acc.Metrics[0].Fields, 1)
	v, ok := acc.Metrics[0].Fields["value"].(bool)
	require.Truef(t, ok, "bool type expected, got '%T' with '%v' value instead", acc.Metrics[0].Fields["value"], acc.Metrics[0].Fields["value"])
	require.True(t, v)

	require.Equal(t, "temperature", acc.Metrics[1].Measurement)
	require.Equal(t, "1/1/1", acc.Metrics[1].Tags["groupaddress"])
	require.Len(t, acc.Metrics[1].Fields, 1)
	v, ok = acc.Metrics[1].Fields["value"].(bool)
	require.Truef(t, ok, "bool type expected, got '%T' with '%v' value instead", acc.Metrics[1].Fields["value"], acc.Metrics[1].Fields["value"])
	require.False(t, v)
}

func TestReconnect(t *testing.T) {
	listener := KNXListener{
		ServiceType: "dummy",
		Measurements: []Measurement{
			{Name: "temperature", Dpt: "1.001", Addresses: []string{"1/1/1"}},
		},
		Log: testutil.Logger{Name: "knx_listener"},
	}
	require.NoError(t, listener.Init())

	var acc testutil.Accumulator

	// Setup the listener to test
	require.NoError(t, listener.Start(&acc))
	defer listener.Stop()
	client := listener.client.(*KNXDummyInterface)

	testMessages := []message{
		{"1/1/1", "1.001", true},
		{"1/1/1", "1.001", false},
		{"1/1/2", "1.001", false},
		{"1/1/2", "1.001", true},
	}

	for _, testcase := range testMessages {
		event := produceKnxEvent(t, testcase.address, testcase.dpt, testcase.value)
		client.Send(*event)
	}

	// Give the accumulator some time to collect the data
	require.Eventuallyf(t, func() bool {
		return acc.NMetrics() >= 2
	}, 3*time.Second, 100*time.Millisecond, "expected 2 metric but got %d", acc.NMetrics())
	require.True(t, listener.connected.Load())

	client.Close()

	require.Eventually(t, func() bool {
		return !listener.connected.Load()
	}, 3*time.Second, 100*time.Millisecond, "no disconnect")
	acc.Lock()
	err := acc.FirstError()
	acc.Unlock()
	require.ErrorContains(t, err, "disconnected from bus")

	require.NoError(t, listener.Gather(&acc))
	require.Eventually(t, func() bool {
		return listener.connected.Load()
	}, 3*time.Second, 100*time.Millisecond, "no reconnect")
	client = listener.client.(*KNXDummyInterface)

	for _, testcase := range testMessages {
		event := produceKnxEvent(t, testcase.address, testcase.dpt, testcase.value)
		client.Send(*event)
	}

	// Give the accumulator some time to collect the data
	require.Eventuallyf(t, func() bool {
		return acc.NMetrics() >= 2
	}, 3*time.Second, 100*time.Millisecond, "expected 2 metric but got %d", acc.NMetrics())
}
