package loggregator_forwarder_agent_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf/plugins/inputs/loggregator_forwarder_agent"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"testing"
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")
)

func TestNoTLS(t *testing.T) {
	tc := buildTestContextWithoutTLS(t)
	stop := start(tc)
	defer stop()

	envBatch := createCounterBatch("the-source", 42)

	waitForMetricsAndStopServer(tc, []string{"some-counter"}, 1, func() error {
		_, err := tc.Client.Send(context.Background(), envBatch)
		return err
	})

	tc.Expect(getMetricByMeasurement(tc.Accumulator, "some-counter")).ToNot(BeNil())
}

func TestInternalMetrics(t *testing.T) {
	tc := buildTestContextWithoutTLS(t, "1s")
	stop := start(tc)
	defer stop()

	tc.Eventually(tc.Accumulator.NMetrics, "5s", "1s").Should(BeNumerically(">=", 8),
		fmt.Sprintf("Telegraf should have recorded %d metrics. Instead it had %d\n", 8, tc.Accumulator.NMetrics()))
	tc.Input.Stop()

	tc.Expect(getMetricByMeasurement(tc.Accumulator, "telegraf_dropped_envelopes")).ToNot(BeNil())
	tc.Expect(getMetricByMeasurement(tc.Accumulator, "telegraf_ingress_envelopes")).ToNot(BeNil())
	tc.Expect(getMetricByMeasurement(tc.Accumulator, "telegraf_egress_envelopes")).ToNot(BeNil())
	tc.Expect(getMetricByMeasurement(tc.Accumulator, "telegraf_dropped_delta_counters")).ToNot(BeNil())
}

func TestParseConfigWithoutTLS(t *testing.T) {
	tc := buildTestContextWithoutTLS(t)
	loggregatorInput := tc.Input
	tc.Expect(loggregatorInput.TLSAllowedCACerts).To(BeEmpty())
	tc.Expect(loggregatorInput.TLSCert).To(Equal(""))
	tc.Expect(loggregatorInput.TLSKey).To(Equal(""))
}

func TestParseConfigWithTLS(t *testing.T) {
	tc := buildTestContext(t)
	loggregatorInput := tc.Input

	tc.Expect(loggregatorInput.TLSAllowedCACerts).To(ConsistOf(pki.CACertPath()))
	tc.Expect(loggregatorInput.TLSCert).To(Equal(pki.ServerCertPath()))
	tc.Expect(loggregatorInput.TLSKey).To(Equal(pki.ServerKeyPath()))
}

func TestSendOverTLS(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	envBatch := createCounterBatch("the-source", 42)
	waitForMetricsAndStopServer(tc, []string{"some-counter"}, 1, func() error {
		_, err := tc.Client.Send(context.Background(), envBatch)
		return err
	})
	tc.Expect(getMetricByMeasurement(tc.Accumulator, "some-counter")).ToNot(BeNil())
}

func TestSenderOverTLS(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	env := createCounter("the-source", 42, make(map[string]string))
	senderClient := safelyGetSenderClient(tc)
	waitForMetricsAndStopServer(tc, []string{"some-counter"}, 1, func() error {
		if senderClient == nil {
			return errors.New("sender Client is nil")
		}
		err := senderClient.Send(env)
		return err
	})

	tc.Expect(getMetricByMeasurement(tc.Accumulator, "some-counter")).ToNot(BeNil())
}

func TestBatchSenderOverTLS(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	envBatch := createCounterBatch("the-source", 42)
	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(tc, []string{"some-counter"}, 1, func() error {
		if senderClient == nil {
			return errors.New("sender Client is nil")
		}
		err := senderClient.Send(envBatch)
		return err
	})

	tc.Expect(getMetricByMeasurement(tc.Accumulator, "some-counter")).ToNot(BeNil())
}

func TestParsesCounterTotal(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()
	envBatch := createCounterBatch("the-source", 42)

	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(tc, []string{"some-counter"}, 1, func() error {
		if senderClient == nil {
			return errors.New("sender Client is nil")
		}
		err := senderClient.Send(envBatch)
		return err
	})

	expectedMetric := getMetricByMeasurement(tc.Accumulator, "some-counter")
	tc.Expect(expectedMetric.Measurement).To(Equal("some-counter"))
	tc.Expect(expectedMetric.Fields["counter"]).To(Equal(uint64(42)))
}

