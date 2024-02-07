package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

type zabbixRequestData struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock"`
}

type zabbixRequest struct {
	Request      string              `json:"request"`
	Data         []zabbixRequestData `json:"data"`
	Clock        int                 `json:"clock"`
	Host         string              `json:"host"`
	HostMetadata string              `json:"host_metadata"`
}

type zabbixLLDValue struct {
	Data []map[string]string `json:"data"`
}

func TestZabbix(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)

	tests := map[string]struct {
		KeyPrefix             string
		AgentActive           bool
		SkipMeasurementPrefix bool
		telegrafMetrics       []telegraf.Metric
		zabbixMetrics         []zabbixRequestData
	}{
		"send one metric with one field and no extra tags, generates one zabbix metric": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"string values representing a float number should be sent in the exact same format": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": "3.1415",
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value",
					Value: "3.1415",
					Clock: 1522082244,
				},
			},
		},
		"send one metric with one string field and no extra tags, generates one zabbix metric": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": "some value",
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value",
					Value: "some value",
					Clock: 1522082244,
				},
			},
		},
		"boolean values are converted to 1 (true) or 0 (false)": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"valueTrue":  true,
						"valueFalse": false,
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.valueTrue",
					Value: "true",
					Clock: 1522082244,
				},
				{
					Host:  "hostname",
					Key:   "name.valueFalse",
					Value: "false",
					Clock: 1522082244,
				},
			},
		},
		"invalid value data is ignored and not sent": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": []int{1, 2},
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{},
		},
		"metrics without host tag use the system hostname": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{},
					map[string]interface{}{
						"value": "x",
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  hostname,
					Key:   "name.value",
					Value: "x",
					Clock: 1522082244,
				},
			},
		},
		"send one metric with extra tags, zabbix metric should be generated with a parameter": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
						"foo":  "bar",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value[bar]",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"send one metric with two extra tags, zabbix parameters should be alfabetically orderer": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host":   "hostname",
						"zparam": "last",
						"aparam": "first",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value[first,last]",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"send one metric with two fields and no extra tags, generates two zabbix metrics": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"valueA": int64(0),
						"valueB": int64(1),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.valueA",
					Value: "0",
					Clock: 1522082244,
				},
				{
					Host:  "hostname",
					Key:   "name.valueB",
					Value: "1",
					Clock: 1522082244,
				},
			},
		},
		"send two metrics with one field and no extra tags, generates two zabbix metrics": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("nameA",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				testutil.MustMetric("nameB",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "nameA.value",
					Value: "0",
					Clock: 1522082244,
				},
				{
					Host:  "hostname",
					Key:   "nameB.value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"send two metrics with different hostname, generates two zabbix metrics for different hosts": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostnameA",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostnameB",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostnameA",
					Key:   "name.value",
					Value: "0",
					Clock: 1522082244,
				},
				{
					Host:  "hostnameB",
					Key:   "name.value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"if key_prefix is configured, zabbix metrics should have that prefix in the key": {
			KeyPrefix: "telegraf.",
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "telegraf.name.value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"if skip_measurement_prefix is configured, zabbix metrics should have to skip that prefix in the key": {
			SkipMeasurementPrefix: true,
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"if AgentActive is configured, zabbix metrics should be sent respecting that protocol": {
			AgentActive: true,
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"metrics should be time sorted, oldest to newest, to avoid zabbix doing extra work when generating trends": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostnameD",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(4444444444, 0),
				),
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostnameC",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(3333333333, 0),
				),
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostnameA",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1111111111, 0),
				),
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostnameB",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(2222222222, 0),
				),
			},
			zabbixMetrics: []zabbixRequestData{
				{
					Host:  "hostnameA",
					Key:   "name.value",
					Value: "0",
					Clock: 1111111111,
				},
				{
					Host:  "hostnameB",
					Key:   "name.value",
					Value: "0",
					Clock: 2222222222,
				},
				{
					Host:  "hostnameC",
					Key:   "name.value",
					Value: "0",
					Clock: 3333333333,
				},
				{
					Host:  "hostnameD",
					Key:   "name.value",
					Value: "0",
					Clock: 4444444444,
				},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// Simulate a Zabbix server to get the data sent. It has a timeout to avoid waiting forever.
			listener, err := net.Listen("tcp", "127.0.0.1:")
			require.NoError(t, err)
			defer listener.Close()

			z := &Zabbix{
				Address:               listener.Addr().String(),
				KeyPrefix:             test.KeyPrefix,
				HostTag:               "host",
				SkipMeasurementPrefix: test.SkipMeasurementPrefix,
				AgentActive:           test.AgentActive,
				LLDSendInterval:       config.Duration(10 * time.Minute),
				Log:                   testutil.Logger{},
			}
			require.NoError(t, z.Init())

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				success := make(chan zabbixRequest, 1)

				go func() {
					success <- listenForZabbixMetric(t, listener, len(test.zabbixMetrics) == 0)
				}()

				// By default we use trappers
				requestType := "sender data"
				if test.AgentActive {
					requestType = "agent data"
				}

				select {
				case request := <-success:
					require.Equal(t, requestType, request.Request)
					compareData(t, test.zabbixMetrics, request.Data)
				case <-time.After(1 * time.Second):
					require.Empty(t, test.zabbixMetrics, "no metrics should be expected if the connection times out")
				}

				wg.Done()
			}()

			require.NoError(t, z.Write(test.telegrafMetrics))

			// Wait for zabbix server emulator to finish
			wg.Wait()
		})
	}
}

