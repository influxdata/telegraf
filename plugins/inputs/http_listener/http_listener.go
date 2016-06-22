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
)

type HttpListener struct {
	ServiceAddress  string
	readTimeout	int
	writeTimeout	int

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

	var server = http.Server{
		Handler:        t.writeHandler,
		ReadTimeout:    t.readTimeout * time.Second,
		WriteTimeout:   t.writeTimeout * time.Second,
	}

	var err = server.Serve(t.listener)

	return err
}

func (t *HttpListener) writeHandler(res http.ResponseWriter, req *http.Request) error {
	body, err := ioutil.ReadAll(req.Body)

	if err == nil {
		var metrics []telegraf.Metric
		for {
			if len(body) == 0 {
				continue
			}
			metrics, err = t.parser.Parse(body)
			if err == nil {
				t.storeMetrics(metrics)
			} else {
				log.Printf("Problem parsing body: [%s], Error: %s\n", string(body), err)
				res.WriteHeader(500)
				res.Write([]byte("ERROR parsing metrics"))
			}
		}
		res.WriteHeader(204)
		res.Write([]byte(""))
	} else {
		log.Printf("Problem reading request: [%s], Error: %s\n", string(body), err)
		res.WriteHeader(500)
		res.Write([]byte("ERROR reading request"))
	}

	return err
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
