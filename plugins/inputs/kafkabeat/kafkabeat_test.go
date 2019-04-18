package kafkabeat

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var kafkabeatTest = NewKafkabeat()

var (
	kafkabeatStatsAccumulator testutil.Accumulator
)

func Test_KafkabeatStats(test *testing.T) {
	//kafkabeatStatsAccumulator.SetDebug(true)
	fakeServer := httptest.NewUnstartedServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				if request.URL.String() == suffixInfo {
					fmt.Fprintf(writer, "%s", string(kafkabeatInfo))
				} else if request.URL.String() == suffixStats {
					fmt.Fprintf(writer, "%s", string(kafkabeatStats))
				} else {
					test.Logf("Unknown URL: " + request.URL.String())
				}
			},
		),
	)
	requestURL, err := url.Parse(kafkabeatTest.URL)
	if err != nil {
		test.Logf("Can't connect to: %s", kafkabeatTest.URL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if kafkabeatTest.client == nil {
		client, err := kafkabeatTest.createHttpClient()

		if err != nil {
			test.Logf("Can't createHttpClient")
		}
		kafkabeatTest.client = client
	}

	err = kafkabeatTest.gatherStats(&kafkabeatStatsAccumulator)
	if err != nil {
		test.Logf("Can't gather stats")
	}

	kafkabeatStatsAccumulator.AssertContainsTaggedFields(
		test,
		"kafkabeat_beat",
		map[string]interface{}{
			"cpu_system_ticks":      float64(1393890),
			"cpu_system_time_ms":    float64(1393890),
			"cpu_total_ticks":       float64(52339260),
			"cpu_total_time_ms":     float64(52339264),
			"cpu_total_value":       float64(52339260),
			"cpu_user_ticks":        float64(50945370),
			"cpu_user_time_ms":      float64(50945374),
			"info_uptime_ms":        float64(65057537),
			"memstats_gc_next":      float64(559016128),
			"memstats_memory_alloc": float64(280509808),
			"memstats_memory_total": float64(4596157344344),
			"memstats_rss":          float64(368422912),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.6.2"),
		},
	)

	kafkabeatStatsAccumulator.AssertContainsTaggedFields(
		test,
		"kafkabeat_libbeat",
		map[string]interface{}{
			"config_module_running":     float64(0),
			"config_module_starts":      float64(0),
			"config_module_stops":       float64(0),
			"config_reloads":            float64(0),
			"output_events_acked":       float64(186307311),
			"output_events_active":      float64(0),
			"output_events_batches":     float64(1753223),
			"output_events_dropped":     float64(0),
			"output_events_duplicates":  float64(0),
			"output_events_failed":      float64(0),
			"output_events_total":       float64(186307311),
			"output_read_bytes":         float64(1248297178),
			"output_read_errors":        float64(0),
			"output_write_bytes":        float64(60016355484),
			"output_write_errors":       float64(0),
			"pipeline_clients":          float64(1),
			"pipeline_events_active":    float64(0),
			"pipeline_events_dropped":   float64(0),
			"pipeline_events_failed":    float64(0),
			"pipeline_events_filtered":  float64(0),
			"pipeline_events_published": float64(186307311),
			"pipeline_events_retry":     float64(106),
			"pipeline_events_total":     float64(186307311),
			"pipeline_queue_acked":      float64(186307311),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.6.2"),
		},
	)

	kafkabeatStatsAccumulator.AssertContainsTaggedFields(
		test,
		"kafkabeat_system",
		map[string]interface{}{
			"cpu_cores":    float64(32),
			"load_1":       float64(10.76),
			"load_15":      float64(7.19),
			"load_5":       float64(10.7),
			"load_norm_1":  float64(0.3363),
			"load_norm_15": float64(0.2247),
			"load_norm_5":  float64(0.3344),
		},
		map[string]string{
			"beat_host":    string("node-6"),
			"beat_id":      string("9c1c8697-acb4-4df0-987d-28197814f785"),
			"beat_name":    string("node-6-test"),
			"beat_version": string("6.6.2"),
		},
	)

}

func Test_KafkabeatRequest(test *testing.T) {
	//kafkabeatStatsAccumulator.SetDebug(true)
	fakeServer := httptest.NewUnstartedServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				if request.URL.String() == suffixInfo {
					fmt.Fprintf(writer, "%s", string(kafkabeatInfo))
				} else if request.URL.String() == suffixStats {
					fmt.Fprintf(writer, "%s", string(kafkabeatStats))
				} else {
					test.Logf("Unknown URL: " + request.URL.String())
				}

				assert.Equal(test, request.Host, "kafkabeat.test.local")
				assert.Equal(test, request.Method, "POST")
				assert.Equal(test, request.Header.Get("Authorization"), "Basic YWRtaW46UFdE")
				assert.Equal(test, request.Header.Get("X-Test"), "test-value")
			},
		),
	)
	requestURL, err := url.Parse(kafkabeatTest.URL)
	if err != nil {
		test.Logf("Can't connect to: %s", kafkabeatTest.URL)
	}
	fakeServer.Listener, _ = net.Listen("tcp", fmt.Sprintf("%s:%s", requestURL.Hostname(), requestURL.Port()))
	fakeServer.Start()
	defer fakeServer.Close()

	if kafkabeatTest.client == nil {
		client, err := kafkabeatTest.createHttpClient()

		if err != nil {
			test.Logf("Can't createHttpClient")
		}
		kafkabeatTest.client = client
	}

	kafkabeatTest.Headers["X-Test"] = "test-value"
	kafkabeatTest.HostHeader = "kafkabeat.test.local"
	kafkabeatTest.Method = "POST"
	kafkabeatTest.Username = "admin"
	kafkabeatTest.Password = "PWD"

	err = kafkabeatTest.gatherStats(&kafkabeatStatsAccumulator)
	if err != nil {
		test.Logf("Can't gather stats")
	}

}
