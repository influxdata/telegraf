package http_listener

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"io/ioutil"

	"github.com/hydrogen18/stoppableListener"
	"strconv"
)

type HttpListener struct {
	ServiceAddress  string
	ReadTimeout	string
	WriteTimeout	string

	sync.Mutex

	listener *stoppableListener.StoppableListener

	parser parsers.Parser
	acc    telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8086"

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
	t.listener, err = stoppableListener.New(rawListener)

	go t.httpListen()

	log.Printf("Started HTTP listener service on %s\n", t.ServiceAddress)

	return err
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
	writeTimeout, err := strconv.ParseInt(t.WriteTimeout, 10, 32)

	var server = http.Server{
		Handler:        t,
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
	}

	err = server.Serve(t.listener)

	return err
}

func (t *HttpListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)

	if err == nil {
		log.Printf("Received request: [%s]\n", string(body))

		var metrics []telegraf.Metric
		if len(body) == 0 {
			log.Printf("No metrics to parse\n")
		} else {
			metrics, err = t.parser.Parse(body)
			if err == nil {
				t.storeMetrics(metrics)
				log.Printf("Persisted %d metrics\n", len(metrics))
			} else {
				log.Printf("Problem parsing body: [%s], Error: %s\n", string(body), err)
				res.WriteHeader(500)
				res.Write([]byte("ERROR parsing metrics"))
			}
			res.WriteHeader(204)
			res.Write([]byte(""))
		}
	} else {
		log.Printf("Problem reading request: [%s], Error: %s\n", string(body), err)
		res.WriteHeader(500)
		res.Write([]byte("ERROR reading request"))
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