func TestParsesCounterTags(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()
	envBatch := createCounterBatchWithTags("the-source", 42, map[string]string{
		"foo":         "bar",
		"instance_id": "dead-beef-dead-beef",
	})

	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(tc, []string{"some-counter"}, 1, func() error {
		if senderClient == nil {
			return errors.New("sender Client is nil")
		}
		err := senderClient.Send(envBatch)
		return err
	})

	expectedMetric := getMetricByMeasurement(tc.Accumulator, "some-counter")
	tc.Expect(expectedMetric.Tags).To(HaveKeyWithValue("foo", "bar"))
	tc.Expect(expectedMetric.Tags).To(HaveKeyWithValue("instance_id", "dead-beef-dead-beef"))
	tc.Expect(expectedMetric.Tags).To(HaveKeyWithValue("source_id", "the-source"))
}

func TestDropCountersWithDeltaMissingTotal(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	badEnv := createCounterWithDelta("the-source", "delta-counter", 0, 5)
	goodEnv := createCounterWithDelta("the-source", "total-counter", 27, 0)
	batch := loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{badEnv, goodEnv},
	}

	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(
		tc,
		[]string{"total-counter", "telegraf_dropped_delta_counters"},
		2,
		func() error {
			if senderClient == nil {
				return errors.New("sender Client is nil")
			}
			err := senderClient.Send(&batch)
			return err
		},
	)

	tc.Expect(getMetricByMeasurement(tc.Accumulator, "delta-counter")).To(BeNil())

	expectedMetric := getMetricByMeasurement(tc.Accumulator, "total-counter")
	tc.Expect(expectedMetric.Fields["counter"]).To(Equal(uint64(27)))
}

func TestIgnoresDelta(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	env := createCounterWithDelta("the-source", "total-counter", 27, 6)
	batch := loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{env},
	}

	senderClient := safelyGetBatchSenderClient(tc)

	waitForMetricsAndStopServer(tc, []string{"total-counter"}, 1, func() error {
		if senderClient == nil {
			return errors.New("sender Client is nil")
		}
		err := senderClient.Send(&batch)
		return err
	})

	expectedMetric := getMetricByMeasurement(tc.Accumulator, "total-counter")
	tc.Expect(expectedMetric.Fields["counter"]).To(Equal(uint64(27)))
}

func TestParsingSourceID(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	sourceIdType1 := createCounter("the-source", 42, nil)
	sourceIdType2 := createCounter("", 42, map[string]string{"origin": "origin-source"})
	sourceIdType3 := &loggregator_v2.Envelope{
		SourceId: "",
		DeprecatedTags: map[string]*loggregator_v2.Value{"origin": {
			Data: &loggregator_v2.Value_Text{"origin-source-v1"},
		}},
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  "total-counter",
				Delta: 6,
				Total: 27,
			},
		},
	}
	batch := loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{sourceIdType1, sourceIdType2, sourceIdType3},
	}
	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(
		tc,
		[]string{"some-counter", "total-counter"},
		3,
		func() error {
			err := senderClient.Send(&batch)
			return err
		},
	)

	var expectedMetrics []*testutil.Metric
	tc.Accumulator.Lock()
	defer tc.Accumulator.Unlock()
	for _, metric := range tc.Accumulator.Metrics {
		if metric.Measurement == "some-counter" || metric.Measurement == "total-counter" {
			expectedMetrics = append(expectedMetrics, metric)
		}
	}

	tc.Expect(expectedMetrics).To(HaveLen(3))
	tc.Expect(expectedMetrics[0].Tags["source_id"]).To(Equal("the-source"))
	tc.Expect(expectedMetrics[1].Tags["source_id"]).To(Equal("origin-source"))
	tc.Expect(expectedMetrics[2].Tags["source_id"]).To(Equal("origin-source-v1"))
}

func TestParsesEachGaugeMetric(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	metrics := map[string]*loggregator_v2.GaugeValue{
		"pressure":    gaugeValue(42.1, "psi"),
		"temperature": gaugeValue(27.8, "degrees fahrenheit"),
	}

	env := gaugeEnvelope("the-source", metrics)
	batch := loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{env},
	}

	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(
		tc,
		[]string{"pressure", "temperature"},
		2,
		func() error {
			if senderClient == nil {
				return errors.New("sender Client is nil")
			}
			err := senderClient.Send(&batch)
			return err
		},
	)

	pressure, temperature := findPressureAndTemperature(tc.Accumulator)

	tc.Expect(pressure.Fields["gauge"]).To(Equal(42.1))
	tc.Expect(temperature.Fields["gauge"]).To(Equal(27.8))
}

