package loggregator_rlp_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"crypto/tls"
	"fmt"
	"github.com/influxdata/telegraf/plugins/inputs/loggregator_rlp"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")
)

type rlpTestContext struct {
	Input       *loggregator_rlp.LoggregatorRLPInput
	Accumulator *testutil.Accumulator
	RLP         *MockRLP
	StopRLP     func()

	*GomegaWithT
}

func (tc *rlpTestContext) teardown() {
	tc.Input.Stop()
	tc.StopRLP()
}

func TestParseConfigWithTLS(t *testing.T) {
	tc := buildTestContext(t, nil, []string{})

	rlpInput := tc.Input

	tc.Expect(rlpInput.TLSCA).To(Equal(pki.CACertPath()))
	tc.Expect(rlpInput.TLSCert).To(Equal(pki.ClientCertPath()))
	tc.Expect(rlpInput.TLSKey).To(Equal(pki.ClientKeyPath()))
}

func TestReceivesAllMetricTypesFromRLP(t *testing.T) {
	tc := buildTestContext(t, createHTTPTimer("source-id"), []string{"counter", "gauge", "timer"})
	defer tc.teardown()

	tc.Expect(tc.Input.Start(tc.Accumulator)).To(Succeed())

	tc.Eventually(tc.RLP.ActualReq, "5s").ShouldNot(BeNil())
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ConsistOf(
		&loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Counter{
				Counter: &loggregator_v2.CounterSelector{},
			},
		},
		&loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Gauge{
				Gauge: &loggregator_v2.GaugeSelector{},
			},
		},
		&loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Timer{
				Timer: &loggregator_v2.TimerSelector{},
			},
		},
	))
}

func TestReceivesSubsetOfMetricTypesFromRLP(t *testing.T) {
	tc := buildTestContext(t, createHTTPTimer("source-id"), []string{"counter"})
	defer tc.teardown()

	tc.Expect(tc.Input.Start(tc.Accumulator)).To(Succeed())

	tc.Eventually(tc.RLP.ActualReq, "5s").ShouldNot(BeNil())
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ConsistOf(
		&loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Counter{
				Counter: &loggregator_v2.CounterSelector{},
			},
		},
	))
}

func TestFiltersMetricsBySourceIds(t *testing.T) {
	tc := buildTestContextWithSourceIdFilters(
		t,
		[]*loggregator_v2.Envelope{
			createHTTPTimer("source1"),
			createHTTPTimer("other-source"),
		},
		[]string{"counter", "gauge", "timer"},
		[]string{"source1", "source2"},
	)
	defer tc.teardown()

	tc.Expect(tc.Input.Start(tc.Accumulator)).To(Succeed())

	tc.Eventually(tc.RLP.ActualReq, "5s").ShouldNot(BeNil())
	tc.Expect(tc.RLP.ActualReq().Selectors).To(HaveLen(6))
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ContainElement(&loggregator_v2.Selector{
		SourceId: "source1",
		Message: &loggregator_v2.Selector_Counter{
			Counter: &loggregator_v2.CounterSelector{},
		},
	}))
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ContainElement(&loggregator_v2.Selector{
		SourceId: "source2",
		Message: &loggregator_v2.Selector_Counter{
			Counter: &loggregator_v2.CounterSelector{},
		},
	}))
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ContainElement(&loggregator_v2.Selector{
		SourceId: "source1",
		Message: &loggregator_v2.Selector_Gauge{
			Gauge: &loggregator_v2.GaugeSelector{},
		},
	}))
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ContainElement(&loggregator_v2.Selector{
		SourceId: "source2",
		Message: &loggregator_v2.Selector_Gauge{
			Gauge: &loggregator_v2.GaugeSelector{},
		},
	}))
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ContainElement(&loggregator_v2.Selector{
		SourceId: "source1",
		Message: &loggregator_v2.Selector_Timer{
			Timer: &loggregator_v2.TimerSelector{},
		},
	}))
	tc.Expect(tc.RLP.ActualReq().Selectors).To(ContainElement(&loggregator_v2.Selector{
		SourceId: "source2",
		Message: &loggregator_v2.Selector_Timer{
			Timer: &loggregator_v2.TimerSelector{},
		},
	}))
}

func TestParsesTimers(t *testing.T) {
	tc := buildTestContext(t, createHTTPTimer("source-id"), []string{"timer"})
	defer tc.teardown()

	tc.Expect(tc.Input.Start(tc.Accumulator)).To(Succeed())

	tc.Eventually(func() bool {
		tc.Accumulator.Lock()
		defer tc.Accumulator.Unlock()

		for _, metric := range tc.Accumulator.Metrics {
			if "http" == metric.Measurement {
				return true
			}
		}

		return false
	}, "5s", "1s").Should(
		BeTrue(),
		"Telegraf should have received http metric",
	)
}

