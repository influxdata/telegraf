package http_listener

import (
	"bytes"
	"compress/gzip"
	"fmt"
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

const MAX_REQUEST_BODY_SIZE = 50 * 1024 * 1024

type HttpListener struct {
	ServiceAddress string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration

	sync.Mutex
	wg sync.WaitGroup

	listener *stoppableListener.StoppableListener

	parser parsers.Parser
	acc    telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## timeouts
  read_timeout = "10s"
  write_timeout = "10s"
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

	t.wg.Wait()

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
	t.wg.Add(1)
	defer t.wg.Done()

	switch req.URL.Path {
	case "/write":
		var http400msg bytes.Buffer
		defer func() {
			if http400msg.Len() > 0 {
				res.Header().Set("Content-Type", "application/json")
				res.Header().Set("X-Influxdb-Version", "1.0")
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, http400msg.String())))
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
		}()

		body := req.Body
		if req.Header.Get("Content-Encoding") == "gzip" {
			b, err := gzip.NewReader(req.Body)
			if err != nil {
				http400msg.WriteString(err.Error() + " ")
				return
			}
			defer b.Close()
			body = b
		}

		allocSize := 512
		if req.ContentLength < MAX_REQUEST_BODY_SIZE {
			allocSize = int(req.ContentLength)
		}
		buf := bytes.NewBuffer(make([]byte, 0, allocSize))
		_, err := buf.ReadFrom(http.MaxBytesReader(res, body, MAX_REQUEST_BODY_SIZE))
		if err != nil {
			log.Printf("E! HttpListener unable to read request body. error: %s\n", err.Error())
			http400msg.WriteString("HttpHandler unable to read from request body: " + err.Error())
			return
		}

		metrics, err := t.parser.Parse(buf.Bytes())
		if err != nil {
			log.Printf("E! HttpListener unable to parse metrics. error: %s \n", err.Error())
			if len(metrics) == 0 {
				http400msg.WriteString(err.Error())
			} else {
				http400msg.WriteString("partial write: " + err.Error())
			}
		}

		for _, m := range metrics {
			t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
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
		// Don't know how to respond to calls to other endpoints
		http.NotFound(res, req)
	}
}

func init() {
	inputs.Add("http_listener", func() telegraf.Input {
		return &HttpListener{}
	})
}
