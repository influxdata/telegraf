package zipkin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type MockTracer struct {
	Data Trace
	Err  error
}

func (m *MockTracer) Record(t Trace) error {
	fmt.Println("Adding trace ", t)
	m.Data = t
	return nil
}

func (m *MockTracer) Error(err error) {
	m.Err = err
}

func TestZipkinServer(t *testing.T) {
	dat, err := ioutil.ReadFile("testdata/threespans.dat")
	if err != nil {
		t.Fatalf("Could not find file %s\n", "test/threespans.dat")
	}

	s := NewServer("/api/v1/spans")
	mockTracer := &MockTracer{}
	s.tracer = mockTracer
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"POST",
		"http://server.local/api/v1/spans",
		ioutil.NopCloser(
			bytes.NewReader(dat)))
	handler := s.SpanHandler
	handler(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("MainHandler did not return StatusNoContent %d", w.Code)
	}

	got := mockTracer.Data

	d := int64(53106)
	d1 := int64(50410)
	d2 := int64(103680)
	parentID1 := int64(22964302721410078)
	want := Trace{
		Span{
			Name:        "Child",
			ID:          "8090652509916334619",
			TraceIDHigh: nil,
			ParentID:    &parentID1,
			Timestamp:   time.Unix(1498688360851331, 0),
			Duration:    &d,
			//note: []Annotation(nil) is different than
			// []Annotation{}
			Annotations: []Annotation(nil),
			BinaryAnnotations: []BinaryAnnotation{
				BinaryAnnotation{
					Key:         "lc",
					Value:       "trivial",
					Host:        "2130706433:0",
					ServiceName: "trivial",
					Type:        "STRING",
				},
			},
		},
		Span{
			Name:        "Child",
			ID:          "103618986556047333",
			TraceIDHigh: nil,
			ParentID:    &parentID1,
			Timestamp:   time.Unix(1498688360904552, 0),
			Duration:    &d1,
			Annotations: []Annotation(nil),
			BinaryAnnotations: []BinaryAnnotation{
				BinaryAnnotation{
					Key:         "lc",
					Value:       "trivial",
					Host:        "2130706433:0",
					ServiceName: "trivial",
					Type:        "STRING",
				},
			},
		},
		Span{
			Name:      "Parent",
			ID:        "22964302721410078",
			Timestamp: time.Unix(1498688360851318, 0),
			Duration:  &d2,
			Annotations: []Annotation{
				Annotation{
					Timestamp:   time.Unix(1498688360851325, 0),
					Value:       "Starting child #0",
					Host:        "2130706433:0",
					ServiceName: "trivial",
				},
				Annotation{
					Timestamp:   time.Unix(1498688360904545, 0),
					Value:       "Starting child #1",
					Host:        "2130706433:0",
					ServiceName: "trivial",
				},
				Annotation{
					Timestamp:   time.Unix(1498688360954992, 0),
					Value:       "A Log",
					Host:        "2130706433:0",
					ServiceName: "trivial",
				},
			},
			BinaryAnnotations: []BinaryAnnotation{
				BinaryAnnotation{
					Key:         "lc",
					Value:       "trivial",
					Host:        "2130706433:0",
					ServiceName: "trivial",
					Type:        "STRING",
				},
			},
		},
	}

	fmt.Println("BINARY ANNOTATIONS FOR TESTING: ")
	fmt.Println(got[0].BinaryAnnotations, want[0].BinaryAnnotations)

	if !reflect.DeepEqual(got, want) {
		t.Fatal("Got != want, Fields weren't unmarshalled correctly")
	}
}
