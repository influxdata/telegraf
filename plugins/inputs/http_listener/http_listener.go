package http_listener

import (
	"io/ioutil"
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

	log.Printf("Started HTTP listener service on %s\n", t.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (t *HttpListener) Stop() {
	t.Lock()
	defer t.Unlock()

	t.listener.Stop()
	t.listener.Close()

	t.wg.Wait()

	log.Println("Stopped HTTP listener service on ", t.ServiceAddress)
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
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Problem reading request: [%s], Error: %s\n", string(body), err)
		http.Error(res, "ERROR reading request", http.StatusInternalServerError)
		return
	}

	switch req.URL.Path {
	case "/write":
		var metrics []telegraf.Metric
		metrics, err = t.parser.Parse(body)
		if err == nil {
			for _, m := range metrics {
				t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
			}
			res.WriteHeader(http.StatusNoContent)
		} else {
			log.Printf("Problem parsing body: [%s], Error: %s\n", string(body), err)
			http.Error(res, "ERROR parsing metrics", http.StatusInternalServerError)
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
