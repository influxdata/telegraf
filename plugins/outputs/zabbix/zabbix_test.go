package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/datadope-io/go-zabbix/v2"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestSuccessfulReceive(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)

	tests := []struct {
		name                  string
		prefix                string
		agentActive           bool
		skipMeasurementPrefix bool
		input                 []telegraf.Metric
		expected              []zabbix.Packet
	}{
		{
			name: "send one metric with one field and no extra tags, generates one zabbix metric",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "name.value",
							Value: "0",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "string values representing a float number should be sent in the exact same format",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": "3.1415",
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "name.value",
							Value: "3.1415",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "send one metric with one string field and no extra tags, generates one zabbix metric",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": "some value",
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "name.value",
							Value: "some value",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "boolean values are converted to 1 (true) or 0 (false)",
			input: []telegraf.Metric{
				metric.New(
					"name",
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
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
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
			},
		},
		{
			name: "metrics without host tag use the system hostname",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{},
					map[string]interface{}{
						"value": "x",
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  hostname,
							Key:   "name.value",
							Value: "x",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "send one metric with extra tags, zabbix metric should be generated with a parameter",
			input: []telegraf.Metric{
				metric.New(
					"name",
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
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "name.value[bar]",
							Value: "0",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "send one metric with two extra tags, zabbix parameters should be alphabetically ordered",
			input: []telegraf.Metric{
				metric.New(
					"name",
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
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "name.value[first,last]",
							Value: "0",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "send one metric with two fields and no extra tags, generates two zabbix metrics",
			input: []telegraf.Metric{
				metric.New(
					"name",
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
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
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
			},
		},
		{
			name: "send two metrics with one field and no extra tags, generates two zabbix metrics",
			input: []telegraf.Metric{
				metric.New(
					"nameA",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				metric.New(
					"nameB",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
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
			},
		},
		{
			name: "send two metrics with different hostname, generates two zabbix metrics for different hosts",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostnameA",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				metric.New(
					"name",
					map[string]string{
						"host": "hostnameB",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
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
			},
		},
		{
			name:   "if key_prefix is configured, zabbix metrics should have that prefix in the key",
			prefix: "telegraf.",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "telegraf.name.value",
							Value: "0",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name:                  "if skip_measurement_prefix is configured, zabbix metrics should have to skip that prefix in the key",
			skipMeasurementPrefix: true,
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "value",
							Value: "0",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name:        "if AgentActive is configured, zabbix metrics should be sent respecting that protocol",
			agentActive: true,
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "agent data",
					Data: []*zabbix.Metric{
						{
							Host:  "hostname",
							Key:   "name.value",
							Value: "0",
							Clock: 1522082244,
						},
					},
				},
			},
		},
		{
			name: "metrics should be time sorted, oldest to newest, to avoid zabbix doing extra work when generating trends",
			input: []telegraf.Metric{
				metric.New(
					"name",
					map[string]string{
						"host": "hostnameD",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(4444444444, 0),
				),
				metric.New(
					"name",
					map[string]string{
						"host": "hostnameC",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(3333333333, 0),
				),
				metric.New(
					"name",
					map[string]string{
						"host": "hostnameA",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1111111111, 0),
				),
				metric.New(
					"name",
					map[string]string{
						"host": "hostnameB",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(2222222222, 0),
				),
			},
			expected: []zabbix.Packet{
				{
					Request: "sender data",
					Data: []*zabbix.Metric{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a Zabbix mock server and start listening
			server, err := newZabbixMockServer()
			require.NoError(t, err)
			defer server.close()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				server.listen()
			}()

			// Setup the plugin
			plugin := &Zabbix{
				Address:               server.addr(),
				KeyPrefix:             tt.prefix,
				HostTag:               "host",
				SkipMeasurementPrefix: tt.skipMeasurementPrefix,
				AgentActive:           tt.agentActive,
				LLDSendInterval:       config.Duration(10 * time.Minute),
				Log:                   testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Connect and write the data
			require.NoError(t, plugin.Connect())
			defer plugin.Close()
			require.NoError(t, plugin.Write(tt.input))

			// Wait for the data to arrive
			require.Eventually(t, func() bool {
				return server.count.Load() > 0
			}, 3*time.Second, 100*time.Millisecond, "nothing received")

			// Stop listening
			server.listener.Close()
			wg.Wait()

			// Check the received data
			server.Lock()
			defer server.Unlock()

			require.Empty(t, server.errs, "server had errors")
			requireRequestDataEqual(t, tt.expected, server.received, false)
		})
	}
}

func TestInvalidData(t *testing.T) {
	input := []telegraf.Metric{
		metric.New(
			"name",
			map[string]string{
				"host": "hostname",
			},
			map[string]interface{}{
				"value": []int{1, 2},
			},
			time.Unix(1522082244, 0),
		),
	}

	// Setup a Zabbix mock server and start listening
	server, err := newZabbixMockServer()
	require.NoError(t, err)
	defer server.close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.listen()
	}()

	// Setup the plugin
	plugin := &Zabbix{
		Address:         server.addr(),
		HostTag:         "host",
		LLDSendInterval: config.Duration(10 * time.Minute),
		Log:             testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Connect and write the data
	require.NoError(t, plugin.Connect())
	defer plugin.Close()
	require.NoError(t, plugin.Write(input))
	require.NoError(t, plugin.Close())

	// Stop listening
	server.listener.Close()
	wg.Wait()

	// Check the received data
	server.Lock()
	defer server.Unlock()

	require.Empty(t, server.errs, "server had errors")
	require.Empty(t, server.received)
}

// TestLLD tests how LLD metrics are sent simulating the time passing.
// LLD is sent each LLDSendInterval. Only new data.
// LLD data is cleared LLDClearInterval.
func TestLLD(t *testing.T) {
	// Telegraf metric which will be sent repeatedly
	m := metric.New(
		"name",
		map[string]string{"host": "hostA", "foo": "bar"},
		map[string]interface{}{"value": int64(0)},
		time.Unix(0, 0),
	)

	mNew := metric.New(
		"name",
		map[string]string{"host": "hostA", "foo": "moo"},
		map[string]interface{}{"value": int64(0)},
		time.Unix(0, 0),
	)

	// Expected Zabbix metric generated
	expected := []zabbix.Packet{
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
				{
					Host:  "hostA",
					Key:   "telegraf.lld.name.foo",
					Value: `{"data":[{"{#FOO}":"bar"}]}`,
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[moo]",
					Value: "0",
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
				{
					Host:  "hostA",
					Key:   "telegraf.lld.name.foo",
					Value: `{"data":[{"{#FOO}":"bar"},{"{#FOO}":"moo"}]}`,
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
			},
		},
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value[bar]",
					Value: "0",
				},
				{
					Host:  "hostA",
					Key:   "telegraf.lld.name.foo",
					Value: `{"data":[{"{#FOO}":"bar"}]}`,
				},
			},
		},
	}

	// Setup a Zabbix mock server and start listening
	server, err := newZabbixMockServer()
	require.NoError(t, err)
	defer server.close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.listen()
	}()

	// Setup plugin
	plugin := &Zabbix{
		Address:          server.addr(),
		KeyPrefix:        "telegraf.",
		HostTag:          "host",
		LLDSendInterval:  config.Duration(10 * time.Minute),
		LLDClearInterval: config.Duration(1 * time.Hour),
		Log:              testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Connect and write the metrics
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// First packet
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Second packet, while time has not surpassed LLDSendInterval
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Simulate time passing for a new LLD send
	plugin.lldLastSend = time.Now().Add(-time.Duration(plugin.LLDSendInterval)).Add(-time.Millisecond)

	// Third packet, time has surpassed LLDSendInterval, metrics + LLD
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Fourth packet
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Simulate time passing for a new LLD send
	plugin.lldLastSend = time.Now().Add(-time.Duration(plugin.LLDSendInterval)).Add(-time.Millisecond)

	// Fifth packet, time has surpassed LLDSendInterval, metrics. No LLD as there is nothing new.
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Sixth packet, new LLD info, but time has not surpassed LLDSendInterval
	require.NoError(t, plugin.Write([]telegraf.Metric{mNew}))

	// Simulate time passing for LLD clear
	plugin.lldLastSend = time.Now().Add(-time.Duration(plugin.LLDClearInterval)).Add(-time.Millisecond)

	// Seventh packet, time has surpassed LLDSendInterval and LLDClearInterval, metrics + LLD.
	// LLD will be cleared.
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Eighth packet, time host not surpassed LLDSendInterval, just metrics.
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Simulate time passing for a new LLD send
	plugin.lldLastSend = time.Now().Add(-time.Duration(plugin.LLDSendInterval)).Add(-time.Millisecond)

	// Ninth packet, time has surpassed LLDSendInterval, metrics + LLD.
	require.NoError(t, plugin.Write([]telegraf.Metric{m}))

	// Wait for the metrics to be received
	require.Eventuallyf(t, func() bool {
		return server.count.Load() >= uint32(len(expected))
	}, 3*time.Second, 100*time.Millisecond, "expected %d got %d", len(expected), server.count.Load())

	// Stop listening
	require.NoError(t, plugin.Close())
	server.listener.Close()
	wg.Wait()

	// Check the received metrics
	server.Lock()
	defer server.Unlock()

	require.Empty(t, server.errs, "server had errors")
	requireRequestDataEqual(t, expected, server.received, true)
}

// TestAutoRegister tests that auto-registration requests are sent to zabbix if enabled
func TestAutoRegister(t *testing.T) {
	now := time.Now()
	input := []telegraf.Metric{
		metric.New(
			"name",
			map[string]string{"host": "hostA"},
			map[string]interface{}{"value": int64(0)},
			now,
		),
		metric.New(
			"name",
			map[string]string{"host": "hostB"},
			map[string]interface{}{"value": int64(42)},
			now,
		),
	}

	expected := []zabbix.Packet{
		{
			Request: "sender data",
			Data: []*zabbix.Metric{
				{
					Host:  "hostA",
					Key:   "telegraf.name.value",
					Value: "0",
					Clock: now.Unix(),
				},
				{
					Host:  "hostB",
					Key:   "telegraf.name.value",
					Value: "42",
					Clock: now.Unix(),
				},
			},
		},
		{
			Request:      "active checks",
			Host:         "hostA",
			HostMetadata: "xxx",
		},
		{
			Request:      "active checks",
			Host:         "hostB",
			HostMetadata: "xxx",
		},
	}

	// Setup a Zabbix mock server and start listening
	server, err := newZabbixMockServer()
	require.NoError(t, err)
	defer server.close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.listen()
	}()

	// Setup plugin
	plugin := &Zabbix{
		Address:                    server.addr(),
		KeyPrefix:                  "telegraf.",
		HostTag:                    "host",
		Autoregister:               "xxx",
		AutoregisterResendInterval: config.Duration(time.Minute * 5),
		Log:                        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Connect and write the metrics
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(input))

	// Wait for the metrics to be received
	require.Eventuallyf(t, func() bool {
		return server.count.Load() >= uint32(len(expected))
	}, 3*time.Second, 100*time.Millisecond, "expected %d got %d", len(expected), server.count.Load())

	// Stop listening
	require.NoError(t, plugin.Close())
	server.listener.Close()
	wg.Wait()

	// Check the received metrics
	server.Lock()
	defer server.Unlock()

	require.Empty(t, server.errs, "server had errors")
	actual := server.received
	sort.SliceStable(expected, func(i, j int) bool { return expected[i].Host < expected[j].Host })
	sort.SliceStable(actual, func(i, j int) bool { return actual[i].Host < actual[j].Host })
	requireRequestDataEqual(t, expected, actual, false)
}

func TestBuildZabbixMetric(t *testing.T) {
	keyPrefix := "prefix."
	hostTag := "host"

	z := &Zabbix{
		KeyPrefix: keyPrefix,
		HostTag:   hostTag,
	}

	zm, err := z.buildZabbixMetric(metric.New(

		"name",
		map[string]string{hostTag: "hostA", "foo": "bar", "a": "b"},
		map[string]interface{}{},
		time.Now()),
		"value",
		1,
	)
	require.NoError(t, err)
	require.Equal(t, keyPrefix+"name.value[b,bar]", zm.Key)

	zm, err = z.buildZabbixMetric(metric.New(

		"name",
		map[string]string{hostTag: "hostA"},
		map[string]interface{}{},
		time.Now()),
		"value",
		1,
	)
	require.NoError(t, err)
	require.Equal(t, keyPrefix+"name.value", zm.Key)
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
			metric := metric.New(

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

func TestCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get all testcase directories
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	outputs.Add("zabbix", func() telegraf.Output {
		return &Zabbix{
			KeyPrefix:                  "telegraf.",
			HostTag:                    "host",
			AutoregisterResendInterval: config.Duration(time.Minute * 30),
			LLDSendInterval:            config.Duration(time.Minute * 10),
			LLDClearInterval:           config.Duration(time.Hour),
		}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			inputFilename := filepath.Join(testcasePath, "input.influx")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Load the input data
			input, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)

			// Read the expected output if any
			var expected []zabbix.Packet
			if _, err := os.Stat(expectedFilename); err == nil {
				buf, err := os.ReadFile(expectedFilename)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(buf, &expected))
			}

			// Read the expected output if any
			var expectedError string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				expectedErrors, err := testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.Len(t, expectedErrors, 1)
				expectedError = expectedErrors[0]
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Outputs, 1)

			// Setup a Zabbix mock server and start listening
			server, err := newZabbixMockServer()
			require.NoError(t, err)
			defer server.close()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				server.listen()
			}()
			defer server.listener.Close()

			// Setup the plugin
			plugin := cfg.Outputs[0].Output.(*Zabbix)
			plugin.Address = server.addr()
			plugin.Log = testutil.Logger{}
			require.NoError(t, plugin.Init())

			// Connect and write the metric(s)
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			err = plugin.Write(input)
			if expectedError != "" {
				require.ErrorContains(t, err, expectedError)
				return
			}
			require.NoError(t, err)

			// Wait for the data to arrive
			require.Eventuallyf(t, func() bool {
				return server.count.Load() >= uint32(len(expected))
			}, 3*time.Second, 100*time.Millisecond, "expected %d got %d", len(expected), server.count.Load())

			server.listener.Close()
			wg.Wait()

			// Check the received data
			server.Lock()
			defer server.Unlock()
			require.Empty(t, server.errs, "server had errors")
			requireRequestDataEqual(t, expected, server.received, false)
		})
	}
}

type zabbixMockServer struct {
	listener net.Listener

	received []zabbix.Packet
	errs     []error
	count    atomic.Uint32
	sync.Mutex
}

func newZabbixMockServer() (*zabbixMockServer, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return &zabbixMockServer{listener: l}, nil
}

func (s *zabbixMockServer) addr() string {
	return s.listener.Addr().String()
}

func (s *zabbixMockServer) close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *zabbixMockServer) listen() {
	for {
		request, err := s.handle()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				s.Lock()
				s.errs = append(s.errs, err)
				s.Unlock()
			}
			return
		}
		s.Lock()
		s.received = append(s.received, request)
		s.Unlock()
		s.count.Store(uint32(len(s.received)))
	}
}

