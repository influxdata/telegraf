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
	"time"

	"github.com/influxdata/telegraf/plugins/parsers/graphite"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDP_MAX_PACKET_SIZE int = 64 * 1024

	defaultFieldName = "value"

	defaultSeparator = "_"
)

var dropwarn = "ERROR: statsd message queue full. " +
	"We have dropped %d messages so far. " +
	"You may want to increase allowed_pending_messages in the config\n"

var prevInstance *Statsd

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
	ConvertNames   bool

	// MetricSeparator is the separator between parts of the metric name.
	MetricSeparator string
	// This flag enables parsing of tags in the dogstatsd extention to the
	// statsd protocol (http://docs.datadoghq.com/guides/dogstatsd/)
	ParseDataDogTags bool

	// UDPPacketSize is deprecated, it's only here for legacy support
	// we now always create 1 max size buffer and then copy only what we need
	// into the in channel
	// see https://github.com/influxdata/telegraf/pull/992
	UDPPacketSize int `toml:"udp_packet_size"`

	sync.Mutex
	wg sync.WaitGroup
	// drops tracks the number of dropped metrics.
	drops int

	// Channel for all incoming statsd packets
	in   chan []byte
	done chan struct{}

	// Cache gauges, counters & sets so they can be aggregated as they arrive
	// gauges and counters map measurement/tags hash -> field name -> metrics
	// sets and timings map measurement/tags hash -> metrics
	gauges   map[string]cachedgauge
	counters map[string]cachedcounter
	sets     map[string]cachedset
	timings  map[string]cachedtimings

	// bucket -> influx templates
	Templates []string

	listener *net.UDPConn

	graphiteParser *graphite.GraphiteParser
}

// One statsd metric, form is <bucket>:<value>|<mtype>|@<samplerate>
type metric struct {
	name       string
	field      string
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
	name   string
	fields map[string]map[int64]bool
	tags   map[string]string
}

type cachedgauge struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

type cachedcounter struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

type cachedtimings struct {
	name   string
	fields map[string]RunningStats
	tags   map[string]string
}

func (_ *Statsd) Description() string {
	return "Statsd Server"
}

const sampleConfig = `
  ## Address and port to host UDP listener on
  service_address = ":8125"
  ## Delete gauges every interval (default=false)
  delete_gauges = false
  ## Delete counters every interval (default=false)
  delete_counters = false
  ## Delete sets every interval (default=false)
  delete_sets = false
  ## Delete timings & histograms every interval (default=true)
  delete_timings = true
  ## Percentiles to calculate for timing & histogram stats
  percentiles = [90]

  ## separator to use between elements of a statsd metric
  metric_separator = "_"

  ## Parses tags in the datadog statsd format
  ## http://docs.datadoghq.com/guides/dogstatsd/
  parse_data_dog_tags = false

  ## Statsd data translation templates, more info can be read here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#graphite
  # templates = [
  #     "cpu.* measurement*"
  # ]

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets
  allowed_pending_messages = 10000

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000
`

func (_ *Statsd) SampleConfig() string {
	return sampleConfig
}

func (s *Statsd) Gather(acc telegraf.Accumulator) error {
	s.Lock()
	defer s.Unlock()
	now := time.Now()

	for _, metric := range s.timings {
		// Defining a template to parse field names for timers allows us to split
		// out multiple fields per timer. In this case we prefix each stat with the
		// field name and store these all in a single measurement.
		fields := make(map[string]interface{})
		for fieldName, stats := range metric.fields {
			var prefix string
			if fieldName != defaultFieldName {
				prefix = fieldName + "_"
			}
			fields[prefix+"mean"] = stats.Mean()
			fields[prefix+"stddev"] = stats.Stddev()
			fields[prefix+"upper"] = stats.Upper()
			fields[prefix+"lower"] = stats.Lower()
			fields[prefix+"count"] = stats.Count()
			for _, percentile := range s.Percentiles {
				name := fmt.Sprintf("%s%v_percentile", prefix, percentile)
				fields[name] = stats.Percentile(percentile)
			}
		}

		acc.AddFields(metric.name, fields, metric.tags, now)
	}
	if s.DeleteTimings {
		s.timings = make(map[string]cachedtimings)
	}

	for _, metric := range s.gauges {
		acc.AddFields(metric.name, metric.fields, metric.tags, now)
	}
	if s.DeleteGauges {
		s.gauges = make(map[string]cachedgauge)
	}

	for _, metric := range s.counters {
		acc.AddFields(metric.name, metric.fields, metric.tags, now)
	}
	if s.DeleteCounters {
		s.counters = make(map[string]cachedcounter)
	}

	for _, metric := range s.sets {
		fields := make(map[string]interface{})
		for field, set := range metric.fields {
			fields[field] = int64(len(set))
		}
		acc.AddFields(metric.name, fields, metric.tags, now)
	}
	if s.DeleteSets {
		s.sets = make(map[string]cachedset)
	}

	return nil
}