// TestLLD tests how LLD metrics are sent simulating the time passing.
// LLD is sent each LLDSendInterval. Only new data.
// LLD data is cleared LLDClearInterval.
func TestLLD(t *testing.T) {
	// Telegraf metric which will be sent repeatedly
	m := testutil.MustMetric(
		"name",
		map[string]string{"host": "hostA", "foo": "bar"},
		map[string]interface{}{"value": int64(0)},
		time.Unix(0, 0),
	)

	mNew := testutil.MustMetric(
		"name",
		map[string]string{"host": "hostA", "foo": "moo"},
		map[string]interface{}{"value": int64(0)},
		time.Unix(0, 0),
	)

	// Expected Zabbix metric generated
	zabbixMetric := zabbixRequestData{
		Host:  "hostA",
		Key:   "telegraf.name.value[bar]",
		Value: "0",
		Clock: 0,
	}

	// Expected Zabbix metric generated
	zabbixMetricNew := zabbixRequestData{
		Host:  "hostA",
		Key:   "telegraf.name.value[moo]",
		Value: "0",
		Clock: 0,
	}

	// Expected Zabbix LLD metric generated
	zabbixLLDMetric := zabbixRequestData{
		Host:  "hostA",
		Key:   "telegraf.lld.name.foo",
		Value: `{"data":[{"{#FOO}":"bar"}]}`,
		Clock: 0,
	}

	// Expected Zabbix LLD metric generated
	zabbixLLDMetricNew := zabbixRequestData{
		Host:  "hostA",
		Key:   "telegraf.lld.name.foo",
		Value: `{"data":[{"{#FOO}":"bar"},{"{#FOO}":"moo"}]}`,
		Clock: 0,
	}

	// Simulate a Zabbix server to get the data sent
	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer listener.Close()

	z := &Zabbix{
		Address:          listener.Addr().String(),
		KeyPrefix:        "telegraf.",
		HostTag:          "host",
		LLDSendInterval:  config.Duration(10 * time.Minute),
		LLDClearInterval: config.Duration(1 * time.Hour),
		Log:              testutil.Logger{},
	}
	require.NoError(t, z.Init())

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Read first packet with two metrics, then the first autoregister packet and the second autoregister packet.
	go func() {
		// First packet with metrics
		request := listenForZabbixMetric(t, listener, false)
		compareData(t, []zabbixRequestData{zabbixMetric}, request.Data)

		// Second packet, while time has not surpassed LLDSendInterval
		request = listenForZabbixMetric(t, listener, false)
		compareData(t, []zabbixRequestData{zabbixMetric}, request.Data)

		// Third packet, time has surpassed LLDSendInterval, metrics + LLD
		request = listenForZabbixMetric(t, listener, false)
		require.Len(t, request.Data, 2, "Expected 2 metrics")
		request.Data[1].Clock = 0 // Ignore lld request clock
		compareData(t, []zabbixRequestData{zabbixMetric, zabbixLLDMetric}, request.Data)

		// Fourth packet with metrics
		request = listenForZabbixMetric(t, listener, false)
		compareData(t, []zabbixRequestData{zabbixMetric}, request.Data)

		// Fifth packet, time has surpassed LLDSendInterval, metrics. No LLD as there is nothing new.
		request = listenForZabbixMetric(t, listener, false)
		compareData(t, []zabbixRequestData{zabbixMetric}, request.Data)

		// Sixth packet, new LLD info, but time has not surpassed LLDSendInterval
		request = listenForZabbixMetric(t, listener, false)
		compareData(t, []zabbixRequestData{zabbixMetricNew}, request.Data)

		// Seventh packet, time has surpassed LLDSendInterval, metrics + LLD.
		// Also, time has surpassed LLDClearInterval, so LLD is cleared.
		request = listenForZabbixMetric(t, listener, false)
		require.Len(t, request.Data, 2, "Expected 2 metrics")
		request.Data[1].Clock = 0 // Ignore lld request clock
		compareData(t, []zabbixRequestData{zabbixMetric, zabbixLLDMetricNew}, request.Data)

		// Eighth packet, time host not surpassed LLDSendInterval, just metrics.
		request = listenForZabbixMetric(t, listener, false)
		compareData(t, []zabbixRequestData{zabbixMetric}, request.Data)

		// Ninth packet, time has surpassed LLDSendInterval, metrics + LLD.
		// Just the info of the zabbixMetric as zabbixMetricNew has not been seen since LLDClearInterval.
		request = listenForZabbixMetric(t, listener, false)
		require.Len(t, request.Data, 2, "Expected 2 metrics")
		request.Data[1].Clock = 0 // Ignore lld request clock
		compareData(t, []zabbixRequestData{zabbixMetric, zabbixLLDMetric}, request.Data)

		wg.Done()
	}()

	// First packet
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Second packet, while time has not surpassed LLDSendInterval
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Simulate time passing for a new LLD send
	z.lldLastSend = time.Now().Add(-time.Duration(z.LLDSendInterval)).Add(-time.Millisecond)

	// Third packet, time has surpassed LLDSendInterval, metrics + LLD
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Fourth packet
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Simulate time passing for a new LLD send
	z.lldLastSend = time.Now().Add(-time.Duration(z.LLDSendInterval)).Add(-time.Millisecond)

	// Fifth packet, time has surpassed LLDSendInterval, metrics. No LLD as there is nothing new.
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Sixth packet, new LLD info, but time has not surpassed LLDSendInterval
	require.NoError(t, z.Write([]telegraf.Metric{mNew}))

	// Simulate time passing for LLD clear
	z.lldLastSend = time.Now().Add(-time.Duration(z.LLDClearInterval)).Add(-time.Millisecond)

	// Seventh packet, time has surpassed LLDSendInterval and LLDClearInterval, metrics + LLD.
	// LLD will be cleared.
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Eighth packet, time host not surpassed LLDSendInterval, just metrics.
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Simulate time passing for a new LLD send
	z.lldLastSend = time.Now().Add(-time.Duration(z.LLDSendInterval)).Add(-time.Millisecond)

	// Ninth packet, time has surpassed LLDSendInterval, metrics + LLD.
	require.NoError(t, z.Write([]telegraf.Metric{m}))

	// Wait for zabbix server emulator to finish
	wg.Wait()
}