func TestParsesGaugeTags(t *testing.T) {
	tc := buildTestContext(t)
	stop := start(tc)
	defer stop()

	tags := map[string]string{
		"foo":         "bar",
		"instance_id": "dead-beef-dead-beef",
	}

	metrics := map[string]*loggregator_v2.GaugeValue{
		"pressure":    gaugeValue(42.1, "psi"),
		"temperature": gaugeValue(27.8, "degrees fahrenheit"),
	}

	env := gaugeEnvelopeWithTags("the-source", metrics, tags)
	batch := loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{env},
	}

	senderClient := safelyGetBatchSenderClient(tc)
	waitForMetricsAndStopServer(
		tc,
		[]string{"pressure", "temperature"},
		2,
		func() error {
			if senderClient == nil {
				return errors.New("sender Client is nil")
			}
			err := senderClient.Send(&batch)
			return err
		},
	)

	pressure, temperature := findPressureAndTemperature(tc.Accumulator)

	tc.Expect(pressure.Tags).To(HaveKeyWithValue("foo", "bar"))
	tc.Expect(pressure.Tags).To(HaveKeyWithValue("instance_id", "dead-beef-dead-beef"))

	tc.Expect(temperature.Tags).To(HaveKeyWithValue("foo", "bar"))
	tc.Expect(temperature.Tags).To(HaveKeyWithValue("instance_id", "dead-beef-dead-beef"))
}

func safelyGetBatchSenderClient(tc *agentTestContext) loggregator_v2.Ingress_BatchSenderClient {
	var senderClient loggregator_v2.Ingress_BatchSenderClient
	tc.Eventually(func() error {
		var err error
		senderClient, err = tc.Client.BatchSender(context.Background(), grpc.EmptyCallOption{})
		return err
	}, "5s", "1s").Should(Succeed())

	return senderClient
}

func safelyGetSenderClient(tc *agentTestContext) loggregator_v2.Ingress_SenderClient {
	var senderClient loggregator_v2.Ingress_SenderClient
	tc.Eventually(func() error {
		var err error
		senderClient, err = tc.Client.Sender(context.Background(), grpc.EmptyCallOption{})
		return err
	}, "5s", "1s").Should(Succeed())

	return senderClient
}

func waitForMetricsAndStopServer(
	tc *agentTestContext,
	expectedMetricNames []string,
	expectedMetrics int,
	sendFunction func() error,
) {
	tc.Eventually(sendFunction, "5s", "1s").Should(Succeed())
	tc.Eventually(func() int {
		metricCount := 0
		tc.Accumulator.Lock()
		defer tc.Accumulator.Unlock()

		for _, metric := range tc.Accumulator.Metrics {
			for _, expectedMetric := range expectedMetricNames {
				if expectedMetric == metric.Measurement {
					metricCount++
				}
			}
		}

		return metricCount
	}, "5s", "1s").Should(
		BeNumerically("==", expectedMetrics),
		fmt.Sprintf("Telegraf should have recorded %d metrics", expectedMetrics),
	)

	tc.Input.Stop()
}

type agentTestContext struct {
	Input       *loggregator_forwarder_agent.LoggregatorForwarderAgentInput
	Accumulator *testutil.Accumulator
	Client      loggregator_v2.IngressClient
	StopClient  func()

	*GomegaWithT
}

func buildTestContext(t *testing.T, options ...interface{}) *agentTestContext {
	port := getFreePort()
	interval := "30s"
	if len(options) > 0 {
		interval = options[0].(string)
	}
	configWithTLS := []byte(fmt.Sprintf(`
  port = %d
  tls_allowed_cacerts = [ "%s" ]
  tls_cert = "%s"
  tls_key = "%s"
  internal_metrics_interval = "%s"
`, port, pki.CACertPath(), pki.ServerCertPath(), pki.ServerKeyPath(), interval))
	input := loggregator_forwarder_agent.NewLoggregator()
	err := toml.Unmarshal(configWithTLS, input)
	if err != nil {
		panic(err)
	}

	ingressClient, stopClient := buildClientWithTLS(input)

	return &agentTestContext{
		Input:       input,
		Accumulator: new(testutil.Accumulator),
		Client:      ingressClient,
		StopClient:  stopClient,

		GomegaWithT: NewGomegaWithT(t),
	}
}

