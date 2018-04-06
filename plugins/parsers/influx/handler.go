package influx

import (
	"bytes"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/prometheus/common/log"
)

type MetricHandler struct {
	builder   *metric.Builder
	metrics   []telegraf.Metric
	precision time.Duration
}

func NewMetricHandler() *MetricHandler {
	return &MetricHandler{
		builder:   metric.NewBuilder(),
		precision: time.Nanosecond,
	}
}

func (h *MetricHandler) SetTimeFunc(f metric.TimeFunc) {
	h.builder.TimeFunc = f
}

func (h *MetricHandler) SetTimePrecision(precision time.Duration) {
	h.builder.TimePrecision = precision
	h.precision = precision
}

func (h *MetricHandler) Metric() (telegraf.Metric, error) {
	return h.builder.Metric()
}

func (h *MetricHandler) SetMeasurement(name []byte) {
	h.builder.SetName(nameUnescape(name))
}

func (h *MetricHandler) AddTag(key []byte, value []byte) {
	tk := unescape(key)
	tv := unescape(value)
	h.builder.AddTag(tk, tv)
}

func (h *MetricHandler) AddInt(key []byte, value []byte) {
	fk := unescape(key)
	fv, err := parseIntBytes(bytes.TrimSuffix(value, []byte("i")), 10, 64)
	if err != nil {
		log.Errorf("E! Received unparseable int value: %q: %v", value, err)
		return
	}
	h.builder.AddField(fk, fv)
}

func (h *MetricHandler) AddUint(key []byte, value []byte) {
	fk := unescape(key)
	fv, err := parseUintBytes(bytes.TrimSuffix(value, []byte("u")), 10, 64)
	if err != nil {
		log.Errorf("E! Received unparseable uint value: %q: %v", value, err)
		return
	}
	h.builder.AddField(fk, fv)
}

func (h *MetricHandler) AddFloat(key []byte, value []byte) {
	fk := unescape(key)
	fv, err := parseFloatBytes(value, 64)
	if err != nil {
		log.Errorf("E! Received unparseable float value: %q: %v", value, err)
		return
	}
	h.builder.AddField(fk, fv)
}

func (h *MetricHandler) AddString(key []byte, value []byte) {
	fk := unescape(key)
	fv := stringFieldUnescape(value)
	h.builder.AddField(fk, fv)
}

func (h *MetricHandler) AddBool(key []byte, value []byte) {
	fk := unescape(key)
	fv, err := parseBoolBytes(value)
	if err != nil {
		log.Errorf("E! Received unparseable boolean value: %q: %v", value, err)
		return
	}
	h.builder.AddField(fk, fv)
}

func (h *MetricHandler) SetTimestamp(tm []byte) {
	v, err := parseIntBytes(tm, 10, 64)
	if err != nil {
		log.Errorf("E! Received unparseable timestamp: %q: %v", tm, err)
		return
	}
	ns := v * int64(h.precision)
	h.builder.SetTime(time.Unix(0, ns))
}

func (h *MetricHandler) Reset() {
	h.builder.Reset()
}
