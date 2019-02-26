package cloud_pubsub_push

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/models"
	"github.com/influxdata/telegraf/plugins/parsers"
)

func TestStartStop(t *testing.T) {
	pubPush := PubSubPush{
		Path: "/",
		sem:  make(chan struct{}, 5),
	}
	pubPush.ctx, pubPush.cancel = context.WithCancel(context.Background())

	dst := make(chan telegraf.Metric, 5)
	acc := agent.NewAccumulator(&testMetricMaker{}, dst)

	err := pubPush.Start(acc)
	require.NoError(t, err)

	pubPush.Stop()
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		body     io.Reader
		status   int
		maxsize  int64
		expected string
		fail     bool
		full     bool
		timeout  bool
	}{
		{
			name:   "bad method get",
			method: "GET",
			path:   "/",
			status: http.StatusMethodNotAllowed,
		},
		{
			name:   "post not found",
			method: "POST",
			path:   "/allthings",
			status: http.StatusNotFound,
		},
		{
			name:   "post large date",
			method: "POST",
			path:   "/",
			status: http.StatusRequestEntityTooLarge,
			body:   strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
		},
		{
			name:    "post valid data",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusNoContent,
			body:    strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
		},
		{
			name:    "post valid data, timeout",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusServiceUnavailable,
			body:    strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
			timeout: true,
		},
		{
			name:    "fail write",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusServiceUnavailable,
			body:    strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
			fail:    true,
		},
		{
			name:    "full buffer",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusServiceUnavailable,
			body:    strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"dGVzdGluZ0dvb2dsZSxzZW5zb3I9Ym1lXzI4MCB0ZW1wX2M9MjMuOTUsaHVtaWRpdHk9NjIuODMgMTUzNjk1Mjk3NDU1MzUxMDIzMQ==","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
			full:    true,
		},
		{
			name:    "post invalid body",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusBadRequest,
			body:    strings.NewReader(`invalid body`),
		},
		{
			name:    "post invalid data",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusBadRequest,
			body:    strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"not base 64 encoded data","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
		},
		{
			name:    "post invalid data format",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusBadRequest,
			body:    strings.NewReader(`{"message":{"attributes":{"deviceId":"myPi","deviceNumId":"2808946627307959","deviceRegistryId":"my-registry","deviceRegistryLocation":"us-central1","projectId":"conference-demos","subFolder":""},"data":"bm90IHZhbGlkIGZvcm1hdHRlZCBkYXRh","messageId":"204004313210337","message_id":"204004313210337","publishTime":"2018-09-14T19:22:54.587Z","publish_time":"2018-09-14T19:22:54.587Z"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
		},
		{
			name:    "post invalid structured body",
			method:  "POST",
			path:    "/",
			maxsize: 500 * 1024 * 1024,
			status:  http.StatusBadRequest,
			body:    strings.NewReader(`{"message":{"attributes":{"thing":1},"data":"bm90IHZhbGlkIGZvcm1hdHRlZCBkYXRh"},"subscription":"projects/conference-demos/subscriptions/my-subscription"}`),
		},
	}

	wg := &sync.WaitGroup{}
	for _, test := range tests {
		req, err := http.NewRequest(test.method, test.path, test.body)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		pubPush := PubSubPush{
			Path: "/",
			MaxBodySize: internal.Size{
				Size: test.maxsize,
			},
			sem:         make(chan struct{}, 1),
			undelivered: make(map[telegraf.TrackingID]chan bool),
		}

		pubPush.ctx, pubPush.cancel = context.WithCancel(context.Background())
		pubPush.WriteTimeout = internal.Duration{
			Duration: time.Second * 2,
		}

		if test.full {
			// fill buffer with fake message
			pubPush.sem <- struct{}{}
		}
		if test.timeout {
			pubPush.WriteTimeout = internal.Duration{
				Duration: time.Second * 0,
			}
		}

		p, _ := parsers.NewParser(&parsers.Config{
			MetricName: "cloud_pubsub_push",
			DataFormat: "influx",
		})
		pubPush.SetParser(p)

		dst := make(chan telegraf.Metric)
		ro := models.NewRunningOutput("test", &testOutput{failWrite: test.fail}, &models.OutputConfig{}, 1, 1)
		pubPush.acc = agent.NewAccumulator(&testMetricMaker{}, dst).WithTracking(1)

		wg.Add(1)
		go func() {
			defer wg.Done()
			pubPush.receiveDelivered()
		}()

		wg.Add(1)
		go func(status int, d chan telegraf.Metric) {
			defer wg.Done()
			for m := range d {
				ro.AddMetric(m)
				for ro.Write() != nil {
					select {
					case <-pubPush.ctx.Done():
						return
					default:
						<-time.After(time.Millisecond * 500)
					}
				}
			}
		}(test.status, dst)

		http.TimeoutHandler(http.HandlerFunc(pubPush.ServeHTTP), pubPush.WriteTimeout.Duration, "timed out").ServeHTTP(rr, req)

		require.Equal(t, test.status, rr.Code, test.name)

		if test.expected != "" {
			require.Equal(t, test.expected, rr.Body.String(), test.name)
		}

		pubPush.cancel()
		close(dst)
		wg.Wait()
	}
}

type testMetricMaker struct{}

func (tm *testMetricMaker) Name() string {
	return "TestPlugin"
}

func (tm *testMetricMaker) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

type testOutput struct {
	// if true, mock a write failure
	failWrite bool
}

func (*testOutput) Connect() error {
	return nil
}

func (*testOutput) Close() error {
	return nil
}

func (*testOutput) Description() string {
	return ""
}

func (*testOutput) SampleConfig() string {
	return ""
}

func (t *testOutput) Write(metrics []telegraf.Metric) error {
	if t.failWrite {
		return fmt.Errorf("failed write")
	}
	return nil
}
