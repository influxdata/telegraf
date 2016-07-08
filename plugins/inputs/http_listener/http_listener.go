package http_listener

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/http_listener/stoppableListener"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type HttpListener struct {
	ServiceAddress string
	ReadTimeout    string
	WriteTimeout   string

	sync.Mutex

	listener *stoppableListener.StoppableListener

	parser parsers.Parser
	acc    telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## timeouts in seconds
  read_timeout = "10"
  write_timeout = "10"
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

	log.Println("Stopped HTTP listener service on ", t.ServiceAddress)
}

// httpListen listens for HTTP requests.
func (t *HttpListener) httpListen() error {

	readTimeout, err := strconv.ParseInt(t.ReadTimeout, 10, 32)
	if err != nil {
		return err
	}
	writeTimeout, err := strconv.ParseInt(t.WriteTimeout, 10, 32)
	if err != nil {
		return err
	}

	var server = http.Server{
		Handler:      t,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
	}

	return server.Serve(t.listener)
}

func (t *HttpListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		log.Printf("Problem reading request: [%s], Error: %s\n", string(body), err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("ERROR reading request"))
	}

	var path = req.URL.Path[1:]

	if path == "write" {
		var metrics []telegraf.Metric
		metrics, err = t.parser.Parse(body)
		if err == nil {
			t.storeMetrics(metrics)
			res.WriteHeader(http.StatusNoContent)
			res.Write([]byte(""))
		} else {
			log.Printf("Problem parsing body: [%s], Error: %s\n", string(body), err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("ERROR parsing metrics"))
		}
	} else if path == "query" {
		// Deliver a dummy response to the query endpoint, as some InfluxDB clients test endpoint availability with a query
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Influxdb-Version", "1.0")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{\"results\":[]}"))
	} else {
		// Don't know how to respond to calls to other endpoints
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte("Not Found"))
	}
}

func (t *HttpListener) storeMetrics(metrics []telegraf.Metric) error {
	t.Lock()
	defer t.Unlock()

	for _, m := range metrics {
		t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
	return nil
}

func init() {
	inputs.Add("http_listener", func() telegraf.Input {
		return &HttpListener{}
	})
}
