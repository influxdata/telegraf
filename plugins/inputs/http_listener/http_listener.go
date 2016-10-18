package http_listener

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

const (
	// DEFAULT_MAX_BODY_SIZE is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	// 500 MB
	DEFAULT_MAX_BODY_SIZE = 500 * 1024 * 1024

	// MAX_LINE_SIZE is the maximum size, in bytes, that can be allocated for
	// a single InfluxDB point.
	// 64 KB
	DEFAULT_MAX_LINE_SIZE = 64 * 1024
)

type HttpListener struct {
	ServiceAddress string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    int64
	MaxLineSize    int

	mu sync.Mutex
	wg sync.WaitGroup

	listener net.Listener

	parser influx.InfluxParser
	acc    telegraf.Accumulator
	pool   *pool
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 536,870,912 bytes (500 mebibytes)
  max_body_size = 0

  ## Maximum line size allowed to be sent in bytes.
  ## 0 means to use the default of 65536 bytes (64 kibibytes)
  max_line_size = 0
`

func (h *HttpListener) SampleConfig() string {
	return sampleConfig
}

func (h *HttpListener) Description() string {
	return "Influx HTTP write listener"
}

func (h *HttpListener) Gather(_ telegraf.Accumulator) error {
	log.Printf("D! The http_listener has created %d buffers", h.pool.ncreated())
	return nil
}

// Start starts the http listener service.
func (h *HttpListener) Start(acc telegraf.Accumulator) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.parser = influx.InfluxParser{}

	if h.MaxBodySize == 0 {
		h.MaxBodySize = DEFAULT_MAX_BODY_SIZE
	}
	if h.MaxLineSize == 0 {
		h.MaxLineSize = DEFAULT_MAX_LINE_SIZE
	}

	h.acc = acc
	h.pool = NewPool(200, h.MaxLineSize)

	var listener, err = net.Listen("tcp", h.ServiceAddress)
	if err != nil {
		return err
	}
	h.listener = listener

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.httpListen()
	}()

	log.Printf("I! Started HTTP listener service on %s\n", h.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (h *HttpListener) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.listener.Close()
	h.wg.Wait()

	log.Println("I! Stopped HTTP listener service on ", h.ServiceAddress)
}

// httpListen sets up an http.Server and calls server.Serve.
// like server.Serve, httpListen will always return a non-nil error, for this
// reason, the error returned should probably be ignored.
// see https://golang.org/pkg/net/http/#Server.Serve
func (h *HttpListener) httpListen() error {
	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

	var server = http.Server{
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
	}

	return server.Serve(h.listener)
}

func (h *HttpListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/write":
		h.serveWrite(res, req)
	case "/query":
		// Deliver a dummy response to the query endpoint, as some InfluxDB
		// clients test endpoint availability with a query
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Influxdb-Version", "1.0")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{\"results\":[]}"))
	case "/ping":
		// respond to ping requests
		res.WriteHeader(http.StatusNoContent)
	default:
		// Don't know how to respond to calls to other endpoints
		http.NotFound(res, req)
	}
}

func (h *HttpListener) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize {
		tooLarge(res)
		return
	}
	now := time.Now()

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		body, err := gzip.NewReader(req.Body)
		defer body.Close()
		if err != nil {
			log.Println("E! " + err.Error())
			badRequest(res)
			return
		}
	}
	body = http.MaxBytesReader(res, body, h.MaxBodySize)

	var return400 bool
	var hangingBytes bool
	buf := h.pool.get()
	defer func() { h.pool.put(buf) }()
	bufStart := 0
	for {
		n, err := io.ReadFull(body, buf[bufStart:])
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			log.Println("E! " + err.Error())
			// problem reading the request body
			badRequest(res)
			return
		}

		if err == io.EOF {
			if return400 {
				badRequest(res)
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
			return
		}

		if hangingBytes {
			i := bytes.IndexByte(buf, '\n')
			if i == -1 {
				// still didn't find a newline, keep scanning
				continue
			}
			// rotate the bit remaining after the first newline to the front of the buffer
			i++ // start copying after the newline
			bufStart = len(buf) - i
			if bufStart > 0 {
				copy(buf, buf[i:])
			}
			hangingBytes = false
			continue
		}

		if err == io.ErrUnexpectedEOF {
			// finished reading the request body
			if err := h.parse(buf[:n+bufStart], now); err != nil {
				log.Println("E! " + err.Error())
				return400 = true
			}
			if return400 {
				badRequest(res)
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
			return
		}

		// if we got down here it means that we filled our buffer, and there
		// are still bytes remaining to be read. So we will parse up until the
		// final newline, then push the rest of the bytes into the next buffer.
		i := bytes.LastIndexByte(buf, '\n')
		if i == -1 {
			// drop any line longer than the max buffer size
			log.Printf("E! http_listener received a single line longer than the maximum of %d bytes",
				len(buf))
			hangingBytes = true
			return400 = true
			bufStart = 0
			continue
		}
		if err := h.parse(buf[:i], now); err != nil {
			log.Println("E! " + err.Error())
			return400 = true
		}
		// rotate the bit remaining after the last newline to the front of the buffer
		i++ // start copying after the newline
		bufStart = len(buf) - i
		if bufStart > 0 {
			copy(buf, buf[i:])
		}
	}
}

func (h *HttpListener) parse(b []byte, t time.Time) error {
	metrics, err := h.parser.ParseWithDefaultTime(b, t)

	for _, m := range metrics {
		h.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}

	return err
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func init() {
	inputs.Add("http_listener", func() telegraf.Input {
		return &HttpListener{
			ServiceAddress: ":8186",
		}
	})
}
