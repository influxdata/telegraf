package statsd

import (
	"errors"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/graphite"
)

const (
	// UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDP_MAX_PACKET_SIZE int = ((64 * 1024) - 8) - 20

	defaultFieldName = "value"

	defaultSeparator           = "_"
	defaultAllowPendingMessage = 10000
)

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

	// drops tracks the number of dropped metrics.
	drops int

	// Kill switch
	done chan struct{}

	// Cache gauges, counters & sets so they can be aggregated as they arrive
	// gauges and counters map measurement/tags hash -> field name -> metrics
	// sets and timings map measurement/tags hash -> metrics
	gauges        chan map[string]cachedgauge
	gaugesReset   chan struct{}
	counters      chan map[string]cachedcounter
	countersReset chan struct{}
	sets          chan map[string]cachedset
	setsReset     chan struct{}
	timings       chan map[string]cachedtimings
	timingsReset  chan struct{}

	// bucket -> influx templates
	Templates []string

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
	strvalue   string
	mtype      string
	additive   bool
	samplerate float64
	tags       map[string]string
}

type cachedset struct {
	name   string
	fields map[string]map[string]bool
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

  ## The following configuration options control when telegraf clears it's cache
  ## of previous values. If set to false, then telegraf will only clear it's
  ## cache when the daemon is restarted.
  ## Reset gauges every interval (default=true)
  delete_gauges = true
  ## Reset counters every interval (default=true)
  delete_counters = true
  ## Reset sets every interval (default=true)
  delete_sets = true
  ## Reset timings & histograms every interval (default=true)
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

type accumulator struct {
	measurement string
	fields      map[string]interface{}
	tags        map[string]string
	t           []time.Time
}

func (s *Statsd) Gather(acc telegraf.Accumulator) error {
	now := time.Now()
	log.Printf("D! Statsd.Gather() started")

	// The order of funcs/maps here is important
	var gathers = []func(time.Time) []*accumulator{s.gatherCounters, s.gatherSets, s.gatherGauges, s.gatherTimings}
	a := make([]*accumulator, 0)
	for pos := range gathers {
		accumulatorArray := gathers[pos](now)
		for p := range accumulatorArray {
			if accumulatorArray[p] != nil {
				a = append(a, accumulatorArray[p])
			}
		}
	}
	log.Printf("D! Statsd.Gather() took %fs", time.Since(now).Seconds())

	now = time.Now()
	for _, v := range a {
		acc.AddFields(v.measurement, v.fields, v.tags, now)
	}
	log.Printf("D! Statsd.Gather().Accumulator took %fs", time.Since(now).Seconds())

	return nil
}

type metricChannels struct {
	gaugeMetricChannel, timingsMetricChannel, countersMetricChannel, setsMetricChannel chan *metric
}

func (s *Statsd) Start(_ telegraf.Accumulator) error {
	s.done = make(chan struct{})
	done := make(chan struct{})
	msg := make(chan []byte, s.AllowedPendingMessages)

	if s.ConvertNames {
		log.Printf("I! WARNING statsd: convert_names config option is deprecated, please use metric_separator instead")
	}

	if s.MetricSeparator == "" {
		s.MetricSeparator = defaultSeparator
	}

	// Start metric ephemeral storage
	mC := new(metricChannels)
	s.gaugesReset, mC.gaugeMetricChannel, s.gauges = s.gaugesC()
	s.timingsReset, mC.timingsMetricChannel, s.timings = s.timingsC(s.PercentileLimit)
	s.countersReset, mC.countersMetricChannel, s.counters = s.countersC()
	s.setsReset, mC.setsMetricChannel, s.sets = s.setsC()

	// Start the line parser
	go func() {
		if err := s.parser(
			mC,
			done,
			msg,
		); err != nil {
			log.Printf("E! Parser exit: %s", err)
		}
	}()
	// Start the udpListern parser
	go func() {
		if err := s.udpListen(done, msg); err != nil {
			log.Printf("E! udpListen exit: %s", err)
		}
	}()
	// Prepare for the end of times
	go func() {
		<-s.done
		close(done)
		close(mC.gaugeMetricChannel)
		close(mC.timingsMetricChannel)
		close(mC.countersMetricChannel)
		close(mC.setsMetricChannel)
	}()

	return nil
}

var dropWarn = "E! Error: statsd message queue full. We have dropped %d messages so far. You may want to increase allowed_pending_messages in the config\n"

// udpListen starts listening for udp packets on the configured port.
func (s *Statsd) udpListen(done chan struct{}, msg chan []byte) error {
	address, err := net.ResolveUDPAddr("udp", s.ServiceAddress)
	if err != nil {
		log.Fatalf("E! ERROR: ResolveUDPAddr - %s", err)
	}
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
	}
	defer func() {
		if err = listener.Close(); err != nil {
			log.Printf("E! Error: %s", err)
		}
	}()
	log.Println("I! Statsd listener listening on: ", listener.LocalAddr().String())

	buf := make([]byte, UDP_MAX_PACKET_SIZE)
	for {
		select {
		case <-done:
			return nil
		default:
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Printf("E! Error: READ: %s\n", err)
				continue
			}
			bufCopy := make([]byte, n)
			copy(bufCopy, buf)

			select {
			case msg <- bufCopy:
				s.drops = len(msg)
			default:
				s.drops++
				if s.drops == 1 || s.AllowedPendingMessages == 0 || s.drops%s.AllowedPendingMessages == 0 {
					log.Printf(dropWarn, s.drops)
				}
			}
		}
	}
}