// TestAutoregister tests that autoregistration requests are sent to zabbix if enabled
func TestAutoregister(t *testing.T) {
	// Simulate a Zabbix server to get the data sent
	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer listener.Close()

	z := &Zabbix{
		Address:                    listener.Addr().String(),
		KeyPrefix:                  "telegraf.",
		HostTag:                    "host",
		SkipMeasurementPrefix:      false,
		AgentActive:                false,
		Autoregister:               "xxx",
		AutoregisterResendInterval: config.Duration(time.Minute * 5),
		Log:                        testutil.Logger{},
	}
	require.NoError(t, z.Init())

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Read first packet with two metrics, then the first autoregister packet and the second autoregister packet.
	go func() {
		// Accept packet with the two metrics sent
		_ = listenForZabbixMetric(t, listener, false)

		// Read the first autoregister packet
		request := listenForZabbixMetric(t, listener, false)
		require.Equal(t, "active checks", request.Request)
		require.Equal(t, "xxx", request.HostMetadata)

		hostsRegistered := []string{request.Host}

		// Read the second autoregister packet
		request = listenForZabbixMetric(t, listener, false)
		require.Equal(t, "active checks", request.Request)
		require.Equal(t, "xxx", request.HostMetadata)

		// Check we have received autoregistration for both hosts
		hostsRegistered = append(hostsRegistered, request.Host)
		require.ElementsMatch(t, []string{"hostA", "hostB"}, hostsRegistered)

		wg.Done()
	}()

	err = z.Write([]telegraf.Metric{
		testutil.MustMetric(
			"name",
			map[string]string{"host": "hostA"},
			map[string]interface{}{"value": int64(0)},
			time.Now(),
		),
		testutil.MustMetric(
			"name",
			map[string]string{"host": "hostB"},
			map[string]interface{}{"value": int64(0)},
			time.Now(),
		),
	})
	require.NoError(t, err)

	// Wait for zabbix server emulator to finish
	wg.Wait()
}

