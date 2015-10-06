package statsd

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdb/influxdb/services/graphite"

	"github.com/influxdb/telegraf/plugins"
)

var dropwarn = "ERROR: Message queue full. Discarding line [%s] " +
	"You may want to increase allowed_pending_messages in the config\n"

type Statsd struct {
	// Address & Port to serve from
	ServiceAddress string

	// Number of messages allowed to queue up in between calls to Gather. If this
	// fills up, packets will get dropped until the next Gather interval is ran.
	AllowedPendingMessages int

	DeleteGauges   bool
	DeleteCounters bool
	DeleteSets     bool

	sync.Mutex

	// Channel for all incoming statsd messages
	in        chan string
	inmetrics chan metric
	done      chan struct{}

	// Cache gauges, counters & sets so they can be aggregated as they arrive
	gauges   map[string]cachedgauge
	counters map[string]cachedcounter
	sets     map[string]cachedset

	// bucket -> influx templates
	Templates []string
}

func NewStatsd() *Statsd {
	s := Statsd{}

	// Make data structures
	s.done = make(chan struct{})
	s.in = make(chan string, s.AllowedPendingMessages)
	s.inmetrics = make(chan metric, s.AllowedPendingMessages)
	s.gauges = make(map[string]cachedgauge)
	s.counters = make(map[string]cachedcounter)
	s.sets = make(map[string]cachedset)

	return &s
}

// One statsd metric, form is <bucket>:<value>|<mtype>|@<samplerate>
type metric struct {
	name       string
	bucket     string
	hash       string
	intvalue   int64
	floatvalue float64
	mtype      string
	additive   bool
	samplerate float64
	tags       map[string]string
}

type cachedset struct {
	name string
	set  map[int64]bool
	tags map[string]string
}

type cachedgauge struct {
	name  string
	value float64
	tags  map[string]string
}

type cachedcounter struct {
	name  string
	value int64
	tags  map[string]string
}

type cachedtiming struct {
	name    string
	timings []float64
	tags    map[string]string
}

func (_ *Statsd) Description() string {
	return "Statsd listener"
}

const sampleConfig = `
    # Address and port to host UDP listener on
    service_address = ":8125"
    # Delete gauges every interval
    delete_gauges = false
    # Delete counters every interval
    delete_counters = false
    # Delete sets every interval
    delete_sets = false

    # Number of messages allowed to queue up, once filled,
    # the statsd server will start dropping packets
    allowed_pending_messages = 10000
`

func (_ *Statsd) SampleConfig() string {
	return sampleConfig
}

func (s *Statsd) Gather(acc plugins.Accumulator) error {
	s.Lock()
	defer s.Unlock()

	items := len(s.inmetrics)
	for i := 0; i < items; i++ {

		m := <-s.inmetrics

		switch m.mtype {
		case "c", "g", "s":
			log.Println("ERROR: Uh oh, this should not have happened")
		case "ms", "h":
			// TODO
		}
	}

	for _, cmetric := range s.gauges {
		acc.Add(cmetric.name, cmetric.value, cmetric.tags)
	}
	if s.DeleteGauges {
		s.gauges = make(map[string]cachedgauge)
	}

	for _, cmetric := range s.counters {
		acc.Add(cmetric.name, cmetric.value, cmetric.tags)
	}
	if s.DeleteCounters {
		s.counters = make(map[string]cachedcounter)
	}

	for _, cmetric := range s.sets {
		acc.Add(cmetric.name, int64(len(cmetric.set)), cmetric.tags)
	}
	if s.DeleteSets {
		s.sets = make(map[string]cachedset)
	}

	return nil
}

func (s *Statsd) Start() error {
	log.Println("Starting up the statsd service")

	// Make data structures
	s.done = make(chan struct{})
	s.in = make(chan string, s.AllowedPendingMessages)
	s.inmetrics = make(chan metric, s.AllowedPendingMessages)
	s.gauges = make(map[string]cachedgauge)
	s.counters = make(map[string]cachedcounter)
	s.sets = make(map[string]cachedset)

	// Start the UDP listener
	go s.udpListen()
	// Start the line parser
	go s.parser()
	return nil
}

// udpListen starts listening for udp packets on the configured port.
func (s *Statsd) udpListen() error {
	address, _ := net.ResolveUDPAddr("udp", s.ServiceAddress)
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
	}
	defer listener.Close()
	log.Println("Statsd listener listening on: ", listener.LocalAddr().String())

	for {
		select {
		case <-s.done:
			return nil
		default:
			buf := make([]byte, 1024)
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Printf("ERROR: %s\n", err.Error())
			}

			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					select {
					case s.in <- line:
					default:
						log.Printf(dropwarn, line)
					}
				}
			}
		}
	}
}

// parser monitors the s.in channel, if there is a line ready, it parses the
// statsd string into a usable metric struct and either aggregates the value
// or pushes it into the s.inmetrics channel.
func (s *Statsd) parser() error {
	for {
		select {
		case <-s.done:
			return nil
		case line := <-s.in:
			s.parseStatsdLine(line)
		}
	}
}

