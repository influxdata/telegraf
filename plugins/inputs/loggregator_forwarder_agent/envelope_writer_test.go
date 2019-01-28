package loggregator_forwarder_agent_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/influxdata/telegraf/plugins/inputs/loggregator_forwarder_agent"
	"github.com/influxdata/telegraf/testutil"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

type envelopeWriterTestContext struct {
	accumulator *testutil.Accumulator
	writer      *loggregator_forwarder_agent.EnvelopeWriter

	*GomegaWithT
}

func setup(t *testing.T) *envelopeWriterTestContext {
	accumulator := &testutil.Accumulator{}
	writer := loggregator_forwarder_agent.NewEnvelopeWriter(accumulator, time.Hour)
	g := NewGomegaWithT(t)
	clearInternalMetrics(g, accumulator)

	return &envelopeWriterTestContext{
		accumulator: accumulator,
		writer:      writer,
		GomegaWithT: g,
	}
}

func (tc *envelopeWriterTestContext) nextMetric() *testutil.Metric {
	tc.Eventually(tc.accumulator.NMetrics).Should(BeEquivalentTo(1))

	tc.accumulator.Lock()
	metric := tc.accumulator.Metrics[0]
	tc.accumulator.Unlock()
	tc.accumulator.ClearMetrics()

	return metric
}

func TestRemovesTagsWithLeadingDoubleUnderscore(t *testing.T) {
	tc := setup(t)

	tc.writer.Write(&loggregator_v2.Envelope{
		Tags: map[string]string{
			"__two_underscores":   "underscores",
			"_one_underscore":     "underscore",
			"middle__underscores": "middle",
			"noUnderscores":       "none",
		},
		DeprecatedTags: map[string]*loggregator_v2.Value{
			"__two_underscores_deprecated": {
				Data: &loggregator_v2.Value_Text{
					Text: "underscores",
				},
			},
			"_one_underscore_deprecated": {
				Data: &loggregator_v2.Value_Text{
					Text: "underscore",
				},
			},
			"middle__underscores_deprecated": {
				Data: &loggregator_v2.Value_Text{
					Text: "middle",
				},
			},
			"noUnderscoresDeprecated": {
				Data: &loggregator_v2.Value_Text{
					Text: "none",
				},
			},
		},
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name: "counter",
			},
		},
	})
	metric := tc.nextMetric()

	tc.Expect(metric.Tags).ToNot(And(
		HaveKey("__two_underscores"),
		HaveKey("__two_underscores_deprecated"),
	))
	tc.Expect(metric.Tags).To(And(
		HaveKeyWithValue("_one_underscore", "underscore"),
		HaveKeyWithValue("middle__underscores", "middle"),
		HaveKeyWithValue("noUnderscores", "none"),
		HaveKeyWithValue("_one_underscore_deprecated", "underscore"),
		HaveKeyWithValue("middle__underscores_deprecated", "middle"),
		HaveKeyWithValue("noUnderscoresDeprecated", "none"),
	))
}

func TestTimersTransformsToGauges_CalculatesDifferenceBetweenStartAndStop(t *testing.T) {
	tc := setup(t)

	now := time.Now()
	tc.writer.Write(buildHttpTimerEnvelope(now))
	metric := tc.nextMetric()

	tc.Expect(metric.Measurement).To(Equal("http"))
	tc.Expect(metric.Fields).To(HaveKeyWithValue("gauge", BeEquivalentTo(7)))
	tc.Expect(metric.Tags).To(Equal(map[string]string{
		"source_id":   "source",
		"instance_id": "instance",
		"deployment":  "deployment_value",
		"job":         "job_value",
		"status_code": "201",
	}))
	tc.Expect(metric.Time).To(BeTemporally("==", now))
}

func TestTimersTransformsToGauges_DropsExtraTags(t *testing.T) {
	tc := setup(t)

	env := buildHttpTimerEnvelope(time.Now())
	env.Tags = map[string]string{
		"source_id":   "source",
		"instance_id": "instance",
		"extra":       "stuff",
		"status_code": "400",
	}
	env.DeprecatedTags["extra-deprecated"] = makeTextValue("stuff")

	tc.writer.Write(env)
	metric := tc.nextMetric()

	tc.Expect(metric.Tags).To(Equal(map[string]string{
		"source_id":   "source",
		"instance_id": "instance",
		"deployment":  "deployment_value",
		"job":         "job_value",
		"status_code": "400",
	}))
}

