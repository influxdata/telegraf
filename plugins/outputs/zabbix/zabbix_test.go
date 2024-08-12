package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"sort"
	"strings"
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

type result struct {
	req zabbixRequest
	err error
}

type zabbixMockServer struct {
	listener          net.Listener
	ignoreAcceptError bool
	results           []result
}

func newZabbixMockServer(addr string, ignoreAcceptError bool) (*zabbixMockServer, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &zabbixMockServer{listener: l, ignoreAcceptError: ignoreAcceptError}, nil
}

func (s *zabbixMockServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *zabbixMockServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
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
		"send one metric with two extra tags, zabbix parameters should be alphabetically ordered": {
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
			server, err := newZabbixMockServer("127.0.0.1:", len(test.zabbixMetrics) == 0)
			require.NoError(t, err)
			defer server.Close()

			z := &Zabbix{
				Address:               server.Addr(),
				KeyPrefix:             test.KeyPrefix,
				HostTag:               "host",
				SkipMeasurementPrefix: test.SkipMeasurementPrefix,
				AgentActive:           test.AgentActive,
				LLDSendInterval:       config.Duration(10 * time.Minute),
				Log:                   testutil.Logger{},
			}
			require.NoError(t, z.Init())

			resCh := make(chan result, 1)
			go func() {
				resCh <- server.listenForSingleRequest()
			}()

			require.NoError(t, z.Write(test.telegrafMetrics))

			// By default, we use trappers
			requestType := "sender data"
			if test.AgentActive {
				requestType = "agent data"
			}

			select {
			case res := <-resCh:
				require.NoError(t, res.err)
				require.Equal(t, requestType, res.req.Request)
				compareData(t, test.zabbixMetrics, res.req.Data)
			case <-time.After(1 * time.Second):
				require.Empty(t, test.zabbixMetrics, "no metrics should be expected if the connection times out")
			}
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
	server, err := newZabbixMockServer("127.0.0.1:", false)
	require.NoError(t, err)
	defer server.Close()
	server.Start()

	z := &Zabbix{
		Address:          server.Addr(),
		KeyPrefix:        "telegraf.",
		HostTag:          "host",
		LLDSendInterval:  config.Duration(10 * time.Minute),
		LLDClearInterval: config.Duration(1 * time.Hour),
		Log:              testutil.Logger{},
	}
	require.NoError(t, z.Init())

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

	require.Eventually(t, func() bool {
		return len(server.results) == 9
	}, 2*time.Second, 50*time.Millisecond, "did not receive 9 results within specific time")

	// Read first packet with two metrics, then the first auto-register packet and the second auto-register packet.
	// First packet with metrics
	require.NoError(t, server.results[0].err)
	compareData(t, []zabbixRequestData{zabbixMetric}, server.results[0].req.Data)

	// Second packet, while time has not surpassed LLDSendInterval
	require.NoError(t, server.results[1].err)
	compareData(t, []zabbixRequestData{zabbixMetric}, server.results[1].req.Data)

	// Third packet, time has surpassed LLDSendInterval, metrics + LLD
	require.NoError(t, server.results[2].err)
	require.Len(t, server.results[2].req.Data, 2, "Expected 2 metrics")
	server.results[2].req.Data[1].Clock = 0 // Ignore lld request clock
	compareData(t, []zabbixRequestData{zabbixMetric, zabbixLLDMetric}, server.results[2].req.Data)

	// Fourth packet with metrics
	require.NoError(t, server.results[3].err)
	compareData(t, []zabbixRequestData{zabbixMetric}, server.results[3].req.Data)

	// Fifth packet, time has surpassed LLDSendInterval, metrics. No LLD as there is nothing new.
	require.NoError(t, server.results[4].err)
	compareData(t, []zabbixRequestData{zabbixMetric}, server.results[4].req.Data)

	// Sixth packet, new LLD info, but time has not surpassed LLDSendInterval
	require.NoError(t, server.results[5].err)
	compareData(t, []zabbixRequestData{zabbixMetricNew}, server.results[5].req.Data)

	// Seventh packet, time has surpassed LLDSendInterval, metrics + LLD.
	// Also, time has surpassed LLDClearInterval, so LLD is cleared.
	require.NoError(t, server.results[6].err)
	require.Len(t, server.results[6].req.Data, 2, "Expected 2 metrics")
	server.results[6].req.Data[1].Clock = 0 // Ignore lld request clock
	compareData(t, []zabbixRequestData{zabbixMetric, zabbixLLDMetricNew}, server.results[6].req.Data)

	// Eighth packet, time host not surpassed LLDSendInterval, just metrics.
	require.NoError(t, server.results[7].err)
	compareData(t, []zabbixRequestData{zabbixMetric}, server.results[7].req.Data)

	// Ninth packet, time has surpassed LLDSendInterval, metrics + LLD.
	// Just the info of the zabbixMetric as zabbixMetricNew has not been seen since LLDClearInterval.
	require.NoError(t, server.results[8].err)
	require.Len(t, server.results[8].req.Data, 2, "Expected 2 metrics")
	server.results[8].req.Data[1].Clock = 0 // Ignore lld request clock
	compareData(t, []zabbixRequestData{zabbixMetric, zabbixLLDMetric}, server.results[8].req.Data)
}

// TestAutoRegister tests that auto-registration requests are sent to zabbix if enabled
func TestAutoRegister(t *testing.T) {
	// Simulate a Zabbix server to get the data sent
	server, err := newZabbixMockServer("127.0.0.1:", false)
	require.NoError(t, err)
	defer server.Close()
	server.Start()

	z := &Zabbix{
		Address:                    server.Addr(),
		KeyPrefix:                  "telegraf.",
		HostTag:                    "host",
		SkipMeasurementPrefix:      false,
		AgentActive:                false,
		Autoregister:               "xxx",
		AutoregisterResendInterval: config.Duration(time.Minute * 5),
		Log:                        testutil.Logger{},
	}
	require.NoError(t, z.Init())
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

	require.Eventually(t, func() bool {
		return len(server.results) == 3
	}, 2*time.Second, 50*time.Millisecond, "did not receive 3 results within specific time")

	// Read first packet with two metrics, then the first auto-register packet and the second auto-register packet.
	// Accept packet with the two metrics sent
	require.NoError(t, server.results[0].err)

	// Read the first auto-register packet
	require.NoError(t, server.results[1].err)
	require.Equal(t, "active checks", server.results[1].req.Request)
	require.Equal(t, "xxx", server.results[1].req.HostMetadata)

	// Read the second auto-register packet
	require.NoError(t, server.results[2].err)
	require.Equal(t, "active checks", server.results[2].req.Request)
	require.Equal(t, "xxx", server.results[2].req.HostMetadata)

	// Check we have received auto-registration for both hosts
	hostsRegistered := []string{server.results[1].req.Host}
	hostsRegistered = append(hostsRegistered, server.results[2].req.Host)
	require.ElementsMatch(t, []string{"hostA", "hostB"}, hostsRegistered)
}

// compareData compares generated data with expected data ignoring slice order if all Clocks are the same.
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
	// That array order depends on the access to a map, so it does not have a defined order.
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

func (s *zabbixMockServer) Start() {
	go func() {
		defer s.listener.Close()
		for {
			res := s.listenForSingleRequest()
			s.results = append(s.results, res)
		}
	}()
}

func (s *zabbixMockServer) listenForSingleRequest() result {
	conn, err := s.listener.Accept()
	if err != nil {
		if s.ignoreAcceptError {
			return result{req: zabbixRequest{}, err: nil}
		}
		return result{req: zabbixRequest{}, err: err}
	}
	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return result{req: zabbixRequest{}, err: err}
	}

	// Obtain request from the mock zabbix server
	// Read protocol header and version
	header := make([]byte, 5)
	_, err = conn.Read(header)
	if err != nil {
		return result{req: zabbixRequest{}, err: err}
	}

	// Read data length
	dataLengthRaw := make([]byte, 8)
	_, err = conn.Read(dataLengthRaw)
	if err != nil {
		return result{req: zabbixRequest{}, err: err}
	}

	dataLength := binary.LittleEndian.Uint64(dataLengthRaw)

	// Read data content
	content := make([]byte, dataLength)
	_, err = conn.Read(content)
	if err != nil {
		return result{req: zabbixRequest{}, err: err}
	}

	// The zabbix output checks that there are not errors
	// Simulated response from the server
	resp := []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\": \"success\", \"info\": \"\"}\n")
	_, err = conn.Write(resp)
	if err != nil {
		return result{req: zabbixRequest{}, err: err}
	}

	// Strip zabbix header and get JSON request
	var request zabbixRequest
	err = json.Unmarshal(content, &request)
	if err != nil {
		return result{req: zabbixRequest{}, err: err}
	}

	return result{req: request, err: nil}
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
	require.Equal(t, keyPrefix+"name.value[b,bar]", zm.Key)

	zm, err = z.buildZabbixMetric(testutil.MustMetric(
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
