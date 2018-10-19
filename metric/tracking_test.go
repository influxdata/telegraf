package metric

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

func mustMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	tm time.Time,
	tp ...telegraf.ValueType,
) telegraf.Metric {
	m, err := New(name, tags, fields, tm, tp...)
	if err != nil {
		panic("mustMetric")
	}
	return m
}

type deliveries struct {
	Info map[telegraf.TrackingID]telegraf.DeliveryInfo
}

func (d *deliveries) onDelivery(info telegraf.DeliveryInfo) {
	d.Info[info.ID()] = info
}

func TestTracking(t *testing.T) {
	tests := []struct {
		name     string
		metric   telegraf.Metric
		actions  func(metric telegraf.Metric)
		accepted int
		rejected int
	}{
		{
			name: "accept",
			metric: mustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
			actions: func(m telegraf.Metric) {
				m.Accept()
			},
			accepted: 1,
			rejected: 0,
		},
		{
			name: "reject",
			metric: mustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
			actions: func(m telegraf.Metric) {
				m.Reject()
			},
			accepted: 0,
			rejected: 1,
		},
		{
			name: "accept copy",
			metric: mustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
			actions: func(m telegraf.Metric) {
				m2 := m.Copy()
				m.Accept()
				m2.Accept()
			},
			accepted: 2,
			rejected: 0,
		},
		{
			name: "copy with accept and done",
			metric: mustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
			actions: func(m telegraf.Metric) {
				m2 := m.Copy()
				m.Accept()
				m2.Remove()
			},
			accepted: 1,
			rejected: 0,
		},
		{
			name: "copy with mixed delivery",
			metric: mustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
			actions: func(m telegraf.Metric) {
				m2 := m.Copy()
				m.Accept()
				m2.Reject()
			},
			accepted: 1,
			rejected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deliveries{
				Info: make(map[telegraf.TrackingID]telegraf.DeliveryInfo),
			}
			metric, id := WithTracking(tt.metric, d.onDelivery)
			tt.actions(metric)

			info := d.Info[id]
			require.Equal(t, tt.accepted, info.Accepted())
			require.Equal(t, tt.rejected, info.Rejected())
		})
	}
}

func TestGroupTracking(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []telegraf.Metric
		actions  func(metrics []telegraf.Metric)
		accepted int
		rejected int
	}{
		{
			name: "accept",
			metrics: []telegraf.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []telegraf.Metric) {
				metrics[0].Accept()
				metrics[1].Accept()
			},
			accepted: 2,
			rejected: 0,
		},
		{
			name: "reject",
			metrics: []telegraf.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []telegraf.Metric) {
				metrics[0].Reject()
				metrics[1].Reject()
			},
			accepted: 0,
			rejected: 2,
		},
		{
			name: "remove",
			metrics: []telegraf.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []telegraf.Metric) {
				metrics[0].Remove()
				metrics[1].Remove()
			},
			accepted: 0,
			rejected: 0,
		},
		{
			name: "mixed",
			metrics: []telegraf.Metric{
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				mustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			actions: func(metrics []telegraf.Metric) {
				metrics[0].Accept()
				metrics[1].Reject()
			},
			accepted: 1,
			rejected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deliveries{
				Info: make(map[telegraf.TrackingID]telegraf.DeliveryInfo),
			}
			metrics, id := WithGroupTracking(tt.metrics, d.onDelivery)
			tt.actions(metrics)

			info := d.Info[id]
			require.Equal(t, tt.accepted, info.Accepted())
			require.Equal(t, tt.rejected, info.Rejected())
		})
	}
}
