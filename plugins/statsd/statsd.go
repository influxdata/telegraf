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

	// Percentiles specifies the percentiles that will be calculated for timing
	// and histogram stats.
	Percentiles     []int
	PercentileLimit int

	DeleteGauges   bool
	DeleteCounters bool
	DeleteSets     bool
	DeleteTimings  bool

	sync.Mutex

	// Channel for all incoming statsd messages
	in   chan string
	done chan struct{}

	// Cache gauges, counters & sets so they can be aggregated as they arrive
	gauges   map[string]cachedgauge
	counters map[string]cachedcounter
	sets     map[string]cachedset
	timings  map[string]cachedtimings

	// bucket -> influx templates
	Templates []string
}

func NewStatsd() *Statsd {
	s := Statsd{}

	// Make data structures
	s.done = make(chan struct{})
	s.in = make(chan string, s.AllowedPendingMessages)
	s.gauges = make(map[string]cachedgauge)
	s.counters = make(map[string]cachedcounter)
	s.sets = make(map[string]cachedset)
	s.timings = make(map[string]cachedtimings)

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

type cachedtimings struct {
	name  string
	stats RunningStats
	tags  map[string]string
}

func (_ *Statsd) Description() string {
	return "Statsd listener"
}

const sampleConfig = `
    # Address and port to host UDP listener on
    service_address = ":8125"
    # Delete gauges every interval (default=false)
    delete_gauges = false
    # Delete counters every interval (default=false)
    delete_counters = false
    # Delete sets every interval (default=false)
    delete_sets = false
    # Delete timings & histograms every interval (default=true)
    delete_timings = true
    # Percentiles to calculate for timing & histogram stats
    percentiles = [90]

    # templates = [
    #     "cpu.* measurement*"
    # ]

    # Number of UDP messages allowed to queue up, once filled,
    # the statsd server will start dropping packets
    allowed_pending_messages = 10000

    # Number of timing/histogram values to track per-measurement in the
    # calculation of percentiles. Raising this limit increases the accuracy
    # of percentiles but also increases the memory usage and cpu time.
    percentile_limit = 1000
`

func (_ *Statsd) SampleConfig() string {
	return sampleConfig
}

func (s *Statsd) Gather(acc plugins.Accumulator) error {
	s.Lock()
	defer s.Unlock()

	for _, metric := range s.timings {
		acc.Add(metric.name+"_mean", metric.stats.Mean(), metric.tags)
		acc.Add(metric.name+"_stddev", metric.stats.Stddev(), metric.tags)
		acc.Add(metric.name+"_upper", metric.stats.Upper(), metric.tags)
		acc.Add(metric.name+"_lower", metric.stats.Lower(), metric.tags)
		acc.Add(metric.name+"_count", metric.stats.Count(), metric.tags)
		for _, percentile := range s.Percentiles {
			name := fmt.Sprintf("%s_percentile_%v", metric.name, percentile)
			acc.Add(name, metric.stats.Percentile(percentile), metric.tags)
		}
	}
	if s.DeleteTimings {
		s.timings = make(map[string]cachedtimings)
	}

	for _, metric := range s.gauges {
		acc.Add(metric.name, metric.value, metric.tags)
	}
	if s.DeleteGauges {
		s.gauges = make(map[string]cachedgauge)
	}

	for _, metric := range s.counters {
		acc.Add(metric.name, metric.value, metric.tags)
	}
	if s.DeleteCounters {
		s.counters = make(map[string]cachedcounter)
	}

	for _, metric := range s.sets {
		acc.Add(metric.name, int64(len(metric.set)), metric.tags)
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
	s.gauges = make(map[string]cachedgauge)
	s.counters = make(map[string]cachedcounter)
	s.sets = make(map[string]cachedset)
	s.timings = make(map[string]cachedtimings)

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
// statsd string into a usable metric struct and aggregates the value
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

	m := metric{}

	// Validate splitting the line on "|"
	pipesplit := strings.Split(line, "|")
	if len(pipesplit) < 2 {
		log.Printf("Error: splitting '|', Unable to parse metric: %s\n", line)
		return errors.New("Error Parsing statsd line")
	} else if len(pipesplit) > 2 {
		sr := pipesplit[2]
		errmsg := "Error: parsing sample rate, %s, it must be in format like: " +
			"@0.1, @0.5, etc. Ignoring sample rate for line: %s\n"
		if strings.Contains(sr, "@") && len(sr) > 1 {
			samplerate, err := strconv.ParseFloat(sr[1:], 64)
			if err != nil {
				log.Printf(errmsg, err.Error(), line)
			} else {
				// sample rate successfully parsed
				m.samplerate = samplerate
			}
		} else {
			log.Printf(errmsg, "", line)
		}
	}

	// Validate metric type
	switch pipesplit[1] {
	case "g", "c", "s", "ms", "h":
		m.mtype = pipesplit[1]
	default:
		log.Printf("Error: Statsd Metric type %s unsupported", pipesplit[1])
		return errors.New("Error Parsing statsd line")
	}

	// Validate splitting the rest of the line on ":"
	colonsplit := strings.Split(pipesplit[0], ":")
	if len(colonsplit) != 2 {
		log.Printf("Error: splitting ':', Unable to parse metric: %s\n", line)
		return errors.New("Error Parsing statsd line")
	}
	m.bucket = colonsplit[0]

	// Parse the value
	if strings.ContainsAny(colonsplit[1], "-+") {
		if m.mtype != "g" {
			log.Printf("Error: +- values are only supported for gauges: %s\n", line)
			return errors.New("Error Parsing statsd line")
		}
		m.additive = true
	}

	switch m.mtype {
	case "g", "ms", "h":
		v, err := strconv.ParseFloat(colonsplit[1], 64)
		if err != nil {
			log.Printf("Error: parsing value to float64: %s\n", line)
			return errors.New("Error Parsing statsd line")
		}
		m.floatvalue = v
	case "c", "s":
		v, err := strconv.ParseInt(colonsplit[1], 10, 64)
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

	// Parse the name & tags from bucket
	m.name, m.tags = s.parseName(m.bucket)
	switch m.mtype {
	case "c":
		m.tags["metric_type"] = "counter"
	case "g":
		m.tags["metric_type"] = "gauge"
	case "s":
		m.tags["metric_type"] = "set"
	case "ms":
		m.tags["metric_type"] = "timing"
	case "h":
		m.tags["metric_type"] = "histogram"
	}

	// Make a unique key for the measurement name/tags
	var tg []string
	for k, v := range m.tags {
		tg = append(tg, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(tg)
	m.hash = fmt.Sprintf("%s%s", strings.Join(tg, ""), m.name)

	s.aggregate(m)
	return nil
}

// parseName parses the given bucket name with the list of bucket maps in the
// config file. If there is a match, it will parse the name of the metric and
// map of tags.
// Return values are (<name>, <tags>)
func (s *Statsd) parseName(bucket string) (string, map[string]string) {
	tags := make(map[string]string)

	bucketparts := strings.Split(bucket, ",")
	// Parse out any tags in the bucket
	if len(bucketparts) > 1 {
		for _, btag := range bucketparts[1:] {
			k, v := parseKeyValue(btag)
			if k != "" {
				tags[k] = v
			}
		}
	}

	o := graphite.Options{
		Separator:   "_",
		Templates:   s.Templates,
		DefaultTags: tags,
	}

	name := bucketparts[0]
	p, err := graphite.NewParserWithOptions(o)
	if err == nil {
		name, tags = p.ApplyTemplate(name)
	}
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "__", -1)

	return name, tags
}

// Parse the key,value out of a string that looks like "key=value"
func parseKeyValue(keyvalue string) (string, string) {
	var key, val string

	split := strings.Split(keyvalue, "=")
	// Must be exactly 2 to get anything meaningful out of them
	if len(split) == 2 {
		key = split[0]
		val = split[1]
	} else if len(split) == 1 {
		val = split[0]
	}

	return key, val
}

// aggregate takes in a metric. It then
// aggregates and caches the current value(s). It does not deal with the
// Delete* options, because those are dealt with in the Gather function.
func (s *Statsd) aggregate(m metric) {
	switch m.mtype {
	case "ms", "h":
		cached, ok := s.timings[m.hash]
		if !ok {
			cached = cachedtimings{
				name: m.name,
				tags: m.tags,
				stats: RunningStats{
					PercLimit: s.PercentileLimit,
				},
			}
		}

		if m.samplerate > 0 {
			for i := 0; i < int(1.0/m.samplerate); i++ {
				cached.stats.AddValue(m.floatvalue)
			}
			s.timings[m.hash] = cached
		} else {
			cached.stats.AddValue(m.floatvalue)
			s.timings[m.hash] = cached
		}
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
}

func init() {
	plugins.Add("statsd", func() plugins.Plugin {
		return &Statsd{}
	})
}
