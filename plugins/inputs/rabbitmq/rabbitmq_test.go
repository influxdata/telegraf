package rabbitmq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
)

func TestRabbitMQGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonFilePath string

		switch r.URL.Path {
		case "/api/overview":
			jsonFilePath = "testdata/overview.json"
		case "/api/nodes":
			jsonFilePath = "testdata/nodes.json"
		case "/api/queues":
			jsonFilePath = "testdata/queues.json"
		case "/api/exchanges":
			jsonFilePath = "testdata/exchanges.json"
		case "/api/healthchecks/node/rabbit@vagrant-ubuntu-trusty-64":
			jsonFilePath = "testdata/healthchecks.json"
		default:
			panic("Cannot handle request")
		}

		data, err := ioutil.ReadFile(jsonFilePath)

		if err != nil {
			panic(fmt.Sprintf("could not read from data file %s", jsonFilePath))
		}

		w.Write(data)
	}))
	defer ts.Close()

	r := &RabbitMQ{
		URL: ts.URL,
	}

	acc := &testutil.Accumulator{}

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)

	overviewMetrics := map[string]interface{}{
		"messages":               5,
		"messages_ready":         32,
		"messages_unacked":       27,
		"messages_acked":         5246,
		"messages_delivered":     5234,
		"messages_delivered_get": 3333,
		"messages_published":     5258,
		"channels":               44,
		"connections":            44,
		"consumers":              65,
		"exchanges":              43,
		"queues":                 62,
		"clustering_listeners":   2,
		"amqp_listeners":         2,
		"return_unroutable":      10,
		"return_unroutable_rate": 3.3,
	}
	compareMetrics(t, overviewMetrics, acc, "rabbitmq_overview")

	queuesMetrics := map[string]interface{}{
		"consumers":                 3,
		"consumer_utilisation":      1.0,
		"memory":                    143776,
		"message_bytes":             3,
		"message_bytes_ready":       4,
		"message_bytes_unacked":     5,
		"message_bytes_ram":         6,
		"message_bytes_persist":     7,
		"messages":                  44,
		"messages_ready":            32,
		"messages_unack":            44,
		"messages_ack":              3457,
		"messages_ack_rate":         9.9,
		"messages_deliver":          22222,
		"messages_deliver_rate":     333.4,
		"messages_deliver_get":      3457,
		"messages_deliver_get_rate": 0.2,
		"messages_publish":          3457,
		"messages_publish_rate":     11.2,
		"messages_redeliver":        33,
		"messages_redeliver_rate":   2.5,
		"idle_since":                "2015-11-01 8:22:14",
	}
	compareMetrics(t, queuesMetrics, acc, "rabbitmq_queue")

	nodeMetrics := map[string]interface{}{
		"disk_free":                 3776,
		"disk_free_limit":           50000000,
		"disk_free_alarm":           0,
		"fd_total":                  1024,
		"fd_used":                   63,
		"mem_limit":                 2503,
		"mem_used":                  159707080,
		"mem_alarm":                 1,
		"proc_total":                1048576,
		"proc_used":                 783,
		"run_queue":                 0,
		"sockets_total":             829,
		"sockets_used":              45,
		"uptime":                    7464827,
		"running":                   1,
		"health_check_status":       1,
		"mnesia_disk_tx_count":      16,
		"mnesia_ram_tx_count":       296,
		"mnesia_disk_tx_count_rate": 1.1,
		"mnesia_ram_tx_count_rate":  2.2,
		"gc_num":                    57280132,
		"gc_bytes_reclaimed":        2533,
		"gc_num_rate":               274.2,
		"gc_bytes_reclaimed_rate":   16490856.3,
		"io_read_avg_time":          983,
		"io_read_avg_time_rate":     88.77,
		"io_read_bytes":             1111,
		"io_read_bytes_rate":        99.99,
		"io_write_avg_time":         134,
		"io_write_avg_time_rate":    4.32,
		"io_write_bytes":            823,
		"io_write_bytes_rate":       32.8,
	}
	compareMetrics(t, nodeMetrics, acc, "rabbitmq_node")

	exchangeMetrics := map[string]interface{}{
		"messages_publish_in":       3678,
		"messages_publish_in_rate":  3.2,
		"messages_publish_out":      3677,
		"messages_publish_out_rate": 5.1,
	}
	compareMetrics(t, exchangeMetrics, acc, "rabbitmq_exchange")
}

func compareMetrics(t *testing.T, expectedMetrics map[string]interface{},
	accumulator *testutil.Accumulator, measurementKey string) {
	measurement, exist := accumulator.Get(measurementKey)

	assert.True(t, exist, "There is measurement %s", measurementKey)
	assert.Equal(t, len(expectedMetrics), len(measurement.Fields))

	for metricName, metricValue := range expectedMetrics {
		actualMetricValue := measurement.Fields[metricName]

		if accumulator.HasStringField(measurementKey, metricName) {
			assert.Equal(t, metricValue, actualMetricValue,
				"Metric name: %s", metricName)
		} else {
			assert.InDelta(t, metricValue, actualMetricValue, 0e5,
				"Metric name: %s", metricName)
		}
	}
}
