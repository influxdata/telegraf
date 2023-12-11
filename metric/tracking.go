package metric

import (
	"runtime"
	"sync/atomic"

	"github.com/influxdata/telegraf"
)

// NotifyFunc is called when a tracking metric is done being processed with
// the tracking information.
type NotifyFunc = func(track telegraf.DeliveryInfo)

// WithTracking adds tracking to the metric and registers the notify function
// to be called when processing is complete.
func WithTracking(metric telegraf.Metric, fn NotifyFunc) (telegraf.Metric, telegraf.TrackingID) {
	return newTrackingMetric(metric, fn)
}

// WithGroupTracking adds tracking to the metrics and registers the notify
// function to be called when processing is complete.
func WithGroupTracking(metric []telegraf.Metric, fn NotifyFunc) ([]telegraf.Metric, telegraf.TrackingID) {
	return newTrackingMetricGroup(metric, fn)
}

var (
	lastID    uint64
	finalizer func(*trackingData)
)

func newTrackingID() telegraf.TrackingID {
	return telegraf.TrackingID(atomic.AddUint64(&lastID, 1))
}

type trackingData struct {
	id          telegraf.TrackingID
	rc          int32
	acceptCount int32
	rejectCount int32
	notifyFunc  NotifyFunc
}

func (d *trackingData) incr() {
	atomic.AddInt32(&d.rc, 1)
}

func (d *trackingData) decr() int32 {
	return atomic.AddInt32(&d.rc, -1)
}

func (d *trackingData) accept() {
	atomic.AddInt32(&d.acceptCount, 1)
}

func (d *trackingData) reject() {
	atomic.AddInt32(&d.rejectCount, 1)
}

func (d *trackingData) notify() {
	d.notifyFunc(
		&deliveryInfo{
			id:       d.id,
			accepted: int(d.acceptCount),
			rejected: int(d.rejectCount),
		},
	)
}

type trackingMetric struct {
	telegraf.Metric
	d *trackingData
}

func newTrackingMetric(metric telegraf.Metric, fn NotifyFunc) (telegraf.Metric, telegraf.TrackingID) {
	m := &trackingMetric{
		Metric: metric,
		d: &trackingData{
			id:          newTrackingID(),
			rc:          1,
			acceptCount: 0,
			rejectCount: 0,
			notifyFunc:  fn,
		},
	}

	if finalizer != nil {
		runtime.SetFinalizer(m.d, finalizer)
	}
	return m, m.d.id
}

func newTrackingMetricGroup(group []telegraf.Metric, fn NotifyFunc) ([]telegraf.Metric, telegraf.TrackingID) {
	d := &trackingData{
		id:          newTrackingID(),
		rc:          0,
		acceptCount: 0,
		rejectCount: 0,
		notifyFunc:  fn,
	}

	for i, m := range group {
		d.incr()
		dm := &trackingMetric{
			Metric: m,
			d:      d,
		}
		group[i] = dm
	}
	if finalizer != nil {
		runtime.SetFinalizer(d, finalizer)
	}

	if len(group) == 0 {
		d.notify()
	}

	return group, d.id
}

func (m *trackingMetric) Copy() telegraf.Metric {
	m.d.incr()
	return &trackingMetric{
		Metric: m.Metric.Copy(),
		d:      m.d,
	}
}

func (m *trackingMetric) Accept() {
	m.d.accept()
	m.decr()
}

func (m *trackingMetric) Reject() {
	m.d.reject()
	m.decr()
}

func (m *trackingMetric) Drop() {
	m.decr()
}

func (m *trackingMetric) decr() {
	v := m.d.decr()
	if v < 0 {
		panic("negative refcount")
	}

	if v == 0 {
		m.d.notify()
	}
}

// Unwrap allows to access the underlying metric directly e.g. for go-templates
func (m *trackingMetric) TrackingID() telegraf.TrackingID {
	return m.d.id
}

// Unwrap allows to access the underlying metric directly e.g. for go-templates
func (m *trackingMetric) Unwrap() telegraf.Metric {
	return m.Metric
}

type deliveryInfo struct {
	id       telegraf.TrackingID
	accepted int
	rejected int
}

func (r *deliveryInfo) ID() telegraf.TrackingID {
	return r.id
}

func (r *deliveryInfo) Delivered() bool {
	return r.rejected == 0
}
