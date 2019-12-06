package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ZabbixRequestData struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock"`
}

type ZabbixRequest struct {
	Request      string              `json:"request"`
	Data         []ZabbixRequestData `json:"data"`
	Clock        int                 `json:"clock"`
	Host         string              `json:"host"`
	HostMetadata string              `json:"host_metadata"`
}

func TestZabbix(t *testing.T) {
	tests := map[string]struct {
		Prefix                string
		AgentActive           bool
		SkipMeasurementPrefix bool
		telegrafMetrics       []telegraf.Metric
		zabbixMetrics         []ZabbixRequestData
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
			zabbixMetrics: []ZabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"float numbers values could be received with padding zeros, but same value once converted": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": float64(3.1415),
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []ZabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.value",
					Value: "3.141500",
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
				{
					Host:  "hostname",
					Key:   "name.valueTrue",
					Value: "1",
					Clock: 1522082244,
				},
				{
					Host:  "hostname",
					Key:   "name.valueFalse",
					Value: "0",
					Clock: 1522082244,
				},
			},
		},
		"strage values are ignored by metric.New()": {
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
			zabbixMetrics: []ZabbixRequestData{},
		},
		"metrics without host are ignored": {
			telegrafMetrics: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{},
					map[string]interface{}{
						"value": "x",
					},
					time.Unix(1522082244, 0),
				),
			},
			zabbixMetrics: []ZabbixRequestData{},
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
		"if prefix is configured, zabbix metrics should have that prefix in the key": {
			Prefix: "telegraf.",
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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
			zabbixMetrics: []ZabbixRequestData{
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

			z := &Zabbix{
				Host:                  "127.0.0.1",
				Port:                  10051,
				Prefix:                test.Prefix,
				SkipMeasurementPrefix: test.SkipMeasurementPrefix,
				AgentActive:           test.AgentActive,
			}

			// Simulate a Zabbix server to get the data sent
			listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", z.Host, z.Port))
			require.NoError(t, lerr)
			defer listener.Close()

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				conn, err := listener.Accept()
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
				resp := []byte("failed: 0\n")
				_, err = conn.Write(resp)
				require.NoError(t, err)

				// Close connection after reading the client data
				conn.Close()

				// Strip zabbix header and get JSON request
				var request ZabbixRequest
				err = json.Unmarshal(content, &request)
				require.NoError(t, err)

				// By default we use trappers
				requestType := "sender data"
				if test.AgentActive {
					requestType = "agent data"
				}

				expectedRequest := ZabbixRequest{
					Request: requestType,
					Clock:   request.Clock,
					Data:    test.zabbixMetrics,
				}

				assert.Equal(t, expectedRequest.Request, request.Request)
				assert.Equal(t, expectedRequest.Clock, request.Clock)
				CompareData(t, expectedRequest.Data, request.Data)

				wg.Done()
			}()

			err := z.Write(test.telegrafMetrics)
			require.NoError(t, err)

			// Wait for zabbix server emulator to finish
			wg.Wait()
		})
	}
}

