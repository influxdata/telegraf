package zipkin

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec/jsonV1"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec/thrift"
)

// SpanHandler is an implementation of a Handler which accepts zipkin thrift
// span data and sends it to the recorder
type SpanHandler struct {
	Path     string
	recorder Recorder
}

// NewSpanHandler returns a new server instance given path to handle
func NewSpanHandler(path string) *SpanHandler {
	return &SpanHandler{
		Path: path,
	}
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set(`Access-Control-Allow-Origin`, origin)
			w.Header().Set(`Access-Control-Allow-Methods`, strings.Join([]string{
				`OPTIONS`,
				`POST`,
			}, ", "))

			w.Header().Set(`Access-Control-Allow-Headers`, strings.Join([]string{
				`Accept`,
				`Accept-Encoding`,
				`Content-Length`,
				`Content-Type`,
			}, ", "))

			w.Header().Set(`Access-Control-Expose-Headers`, strings.Join([]string{
				`Date`,
			}, ", "))
		}

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	}
}

// Register implements the Service interface. Register accepts zipkin thrift data
// POSTed to the path of the mux router
func (s *SpanHandler) Register(router *mux.Router, recorder Recorder) error {
	handler := cors(http.HandlerFunc(s.Spans))
	router.Handle(s.Path, handler).Methods("POST", "OPTIONS")
	s.recorder = recorder
	return nil
}

// Spans handles zipkin thrift spans
func (s *SpanHandler) Spans(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body := r.Body
	var err error
	// Handle gzip decoding of the body
	if r.Header.Get("Content-Encoding") == "gzip" {
		body, err = gzip.NewReader(r.Body)
		if err != nil {
			s.recorder.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer body.Close()
	}

	decoder, err := ContentDecoder(r)
	if err != nil {
		s.recorder.Error(err)
		w.WriteHeader(http.StatusUnsupportedMediaType)
	}

	octets, err := ioutil.ReadAll(body)
	if err != nil {
		s.recorder.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	spans, err := decoder.Decode(octets)
	if err != nil {
		s.recorder.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	trace, err := codec.NewTrace(spans)
	if err != nil {
		s.recorder.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = s.recorder.Record(trace); err != nil {
		s.recorder.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ContentDecoer returns a Decoder that is able to produce Traces from bytes.
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
// If a Content-Type is not set, zipkin assumes application/json
func ContentDecoder(r *http.Request) (codec.Decoder, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return &jsonV1.JSON{}, nil
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == "application/json" {
			return &jsonV1.JSON{}, nil
		} else if t == "application/x-thrift" {
			return &thrift.Thrift{}, nil
		}
	}
	return nil, fmt.Errorf("Unknown Content-Type: %s", contentType)
}
