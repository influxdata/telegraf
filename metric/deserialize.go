package metric

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
)

// storage for tracking data that can't be serialized to disk
var (
	// grouped tracking metrics means that ID->Data association is not one to one,
	// many metrics could be associated with one tracking ID so we cannot just
	// clear this every time in FromBytes.
	trackingStore = make(map[telegraf.TrackingID]telegraf.TrackingData)
	mu            = sync.Mutex{}

	// ErrSkipTracking indicates that tracking information could not be found after
	// deserializing a metric from bytes. In this case we should skip the metric
	// and continue as if it does not exist.
	ErrSkipTracking = errors.New("metric tracking data not found")
)

type serializedMetric struct {
	M   telegraf.Metric
	TID telegraf.TrackingID
}

func ToBytes(m telegraf.Metric) ([]byte, error) {
	var sm serializedMetric
	if um, ok := m.(telegraf.UnwrappableMetric); ok {
		sm.M = um.Unwrap()
	} else {
		sm.M = m
	}

	if tm, ok := m.(telegraf.TrackingMetric); ok {
		sm.TID = tm.TrackingID()
		mu.Lock()
		trackingStore[sm.TID] = tm.TrackingData()
		mu.Unlock()
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(&sm); err != nil {
		return nil, fmt.Errorf("failed to encode metric to bytes: %w", err)
	}
	return buf.Bytes(), nil
}

func FromBytes(b []byte) (telegraf.Metric, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)

	var sm *serializedMetric
	if err := decoder.Decode(&sm); err != nil {
		return nil, fmt.Errorf("failed to decode metric from bytes: %w", err)
	}

	// Not a tracking metric
	if sm.TID == 0 {
		return sm.M, nil
	}

	// Try to lookup the tracking ID in the tracking-data store. If we cannot
	// find it, this is likely a left-ever from a previous read and we should
	// skip the tracking metric.
	mu.Lock()
	defer mu.Unlock()
	td, found := trackingStore[sm.TID]
	if !found {
		return sm.M, ErrSkipTracking
	}

	// Add back the tracking information to the metric
	return &trackingMetric{Metric: sm.M, d: td.(*trackingData)}, nil
}