func TestAutoRegister(t *testing.T) {
	z := &Zabbix{
		Host:                   "127.0.0.1",
		Port:                   10051,
		Prefix:                 "telegraf.",
		SkipMeasurementPrefix:  false,
		AgentActive:            false,
		Autoregister:           "xxx",
		AutoregisterSendPeriod: internal.Duration{Duration: time.Minute * 5},
		autoregisterLastSend:   map[string]time.Time{},
	}

	// Simulate a Zabbix server to get the data sent
	listener, lerr := net.Listen("tcp", fmt.Sprintf("%v:%v", z.Host, z.Port))
	require.NoError(t, lerr)
	defer listener.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		/*
		* Accept packet with the two metrics sent
		 */
		conn, err := listener.Accept()
		require.NoError(t, err)

		// The zabbix output checks that there are not errors
		resp := []byte("failed: 0\n")
		_, err = conn.Write(resp)
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

		// Close connection after reading the client data
		conn.Close()

		/*
		* Read the first autoregister packet
		 */
		conn, err = listener.Accept()
		require.NoError(t, err)

		// Obtain request from the mock zabbix server
		// Read protocol header and version
		header = make([]byte, 5)
		_, err = conn.Read(header)
		require.NoError(t, err)

		// Read data length
		dataLengthRaw = make([]byte, 8)
		_, err = conn.Read(dataLengthRaw)
		require.NoError(t, err)

		dataLength = binary.LittleEndian.Uint64(dataLengthRaw)

		// Read data content
		content = make([]byte, dataLength)
		_, err = conn.Read(content)
		require.NoError(t, err)

		// The zabbix output checks that there are not errors
		// Simulated response from the server
		resp = []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\": \"success\", \"info\": \"\"}\n")
		_, err = conn.Write(resp)
		require.NoError(t, err)

		// Close connection after reading the client data
		conn.Close()

		// Strip zabbix header and get JSON request
		var request ZabbixRequest
		err = json.Unmarshal(content, &request)
		require.NoError(t, err)

		assert.Equal(t, "active checks", request.Request)
		assert.Equal(t, "xxx", request.HostMetadata)

		hostsRegistered := []string{request.Host}

		/*
		* Read the second autoregister packet
		 */
		conn, err = listener.Accept()
		require.NoError(t, err)

		// Obtain request from the mock zabbix server
		// Read protocol header and version
		header = make([]byte, 5)
		_, err = conn.Read(header)
		require.NoError(t, err)

		// Read data length
		dataLengthRaw = make([]byte, 8)
		_, err = conn.Read(dataLengthRaw)
		require.NoError(t, err)

		dataLength = binary.LittleEndian.Uint64(dataLengthRaw)

		// Read data content
		content = make([]byte, dataLength)
		_, err = conn.Read(content)
		require.NoError(t, err)

		// The zabbix output checks that there are not errors
		resp = []byte("ZBXD\x01\x00\x00\x00\x00\x00\x00\x00\x00{\"response\": \"success\", \"info\": \"\"}\n")
		_, err = conn.Write(resp)
		require.NoError(t, err)

		// Close connection after reading the client data
		conn.Close()

		// Strip zabbix header and get JSON request
		err = json.Unmarshal(content, &request)
		require.NoError(t, err)

		assert.Equal(t, "active checks", request.Request)
		assert.Equal(t, "xxx", request.HostMetadata)

		// Check we have received autoregistration for both hosts
		hostsRegistered = append(hostsRegistered, request.Host)
		assert.ElementsMatch(t, []string{"hostA", "hostB"}, hostsRegistered)

		conn.Close()
		wg.Done()
	}()

	err := z.Write([]telegraf.Metric{
		testutil.MustMetric("name", map[string]string{"host": "hostA"}, map[string]interface{}{"value": int64(0)}, time.Now()),
		testutil.MustMetric("name", map[string]string{"host": "hostB"}, map[string]interface{}{"value": int64(0)}, time.Now()),
	})
	require.NoError(t, err)

	// Wait for zabbix server emulator to finish
	wg.Wait()
}

// CompareData compares generated data with expected data ignoring slice order if all Clocks are
// the same.
// This is useful for metrics with several fields that should produce several Zabbix values that
// could not be sorted by clock
func CompareData(t *testing.T, expected []ZabbixRequestData, data []ZabbixRequestData) {
	var clock int64
	sameClock := true

	// Check if all clocks are the same
	for i := 0; i < len(data); i++ {
		if i == 0 {
			clock = data[i].Clock
		} else {
			if clock != data[i].Clock {
				sameClock = false
				break
			}
		}
	}

	if sameClock {
		assert.ElementsMatch(t, expected, data)
	} else {
		assert.Equal(t, expected, data)
	}
}
