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
	default:
		return fmt.Errorf("unknown type '%T' when setting value for DPT", value)
	}
	return nil
}

type TestMessage struct {
	address string
	dpt     string
	value   interface{}
}

func ProduceKnxEvent(t *testing.T, address string, datapoint string, value interface{}) *knx.GroupEvent {
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
	var testcases = []TestMessage{
		{"1/0/1", "1.001", true},
		{"1/0/2", "1.002", false},
		{"1/0/3", "1.003", true},
		{"1/0/9", "1.009", false},
		{"1/1/0", "1.010", true},
		{"5/0/1", "5.001", 12.157},
		{"5/0/3", "5.003", 121.412},
		{"5/0/4", "5.004", uint64(25)},
		{"9/0/1", "9.001", 18.56},
		{"9/0/4", "9.004", 243.84},
		{"9/0/5", "9.005", 12.01},
		{"9/0/7", "9.007", 59.32},
		{"13/0/1", "13.001", int64(-15)},
		{"13/0/2", "13.002", int64(183)},
		{"13/1/0", "13.010", int64(-141)},
		{"13/1/1", "13.011", int64(277)},
		{"13/1/2", "13.012", int64(-4096)},
		{"13/1/3", "13.013", int64(8192)},
		{"13/1/4", "13.014", int64(-65536)},
		{"13/1/5", "13.015", int64(2147483647)},
		{"14/0/0", "14.000", -1.31},
		{"14/0/1", "14.001", 0.44},
		{"14/0/2", "14.002", 32.08},
		// {"14/0/3", "14.003", 92.69},
		// {"14/0/4", "14.004", 1.00794},
		{"14/1/0", "14.010", 5963.78},
		{"14/1/1", "14.011", 150.95},
	}
	acc := &testutil.Accumulator{}

	// Setup the unit-under-test
	measurements := make([]Measurement, 0, len(testcases))
	for _, testcase := range testcases {
		measurements = append(measurements, Measurement{"test", testcase.dpt, []string{testcase.address}})
	}
	listener := KNXListener{
		ServiceType:  "dummy",
		Measurements: measurements,
		Log:          testutil.Logger{Name: "knx_listener"},
	}

	// Setup the listener to test
	err := listener.Start(acc)
	require.NoError(t, err)
	client := listener.client.(*KNXDummyInterface)

	tstart := time.Now()

	// Send the defined test data
	for _, testcase := range testcases {
		event := ProduceKnxEvent(t, testcase.address, testcase.dpt, testcase.value)
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
		switch v := testcases[i].value.(type) {
		case bool, int64, uint64:
			require.Equal(t, v, m.Fields["value"])
		case float64:
			require.InDelta(t, v, m.Fields["value"], epsilon)
		}
		require.True(t, !tstop.Before(m.Time))
		require.True(t, !tstart.After(m.Time))
	}
}

func TestRegularReceives_MultipleMessages(t *testing.T) {
	listener := KNXListener{
		ServiceType: "dummy",
		Measurements: []Measurement{
			{"temperature", "1.001", []string{"1/1/1"}},
		},
		Log: testutil.Logger{Name: "knx_listener"},
	}

	acc := &testutil.Accumulator{}

	// Setup the listener to test
	err := listener.Start(acc)
	require.NoError(t, err)
	client := listener.client.(*KNXDummyInterface)

	testMessages := []TestMessage{
		{"1/1/1", "1.001", true},
		{"1/1/1", "1.001", false},
		{"1/1/2", "1.001", false},
		{"1/1/2", "1.001", true},
	}

	for _, testcase := range testMessages {
		event := ProduceKnxEvent(t, testcase.address, testcase.dpt, testcase.value)
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
	require.Equal(t, true, acc.Metrics[0].Fields["value"])

	require.Equal(t, "temperature", acc.Metrics[1].Measurement)
	require.Equal(t, "1/1/1", acc.Metrics[1].Tags["groupaddress"])
	require.Len(t, acc.Metrics[1].Fields, 1)
	require.Equal(t, false, acc.Metrics[1].Fields["value"])
}