// parser monitors the msg channel, if there is a packet ready, it parses the
// packet into statsd strings and then calls parseStatsdLine, which parses a
// single statsd metric into a struct.
func (s *Statsd) parser(
	mC *metricChannels,
	done chan struct{},
	msg chan []byte,
) error {
	for {
		select {
		case <-done:
			return nil
		case packet := <-msg:
			lines := strings.Split(string(packet), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					s.parseStatsdLine(
						mC,
						line,
					)
				}
			}
		}
	}
}

var (
	errParseStatsdLine = errors.New("E! Error: Parsing statsd line")
	errmsg             = "E! Error: parsing sample rate, %q, it must be in format like: @0.1, @0.5, etc. Ignoring sample rate for line: %s\n"
)

// parseStatsdLine will parse the given statsd line, validating it as it goes.
// If the line is valid, it will be cached for the next call to Gather()
func (s *Statsd) parseStatsdLine(
	mC *metricChannels,
	line string,
) error {
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
					ts := strings.SplitN(tag, ":", 2)
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
		log.Printf("E! Error: splitting ':', Unable to parse metric: %q\n", line)
		return errParseStatsdLine
	}

	// Extract bucket name from individual metric bits
	bucketName, bits := bits[0], bits[1:]

	// Add a metric for each bit available
	for _, bit := range bits {
		m := &metric{}

		m.bucket = bucketName

		// Validate splitting the bit on "|"
		pipesplit := strings.Split(bit, "|")
		l := len(pipesplit)
		switch {
		case l < 2:
			log.Printf("E! Error: splitting '|', Unable to parse metric: %q\n", line)
			return errParseStatsdLine
		case l > 2:
			samplerate := pipesplit[2]
			if strings.Contains(samplerate, "@") && len(samplerate) > 1 {
				sr, err := strconv.ParseFloat(samplerate[1:], 64)
				if err != nil {
					log.Printf(errmsg, err.Error(), line)
				} else {
					// sample rate successfully parsed
					m.samplerate = sr
				}
			} else {
				log.Printf("E! Error: Missing @ or too short, sampleRate: %q line: %q\n", samplerate, line)
			}
		}

		// Parse the value
		if strings.HasPrefix(pipesplit[0], "-") || strings.HasPrefix(pipesplit[0], "+") {
			if pipesplit[1] != "g" && pipesplit[1] != "c" {
				log.Printf("E! Error: +- values are only supported for gauges & counters: %q\n", line)
				return errParseStatsdLine
			}
			m.additive = true
		}

		switch pipesplit[1] {
		case "g":
			m.mtype = pipesplit[1]
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				log.Printf("E! Error: parsing value to float64: %q\n", line)
				return errParseStatsdLine
			}
			m.floatvalue = v
			mC.gaugeMetricChannel <- s.parseNameAndTags(m, lineTags)
		case "ms", "h":
			m.mtype = pipesplit[1]
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				log.Printf("E! Error: parsing value to float64: %q\n", line)
				return errParseStatsdLine
			}
			m.floatvalue = v
			mC.timingsMetricChannel <- s.parseNameAndTags(m, lineTags)
		case "c":
			m.mtype = pipesplit[1]
			var v int64
			v, err := strconv.ParseInt(pipesplit[0], 10, 64)
			if err != nil {
				v2, err2 := strconv.ParseFloat(pipesplit[0], 64)
				if err2 != nil {
					log.Printf("E! Error: parsing value to int64: %q\n", line)
					return errParseStatsdLine
				}
				v = int64(v2)
			}
			// If a sample rate is given with a counter, divide value by the rate
			if m.samplerate != 0 && m.mtype == "c" {
				v = int64(float64(v) / m.samplerate)
			}
			m.intvalue = v
			mC.countersMetricChannel <- s.parseNameAndTags(m, lineTags)
		case "s":
			m.mtype = pipesplit[1]
			m.strvalue = pipesplit[0]
			mC.setsMetricChannel <- s.parseNameAndTags(m, lineTags)
		default:
			log.Printf("E! Error: Statsd Metric type %q unsupported, line: %q", pipesplit[1], line)
			return errParseStatsdLine
		}
	}

	return nil
}

// Parse the name & tags from bucket
func (s *Statsd) parseNameAndTags(m *metric, lineTags map[string]string) *metric {
	m.name, m.field, m.tags = s.parseName(m.bucket)
	if m.mtype == "c" {
		m.tags["metric_type"] = "counter"
	}
	if m.mtype == "g" {
		m.tags["metric_type"] = "gauge"
	}
	if m.mtype == "s" {
		m.tags["metric_type"] = "set"
	}
	if m.mtype == "ms" {
		m.tags["metric_type"] = "timing"
	}
	if m.mtype == "h" {
		m.tags["metric_type"] = "histogram"
	}

	// Data dog tags
	if len(lineTags) > 0 {
		for k, v := range lineTags {
			m.tags[k] = v
		}
	}

	// Make a unique key for the measurement name/tags
	tg := make([]string, len(m.tags))
	var i int
	for k, v := range m.tags {
		tg[i] = strings.Join([]string{k, v}, "=")
		i++
	}
	sort.Strings(tg)
	m.hash = strings.Join([]string{strings.Join(tg, ""), m.name}, "")

	return m
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

func (s *Statsd) Stop() {
	close(s.done)
	log.Println("I! Stopped the statsd service")
}

func init() {
	inputs.Add("statsd", func() telegraf.Input {
		return &Statsd{
			ServiceAddress:         ":8125",
			MetricSeparator:        "_",
			AllowedPendingMessages: defaultAllowPendingMessage,
			DeleteCounters:         true,
			DeleteGauges:           true,
			DeleteSets:             true,
			DeleteTimings:          true,
		}
	})
}
