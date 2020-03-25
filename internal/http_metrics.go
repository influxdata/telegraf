package internal

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/selfstat"
)

func MeasuringHandler(name string, tags map[string]string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &measuringHandler{
			requestTimeNs:         selfstat.RegisterTiming(name, "request_time_ns", tags),
			statusResponseWrapper: responseWriterWrapper(name, tags),
			next:                  h,
		}
	}
}

type measuringHandler struct {
	requestTimeNs         selfstat.Stat
	statusResponseWrapper func(rw http.ResponseWriter) http.ResponseWriter
	next                  http.Handler
}

func (mh *measuringHandler) requestFinished(startedAt time.Time) {
	mh.requestTimeNs.Incr(time.Now().Sub(startedAt).Nanoseconds())
}

// satisfies the http.Handler interface
func (mh *measuringHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()
	defer mh.requestFinished(start)

	mh.next.ServeHTTP(mh.statusResponseWrapper(rw), req)
}

type measuringResponseWriter struct {
	requestCountByClass map[int]selfstat.Stat

	next http.ResponseWriter
}

func makeClassCounter(name string, tags map[string]string, code int) selfstat.Stat {
	newTags := make(map[string]string)
	for k, v := range tags {
		newTags[k] = v
	}
	newTags["responseClass"] = strings.Join([]string{strconv.Itoa(code), "xx"}, "")
	return selfstat.Register(name, "request_count", newTags)
}

func responseWriterWrapper(name string, tags map[string]string) func(rw http.ResponseWriter) http.ResponseWriter {
	return func(rw http.ResponseWriter) http.ResponseWriter {
		counters := make(map[int]selfstat.Stat)
		for i := 0; i < 6; i++ {
			counters[i] = makeClassCounter(name, tags, i)
		}
		return &measuringResponseWriter{
			requestCountByClass: counters,
			next:                rw,
		}
	}
}

func (mrw *measuringResponseWriter) requestCodeHeaderWritten(code int) {
	var cl int = code / 100
	counter, ok := mrw.requestCountByClass[cl]
	if ok {
		counter.Incr(1)
	} else {
		mrw.requestCountByClass[0].Incr(1)
	}
}

func (mrw *measuringResponseWriter) Header() http.Header {
	return mrw.next.Header()
}

func (mrw *measuringResponseWriter) Write(data []byte) (int, error) {
	return mrw.next.Write(data)
}

func (mrw *measuringResponseWriter) WriteHeader(code int) {
	defer mrw.requestCodeHeaderWritten(code)
	mrw.next.WriteHeader(code)
}
