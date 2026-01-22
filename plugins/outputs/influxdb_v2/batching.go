package influxdb_v2

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/ratelimiter"
)

type batch struct {
	bucket    string
	metrics   []telegraf.Metric
	indices   []int
	payload   []byte
	processed bool
	err       error
}

func createBatches(metrics []telegraf.Metric, bucket string, size int) []*batch {
	number := len(metrics) / size
	if len(metrics)%size > 0 {
		number++
	}
	batches := make([]*batch, 0, number)
	for i := range number {
		begin := i * size
		end := min(begin+size, len(metrics))
		b := &batch{
			bucket:  bucket,
			metrics: metrics[begin:end],
			indices: make([]int, end-begin),
		}

		for j := range b.metrics {
			b.indices[j] = begin + j
		}
		batches = append(batches, b)
	}
	return batches
}

func createBatchesFromTag(metrics []telegraf.Metric, tag, fallback string, size int, exclude bool) []*batch {
	// Initial guess for the number of batches, there might be more depending
	// on the number of unique buckets, but there must be at least that many
	collector := make(map[string]*batch)
	for i, metric := range metrics {
		bucket, ok := metric.GetTag(tag)
		if !ok {
			bucket = fallback
		} else if exclude {
			// Avoid modifying the metric if we do remove the tag
			metric = metric.Copy()
			metric.Accept()
			metric.RemoveTag(tag)
		}

		// Create a new bucket if there is none yet
		b, found := collector[bucket]
		if !found {
			b = &batch{bucket: bucket}
		}
		b.metrics = append(b.metrics, metric)
		b.indices = append(b.indices, i)
		collector[bucket] = b
	}

	batches := make([]*batch, 0, len(collector))
	for _, b := range collector {
		if len(b.metrics) <= size {
			batches = append(batches, b)
			continue
		}
		// The batch is larger than it should be, so split it
		var begin int
		for begin < len(b.metrics) {
			end := min(begin+size, len(b.metrics))
			batches = append(batches, &batch{
				bucket:  b.bucket,
				metrics: b.metrics[begin:end],
				indices: b.indices[begin:end],
			})
			begin = end
		}
	}

	return batches
}

func (b *batch) split() (first, second *batch) {
	midpoint := len(b.metrics) / 2

	return &batch{
			bucket:  b.bucket,
			metrics: b.metrics[:midpoint],
			indices: b.indices[:midpoint],
		},
		&batch{
			bucket:  b.bucket,
			metrics: b.metrics[midpoint:],
			indices: b.indices[midpoint:],
		}
}

func (b *batch) serialize(serializer ratelimiter.Serializer, limit int64, encoder internal.ContentEncoder) (int64, error) {
	// Serialize the metrics with the remaining limit,
	body, serr := serializer.SerializeBatch(b.metrics, limit)
	if serr != nil && !errors.Is(serr, internal.ErrSizeLimitReached) {
		// When only part of the metrics failed to be serialized we should remove
		// them from the normal handling and mark them as rejected for upstream
		// to pass on this information
		var werr *internal.PartialWriteError
		if errors.As(serr, &werr) {
			for i, idx := range slices.Backward(werr.MetricsReject) {
				werr.MetricsReject[i] = b.indices[idx]
				b.indices = slices.Delete(b.indices, idx, idx+1)
			}
			serr = werr
		}
	}

	// Exit early if nothing was serialized
	if len(body) == 0 {
		return 0, serr
	}

	// Encode the content if requested
	if encoder != nil {
		enc, err := encoder.Encode(body)
		if err != nil {
			return 0, fmt.Errorf("encoding failed: %w", err)
		}
		b.payload = bytes.Clone(enc)
	} else {
		b.payload = bytes.Clone(body)
	}

	return int64(len(body)), serr
}
