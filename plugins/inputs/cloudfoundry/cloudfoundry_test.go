package cloudfoundry

import (
	"sync"
	"testing"
	"time"

	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/cloudfoundry/fakes"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var (
	validClientCredential = ClientConfig{
		GatewayAddress: "https://gateway.addr",
		APIAddress:     "https://api.addr",
		ClientID:       "client_id",
		ClientSecret:   "client_secret",
	}
)

var (
	// soon is the timeout used for testing async actions that occur quickly
	soon = time.Millisecond * 250
	// tick is the period of checking async assertions
	tick = time.Millisecond * 100
)

func TestValidateMissingGatewayOnInit(t *testing.T) {
	rec, _ := withFakeClient(&Cloudfoundry{})

	err := rec.Init()
	require.EqualError(t, err, "must provide a valid gateway_address")
}

func TestValidateMissingAPIOnInit(t *testing.T) {
	rec, _ := withFakeClient(&Cloudfoundry{
		ClientConfig: ClientConfig{
			GatewayAddress: "https://gateway.addr",
		},
	})

	err := rec.Init()
	require.EqualError(t, err, "must provide a valid api_address")
}

func TestValidateMissingCredentialsOnInit(t *testing.T) {
	rec, _ := withFakeClient(&Cloudfoundry{
		ClientConfig: ClientConfig{
			GatewayAddress: "https://gateway.addr",
			APIAddress:     "https://api.addr",
		},
	})

	err := rec.Init()
	require.EqualError(t, err, "must provide either username/password or client_id/client_secret authentication")
}

func TestValidateMetricTypeOnInit(t *testing.T) {
	rec, _ := withFakeClient(&Cloudfoundry{
		ClientConfig: validClientCredential,
		Types:        []string{"junk"},
	})

	err := rec.Init()
	require.EqualError(t, err, "invalid metric type 'junk' must be one of [counter timer gauge event log]")
}

func TestValidConfig(t *testing.T) {
	rec, _ := withFakeClient(&Cloudfoundry{
		ClientConfig: validClientCredential,
		SourceID:     "",
	})

	err := rec.Init()
	require.NoError(t, err)
}

func TestStartStopTerminates(t *testing.T) {
	rec, _ := withFakeClient(&Cloudfoundry{
		ClientConfig: validClientCredential,
	})

	err := rec.Start(&testutil.Accumulator{})
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		rec.Stop()
		wg.Done()
	}()
	wg.Wait() // should terminate
}

func TestRequestedMetricTypes(t *testing.T) {
	typesAndSelectors := []struct {
		RequestedType     string
		ExpectedSelectors []*loggregator_v2.Selector
	}{
		{
			RequestedType: Log,
			ExpectedSelectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Log{
						Log: &loggregator_v2.LogSelector{},
					},
				},
			},
		},
		{
			RequestedType: Counter,
			ExpectedSelectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{},
					},
				},
			},
		},
		{
			RequestedType: Gauge,
			ExpectedSelectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Gauge{
						Gauge: &loggregator_v2.GaugeSelector{},
					},
				},
			},
		},
		{
			RequestedType: Timer,
			ExpectedSelectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Timer{
						Timer: &loggregator_v2.TimerSelector{},
					},
				},
			},
		},
		{
			RequestedType: Event,
			ExpectedSelectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Event{
						Event: &loggregator_v2.EventSelector{},
					},
				},
			},
		},
	}
	require.Lenf(t,
		typesAndSelectors,
		len(validMetricTypes),
		"should test for all supported metric types",
	)

	for _, tc := range typesAndSelectors {
		t.Run(tc.RequestedType, func(t *testing.T) {
			rec, fakeClient := withFakeClient(&Cloudfoundry{
				ClientConfig: validClientCredential,
				Types:        []string{tc.RequestedType},
			})

			err := rec.Start(&testutil.Accumulator{})
			require.NoError(t, err)
			defer rec.Stop()

			require.Eventually(t, func() bool {
				return fakeClient.StreamCallCount() > 0
			}, soon, tick, "expected stream to be called")

			ctx, req := fakeClient.StreamArgsForCall(0)
			require.NotNil(t, ctx)
			require.NotNil(t, req)

			require.ElementsMatchf(t,
				req.Selectors,
				tc.ExpectedSelectors,
				"expected request for stream with '%s' enabled to contain correct selector",
				tc.RequestedType,
			)
		})
	}
}

