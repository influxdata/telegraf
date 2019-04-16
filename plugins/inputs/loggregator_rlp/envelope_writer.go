package loggregator_rlp

import (
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gendiodes "code.cloudfoundry.org/go-diodes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/loggregator-agent/pkg/diodes"
	"github.com/influxdata/telegraf"
)

var (
	validMetricNameExpression = regexp.MustCompile(`^[a-zA-Z_:\-.][a-zA-Z0-9_:\-.]*$`)
	validLabelExpression      = regexp.MustCompile(`^[a-zA-Z_\-.][a-zA-Z0-9_\-.]*$`)
)

const (
	droppedMetricName              = "telegraf_dropped_envelopes"
	ingressMetricName              = "telegraf_ingress_envelopes"
	egressMetricName               = "telegraf_egress_envelopes"
	droppedDeltaCountersMetricName = "telegraf_dropped_delta_counters"
)

type EnvelopeWriter struct {
	acc            telegraf.Accumulator
	stop           chan struct{}
	wg             sync.WaitGroup
	envelopesQueue *diodes.OneToOneEnvelopeV2

	ingressedEnvelopes   *uint64
	egressedEnvelopes    *uint64
	droppedEnvelopes     *uint64
	droppedDeltaCounters *uint64
}

func NewEnvelopeWriter(acc telegraf.Accumulator, internalMetricsInterval time.Duration) *EnvelopeWriter {
	w := &EnvelopeWriter{
		acc:  acc,
		stop: make(chan struct{}),

		ingressedEnvelopes:   new(uint64),
		egressedEnvelopes:    new(uint64),
		droppedEnvelopes:     new(uint64),
		droppedDeltaCounters: new(uint64),
	}

	w.envelopesQueue = diodes.NewOneToOneEnvelopeV2(10000, gendiodes.AlertFunc(func(missed int) {
		w.incrementMetric(w.droppedEnvelopes, uint64(missed))
	}))

	go w.handleEnvelopes()
	go w.periodicallyReportInternalMetrics(internalMetricsInterval)

	return w
}

func (w *EnvelopeWriter) periodicallyReportInternalMetrics(interval time.Duration) {
	w.reportInternalMetrics()

	intervalTick := time.Tick(interval)
	for {
		select {
		case <-w.stop:
			return
		case <-intervalTick:
			w.wg.Add(1)
			w.reportInternalMetrics()
			w.wg.Done()
		}
	}
}

func (w *EnvelopeWriter) Write(env *loggregator_v2.Envelope) {
	w.envelopesQueue.Set(env)
	w.incrementMetric(w.ingressedEnvelopes, 1)
}

func (w *EnvelopeWriter) handleEnvelopes() {
	for {
		select {
		case <-w.stop:
			return
		default:
			e, ok := w.envelopesQueue.TryNext()
			if !ok {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			w.wg.Add(1)
			w.handleEnvelope(e)
			w.wg.Done()
		}
	}
}

func (w *EnvelopeWriter) handleEnvelope(env *loggregator_v2.Envelope) {
	switch env.GetMessage().(type) {
	case *loggregator_v2.Envelope_Counter:
		w.writeCounter(env)
	case *loggregator_v2.Envelope_Gauge:
		w.writeGauge(env)
	case *loggregator_v2.Envelope_Timer:
		w.writeTimer(env)
	default:
		return
	}

	w.incrementMetric(w.egressedEnvelopes, 1)
}

func (w *EnvelopeWriter) writeCounter(env *loggregator_v2.Envelope) {
	ts := time.Unix(0, env.GetTimestamp())
	counter := env.GetCounter()
	if !validName(counter.GetName()) {
		return
	}

	if counter.GetTotal() == 0 && counter.GetDelta() > 0 {
		w.incrementMetric(w.droppedDeltaCounters, 1)
		return
	}

	fields := map[string]interface{}{
		"counter": counter.GetTotal(),
	}

	w.acc.AddCounter(counter.GetName(), fields, buildTags(env), ts)
}

func (w *EnvelopeWriter) writeGauge(env *loggregator_v2.Envelope) {
	ts := time.Unix(0, env.GetTimestamp())
	for name, metric := range env.GetGauge().GetMetrics() {
		if !validName(name) {
			continue
		}

		fields := map[string]interface{}{
			"gauge": metric.GetValue(),
		}

		w.acc.AddGauge(name, fields, buildTags(env), ts)
	}
}

func (w *EnvelopeWriter) writeTimer(env *loggregator_v2.Envelope) {
	timer := env.GetTimer()
	if !validName(timer.GetName()) {
		return
	}

	difference := float64(timer.Stop-timer.Start) / float64(time.Second)
	w.acc.AddGauge(
		timer.GetName(),
		map[string]interface{}{
			"gauge": difference,
		},
		buildTimerTags(env),
		time.Now(),
	)
}

func validName(name string) bool {
	return validMetricNameExpression.MatchString(name)
}

func buildTags(env *loggregator_v2.Envelope) map[string]string {
	tags := map[string]string{
		"source_id":   sourceID(env),
		"instance_id": env.GetInstanceId(),
	}

	for tagName, tagValue := range env.GetDeprecatedTags() {
		if isIgnoredTag(tagName) {
			continue
		}

		tags[tagName] = tagValue.GetText()
	}

	for tagName, tagValue := range env.GetTags() {
		if isIgnoredTag(tagName) {
			continue
		}

		tags[tagName] = tagValue
	}

	return tags
}

func buildTimerTags(env *loggregator_v2.Envelope) map[string]string {
	return map[string]string{
		"source_id":   sourceID(env),
		"instance_id": env.GetInstanceId(),
		"job":         getTag(env, "job"),
		"deployment":  getTag(env, "deployment"),
		"status_code": getTag(env, "status_code"),
	}
}

func getTag(env *loggregator_v2.Envelope, tagName string) string {
	if tag, ok := env.GetTags()[tagName]; ok {
		return tag
	}

	return env.GetDeprecatedTags()[tagName].GetText()
}

func sourceID(env *loggregator_v2.Envelope) string {
	if env.GetSourceId() != "" {
		return env.GetSourceId()
	}

	return getTag(env, "origin")
}

func isIgnoredTag(tagName string) bool {
	return strings.HasPrefix(tagName, "__") || !validLabelExpression.MatchString(tagName)
}

func (w *EnvelopeWriter) Stop() {
	close(w.stop)
	w.wg.Wait()
}

func (w *EnvelopeWriter) reportInternalMetrics() {
	w.reportCounter(egressMetricName, w.egressedEnvelopes)
	w.reportCounter(ingressMetricName, w.ingressedEnvelopes)
	w.reportCounter(droppedMetricName, w.droppedEnvelopes)
	w.reportCounter(droppedDeltaCountersMetricName, w.droppedDeltaCounters)
}

func (w *EnvelopeWriter) reportCounter(name string, valueAddr *uint64) {
	value := atomic.LoadUint64(valueAddr)
	fields := map[string]interface{}{
		"counter": value,
	}
	w.acc.AddCounter(name, fields, nil)
}

func (w *EnvelopeWriter) incrementMetric(metric *uint64, delta uint64) {
	atomic.AddUint64(metric, delta)
}
