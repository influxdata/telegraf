package rabbitmq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"io/ioutil"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		case "/api/federation-links":
			jsonFilePath = "testdata/federation-links.json"
		case "/api/nodes/rabbit@vagrant-ubuntu-trusty-64/memory":
			jsonFilePath = "testdata/memory.json"
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
		"slave_nodes":               1,
		"synchronised_slave_nodes":  1,
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
		"mem_connection_readers":    1234,
		"mem_connection_writers":    5678,
		"mem_connection_channels":   1133,
		"mem_connection_other":      2840,
		"mem_queue_procs":           2840,
		"mem_queue_slave_procs":     0,
		"mem_plugins":               1755976,
		"mem_other_proc":            23056584,
		"mem_metrics":               196536,
		"mem_mgmt_db":               491272,
		"mem_mnesia":                115600,
		"mem_other_ets":             2121872,
		"mem_binary":                418848,
		"mem_msg_index":             42848,
		"mem_code":                  25179322,
		"mem_atom":                  1041593,
		"mem_other_system":          14741981,
		"mem_allocated_unused":      38208528,
		"mem_reserved_unallocated":  0,
		"mem_total":                 83025920,
	}
	compareMetrics(t, nodeMetrics, acc, "rabbitmq_node")

	exchangeMetrics := map[string]interface{}{
		"messages_publish_in":       3678,
		"messages_publish_in_rate":  3.2,
		"messages_publish_out":      3677,
		"messages_publish_out_rate": 5.1,
	}
	compareMetrics(t, exchangeMetrics, acc, "rabbitmq_exchange")

	federationLinkMetrics := map[string]interface{}{
		"acks_uncommitted":           1,
		"consumers":                  2,
		"messages_unacknowledged":    3,
		"messages_uncommitted":       4,
		"messages_unconfirmed":       5,
		"messages_confirm":           67,
		"messages_publish":           890,
		"messages_return_unroutable": 1,
	}
	compareMetrics(t, federationLinkMetrics, acc, "rabbitmq_federation")
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
