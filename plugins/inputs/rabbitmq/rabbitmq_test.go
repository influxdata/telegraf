package rabbitmq

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestRabbitMQGeneratesMetricsSet1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonFilePath string

		switch r.URL.Path {
		case "/api/overview":
			jsonFilePath = "testdata/set1/overview.json"
		case "/api/nodes":
			jsonFilePath = "testdata/set1/nodes.json"
		case "/api/queues":
			jsonFilePath = "testdata/set1/queues.json"
		case "/api/exchanges":
			jsonFilePath = "testdata/set1/exchanges.json"
		case "/api/federation-links":
			jsonFilePath = "testdata/set1/federation-links.json"
		case "/api/nodes/rabbit@vagrant-ubuntu-trusty-64/memory":
			jsonFilePath = "testdata/set1/memory.json"
		default:
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("unknown path %q", r.URL.Path)
			return
		}

		data, err := os.ReadFile(jsonFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Could not read from data file %q: %v", jsonFilePath, err)
			return
		}

		if _, err = w.Write(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))
	defer ts.Close()

	// Define test cases
	expected := []telegraf.Metric{
		testutil.MustMetric("rabbitmq_overview",
			map[string]string{
				"url": ts.URL,
			},
			map[string]interface{}{
				"messages":               int64(50),
				"messages_ready":         int64(72),
				"messages_unacked":       int64(94),
				"messages_acked":         int64(5246),
				"messages_delivered":     int64(5234),
				"messages_delivered_get": int64(3333),
				"messages_published":     int64(5258),
				"channels":               int64(44),
				"connections":            int64(44),
				"consumers":              int64(70),
				"exchanges":              int64(43),
				"queues":                 int64(66),
				"clustering_listeners":   int64(2),
				"amqp_listeners":         int64(2),
				"return_unroutable":      int64(10),
				"return_unroutable_rate": float64(3.3),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "false",
				"node":        "rabbit@rmqlocal-0.rmqlocal.ankorabbitstatefulset3.svc.cluster.local",
				"queue":       "reply_a716f0523cd44941ad2ea6ce4a3869c3",
				"url":         ts.URL,
				"vhost":       "sorandomsorandom",
			},
			map[string]interface{}{
				"consumers":                 int64(3),
				"consumer_utilisation":      float64(1.0),
				"head_message_timestamp":    int64(1446362534),
				"memory":                    int64(143776),
				"message_bytes":             int64(3),
				"message_bytes_ready":       int64(4),
				"message_bytes_unacked":     int64(5),
				"message_bytes_ram":         int64(6),
				"message_bytes_persist":     int64(7),
				"messages":                  int64(44),
				"messages_ready":            int64(32),
				"messages_unack":            int64(44),
				"messages_ack":              int64(3457),
				"messages_ack_rate":         float64(9.9),
				"messages_deliver":          int64(22222),
				"messages_deliver_rate":     float64(333.4),
				"messages_deliver_get":      int64(3457),
				"messages_deliver_get_rate": float64(0.2),
				"messages_publish":          int64(3457),
				"messages_publish_rate":     float64(11.2),
				"messages_redeliver":        int64(33),
				"messages_redeliver_rate":   float64(2.5),
				"idle_since":                "2015-11-01 8:22:14",
				"slave_nodes":               int64(1),
				"synchronised_slave_nodes":  int64(1),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"node":        "rabbit@rmqlocal-0.rmqlocal.ankorabbitstatefulset3.svc.cluster.local",
				"queue":       "stream_queue_example",
				"url":         ts.URL,
				"vhost":       "sorandomsorandom",
			},
			map[string]interface{}{
				"consumers":                 int64(2),
				"consumer_utilisation":      float64(0.8),
				"head_message_timestamp":    int64(1446366000),
				"memory":                    int64(95000),
				"message_bytes":             int64(1),
				"message_bytes_ready":       int64(2),
				"message_bytes_unacked":     int64(3),
				"message_bytes_ram":         int64(4),
				"message_bytes_persist":     int64(5),
				"messages":                  int64(25),
				"messages_ready":            int64(20),
				"messages_unack":            int64(25),
				"messages_ack":              int64(1500),
				"messages_ack_rate":         float64(5.5),
				"messages_deliver":          int64(10000),
				"messages_deliver_rate":     float64(200.0),
				"messages_deliver_get":      int64(1500),
				"messages_deliver_get_rate": float64(0.1),
				"messages_publish":          int64(1500),
				"messages_publish_rate":     float64(8.5),
				"messages_redeliver":        int64(15),
				"messages_redeliver_rate":   float64(1.2),
				"idle_since":                "2015-11-01 9:15:30",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"node":        "rabbit@rmqlocal-0.rmqlocal.ankorabbitstatefulset3.svc.cluster.local",
				"queue":       "no-type-queue",
				"url":         ts.URL,
				"vhost":       "sorandomsorandom",
			},
			map[string]interface{}{
				"consumers":                 int64(1),
				"consumer_utilisation":      float64(0.9),
				"head_message_timestamp":    int64(1446372000),
				"memory":                    int64(50000),
				"message_bytes":             int64(1),
				"message_bytes_ready":       int64(1),
				"message_bytes_unacked":     int64(1),
				"message_bytes_ram":         int64(1),
				"message_bytes_persist":     int64(2),
				"messages":                  int64(10),
				"messages_ready":            int64(8),
				"messages_unack":            int64(10),
				"messages_ack":              int64(500),
				"messages_ack_rate":         float64(2.0),
				"messages_deliver":          int64(2000),
				"messages_deliver_rate":     float64(50.0),
				"messages_deliver_get":      int64(500),
				"messages_deliver_get_rate": float64(0.1),
				"messages_publish":          int64(500),
				"messages_publish_rate":     float64(3.0),
				"messages_redeliver":        int64(5),
				"messages_redeliver_rate":   float64(0.5),
				"idle_since":                "2015-11-01 10:30:00",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"node":        "rabbit@rmqlocal-0.rmqlocal.ankorabbitstatefulset3.svc.cluster.local",
				"queue":       "invalid-type-queue",
				"url":         ts.URL,
				"vhost":       "sorandomsorandom",
			},
			map[string]interface{}{
				"consumers":                 int64(2),
				"consumer_utilisation":      float64(0.85),
				"head_message_timestamp":    int64(1446375000),
				"memory":                    int64(75000),
				"message_bytes":             int64(1),
				"message_bytes_ready":       int64(1),
				"message_bytes_unacked":     int64(2),
				"message_bytes_ram":         int64(2),
				"message_bytes_persist":     int64(3),
				"messages":                  int64(15),
				"messages_ready":            int64(12),
				"messages_unack":            int64(15),
				"messages_ack":              int64(750),
				"messages_ack_rate":         float64(3.0),
				"messages_deliver":          int64(3000),
				"messages_deliver_rate":     float64(75.0),
				"messages_deliver_get":      int64(750),
				"messages_deliver_get_rate": float64(0.1),
				"messages_publish":          int64(750),
				"messages_publish_rate":     float64(4.5),
				"messages_redeliver":        int64(8),
				"messages_redeliver_rate":   float64(0.8),
				"idle_since":                "2015-11-01 11:45:00",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_node",
			map[string]string{
				"node": "rabbit@vagrant-ubuntu-trusty-64",
				"url":  ts.URL,
			},
			map[string]interface{}{
				"disk_free":                 int64(3776),
				"disk_free_limit":           int64(50000000),
				"disk_free_alarm":           int64(0),
				"fd_total":                  int64(1024),
				"fd_used":                   int64(63),
				"mem_limit":                 int64(2503),
				"mem_used":                  int64(159707080),
				"mem_alarm":                 int64(1),
				"proc_total":                int64(1048576),
				"proc_used":                 int64(783),
				"run_queue":                 int64(0),
				"sockets_total":             int64(829),
				"sockets_used":              int64(45),
				"uptime":                    int64(7464827),
				"running":                   int64(1),
				"mnesia_disk_tx_count":      int64(16),
				"mnesia_ram_tx_count":       int64(296),
				"mnesia_disk_tx_count_rate": float64(1.1),
				"mnesia_ram_tx_count_rate":  float64(2.2),
				"gc_num":                    int64(57280132),
				"gc_bytes_reclaimed":        int64(2533),
				"gc_num_rate":               float64(274.2),
				"gc_bytes_reclaimed_rate":   float64(16490856.3),
				"io_read_avg_time":          float64(983.0),
				"io_read_avg_time_rate":     float64(88.77),
				"io_read_bytes":             int64(1111),
				"io_read_bytes_rate":        float64(99.99),
				"io_write_avg_time":         float64(134.0),
				"io_write_avg_time_rate":    float64(4.32),
				"io_write_bytes":            int64(823),
				"io_write_bytes_rate":       float64(32.8),
				"mem_connection_readers":    int64(1234),
				"mem_connection_writers":    int64(5678),
				"mem_connection_channels":   int64(1133),
				"mem_connection_other":      int64(2840),
				"mem_queue_procs":           int64(2840),
				"mem_queue_slave_procs":     int64(0),
				"mem_plugins":               int64(1755976),
				"mem_other_proc":            int64(23056584),
				"mem_metrics":               int64(196536),
				"mem_mgmt_db":               int64(491272),
				"mem_mnesia":                int64(115600),
				"mem_other_ets":             int64(2121872),
				"mem_binary":                int64(418848),
				"mem_msg_index":             int64(42848),
				"mem_code":                  int64(25179322),
				"mem_atom":                  int64(1041593),
				"mem_other_system":          int64(14741981),
				"mem_allocated_unused":      int64(38208528),
				"mem_reserved_unallocated":  int64(0),
				"mem_total":                 int64(83025920),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "true",
				"durable":     "false",
				"exchange":    "reply_a716f0523cd44941ad2ea6ce4a3869c3",
				"internal":    "false",
				"type":        "direct",
				"url":         ts.URL,
				"vhost":       "sorandomsorandom",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(3678),
				"messages_publish_in_rate":  float64(3.2),
				"messages_publish_out":      int64(3677),
				"messages_publish_out_rate": float64(5.1),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_federation",
			map[string]string{
				"queue":          "exampleLocalQueue",
				"type":           "queue",
				"upstream":       "ExampleFederationUpstream",
				"upstream_queue": "exampleUpstreamQueue",
				"url":            ts.URL,
				"vhost":          "/",
			},
			map[string]interface{}{
				"acks_uncommitted":           int64(1),
				"consumers":                  int64(2),
				"messages_unacknowledged":    int64(3),
				"messages_uncommitted":       int64(4),
				"messages_unconfirmed":       int64(5),
				"messages_confirm":           int64(67),
				"messages_publish":           int64(890),
				"messages_return_unroutable": int64(1),
			},
			time.Unix(0, 0),
		),
	}

	// Run the test
	plugin := &RabbitMQ{
		URL: ts.URL,
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	acc.Wait(len(expected))
	require.Empty(t, acc.Errors)

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestRabbitMQGeneratesMetricsSet2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonFilePath string

		switch r.URL.Path {
		case "/api/overview":
			jsonFilePath = "testdata/set2/overview.json"
		case "/api/nodes":
			jsonFilePath = "testdata/set2/nodes.json"
		case "/api/queues":
			jsonFilePath = "testdata/set2/queues.json"
		case "/api/exchanges":
			jsonFilePath = "testdata/set2/exchanges.json"
		case "/api/federation-links":
			jsonFilePath = "testdata/set2/federation-links.json"
		case "/api/nodes/rabbit@rmqserver/memory":
			jsonFilePath = "testdata/set2/memory.json"
		default:
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("unknown path %q", r.URL.Path)
			return
		}

		data, err := os.ReadFile(jsonFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Could not read from data file %q: %v", jsonFilePath, err)
			return
		}

		if _, err = w.Write(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))
	defer ts.Close()

	// Define test cases
	expected := []telegraf.Metric{
		testutil.MustMetric("rabbitmq_overview",
			map[string]string{
				"url": ts.URL,
			},
			map[string]interface{}{
				"messages":               int64(30),
				"messages_ready":         int64(30),
				"messages_unacked":       int64(0),
				"messages_acked":         int64(3736443),
				"messages_delivered":     int64(3736446),
				"messages_delivered_get": int64(3736446),
				"messages_published":     int64(770025),
				"channels":               int64(43),
				"connections":            int64(43),
				"consumers":              int64(40),
				"exchanges":              int64(8),
				"queues":                 int64(37),
				"clustering_listeners":   int64(1),
				"amqp_listeners":         int64(2),
				"return_unroutable":      int64(0),
				"return_unroutable_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "false",
				"node":        "rabbit@rmqserver",
				"queue":       "39fd2caf-63e5-41e3-c15a-ba8fa11434b2",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"consumers":                 int64(1),
				"consumer_utilisation":      float64(1.0),
				"memory":                    int64(15840),
				"message_bytes":             int64(0),
				"message_bytes_ready":       int64(0),
				"message_bytes_unacked":     int64(0),
				"message_bytes_ram":         int64(0),
				"message_bytes_persist":     int64(0),
				"messages":                  int64(0),
				"messages_ready":            int64(0),
				"messages_unack":            int64(0),
				"messages_ack":              int64(180),
				"messages_ack_rate":         float64(0.0),
				"messages_deliver":          int64(180),
				"messages_deliver_rate":     float64(0.0),
				"messages_deliver_get":      int64(180),
				"messages_deliver_get_rate": float64(0.0),
				"messages_publish":          int64(180),
				"messages_publish_rate":     float64(0.0),
				"messages_redeliver":        int64(0),
				"messages_redeliver_rate":   float64(0.0),
				"idle_since":                "2021-06-28 15:54:14",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "false",
				"node":        "rabbit@rmqserver",
				"queue":       "39fd2cb4-aa2d-c08b-457a-62d0893523a1",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"consumers":                 int64(1),
				"consumer_utilisation":      float64(1.0),
				"memory":                    int64(15600),
				"message_bytes":             int64(0),
				"message_bytes_ready":       int64(0),
				"message_bytes_unacked":     int64(0),
				"message_bytes_ram":         int64(0),
				"message_bytes_persist":     int64(0),
				"messages":                  int64(0),
				"messages_ready":            int64(0),
				"messages_unack":            int64(0),
				"messages_ack":              int64(177),
				"messages_ack_rate":         float64(0.0),
				"messages_deliver":          int64(177),
				"messages_deliver_rate":     float64(0.0),
				"messages_deliver_get":      int64(177),
				"messages_deliver_get_rate": float64(0.0),
				"messages_publish":          int64(177),
				"messages_publish_rate":     float64(0.0),
				"messages_redeliver":        int64(0),
				"messages_redeliver_rate":   float64(0.0),
				"idle_since":                "2021-06-28 15:54:14",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "false",
				"node":        "rabbit@rmqserver",
				"queue":       "39fd2cb5-3820-e01b-6e20-ba29d5553fc3",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"consumers":                 int64(1),
				"consumer_utilisation":      float64(1.0),
				"memory":                    int64(15584),
				"message_bytes":             int64(0),
				"message_bytes_ready":       int64(0),
				"message_bytes_unacked":     int64(0),
				"message_bytes_ram":         int64(0),
				"message_bytes_persist":     int64(0),
				"messages":                  int64(0),
				"messages_ready":            int64(0),
				"messages_unack":            int64(0),
				"messages_ack":              int64(175),
				"messages_ack_rate":         float64(0.0),
				"messages_deliver":          int64(175),
				"messages_deliver_rate":     float64(0.0),
				"messages_deliver_get":      int64(175),
				"messages_deliver_get_rate": float64(0.0),
				"messages_publish":          int64(175),
				"messages_publish_rate":     float64(0.0),
				"messages_redeliver":        int64(0),
				"messages_redeliver_rate":   float64(0.0),
				"idle_since":                "2021-06-28 15:54:15",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_node",
			map[string]string{
				"node": "rabbit@rmqserver",
				"url":  ts.URL,
			},
			map[string]interface{}{
				"disk_free":                 int64(25086496768),
				"disk_free_limit":           int64(50000000),
				"disk_free_alarm":           int64(0),
				"fd_total":                  int64(65536),
				"fd_used":                   int64(78),
				"mem_limit":                 int64(1717546188),
				"mem_used":                  int64(387645440),
				"mem_alarm":                 int64(0),
				"proc_total":                int64(1048576),
				"proc_used":                 int64(1128),
				"run_queue":                 int64(1),
				"sockets_total":             int64(58893),
				"sockets_used":              int64(43),
				"uptime":                    int64(4150152129),
				"running":                   int64(1),
				"mnesia_disk_tx_count":      int64(103),
				"mnesia_ram_tx_count":       int64(2257),
				"mnesia_disk_tx_count_rate": float64(0.0),
				"mnesia_ram_tx_count_rate":  float64(0.0),
				"gc_num":                    int64(329526389),
				"gc_bytes_reclaimed":        int64(13660012170840),
				"gc_num_rate":               float64(125.2),
				"gc_bytes_reclaimed_rate":   float64(6583379.2),
				"io_read_avg_time":          float64(0.0),
				"io_read_avg_time_rate":     float64(0.0),
				"io_read_bytes":             int64(1),
				"io_read_bytes_rate":        float64(0.0),
				"io_write_avg_time":         float64(0.0),
				"io_write_avg_time_rate":    float64(0.0),
				"io_write_bytes":            int64(193066),
				"io_write_bytes_rate":       float64(0.0),
				"mem_connection_readers":    int64(1246768),
				"mem_connection_writers":    int64(72108),
				"mem_connection_channels":   int64(308588),
				"mem_connection_other":      int64(4883596),
				"mem_queue_procs":           int64(780996),
				"mem_queue_slave_procs":     int64(0),
				"mem_plugins":               int64(11932828),
				"mem_other_proc":            int64(39203520),
				"mem_metrics":               int64(626932),
				"mem_mgmt_db":               int64(3341264),
				"mem_mnesia":                int64(396016),
				"mem_other_ets":             int64(3771384),
				"mem_binary":                int64(209324208),
				"mem_msg_index":             int64(32648),
				"mem_code":                  int64(32810827),
				"mem_atom":                  int64(1458513),
				"mem_other_system":          int64(14284124),
				"mem_allocated_unused":      int64(61026048),
				"mem_reserved_unallocated":  int64(0),
				"mem_total":                 int64(385548288),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "false",
				"node":        "rabbit@rmqserver",
				"queue":       "no-type-queue-set2",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"consumers":                 int64(1),
				"consumer_utilisation":      float64(1.0),
				"memory":                    int64(12000),
				"message_bytes":             int64(0),
				"message_bytes_ready":       int64(0),
				"message_bytes_unacked":     int64(0),
				"message_bytes_ram":         int64(0),
				"message_bytes_persist":     int64(0),
				"messages":                  int64(0),
				"messages_ready":            int64(0),
				"messages_unack":            int64(0),
				"messages_ack":              int64(100),
				"messages_ack_rate":         float64(0.0),
				"messages_deliver":          int64(100),
				"messages_deliver_rate":     float64(0.0),
				"messages_deliver_get":      int64(100),
				"messages_deliver_get_rate": float64(0.0),
				"messages_publish":          int64(100),
				"messages_publish_rate":     float64(0.0),
				"messages_redeliver":        int64(0),
				"messages_redeliver_rate":   float64(0.0),
				"idle_since":                "2021-06-28 15:54:17",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_queue",
			map[string]string{
				"auto_delete": "false",
				"durable":     "false",
				"node":        "rabbit@rmqserver",
				"queue":       "invalid-type-queue-set2",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"consumers":                 int64(1),
				"consumer_utilisation":      float64(1.0),
				"memory":                    int64(15000),
				"message_bytes":             int64(0),
				"message_bytes_ready":       int64(0),
				"message_bytes_unacked":     int64(0),
				"message_bytes_ram":         int64(0),
				"message_bytes_persist":     int64(0),
				"messages":                  int64(0),
				"messages_ready":            int64(0),
				"messages_unack":            int64(0),
				"messages_ack":              int64(200),
				"messages_ack_rate":         float64(0.0),
				"messages_deliver":          int64(200),
				"messages_deliver_rate":     float64(0.0),
				"messages_deliver_get":      int64(200),
				"messages_deliver_get_rate": float64(0.0),
				"messages_publish":          int64(200),
				"messages_publish_rate":     float64(0.0),
				"messages_redeliver":        int64(0),
				"messages_redeliver_rate":   float64(0.0),
				"idle_since":                "2021-06-28 16:30:45",
				"slave_nodes":               int64(0),
				"synchronised_slave_nodes":  int64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "",
				"internal":    "false",
				"type":        "direct",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(284725),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(284572),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "amq.direct",
				"internal":    "false",
				"type":        "direct",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(0),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(0),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "amq.fanout",
				"internal":    "false",
				"type":        "fanout",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(0),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(0),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "amq.headers",
				"internal":    "false",
				"type":        "headers",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(0),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(0),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "amq.match",
				"internal":    "false",
				"type":        "headers",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(0),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(0),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "amq.rabbitmq.trace",
				"internal":    "true",
				"type":        "topic",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(0),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(0),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "false",
				"durable":     "true",
				"exchange":    "amq.topic",
				"internal":    "false",
				"type":        "topic",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(0),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(0),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("rabbitmq_exchange",
			map[string]string{
				"auto_delete": "true",
				"durable":     "false",
				"exchange":    "Exchange",
				"internal":    "false",
				"type":        "topic",
				"url":         ts.URL,
				"vhost":       "/",
			},
			map[string]interface{}{
				"messages_publish_in":       int64(18006),
				"messages_publish_in_rate":  float64(0.0),
				"messages_publish_out":      int64(60798),
				"messages_publish_out_rate": float64(0.0),
			},
			time.Unix(0, 0),
		),
	}
	expectedErrors := []error{
		errors.New("error response trying to get \"/api/federation-links\": \"Object Not Found\" (reason: \"Not Found\")"),
	}

	// Run the test
	plugin := &RabbitMQ{
		URL: ts.URL,
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	acc.Wait(len(expected))
	require.Len(t, acc.Errors, len(expectedErrors))
	require.ElementsMatch(t, expectedErrors, acc.Errors)

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestRabbitMQMetricFilerts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("unknown path %q", r.URL.Path), http.StatusNotFound)
	}))
	defer ts.Close()

	metricErrors := map[string]error{
		"exchange":   errors.New("getting \"/api/exchanges\" failed: 404 Not Found"),
		"federation": errors.New("getting \"/api/federation-links\" failed: 404 Not Found"),
		"node":       errors.New("getting \"/api/nodes\" failed: 404 Not Found"),
		"overview":   errors.New("getting \"/api/overview\" failed: 404 Not Found"),
		"queue":      errors.New("getting \"/api/queues\" failed: 404 Not Found"),
	}

	// Include test
	for name, expected := range metricErrors {
		plugin := &RabbitMQ{
			URL:           ts.URL,
			Log:           testutil.Logger{},
			MetricInclude: []string{name},
		}
		require.NoError(t, plugin.Init())

		acc := &testutil.Accumulator{}
		require.NoError(t, plugin.Gather(acc))
		require.Len(t, acc.Errors, 1)
		require.ElementsMatch(t, []error{expected}, acc.Errors)
	}

	// Exclude test
	for name := range metricErrors {
		// Exclude the current metric error from the list of expected errors
		var expected []error
		for n, e := range metricErrors {
			if n != name {
				expected = append(expected, e)
			}
		}
		plugin := &RabbitMQ{
			URL:           ts.URL,
			Log:           testutil.Logger{},
			MetricExclude: []string{name},
		}
		require.NoError(t, plugin.Init())

		acc := &testutil.Accumulator{}
		require.NoError(t, plugin.Gather(acc))
		require.Len(t, acc.Errors, len(expected))
		require.ElementsMatch(t, expected, acc.Errors)
	}
}

func TestRabbitMQQueueTypeTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonFilePath string

		switch r.URL.Path {
		case "/api/overview":
			jsonFilePath = "testdata/set1/overview.json"
		case "/api/nodes":
			jsonFilePath = "testdata/set1/nodes.json"
		case "/api/queues":
			jsonFilePath = "testdata/set1/queues.json"
		case "/api/exchanges":
			jsonFilePath = "testdata/set1/exchanges.json"
		case "/api/federation-links":
			jsonFilePath = "testdata/set1/federation-links.json"
		case "/api/nodes/rabbit@vagrant-ubuntu-trusty-64/memory":
			jsonFilePath = "testdata/set1/memory.json"
		default:
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("unknown path %q", r.URL.Path)
			return
		}

		data, err := os.ReadFile(jsonFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Could not read from data file %q: %v", jsonFilePath, err)
			return
		}

		if _, err = w.Write(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))
	defer ts.Close()

	// Test with IncludeQueueTypeTag = false (default)
	plugin := &RabbitMQ{
		URL: ts.URL,
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Gather(acc))

	acc.Wait(4)
	require.Empty(t, acc.Errors)

	// Check that queue metrics don't have type tag when disabled
	queueMetrics := acc.GetTelegrafMetrics()
	for _, metric := range queueMetrics {
		if metric.Name() == "rabbitmq_queue" {
			require.False(t, metric.HasTag("type"), "Queue metric should not have type tag when IncludeQueueTypeTag is false")
		}
	}

	// Test with IncludeQueueTypeTag = true
	pluginWithType := &RabbitMQ{
		URL:                 ts.URL,
		IncludeQueueTypeTag: true,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, pluginWithType.Init())

	accWithType := &testutil.Accumulator{}
	require.NoError(t, pluginWithType.Gather(accWithType))

	accWithType.Wait(4)
	require.Empty(t, accWithType.Errors)

	// Check that queue metrics have type tag when enabled and verify specific queue types
	queueMetricsWithType := accWithType.GetTelegrafMetrics()
	queueTypeMap := make(map[string]string) // queue name -> queue type

	for _, metric := range queueMetricsWithType {
		if metric.Name() == "rabbitmq_queue" {
			require.True(t, metric.HasTag("type"), "Queue metric should have type tag when IncludeQueueTypeTag is true")

			queueName, _ := metric.GetTag("queue")
			queueType, _ := metric.GetTag("type")
			queueTypeMap[queueName] = queueType
		}
	}

	require.Len(t, queueTypeMap, 4, "Should have metrics for all 4 queues in test data")
	require.Equal(t, "quorum", queueTypeMap["reply_a716f0523cd44941ad2ea6ce4a3869c3"], "Quorum queue should have type 'quorum'")
	require.Equal(t, "stream", queueTypeMap["stream_queue_example"], "Stream queue should have type 'stream'")
	require.Equal(t, "classic", queueTypeMap["no-type-queue"], "Queue without type field should default to 'classic'")
	require.Equal(t, "invalid_type", queueTypeMap["invalid-type-queue"], "Queue with invalid type should preserve the invalid type")
}