func TestReadEnvelopeStream(t *testing.T) {
	rec, fakeClient := withFakeClient(&Cloudfoundry{
		ClientConfig: validClientCredential,
	})
	acc := &testutil.Accumulator{}
	ts := time.Now()

	// setup fake stream
	envs := make(chan []*loggregator_v2.Envelope, 2)
	fakeClient.StreamReturnsOnCall(0, func() []*loggregator_v2.Envelope {
		select {
		case <-rec.ctx.Done():
			return nil
		case env := <-envs:
			return env
		}
	})

	// setup a dummy Event metric
	msg := &loggregator_v2.Envelope_Event{
		Event: &loggregator_v2.Event{
			Title: "event",
			Body:  "ok",
		},
	}

	// populate stream with first batch
	envs <- []*loggregator_v2.Envelope{
		{
			SourceId:   "source-1",
			InstanceId: "instance-1",
			Timestamp:  ts.UnixNano(),
			Message:    msg,
			Tags: map[string]string{
				"app_name": "app-1",
			},
		},
		{
			SourceId:   "source-2",
			InstanceId: "instance-2",
			Timestamp:  ts.UnixNano(),
			Message:    msg,
			Tags: map[string]string{
				"app_name": "app-2",
			},
		},
	}

	// populate stream with a second batch
	envs <- []*loggregator_v2.Envelope{
		{
			SourceId:   "source-3",
			InstanceId: "instance-3",
			Timestamp:  ts.UnixNano(),
			Message:    msg,
			Tags: map[string]string{
				"app_name": "app-3",
			},
		},
	}

	// connect
	err := rec.Start(acc)
	require.NoError(t, err)
	defer rec.Stop()

	// wait for all 3 metrics to arrive
	acc.Wait(3)

	want := []telegraf.Metric{
		testutil.MustMetric(
			"cloudfoundry",
			map[string]string{
				"source_id":   "source-1",
				"instance_id": "instance-1",
				"app_name":    "app-1",
			},
			map[string]interface{}{
				"title": "event",
				"body":  "ok",
			},
			ts.UTC(),
		),
		testutil.MustMetric(
			"cloudfoundry",
			map[string]string{
				"source_id":   "source-2",
				"instance_id": "instance-2",
				"app_name":    "app-2",
			},
			map[string]interface{}{
				"title": "event",
				"body":  "ok",
			},
			ts.UTC(),
		),
		testutil.MustMetric(
			"cloudfoundry",
			map[string]string{
				"source_id":   "source-3",
				"instance_id": "instance-3",
				"app_name":    "app-3",
			},
			map[string]interface{}{
				"title": "event",
				"body":  "ok",
			},
			ts.UTC(),
		),
	}
	testutil.RequireMetricsEqual(t, want, acc.GetTelegrafMetrics())
}

func TestReconnectEnvelopeStream(t *testing.T) {
	rec, fakeClient := withFakeClient(&Cloudfoundry{
		ClientConfig: validClientCredential,
	})
	acc := &testutil.Accumulator{}
	ts := time.Now()

	// setup fake streams
	streamOne := make(chan []*loggregator_v2.Envelope, 1)
	fakeClient.StreamReturnsOnCall(0, func() []*loggregator_v2.Envelope {
		select {
		case <-rec.ctx.Done():
			return nil
		case env := <-streamOne:
			return env
		}
	})
	streamTwo := make(chan []*loggregator_v2.Envelope, 1)
	fakeClient.StreamReturnsOnCall(1, func() []*loggregator_v2.Envelope {
		select {
		case <-rec.ctx.Done():
			return nil
		case env := <-streamTwo:
			return env
		}
	})

	// setup a dummy Event metric
	msg := &loggregator_v2.Envelope_Event{
		Event: &loggregator_v2.Event{
			Title: "event",
			Body:  "ok",
		},
	}

	// populate streamOne with an event batch
	streamOne <- []*loggregator_v2.Envelope{
		{
			SourceId:   "source-1",
			InstanceId: "instance-1",
			Timestamp:  ts.UnixNano(),
			Message:    msg,
			Tags: map[string]string{
				"app_name": "app-1",
			},
		},
	}

	// populate streamTwo with an event batch
	streamTwo <- []*loggregator_v2.Envelope{
		{
			SourceId:   "source-2",
			InstanceId: "instance-2",
			Timestamp:  ts.UnixNano(),
			Message:    msg,
			Tags: map[string]string{
				"app_name": "app-2",
			},
		},
	}

	// connect
	err := rec.Start(acc)
	require.NoError(t, err)
	defer rec.Stop()

	// wait for 1 metric to arrive
	acc.Wait(1)

	// stream should have been "connected" once so far
	require.Equal(t, 1, fakeClient.StreamCallCount())

	// expect 1st event from initial connection
	wantFromStreamOne := []telegraf.Metric{
		testutil.MustMetric(
			"cloudfoundry",
			map[string]string{
				"source_id":   "source-1",
				"instance_id": "instance-1",
				"app_name":    "app-1",
			},
			map[string]interface{}{
				"title": "event",
				"body":  "ok",
			},
			ts.UTC(),
		),
	}
	testutil.RequireMetricsEqual(t, wantFromStreamOne, acc.GetTelegrafMetrics())

	// close stream one forcing a reconnect
	acc.ClearMetrics()
	close(streamOne)

	// wait for 1 more metric to arrive
	acc.Wait(1)

	// stream should have been "connected" twice so far
	require.Equal(t, 2, fakeClient.StreamCallCount())

	// expect 2nd event after reconnection
	wantFromStreamTwo := []telegraf.Metric{
		testutil.MustMetric(
			"cloudfoundry",
			map[string]string{
				"source_id":   "source-2",
				"instance_id": "instance-2",
				"app_name":    "app-2",
			},
			map[string]interface{}{
				"title": "event",
				"body":  "ok",
			},
			ts.UTC(),
		),
	}
	testutil.RequireMetricsEqual(t, wantFromStreamTwo, acc.GetTelegrafMetrics())
}

// newTestCloudfoundryPlugin returns a Cloudfoundry input plugin
// with a mock cloudfoundry client for testing
func withFakeClient(c *Cloudfoundry) (*Cloudfoundry, *fakes.FakeCloudfoundryClient) {
	fakeClient := &fakes.FakeCloudfoundryClient{}
	fakeClient.StreamReturns(func() []*loggregator_v2.Envelope {
		return nil
	})
	c.NewClient = func(cfg ClientConfig, errs chan error) CloudfoundryClient {
		return fakeClient
	}
	c.Log = testutil.Logger{}
	c.RetryInterval.Duration = time.Millisecond * 10
	return c, fakeClient
}