func TestParsesCounters(t *testing.T) {
	tc := buildTestContext(t, createCounter(), []string{"counter"})
	defer tc.teardown()

	tc.Expect(tc.Input.Start(tc.Accumulator)).To(Succeed())

	tc.Eventually(func() bool {
		tc.Accumulator.Lock()
		defer tc.Accumulator.Unlock()

		for _, metric := range tc.Accumulator.Metrics {
			if "counter" == metric.Measurement {
				return true
			}
		}

		return false
	}, "5s", "1s").Should(
		BeTrue(),
		"Telegraf should have received http metric",
	)
}

func TestParsesGauges(t *testing.T) {
	tc := buildTestContext(t, createGauge(), []string{"gauge"})
	defer tc.teardown()

	tc.Expect(tc.Input.Start(tc.Accumulator)).To(Succeed())

	tc.Eventually(func() bool {
		tc.Accumulator.Lock()
		defer tc.Accumulator.Unlock()

		for _, metric := range tc.Accumulator.Metrics {
			if "gauge" == metric.Measurement {
				return true
			}
		}

		return false
	}, "5s", "1s").Should(
		BeTrue(),
		"Telegraf should have received http metric",
	)
}

func buildTestContextWithSourceIdFilters(t *testing.T, envelopeResponse []*loggregator_v2.Envelope, envelopeTypes, sourceIdFilters []string, options ...interface{}) *rlpTestContext {
	mockRlp, stopRLP := buildRLPWithTLS(envelopeResponse)

	interval := "30s"
	if len(options) > 0 {
		interval = options[0].(string)
	}

	for i, envelopeType := range envelopeTypes {
		envelopeTypes[i] = fmt.Sprintf("\"%s\"", envelopeType)
	}

	for i, sourceId := range sourceIdFilters {
		sourceIdFilters[i] = fmt.Sprintf("\"%s\"", sourceId)
	}

	envelopeTypesString := strings.Join(envelopeTypes, ",")
	sourceIdFiltersString := strings.Join(sourceIdFilters, ",")
	configWithTLS := []byte(fmt.Sprintf(`
  rlp_address = "%s"
  tls_common_name = "localhost"
  tls_ca = "%s"
  tls_cert = "%s"
  tls_key = "%s"
  internal_metrics_interval = "%s"
  diode_buffer_size = %d
  envelope_types = [%s]
  source_id_filters = [%s]
`, mockRlp.Addr,
		pki.CACertPath(),
		pki.ClientCertPath(),
		pki.ClientKeyPath(),
		interval,
		10000,
		envelopeTypesString,
		sourceIdFiltersString,
	))
	input := loggregator_rlp.NewLoggregatorRLP()
	err := toml.Unmarshal(configWithTLS, input)
	if err != nil {
		panic(err)
	}

	return &rlpTestContext{
		Input:       input,
		Accumulator: new(testutil.Accumulator),
		RLP:         mockRlp,
		StopRLP:     stopRLP,
		GomegaWithT: NewGomegaWithT(t),
	}
}

func buildTestContext(t *testing.T, envelopeResponse *loggregator_v2.Envelope, envelopeTypes []string, options ...interface{}) *rlpTestContext {
	return buildTestContextWithSourceIdFilters(t, []*loggregator_v2.Envelope{envelopeResponse}, envelopeTypes, []string{}, options...)
}

func buildRLPWithTLS(envelopeResponse []*loggregator_v2.Envelope) (*MockRLP, func()) {
	tlsConfig, err := pki.TLSServerConfig().TLSConfig()
	if err != nil {
		panic(err)
	}

	return buildRLP(envelopeResponse, tlsConfig)
}

func createGauge() *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: "source_id",
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					"gauge": {
						Value: 49,
						Unit:  "unit",
					},
				},
			},
		},
	}
}

func createCounter() *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: "source_id",
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  "counter",
				Total: 6,
			},
		},
	}
}

func createHTTPTimer(sourceId string) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: sourceId,
		Message: &loggregator_v2.Envelope_Timer{
			Timer: &loggregator_v2.Timer{
				Name:  "http",
				Start: 1e9,
				Stop:  7e9,
			},
		},
	}
}

func buildRLP(envelopeResponse []*loggregator_v2.Envelope, tlsCfg *tls.Config) (*MockRLP, func()) {
	rlp := NewMockRlp(envelopeResponse, tlsCfg)
	rlp.Start()

	return rlp, rlp.Stop
}