func TestTimersTransformsToGauges_PreservesSubSecondAccuracy(t *testing.T) {
	tc := setup(t)

	env := buildHttpTimerEnvelope(time.Now())
	env.GetTimer().Stop = int64(float64(7.5) * float64(time.Second))

	tc.writer.Write(env)
	metric := tc.nextMetric()

	tc.Expect(metric.Fields).To(HaveKeyWithValue("gauge", 6.5))
}

func TestTimersTransformsToGauges_DefaultsToSourceIdFromOriginTag(t *testing.T) {
	tc := setup(t)
	env := buildHttpTimerEnvelope(time.Now())
	env.SourceId = ""
	env.Tags["origin"] = "origin_value"
	env.DeprecatedTags["origin"] = makeTextValue("deprecated_origin_value")
	tc.writer.Write(env)
	tc.Expect(tc.nextMetric().Tags).To(HaveKeyWithValue("source_id", "origin_value"))

	env2 := buildHttpTimerEnvelope(time.Now())
	env2.SourceId = ""
	env2.DeprecatedTags["origin"] = makeTextValue("deprecated_origin_value")
	tc.writer.Write(env2)
	tc.Expect(tc.nextMetric().Tags).To(HaveKeyWithValue("source_id", "deprecated_origin_value"))
}

func TestIgnoreMetricsWithInvalidNames(t *testing.T) {
	tc := setup(t)
	tc.writer.Write(buildCounterEnvelope("7_invalid"))
	tc.writer.Write(buildCounterEnvelope("*_invalid"))
	tc.writer.Write(buildCounterEnvelope("invali&d"))

	tc.Consistently(tc.accumulator.NMetrics).Should(BeZero())
}

func TestIgnoreLabelsWithInvalidNames(t *testing.T) {
	tc := setup(t)
	tc.writer.Write(buildCounterEnvelopeWithTags("invalid_labels", map[string]string{
		"7_invalid": "value",
		"*invalid":  "value",
		"invali&d":  "value",
		"invalid:":  "value",
	}))

	metric := tc.nextMetric()
	tc.Expect(metric.Tags).ToNot(HaveKey("7_invalid"))
	tc.Expect(metric.Tags).ToNot(HaveKey("*invalid"))
	tc.Expect(metric.Tags).ToNot(HaveKey("invali&d"))
	tc.Expect(metric.Tags).ToNot(HaveKey("invalid:"))
}

func clearInternalMetrics(g *GomegaWithT, accumulator *testutil.Accumulator) {
	g.Eventually(accumulator.NMetrics).Should(BeEquivalentTo(4))
	accumulator.ClearMetrics()
}

func buildHttpTimerEnvelope(t time.Time) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  t.UnixNano(),
		Message: &loggregator_v2.Envelope_Timer{
			Timer: &loggregator_v2.Timer{
				Name:  "http",
				Start: 1 * int64(time.Second),
				Stop:  8 * int64(time.Second),
			},
		},
		Tags: map[string]string{},
		DeprecatedTags: map[string]*loggregator_v2.Value{
			"status_code": makeTextValue("201"),
			"job":         makeTextValue("job_value"),
			"deployment":  makeTextValue("deployment_value"),
		},
	}
}

func buildCounterEnvelope(name string) *loggregator_v2.Envelope {
	return buildCounterEnvelopeWithTags(name, map[string]string{})
}

func buildCounterEnvelopeWithTags(name string, tags map[string]string) *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  time.Now().UnixNano(),
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  name,
				Total: 7,
			},
		},
		Tags: tags,
		DeprecatedTags: map[string]*loggregator_v2.Value{
			"status_code": makeTextValue("201"),
			"job":         makeTextValue("job_value"),
			"deployment":  makeTextValue("deployment_value"),
		},
	}
}

func makeTextValue(value string) *loggregator_v2.Value {
	return &loggregator_v2.Value{
		Data: &loggregator_v2.Value_Text{
			Text: value,
		},
	}
}
