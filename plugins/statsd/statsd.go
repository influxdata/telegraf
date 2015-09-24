package statsd

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

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
	gauges   map[string]cachedmetric
	counters map[string]cachedmetric
	sets     map[string]cachedmetric

	Mappings []struct {
		Match  string
		Name   string
		Tagmap map[string]int
	}
}

// One statsd metric, form is <bucket>:<value>|<mtype>|@<samplerate>
type metric struct {
	name       string
	bucket     string
	value      int64
	mtype      string
	additive   bool
	samplerate float64
	tags       map[string]string
}

// cachedmetric is a subset of metric used specifically for storing cached
// gauges and counters, ready for sending to InfluxDB.
type cachedmetric struct {
	value int64
	tags  map[string]string
	set   map[int64]bool
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

	values := make(map[string]int64)
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

	for name, cmetric := range s.gauges {
		acc.Add(name, cmetric.value, cmetric.tags)
	}
	if s.DeleteGauges {
		s.gauges = make(map[string]cachedmetric)
	}

	for name, cmetric := range s.counters {
		acc.Add(name, cmetric.value, cmetric.tags)
	}
	if s.DeleteCounters {
		s.counters = make(map[string]cachedmetric)
	}

	for name, cmetric := range s.sets {
		acc.Add(name, cmetric.value, cmetric.tags)
	}
	if s.DeleteSets {
		s.sets = make(map[string]cachedmetric)
	}

	for name, value := range values {
		acc.Add(name, value, nil)
	}
	return nil
}

func (s *Statsd) Start() error {
	log.Println("Starting up the statsd service")

	// Make data structures
	s.done = make(chan struct{})
	s.in = make(chan string, s.AllowedPendingMessages)
	s.inmetrics = make(chan metric, s.AllowedPendingMessages)
	s.gauges = make(map[string]cachedmetric)
	s.counters = make(map[string]cachedmetric)
	s.sets = make(map[string]cachedmetric)

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
func (s *Statsd) parseStatsdLine(line string) {
	s.Lock()
	defer s.Unlock()

	// Validate splitting the line on "|"
	m := metric{}
	parts1 := strings.Split(line, "|")
	if len(parts1) < 2 {
		log.Printf("Error splitting '|', Unable to parse metric: %s\n", line)
		return
	} else if len(parts1) > 2 {
		sr := parts1[2]
		if strings.Contains(sr, "@") && len(sr) > 1 {
			samplerate, err := strconv.ParseFloat(sr[1:], 64)
			if err != nil {
				log.Printf("Error parsing sample rate: %s\n", err.Error())
			} else {
				m.samplerate = samplerate
			}
		} else {
			msg := "Error parsing sample rate, it must be in format like: " +
				"@0.1, @0.5, etc. Ignoring sample rate for line: %s\n"
			log.Printf(msg, line)
		}
	}

	// Validate metric type
	switch parts1[1] {
	case "g", "c", "s", "ms", "h":
		m.mtype = parts1[1]
	default:
		log.Printf("Statsd Metric type %s unsupported", parts1[1])
		return
	}

	// Validate splitting the rest of the line on ":"
	parts2 := strings.Split(parts1[0], ":")
	if len(parts2) != 2 {
		log.Printf("Error splitting ':', Unable to parse metric: %s\n", line)
		return
	}
	m.bucket = parts2[0]

	// Parse the value
	if strings.ContainsAny(parts2[1], "-+") {
		if m.mtype != "g" {
			log.Printf("Error: +- values are only supported for gauges: %s\n", line)
			return
		}
		m.additive = true
	}
	v, err := strconv.ParseInt(parts2[1], 10, 64)
	if err != nil {
		log.Printf("Error: parsing value to int64: %s\n", line)
		return
	}
	// If a sample rate is given with a counter, divide value by the rate
	if m.samplerate != 0 && m.mtype == "c" {
		v = int64(float64(v) / m.samplerate)
	}
	m.value = v

	// Parse the name
	m.name, m.tags = s.parseName(m)

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
}

// parseName parses the given bucket name with the list of bucket maps in the
// config file. If there is a match, it will parse the name of the metric and
// map of tags.
// Return values are (<name>, <tags>)
func (s *Statsd) parseName(m metric) (string, map[string]string) {
	var tags map[string]string
	name := strings.Replace(m.bucket, ".", "_", -1)
	name = strings.Replace(name, "-", "__", -1)

	for _, bm := range s.Mappings {
		if bucketglob(bm.Match, m.bucket) {
			tags = make(map[string]string)
			bparts := strings.Split(m.bucket, ".")
			for name, index := range bm.Tagmap {
				if index >= len(bparts) {
					log.Printf("ERROR: Index %d out of range for bucket %s\n",
						index, m.bucket)
					continue
				}
				tags[name] = bparts[index]
			}
			if bm.Name != "" {
				name = bm.Name
			}
		}
	}

	switch m.mtype {
	case "c":
		name = name + "_counter"
	case "g":
		name = name + "_gauge"
	case "s":
		name = name + "_set"
	case "ms", "h":
		name = name + "_timer"
	}

	return name, tags
}

func bucketglob(pattern, bucket string) bool {
	pparts := strings.Split(pattern, ".")
	bparts := strings.Split(bucket, ".")
	if len(pparts) != len(bparts) {
		return false
	}

	for i, _ := range pparts {
		if pparts[i] == "*" || pparts[i] == bparts[i] {
			continue
		} else {
			return false
		}
	}
	return true
}

// aggregate takes in a metric of type "counter", "gauge", or "set". It then
// aggregates and caches the current value. It does not deal with the
// DeleteCounters, DeleteGauges or DeleteSets options, because those are dealt
// with in the Gather function.
func (s *Statsd) aggregate(m metric) {
	switch m.mtype {
	case "c":
		cached, ok := s.counters[m.name]
		if !ok {
			s.counters[m.name] = cachedmetric{
				value: m.value,
				tags:  m.tags,
			}
		} else {
			cached.value += m.value
			cached.tags = m.tags
			s.counters[m.name] = cached
		}
	case "g":
		cached, ok := s.gauges[m.name]
		if !ok {
			s.gauges[m.name] = cachedmetric{
				value: m.value,
				tags:  m.tags,
			}
		} else {
			if m.additive {
				cached.value = cached.value + m.value
			} else {
				cached.value = m.value
			}
			cached.tags = m.tags
			s.gauges[m.name] = cached
		}
	case "s":
		cached, ok := s.sets[m.name]
		if !ok {
			// Completely new metric (initialize with count of 1)
			s.sets[m.name] = cachedmetric{
				value: 1,
				tags:  m.tags,
				set:   map[int64]bool{m.value: true},
			}
		} else {
			_, ok := s.sets[m.name].set[m.value]
			if !ok {
				// Metric exists, but value has not been counted
				cached.value += 1
				cached.set[m.value] = true
				s.sets[m.name] = cached
			}
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
