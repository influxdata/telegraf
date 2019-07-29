package influx

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type MetricHandler struct {
	builder   *metric.Builder
	err       error
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
	m, err := h.builder.Metric()
	h.builder.Reset()
	return m, err
}

func (h *MetricHandler) SetMeasurement(name []byte) error {
	h.builder.SetName(nameUnescape(name))
	return nil
}

func (h *MetricHandler) AddTag(key []byte, value []byte) error {
	tk := unescape(key)
	tv := unescape(value)
	h.builder.AddTag(tk, tv)
	return nil
}

func (h *MetricHandler) AddInt(key []byte, value []byte) error {
	fk := unescape(key)
	fv, err := parseIntBytes(bytes.TrimSuffix(value, []byte("i")), 10, 64)
	if err != nil {
		if numerr, ok := err.(*strconv.NumError); ok {
			return numerr.Err
		}
		return err
	}
	h.builder.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddUint(key []byte, value []byte) error {
	fk := unescape(key)
	fv, err := parseUintBytes(bytes.TrimSuffix(value, []byte("u")), 10, 64)
	if err != nil {
		if numerr, ok := err.(*strconv.NumError); ok {
			return numerr.Err
		}
		return err
	}
	h.builder.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddFloat(key []byte, value []byte) error {
	fk := unescape(key)
	fv, err := parseFloatBytes(value, 64)
	if err != nil {
		if numerr, ok := err.(*strconv.NumError); ok {
			return numerr.Err
		}
		return err
	}
	h.builder.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddString(key []byte, value []byte) error {
	fk := unescape(key)
	fv := stringFieldUnescape(value)
	h.builder.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) AddBool(key []byte, value []byte) error {
	fk := unescape(key)
	fv, err := parseBoolBytes(value)
	if err != nil {
		return errors.New("unparseable bool")
	}
	h.builder.AddField(fk, fv)
	return nil
}

func (h *MetricHandler) SetTimestamp(tm []byte) error {
	v, err := parseIntBytes(tm, 10, 64)
	if err != nil {
		if numerr, ok := err.(*strconv.NumError); ok {
			return numerr.Err
		}
		return err
	}
	ns := v * int64(h.precision)
	h.builder.SetTime(time.Unix(0, ns))
	return nil
}

func (h *MetricHandler) Reset() {
	h.builder.Reset()
}
