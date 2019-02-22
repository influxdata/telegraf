package loggregator_rlp_test

import (
	"crypto/tls"
	"fmt"
	"testing"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/influxdata/telegraf/plugins/inputs/loggregator_rlp"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	. "github.com/onsi/gomega"
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
	tc := buildTestContext(t, nil)

	rlpInput := tc.Input

	tc.Expect(rlpInput.TLSCA).To(Equal(pki.CACertPath()))
	tc.Expect(rlpInput.TLSCert).To(Equal(pki.ClientCertPath()))
	tc.Expect(rlpInput.TLSKey).To(Equal(pki.ClientKeyPath()))
}

func TestReceivesSelectedMetricsFromRLP(t *testing.T) {
	tc := buildTestContext(t, createHTTPTimer())
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

func TestParsesTimers(t *testing.T) {
	tc := buildTestContext(t, createHTTPTimer())
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
	tc := buildTestContext(t, createCounter())
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
	tc := buildTestContext(t, createGauge())
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

func buildTestContext(t *testing.T, envelopeResponse *loggregator_v2.Envelope, options ...interface{}) *rlpTestContext {
	mockRlp, stopRLP := buildRLPWithTLS(envelopeResponse)

	interval := "30s"
	if len(options) > 0 {
		interval = options[0].(string)
	}
	configWithTLS := []byte(fmt.Sprintf(`
  rlp_address = "%s"
  tls_common_name = "localhost"
  tls_ca = "%s"
  tls_cert = "%s"
  tls_key = "%s"
  internal_metrics_interval = "%s"
`,
		mockRlp.Addr,
		pki.CACertPath(),
		pki.ClientCertPath(),
		pki.ClientKeyPath(),
		interval,
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

func buildRLPWithTLS(envelopeResponse *loggregator_v2.Envelope) (*MockRLP, func()) {
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

func createHTTPTimer() *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId: "source_id",
		Message: &loggregator_v2.Envelope_Timer{
			Timer: &loggregator_v2.Timer{
				Name:  "http",
				Start: 1e9,
				Stop:  7e9,
			},
		},
	}
}

func buildRLP(envelopeResponse *loggregator_v2.Envelope, tlsCfg *tls.Config) (*MockRLP, func()) {
	rlp := NewMockRlp(envelopeResponse, tlsCfg)
	rlp.Start()

	return rlp, rlp.Stop
}
