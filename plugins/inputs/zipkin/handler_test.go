package zipkin

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin/trace"
)

type mockRecorder struct {
	data trace.Trace
	err  error
}

func (m *mockRecorder) record(t trace.Trace) error {
	m.data = t
	return nil
}

func (m *mockRecorder) error(err error) {
	m.err = err
}

func TestSpanHandler(t *testing.T) {
	dat, err := os.ReadFile("testdata/threespans.dat")
	if err != nil {
		t.Fatalf("Could not find file %s\n", "testdata/threespans.dat")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"POST",
		"http://server.local/api/v1/spans",
		io.NopCloser(
			bytes.NewReader(dat)))

	r.Header.Set("Content-Type", "application/x-thrift")
	handler := newSpanHandler("/api/v1/spans")
	mockRecorder := &mockRecorder{}
	handler.recorder = mockRecorder

	handler.spans(w, r)
	require.Equal(t, http.StatusNoContent, w.Code)

	got := mockRecorder.data

	parentID := strconv.FormatInt(22964302721410078, 16)
	want := trace.Trace{
		{
			Name:        "Child",
			ID:          "7047c59776af8a1b",
			TraceID:     "22c4fc8ab3669045",
			ParentID:    parentID,
			Timestamp:   time.Unix(0, 1498688360851331*int64(time.Microsecond)).UTC(),
			Duration:    time.Duration(53106) * time.Microsecond,
			ServiceName: "trivial",
			Annotations: make([]trace.Annotation, 0),
			BinaryAnnotations: []trace.BinaryAnnotation{
				{
					Key:         "lc",
					Value:       "trivial",
					Host:        "127.0.0.1",
					ServiceName: "trivial",
				},
			},
		},
		{
			Name:        "Child",
			ID:          "17020eb55a8bfe5",
			TraceID:     "22c4fc8ab3669045",
			ParentID:    parentID,
			Timestamp:   time.Unix(0, 1498688360904552*int64(time.Microsecond)).UTC(),
			Duration:    time.Duration(50410) * time.Microsecond,
			ServiceName: "trivial",
			Annotations: make([]trace.Annotation, 0),
			BinaryAnnotations: []trace.BinaryAnnotation{
				{
					Key:         "lc",
					Value:       "trivial",
					Host:        "127.0.0.1",
					ServiceName: "trivial",
				},
			},
		},
		{
			Name:        "Parent",
			ID:          "5195e96239641e",
			TraceID:     "22c4fc8ab3669045",
			ParentID:    parentID,
			Timestamp:   time.Unix(0, 1498688360851318*int64(time.Microsecond)).UTC(),
			Duration:    time.Duration(103680) * time.Microsecond,
			ServiceName: "trivial",
			Annotations: []trace.Annotation{
				{
					Timestamp:   time.Unix(0, 1498688360851325*int64(time.Microsecond)).UTC(),
					Value:       "Starting child #0",
					Host:        "127.0.0.1",
					ServiceName: "trivial",
				},
				{
					Timestamp:   time.Unix(0, 1498688360904545*int64(time.Microsecond)).UTC(),
					Value:       "Starting child #1",
					Host:        "127.0.0.1",
					ServiceName: "trivial",
				},
				{
					Timestamp:   time.Unix(0, 1498688360954992*int64(time.Microsecond)).UTC(),
					Value:       "A Log",
					Host:        "127.0.0.1",
					ServiceName: "trivial",
				},
			},
			BinaryAnnotations: []trace.BinaryAnnotation{
				{
					Key:         "lc",
					Value:       "trivial",
					Host:        "127.0.0.1",
					ServiceName: "trivial",
				},
			},
		},
	}

	require.Equal(t, want, got)
}
