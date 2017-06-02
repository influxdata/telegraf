package metric

import (
	"io"

	"github.com/influxdata/telegraf"
)

type state int

const (
	_ state = iota
	// normal state copies whole metrics into the given buffer until we can't
	// fit the next metric.
	normal
	// split state means that we have a metric that we were able to split, so
	// that we can fit it into multiple metrics (and calls to Read)
	split
	// overflow state means that we have a metric that didn't fit into a single
	// buffer, and needs to be split across multiple calls to Read.
	overflow
	// splitOverflow state means that a split metric didn't fit into a single
	// buffer, and needs to be split across multiple calls to Read.
	splitOverflow
	// done means we're done reading metrics, and now always return (0, io.EOF)
	done
)

type reader struct {
	metrics      []telegraf.Metric
	splitMetrics []telegraf.Metric
	buf          []byte
	state        state

	// metric index
	iM int
	// split metric index
	iSM int
	// buffer index
	iB int
}

func NewReader(metrics []telegraf.Metric) io.Reader {
	return &reader{
		metrics: metrics,
		state:   normal,
	}
}

func (r *reader) Read(p []byte) (n int, err error) {
	var i int
	switch r.state {
	case done:
		return 0, io.EOF
	case normal:
		for {
			// this for-loop is the sunny-day scenario, where we are given a
			// buffer that is large enough to hold at least a single metric.
			// all of the cases below it are edge-cases.
			if r.metrics[r.iM].Len() <= len(p[i:]) {
				i += r.metrics[r.iM].SerializeTo(p[i:])
			} else {
				break
			}
			r.iM++
			if r.iM == len(r.metrics) {
				r.state = done
				return i, io.EOF
			}
		}

		// if we haven't written any bytes, check if we can split the current
		// metric into multiple full metrics at a smaller size.
		if i == 0 {
			tmp := r.metrics[r.iM].Split(len(p))
			if len(tmp) > 1 {
				r.splitMetrics = tmp
				r.state = split
				if r.splitMetrics[0].Len() <= len(p) {
					i += r.splitMetrics[0].SerializeTo(p)
					r.iSM = 1
				} else {
					// splitting didn't quite work, so we'll drop down and
					// overflow the metric.
					r.state = normal
					r.iSM = 0
				}
			}
		}

		// if we haven't written any bytes and we're not at the end of the metrics
		// slice, then it means we have a single metric that is larger than the
		// provided buffer.
		if i == 0 {
			r.buf = r.metrics[r.iM].Serialize()
			i += copy(p, r.buf[r.iB:])
			r.iB += i
			r.state = overflow
		}

	case split:
		if r.splitMetrics[r.iSM].Len() <= len(p) {
			// write the current split metric
			i += r.splitMetrics[r.iSM].SerializeTo(p)
			r.iSM++
			if r.iSM >= len(r.splitMetrics) {
				// done writing the current split metrics
				r.iSM = 0
				r.iM++
				if r.iM == len(r.metrics) {
					r.state = done
					return i, io.EOF
				}
				r.state = normal
			}
		} else {
			// This would only happen if we split the metric, and then a
			// subsequent buffer was smaller than the initial one given,
			// so that our split metric no longer fits.
			r.buf = r.splitMetrics[r.iSM].Serialize()
			i += copy(p, r.buf[r.iB:])
			r.iB += i
			r.state = splitOverflow
		}

	case splitOverflow:
		i = copy(p, r.buf[r.iB:])
		r.iB += i
		if r.iB >= len(r.buf) {
			r.iB = 0
			r.iSM++
			if r.iSM == len(r.splitMetrics) {
				r.iM++
				if r.iM == len(r.metrics) {
					r.state = done
					return i, io.EOF
				}
				r.state = normal
			} else {
				r.state = split
			}
		}

	case overflow:
		i = copy(p, r.buf[r.iB:])
		r.iB += i
		if r.iB >= len(r.buf) {
			r.iB = 0
			r.iM++
			if r.iM == len(r.metrics) {
				r.state = done
				return i, io.EOF
			}
			r.state = normal
		}
	}

	return i, nil
}