func (s *Statsd) Start(_ telegraf.Accumulator) error {
	// Make data structures
	s.done = make(chan struct{})
	s.in = make(chan []byte, s.AllowedPendingMessages)

	if prevInstance == nil {
		s.gauges = make(map[string]cachedgauge)
		s.counters = make(map[string]cachedcounter)
		s.sets = make(map[string]cachedset)
		s.timings = make(map[string]cachedtimings)
	} else {
		s.gauges = prevInstance.gauges
		s.counters = prevInstance.counters
		s.sets = prevInstance.sets
		s.timings = prevInstance.timings
	}

	if s.ConvertNames {
		log.Printf("WARNING statsd: convert_names config option is deprecated," +
			" please use metric_separator instead")
	}

	if s.MetricSeparator == "" {
		s.MetricSeparator = defaultSeparator
	}

	s.wg.Add(2)
	// Start the UDP listener
	go s.udpListen()
	// Start the line parser
	go s.parser()
	log.Printf("Started the statsd service on %s\n", s.ServiceAddress)
	prevInstance = s
	return nil
}

// udpListen starts listening for udp packets on the configured port.
func (s *Statsd) udpListen() error {
	defer s.wg.Done()
	var err error
	address, _ := net.ResolveUDPAddr("udp", s.ServiceAddress)
	s.listener, err = net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
	}
	log.Println("Statsd listener listening on: ", s.listener.LocalAddr().String())

	buf := make([]byte, UDP_MAX_PACKET_SIZE)
	for {
		select {
		case <-s.done:
			return nil
		default:
			n, _, err := s.listener.ReadFromUDP(buf)
			if err != nil && !strings.Contains(err.Error(), "closed network") {
				log.Printf("ERROR READ: %s\n", err.Error())
				continue
			}
			bufCopy := make([]byte, n)
			copy(bufCopy, buf[:n])

			select {
			case s.in <- bufCopy:
			default:
				s.drops++
				if s.drops == 1 || s.drops%s.AllowedPendingMessages == 0 {
					log.Printf(dropwarn, s.drops)
				}
			}
		}
	}
}

// parser monitors the s.in channel, if there is a packet ready, it parses the
// packet into statsd strings and then calls parseStatsdLine, which parses a
// single statsd metric into a struct.
func (s *Statsd) parser() error {
	defer s.wg.Done()
	var packet []byte
	for {
		select {
		case <-s.done:
			return nil
		case packet = <-s.in:
			lines := strings.Split(string(packet), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					s.parseStatsdLine(line)
				}
			}
		}
	}
}