func (s *zabbixMockServer) handle() (zabbix.Packet, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return zabbix.Packet{}, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return zabbix.Packet{}, err
	}

	// Obtain request from the mock zabbix server
	// Read protocol header and version
	header := make([]byte, 5)
	if _, err := conn.Read(header); err != nil {
		return zabbix.Packet{}, err
	}

	// Read data length
	dataLengthRaw := make([]byte, 8)
	if _, err := conn.Read(dataLengthRaw); err != nil {
		return zabbix.Packet{}, err
	}
	dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

	// Read data content
	content := make([]byte, dataLength)
	if _, err := conn.Read(content); err != nil {
		return zabbix.Packet{}, err
	}

	// The zabbix output checks that there are not errors
	// Simulated response from the server
	resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\": \"success\", \"info\": \"\"}\n")
	if _, err := conn.Write(resp); err != nil {
		return zabbix.Packet{}, err
	}

	// Strip zabbix header and get JSON request
	var request zabbix.Packet
	if err := json.Unmarshal(content, &request); err != nil {
		return zabbix.Packet{}, err
	}

	return request, nil
}

type lldValue struct {
	Data []map[string]string `json:"data"`
}

func requireRequestDataEqual(t *testing.T, expected, actual []zabbix.Packet, ignoreClock bool) {
	t.Helper()
	require.Len(t, actual, len(expected))
	for i, expectedReq := range expected {
		actualReq := actual[i]
		require.Equalf(t, expectedReq.Request, actualReq.Request, "different request types in request %d", i)
		require.Equalf(t, expectedReq.Host, actualReq.Host, "different host in request %d", i)
		require.Equalf(t, expectedReq.HostMetadata, actualReq.HostMetadata, "different hostmetadata in request %d", i)

		// Check the elements
		require.Len(t, actualReq.Data, len(expectedReq.Data))

		less := func(a, b *zabbix.Metric) bool {
			if a.Key == b.Key {
				if a.Clock == b.Clock {
					return a.Value < b.Value
				}
				return a.Clock < b.Clock
			}
			return a.Key < b.Key
		}
		sort.SliceStable(actualReq.Data, func(i, j int) bool { return less(actualReq.Data[i], actualReq.Data[j]) })
		sort.SliceStable(expectedReq.Data, func(i, j int) bool { return less(expectedReq.Data[i], expectedReq.Data[j]) })
		for j, expectedData := range expectedReq.Data {
			actualData := actualReq.Data[j]
			require.Equalf(t, expectedData.Key, actualData.Key, "different key in request %d, data %d", i, j)
			require.Equalf(t, expectedData.Host, actualData.Host, "different host in request %d, data %d", i, j)
			if !ignoreClock {
				require.Equalf(t, expectedData.Clock, actualData.Clock, "different clock in request %d, data %d", i, j)
			}
			if strings.HasPrefix(expectedData.Value, "{") {
				var actualValue, expectedValue lldValue
				require.NoError(t, json.Unmarshal([]byte(actualData.Value), &actualValue))
				require.NoError(t, json.Unmarshal([]byte(expectedData.Value), &expectedValue))
				require.ElementsMatchf(t, expectedValue.Data, actualValue.Data, "different value in request %d, data %d", i, j)
			} else {
				require.Equalf(t, expectedData.Value, actualData.Value, "different value in request %d, data %d", i, j)
			}
		}
	}
}