// compareData compares generated data with expected data ignoring slice order if all Clocks are
// the same.
// This is useful for metrics with several fields that should produce several Zabbix values that
// could not be sorted by clock
func compareData(t *testing.T, expected []zabbixRequestData, data []zabbixRequestData) {
	t.Helper()

	var clock int64

	sameClock := true

	// Check if all clocks are the same
	for i := 0; i < len(data); i++ {
		if i == 0 {
			clock = data[i].Clock
		} else if clock != data[i].Clock {
			sameClock = false

			break
		}
	}

	// Zabbix requests with LLD data contains a JSON value with an array of dictionaries.
	// That array order depends in the access to a map, so it does not have a defined order.
	// To compare the data, we need to sort the array of dictionaries.
	// Before comparing the requests, sort those values.
	// To detect if a request contains LLD data, try to unmarshal it to a ZabbixLLDValue.
	// If it could be unmarshalled, sort the slice and marshal it again.
	for i := 0; i < len(data); i++ {
		var lldValue zabbixLLDValue

		err := json.Unmarshal([]byte(data[i].Value), &lldValue)
		if err == nil {
			sort.Slice(lldValue.Data, func(i, j int) bool {
				// Generate a global order based on the keys and values present in the map
				keysValuesI := make([]string, 0, len(lldValue.Data[i])*2)
				keysValuesJ := make([]string, 0, len(lldValue.Data[j])*2)
				for k, v := range lldValue.Data[i] {
					keysValuesI = append(keysValuesI, k, v)
				}
				for k, v := range lldValue.Data[j] {
					keysValuesJ = append(keysValuesJ, k, v)
				}

				sort.Strings(keysValuesI)
				sort.Strings(keysValuesJ)

				return strings.Join(keysValuesI, "") < strings.Join(keysValuesJ, "")
			})
			sortedValue, err := json.Marshal(lldValue)
			require.NoError(t, err)

			data[i].Value = string(sortedValue)
		}
	}

	if sameClock {
		require.ElementsMatch(t, expected, data)
	} else {
		require.Equal(t, expected, data)
	}
}

// listenForZabbixMetric starts a TCP server listening for one Zabbix metric.
// ignoreAcceptError is used to ignore the error when the server is closed.
func listenForZabbixMetric(t *testing.T, listener net.Listener, ignoreAcceptError bool) zabbixRequest {
	t.Helper()

	conn, err := listener.Accept()
	if err != nil && ignoreAcceptError {
		return zabbixRequest{}
	}

	require.NoError(t, err)

	// Obtain request from the mock zabbix server
	// Read protocol header and version
	header := make([]byte, 5)
	_, err = conn.Read(header)
	require.NoError(t, err)

	// Read data length
	dataLengthRaw := make([]byte, 8)
	_, err = conn.Read(dataLengthRaw)
	require.NoError(t, err)

	dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

	// Read data content
	content := make([]byte, dataLength)
	_, err = conn.Read(content)
	require.NoError(t, err)

	// The zabbix output checks that there are not errors
	// Simulated response from the server
	resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\": \"success\", \"info\": \"\"}\n")
	_, err = conn.Write(resp)
	require.NoError(t, err)

	// Close connection after reading the client data
	conn.Close()

	// Strip zabbix header and get JSON request
	var request zabbixRequest
	require.NoError(t, json.Unmarshal(content, &request))

	return request
}

func TestBuildZabbixMetric(t *testing.T) {
	keyPrefix := "prefix."
	hostTag := "host"

	z := &Zabbix{
		KeyPrefix: keyPrefix,
		HostTag:   hostTag,
	}

	zm, err := z.buildZabbixMetric(testutil.MustMetric(
		"name",
		map[string]string{hostTag: "hostA", "foo": "bar", "a": "b"},
		map[string]interface{}{},
		time.Now()),
		"value",
		1,
	)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%sname.value[b,bar]", keyPrefix), zm.Key)

	zm, err = z.buildZabbixMetric(testutil.MustMetric(
		"name",
		map[string]string{hostTag: "hostA"},
		map[string]interface{}{},
		time.Now()),
		"value",
		1,
	)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%sname.value", keyPrefix), zm.Key)
}

func TestGetHostname(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)

	tests := map[string]struct {
		HostTag string
		Host    string
		Tags    map[string]string
		Result  string
	}{
		"metric with host tag": {
			HostTag: "host",
			Tags: map[string]string{
				"host": "bar",
			},
			Result: "bar",
		},
		"metric with host tag changed": {
			HostTag: "source",
			Tags: map[string]string{
				"source": "bar",
			},
			Result: "bar",
		},
		"metric with no host tag": {
			Tags:   map[string]string{},
			Result: hostname,
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			metric := testutil.MustMetric(
				"name",
				test.Tags,
				map[string]interface{}{},
				time.Now(),
			)

			host, err := getHostname(test.HostTag, metric)
			require.NoError(t, err)
			require.Equal(t, test.Result, host)
		})
	}
}
