package http_listener

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/http_listener/stoppableListener"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const (
	// DEFAULT_REQUEST_BODY_MAX is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	// 1 GB
	DEFAULT_REQUEST_BODY_MAX = 1 * 1000 * 1000 * 1000

	// MAX_ALLOCATION_SIZE is the maximum size, in bytes, of a single allocation
	// of bytes that will be made handling a single HTTP request.
	// 15 MB
	MAX_ALLOCATION_SIZE = 10 * 1000 * 1000
)

type HttpListener struct {
	ServiceAddress string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    int64

	sync.Mutex

	listener *stoppableListener.StoppableListener

	parser parsers.Parser
	acc    telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## timeouts
  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 1,000,000,000 bytes (1 gigabyte)
  max_body_size = 0
`

func (t *HttpListener) SampleConfig() string {
	return sampleConfig
}

func (t *HttpListener) Description() string {
	return "Influx HTTP write listener"
}

func (t *HttpListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (t *HttpListener) SetParser(parser parsers.Parser) {
	t.parser = parser
}

// Start starts the http listener service.
func (t *HttpListener) Start(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	if t.MaxBodySize == 0 {
		t.MaxBodySize = DEFAULT_REQUEST_BODY_MAX
	}

	t.acc = acc

	var rawListener, err = net.Listen("tcp", t.ServiceAddress)
	if err != nil {
		return err
	}
	t.listener, err = stoppableListener.New(rawListener)
	if err != nil {
		return err
	}

	go t.httpListen()

	log.Printf("I! Started HTTP listener service on %s\n", t.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (t *HttpListener) Stop() {
	t.Lock()
	defer t.Unlock()

	t.listener.Stop()
	t.listener.Close()

	log.Println("I! Stopped HTTP listener service on ", t.ServiceAddress)
}

// httpListen listens for HTTP requests.
func (t *HttpListener) httpListen() error {
	if t.ReadTimeout.Duration < time.Second {
		t.ReadTimeout.Duration = time.Second * 10
	}
	if t.WriteTimeout.Duration < time.Second {
		t.WriteTimeout.Duration = time.Second * 10
	}

	var server = http.Server{
		Handler:      t,
		ReadTimeout:  t.ReadTimeout.Duration,
		WriteTimeout: t.WriteTimeout.Duration,
	}

	return server.Serve(t.listener)
}

func (t *HttpListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/write":
		var msg413 bytes.Buffer
		var msg400 bytes.Buffer
		defer func() {
			if msg413.Len() > 0 {
				res.WriteHeader(http.StatusRequestEntityTooLarge)
				res.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, msg413.String())))
			} else if msg400.Len() > 0 {
				res.Header().Set("Content-Type", "application/json")
				res.Header().Set("X-Influxdb-Version", "1.0")
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, msg400.String())))
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
		}()

		// Check that the content length is not too large for us to handle.
		if req.ContentLength > t.MaxBodySize {
			msg413.WriteString("http: request body too large")
			return
		}

		// Handle gzip request bodies
		var body io.ReadCloser
		var err error
		if req.Header.Get("Content-Encoding") == "gzip" {
			body, err = gzip.NewReader(http.MaxBytesReader(res, req.Body, t.MaxBodySize))
			if err != nil {
				msg400.WriteString(err.Error() + " ")
				return
			}
		} else {
			body = http.MaxBytesReader(res, req.Body, t.MaxBodySize)
		}

		var buffer *bytes.Buffer
		if req.ContentLength < MAX_ALLOCATION_SIZE {
			// if the content length is less than the max allocation size, then
			// read in the whole request at once:
			buffer = bytes.NewBuffer(make([]byte, 0, req.ContentLength+1))
			_, err := buffer.ReadFrom(body)
			if err != nil {
				msg := "E! "
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						msg += "Read timeout error, you may want to increase the read_timeout setting. "
					}
				}
				log.Printf(msg + err.Error())
				msg400.WriteString("Error reading request body: " + err.Error())
				return
			}
		} else {
			// If the body is larger than the max allocation size then set the
			// maximum size of the buffer that we will allocate at a time.
			// The following loop goes through the request body byte-by-byte.
			// If there is a newline within 256 kilobytes of the end of the body
			// we will attempt to parse metrics, reset the buffer, and continue.
			buffer = bytes.NewBuffer(make([]byte, 0, MAX_ALLOCATION_SIZE))
			reader := bufio.NewReader(body)
			for {
				b, err := reader.ReadByte()
				if err != nil {
					if err != io.EOF {
						msg := "E! "
						if netErr, ok := err.(net.Error); ok {
							if netErr.Timeout() {
								msg += "Read timeout error, you may want to increase the read_timeout setting. "
							}
						} else {
							// if it's not an EOF or a net.Error, then it's almost certainly a
							// tooLarge error coming from http.MaxBytesReader. It's unlikely
							// that this code path will get hit because the client should
							// be setting the ContentLength header, unless it's malicious.
							msg413.WriteString(err.Error())
						}
						log.Printf(msg + err.Error())
						return
					}
					break
				}
				// returned error is always nil:
				// https://golang.org/pkg/bytes/#Buffer.WriteByte
				buffer.WriteByte(b)
				// if we have a newline and we're nearing the end of the buffer,
				// do a write and continue with a fresh buffer.
				if buffer.Len() > MAX_ALLOCATION_SIZE-256*1000 && b == '\n' {
					t.parse(buffer.Bytes(), &msg400)
					buffer.Reset()
				} else if buffer.Len() == buffer.Cap() {
					// we've reached the end of our buffer without finding a newline
					// in the body, so we insert a newline here and attempt to parse.
					if buffer.Len() == 0 {
						continue
					}
					buffer.WriteByte('\n')
					t.parse(buffer.Bytes(), &msg400)
					buffer.Reset()
				}
			}
		}

		if buffer.Len() != 0 {
			t.parse(buffer.Bytes(), &msg400)
		}
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
		http.NotFound(res, req)
	}
}

func (t *HttpListener) parse(b []byte, errmsg *bytes.Buffer) {
	metrics, err := t.parser.Parse(b)
	if err != nil {
		if len(metrics) == 0 {
			errmsg.WriteString(err.Error())
		} else {
			errmsg.WriteString("partial write: " + err.Error())
		}
	}

	for _, m := range metrics {
		t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
}

func init() {
	inputs.Add("http_listener", func() telegraf.Input {
		return &HttpListener{}
	})
}
