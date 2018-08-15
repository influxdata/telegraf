package influx

import (
	"bytes"
	"io"
	"log"

	"github.com/influxdata/telegraf"
)

// reader is an io.Reader for line protocol.
type reader struct {
	metrics    []telegraf.Metric
	serializer *Serializer
	offset     int
	buf        *bytes.Buffer
}

// NewReader creates a new reader over the given metrics.
func NewReader(metrics []telegraf.Metric, serializer *Serializer) io.Reader {
	return &reader{
		metrics:    metrics,
		serializer: serializer,
		offset:     0,
		buf:        bytes.NewBuffer(make([]byte, 0, serializer.maxLineBytes)),
	}
}

// SetMetrics changes the metrics to be read.
func (r *reader) SetMetrics(metrics []telegraf.Metric) {
	r.metrics = metrics
	r.offset = 0
	r.buf.Reset()
}

// Read reads up to len(p) bytes of the current metric into p, each call will
// only serialize at most one metric so the number of bytes read may be less
// than p.  Subsequent calls to Read will read the next metric until all are
// emitted.  If a metric cannot be serialized, an error will be returned, you
// may resume with the next metric by calling Read again.  When all metrics
// are emitted the err is io.EOF.
func (r *reader) Read(p []byte) (int, error) {
	if r.buf.Len() > 0 {
		return r.buf.Read(p)
	}

	if r.offset >= len(r.metrics) {
		return 0, io.EOF
	}

	for _, metric := range r.metrics[r.offset:] {
		_, err := r.serializer.Write(r.buf, metric)
		r.offset += 1
		if err != nil {
			r.buf.Reset()
			if err != nil {
				// Since we are serializing multiple metrics, don't fail the
				// the entire batch just because of one unserializable metric.
				log.Printf("E! [serializers.influx] could not serialize metric: %v; discarding metric", err)
				continue
			}
		}
		break
	}

	return r.buf.Read(p)
}