// parseStatsdLine will parse the given statsd line, validating it as it goes.
// If the line is valid, it will be cached for the next call to Gather()
func (s *Statsd) parseStatsdLine(line string) error {
	s.Lock()
	defer s.Unlock()

	lineTags := make(map[string]string)
	if s.ParseDataDogTags {
		recombinedSegments := make([]string, 0)
		// datadog tags look like this:
		// users.online:1|c|@0.5|#country:china,environment:production
		// users.online:1|c|#sometagwithnovalue
		// we will split on the pipe and remove any elements that are datadog
		// tags, parse them, and rebuild the line sans the datadog tags
		pipesplit := strings.Split(line, "|")
		for _, segment := range pipesplit {
			if len(segment) > 0 && segment[0] == '#' {
				// we have ourselves a tag; they are comma separated
				tagstr := segment[1:]
				tags := strings.Split(tagstr, ",")
				for _, tag := range tags {
					ts := strings.Split(tag, ":")
					var k, v string
					switch len(ts) {
					case 1:
						// just a tag
						k = ts[0]
						v = ""
					case 2:
						k = ts[0]
						v = ts[1]
					}
					if k != "" {
						lineTags[k] = v
					}
				}
			} else {
				recombinedSegments = append(recombinedSegments, segment)
			}
		}
		line = strings.Join(recombinedSegments, "|")
	}

	// Validate splitting the line on ":"
	bits := strings.Split(line, ":")
	if len(bits) < 2 {
		log.Printf("Error: splitting ':', Unable to parse metric: %s\n", line)
		return errors.New("Error Parsing statsd line")
	}

	// Extract bucket name from individual metric bits
	bucketName, bits := bits[0], bits[1:]

	// Add a metric for each bit available
	for _, bit := range bits {
		m := metric{}

		m.bucket = bucketName

		// Validate splitting the bit on "|"
		pipesplit := strings.Split(bit, "|")
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

		// Parse the value
		if strings.HasPrefix(pipesplit[0], "-") || strings.HasPrefix(pipesplit[0], "+") {
			if m.mtype != "g" {
				log.Printf("Error: +- values are only supported for gauges: %s\n", line)
				return errors.New("Error Parsing statsd line")
			}
			m.additive = true
		}

		switch m.mtype {
		case "g", "ms", "h":
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				log.Printf("Error: parsing value to float64: %s\n", line)
				return errors.New("Error Parsing statsd line")
			}
			m.floatvalue = v
		case "c", "s":
			var v int64
			v, err := strconv.ParseInt(pipesplit[0], 10, 64)
			if err != nil {
				v2, err2 := strconv.ParseFloat(pipesplit[0], 64)
				if err2 != nil {
					log.Printf("Error: parsing value to int64: %s\n", line)
					return errors.New("Error Parsing statsd line")
				}
				v = int64(v2)
			}
			// If a sample rate is given with a counter, divide value by the rate
			if m.samplerate != 0 && m.mtype == "c" {
				v = int64(float64(v) / m.samplerate)
			}
			m.intvalue = v
		}

		// Parse the name & tags from bucket
		m.name, m.field, m.tags = s.parseName(m.bucket)
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

		if len(lineTags) > 0 {
			for k, v := range lineTags {
				m.tags[k] = v
			}
		}

		// Make a unique key for the measurement name/tags
		var tg []string
		for k, v := range m.tags {
			tg = append(tg, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(tg)
		m.hash = fmt.Sprintf("%s%s", strings.Join(tg, ""), m.name)

		s.aggregate(m)
	}

	return nil
}

// parseName parses the given bucket name with the list of bucket maps in the
// config file. If there is a match, it will parse the name of the metric and
// map of tags.
// Return values are (<name>, <field>, <tags>)
func (s *Statsd) parseName(bucket string) (string, string, map[string]string) {
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

	var field string
	name := bucketparts[0]

	p := s.graphiteParser
	var err error

	if p == nil || s.graphiteParser.Separator != s.MetricSeparator {
		p, err = graphite.NewGraphiteParser(s.MetricSeparator, s.Templates, nil)
		s.graphiteParser = p
	}

	if err == nil {
		p.DefaultTags = tags
		name, tags, field, _ = p.ApplyTemplate(name)
	}

	if s.ConvertNames {
		name = strings.Replace(name, ".", "_", -1)
		name = strings.Replace(name, "-", "__", -1)
	}
	if field == "" {
		field = defaultFieldName
	}

	return name, field, tags
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
		// Check if the measurement exists
		cached, ok := s.timings[m.hash]
		if !ok {
			cached = cachedtimings{
				name:   m.name,
				fields: make(map[string]RunningStats),
				tags:   m.tags,
			}
		}
		// Check if the field exists. If we've not enabled multiple fields per timer
		// this will be the default field name, eg. "value"
		field, ok := cached.fields[m.field]
		if !ok {
			field = RunningStats{
				PercLimit: s.PercentileLimit,
			}
		}
		if m.samplerate > 0 {
			for i := 0; i < int(1.0/m.samplerate); i++ {
				field.AddValue(m.floatvalue)
			}
		} else {
			field.AddValue(m.floatvalue)
		}
		cached.fields[m.field] = field
		s.timings[m.hash] = cached
	case "c":
		// check if the measurement exists
		_, ok := s.counters[m.hash]
		if !ok {
			s.counters[m.hash] = cachedcounter{
				name:   m.name,
				fields: make(map[string]interface{}),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = s.counters[m.hash].fields[m.field]
		if !ok {
			s.counters[m.hash].fields[m.field] = int64(0)
		}
		s.counters[m.hash].fields[m.field] =
			s.counters[m.hash].fields[m.field].(int64) + m.intvalue
	case "g":
		// check if the measurement exists
		_, ok := s.gauges[m.hash]
		if !ok {
			s.gauges[m.hash] = cachedgauge{
				name:   m.name,
				fields: make(map[string]interface{}),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = s.gauges[m.hash].fields[m.field]
		if !ok {
			s.gauges[m.hash].fields[m.field] = float64(0)
		}
		if m.additive {
			s.gauges[m.hash].fields[m.field] =
				s.gauges[m.hash].fields[m.field].(float64) + m.floatvalue
		} else {
			s.gauges[m.hash].fields[m.field] = m.floatvalue
		}
	case "s":
		// check if the measurement exists
		_, ok := s.sets[m.hash]
		if !ok {
			s.sets[m.hash] = cachedset{
				name:   m.name,
				fields: make(map[string]map[int64]bool),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = s.sets[m.hash].fields[m.field]
		if !ok {
			s.sets[m.hash].fields[m.field] = make(map[int64]bool)
		}
		s.sets[m.hash].fields[m.field][m.intvalue] = true
	}
}

func (s *Statsd) Stop() {
	s.Lock()
	defer s.Unlock()
	log.Println("Stopping the statsd service")
	close(s.done)
	s.listener.Close()
	s.wg.Wait()
	close(s.in)
}

func init() {
	inputs.Add("statsd", func() telegraf.Input {
		return &Statsd{
			MetricSeparator: "_",
		}
	})
}