func buildTestContextWithoutTLS(t *testing.T, options ...interface{}) *agentTestContext {
	port := getFreePort()
	interval := "30s"
	if len(options) > 0 {
		interval = options[0].(string)
	}
	configWithoutTLS := []byte(fmt.Sprintf(`
  port = %d
  internal_metrics_interval = "%s"	
`, port, interval))
	input := loggregator_forwarder_agent.NewLoggregator()
	err := toml.Unmarshal(configWithoutTLS, input)
	if err != nil {
		panic(err)
	}

	ingressClient, stopClient := buildClientWithoutTLS(input)

	return &agentTestContext{
		Input:       input,
		Accumulator: new(testutil.Accumulator),
		Client:      ingressClient,
		StopClient:  stopClient,

		GomegaWithT: NewGomegaWithT(t),
	}
}

func start(tc *agentTestContext) func() {
	if err := tc.Input.Start(tc.Accumulator); err != nil {
		panic(err)
	}

	return tc.StopClient
}

func buildClientWithoutTLS(input *loggregator_forwarder_agent.LoggregatorForwarderAgentInput) (loggregator_v2.IngressClient, func()) {
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", input.Port), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return buildClient(conn)
}

func buildClientWithTLS(input *loggregator_forwarder_agent.LoggregatorForwarderAgentInput) (loggregator_v2.IngressClient, func()) {
	tlsConfig, err := pki.TLSClientConfig().TLSConfig()
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", input.Port), grpc.WithTransportCredentials(
		credentials.NewTLS(tlsConfig),
	))
	if err != nil {
		panic(err)
	}

	return buildClient(conn)
}

func buildClient(conn *grpc.ClientConn) (loggregator_v2.IngressClient, func()) {
	return loggregator_v2.NewIngressClient(conn), func() {
		err := conn.Close()
		if err != nil {
			log.Fatal()
		}
	}
}

func createCounterBatch(sourceID string, total int) *loggregator_v2.EnvelopeBatch {
	env := createCounter(sourceID, total, make(map[string]string))
	return &loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{env},
	}
}

func createCounterBatchWithTags(sourceID string, total int, tags map[string]string) *loggregator_v2.EnvelopeBatch {
	env := createCounter(sourceID, total, tags)
	return &loggregator_v2.EnvelopeBatch{
		Batch: []*loggregator_v2.Envelope{env},
	}
}

func createCounterWithDelta(sourceID string, name string, total int, delta int) *loggregator_v2.Envelope {
	env := &loggregator_v2.Envelope{
		SourceId: sourceID,
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  name,
				Delta: uint64(delta),
				Total: uint64(total),
			},
		},
	}
	return env
}

func createCounterWithName(name string, sourceID string, total int, tags map[string]string) *loggregator_v2.Envelope {
	env := &loggregator_v2.Envelope{
		SourceId: sourceID,
		Tags:     tags,
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  name,
				Total: uint64(total),
			},
		},
	}
	return env
}

func createCounter(sourceID string, total int, tags map[string]string) *loggregator_v2.Envelope {
	return createCounterWithName("some-counter", sourceID, total, tags)
}

func gaugeEnvelope(sourceId string, metrics map[string]*loggregator_v2.GaugeValue) *loggregator_v2.Envelope {
	return gaugeEnvelopeWithTags(sourceId, metrics, make(map[string]string))
}

func gaugeEnvelopeWithTags(sourceId string, metrics map[string]*loggregator_v2.GaugeValue, tags map[string]string) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: sourceId,
		Tags:     tags,
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: metrics,
			},
		},
	}
}

func gaugeValue(value float64, unit string) *loggregator_v2.GaugeValue {
	return &loggregator_v2.GaugeValue{
		Value: value,
		Unit:  unit,
	}
}

func getFreePort() int {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func findPressureAndTemperature(acc *testutil.Accumulator) (*testutil.Metric, *testutil.Metric) {
	return getMetricByMeasurement(acc, "pressure"), getMetricByMeasurement(acc, "temperature")
}

func getMetricByMeasurement(acc *testutil.Accumulator, measurement string) *testutil.Metric {
	acc.Lock()
	defer acc.Unlock()
	for _, metric := range acc.Metrics {
		if metric.Measurement == measurement {
			return metric
		}
	}

	return nil
}
