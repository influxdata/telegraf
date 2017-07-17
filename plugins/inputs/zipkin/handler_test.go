package zipkin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type MockRecorder struct {
	Data Trace
	Err  error
}

func (m *MockRecorder) Record(t Trace) error {
	m.Data = t
	return nil
}

func (m *MockRecorder) Error(err error) {
	m.Err = err
}

func TestSpanHandler(t *testing.T) {
	dat, err := ioutil.ReadFile("testdata/threespans.dat")
	if err != nil {
		t.Fatalf("Could not find file %s\n", "testdata/threespans.dat")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"POST",
		"http://server.local/api/v1/spans",
		ioutil.NopCloser(
			bytes.NewReader(dat)))

	handler := NewSpanHandler("/api/v1/spans")
	mockRecorder := &MockRecorder{}
	handler.recorder = mockRecorder

	handler.Spans(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("MainHandler did not return StatusNoContent %d", w.Code)
	}

	got := mockRecorder.Data

	parentID := strconv.FormatInt(22964302721410078, 10)
	want := Trace{
		Span{
			Name:        "Child",
			ID:          "8090652509916334619",
			TraceID:     "2505404965370368069",
			ParentID:    parentID,
			Timestamp:   time.Unix(0, 1498688360851331*int64(time.Microsecond)),
			Duration:    time.Duration(53106) * time.Microsecond,
			Annotations: nil,
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
			TraceID:     "2505404965370368069",
			ParentID:    parentID,
			Timestamp:   time.Unix(0, 1498688360904552*int64(time.Microsecond)),
			Duration:    time.Duration(50410) * time.Microsecond,
			Annotations: nil,
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
			TraceID:   "2505404965370368069",
			ParentID:  "22964302721410078",
			Timestamp: time.Unix(0, 1498688360851318*int64(time.Microsecond)),
			Duration:  time.Duration(103680) * time.Microsecond,
			Annotations: []Annotation{
				Annotation{
					Timestamp:   time.Unix(0, 1498688360851325*int64(time.Microsecond)),
					Value:       "Starting child #0",
					Host:        "2130706433:0",
					ServiceName: "trivial",
				},
				Annotation{
					Timestamp:   time.Unix(0, 1498688360904545*int64(time.Microsecond)),
					Value:       "Starting child #1",
					Host:        "2130706433:0",
					ServiceName: "trivial",
				},
				Annotation{
					Timestamp:   time.Unix(0, 1498688360954992*int64(time.Microsecond)),
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

	for i, s := range got {
		if !reflect.DeepEqual(s, want[i]) {
			fmt.Printf("index %d wasn't equal", i)
			fmt.Println(s, want[i])
			t.Fatal("Got != want, Fields weren't unmarshalled correctly")
		}
	}
}