// parseStatsdLine will parse the given statsd line, validating it as it goes.
// If the line is valid, it will be cached for the next call to Gather()
func (s *Statsd) parseStatsdLine(line string) error {
	s.Lock()
	defer s.Unlock()

	// Validate splitting the line on "|"
	m := metric{}
	parts1 := strings.Split(line, "|")
	if len(parts1) < 2 {
		log.Printf("Error: splitting '|', Unable to parse metric: %s\n", line)
		return errors.New("Error Parsing statsd line")
	} else if len(parts1) > 2 {
		sr := parts1[2]
		errmsg := "Error: parsing sample rate, %s, it must be in format like: " +
			"@0.1, @0.5, etc. Ignoring sample rate for line: %s\n"
		if strings.Contains(sr, "@") && len(sr) > 1 {
			samplerate, err := strconv.ParseFloat(sr[1:], 64)
			if err != nil {
				log.Printf(errmsg, err.Error(), line)
			} else {
				m.samplerate = samplerate
			}
		} else {
			log.Printf(errmsg, "", line)
		}
	}

	// Validate metric type
	switch parts1[1] {
	case "g", "c", "s", "ms", "h":
		m.mtype = parts1[1]
	default:
		log.Printf("Error: Statsd Metric type %s unsupported", parts1[1])
		return errors.New("Error Parsing statsd line")
	}

	// Validate splitting the rest of the line on ":"
	parts2 := strings.Split(parts1[0], ":")
	if len(parts2) != 2 {
		log.Printf("Error: splitting ':', Unable to parse metric: %s\n", line)
		return errors.New("Error Parsing statsd line")
	}
	m.bucket = parts2[0]

	// Parse the value
	if strings.ContainsAny(parts2[1], "-+") {
		if m.mtype != "g" {
			log.Printf("Error: +- values are only supported for gauges: %s\n", line)
			return errors.New("Error Parsing statsd line")
		}
		m.additive = true
	}

	switch m.mtype {
	case "g", "ms", "h":
		v, err := strconv.ParseFloat(parts2[1], 64)
		if err != nil {
			log.Printf("Error: parsing value to float64: %s\n", line)
			return errors.New("Error Parsing statsd line")
		}
		m.floatvalue = v
	case "c", "s":
		v, err := strconv.ParseInt(parts2[1], 10, 64)
		if err != nil {
			log.Printf("Error: parsing value to int64: %s\n", line)
			return errors.New("Error Parsing statsd line")
		}
		// If a sample rate is given with a counter, divide value by the rate
		if m.samplerate != 0 && m.mtype == "c" {
			v = int64(float64(v) / m.samplerate)
		}
		m.intvalue = v
	}

	// Parse the name
	m.name, m.tags = s.parseName(m)

	// Make a unique key for the measurement name/tags
	var tg []string
	for k, v := range m.tags {
		tg = append(tg, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(tg)
	m.hash = fmt.Sprintf("%s%s", strings.Join(tg, ""), m.name)

	switch m.mtype {
	// Aggregate gauges, counters and sets as we go
	case "g", "c", "s":
		s.aggregate(m)
	// Timers get processed at flush time
	default:
		select {
		case s.inmetrics <- m:
		default:
			log.Printf(dropwarn, line)
		}
	}
	return nil
}

// parseName parses the given bucket name with the list of bucket maps in the
// config file. If there is a match, it will parse the name of the metric and
// map of tags.
// Return values are (<name>, <tags>)
func (s *Statsd) parseName(m metric) (string, map[string]string) {
	name := m.bucket
	tags := make(map[string]string)

	o := graphite.Options{
		Separator: "_",
		Templates: s.Templates,
	}

	p, err := graphite.NewParserWithOptions(o)
	if err == nil {
		name, tags = p.ApplyTemplate(m.bucket)
	}
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "__", -1)

	switch m.mtype {
	case "c":
		tags["metric_type"] = "counter"
	case "g":
		tags["metric_type"] = "gauge"
	case "s":
		tags["metric_type"] = "set"
	case "ms", "h":
		tags["metric_type"] = "timer"
	}

	return name, tags
}

// aggregate takes in a metric of type "counter", "gauge", or "set". It then
// aggregates and caches the current value. It does not deal with the
// DeleteCounters, DeleteGauges or DeleteSets options, because those are dealt
// with in the Gather function.
func (s *Statsd) aggregate(m metric) {
	switch m.mtype {
	case "c":
		cached, ok := s.counters[m.hash]
		if !ok {
			s.counters[m.hash] = cachedcounter{
				name:  m.name,
				value: m.intvalue,
				tags:  m.tags,
			}
		} else {
			cached.value += m.intvalue
			cached.tags = m.tags
			s.counters[m.hash] = cached
		}
	case "g":
		cached, ok := s.gauges[m.hash]
		if !ok {
			s.gauges[m.hash] = cachedgauge{
				name:  m.name,
				value: m.floatvalue,
				tags:  m.tags,
			}
		} else {
			if m.additive {
				cached.value = cached.value + m.floatvalue
			} else {
				cached.value = m.floatvalue
			}
			cached.tags = m.tags
			s.gauges[m.hash] = cached
		}
	case "s":
		cached, ok := s.sets[m.hash]
		if !ok {
			// Completely new metric (initialize with count of 1)
			s.sets[m.hash] = cachedset{
				name: m.name,
				tags: m.tags,
				set:  map[int64]bool{m.intvalue: true},
			}
		} else {
			cached.set[m.intvalue] = true
			s.sets[m.hash] = cached
		}
	}
}

func (s *Statsd) Stop() {
	s.Lock()
	defer s.Unlock()
	log.Println("Stopping the statsd service")
	close(s.done)
	close(s.in)
	close(s.inmetrics)
}

func init() {
	plugins.Add("statsd", func() plugins.Plugin {
		return &Statsd{}
	})
}
