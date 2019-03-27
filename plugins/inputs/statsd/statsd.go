package statsd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/graphite"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	// UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDP_MAX_PACKET_SIZE int = 64 * 1024

	defaultFieldName = "value"

	defaultProtocol = "udp"

	defaultSeparator           = "_"
	defaultAllowPendingMessage = 10000
	MaxTCPConnections          = 250
)

var dropwarn = "E! Error: statsd message queue full. " +
	"We have dropped %d messages so far. " +
	"You may want to increase allowed_pending_messages in the config\n"

var malformedwarn = "E! Statsd over TCP has received %d malformed packets" +
	" thus far."

type Statsd struct {
	// Protocol used on listener - udp or tcp
	Protocol string `toml:"protocol"`

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
	// This flag enables parsing of tags in the dogstatsd extension to the
	// statsd protocol (http://docs.datadoghq.com/guides/dogstatsd/)
	ParseDataDogTags bool

	// UDPPacketSize is deprecated, it's only here for legacy support
	// we now always create 1 max size buffer and then copy only what we need
	// into the in channel
	// see https://github.com/influxdata/telegraf/pull/992
	UDPPacketSize int `toml:"udp_packet_size"`

	ReadBufferSize int `toml:"read_buffer_size"`

	sync.Mutex
	// Lock for preventing a data race during resource cleanup
	cleanup sync.Mutex
	wg      sync.WaitGroup
	// accept channel tracks how many active connections there are, if there
	// is an available bool in accept, then we are below the maximum and can
	// accept the connection
	accept chan bool
	// drops tracks the number of dropped metrics.
	drops int
	// malformed tracks the number of malformed packets
	malformed int

	// Channel for all incoming statsd packets
	in   chan *bytes.Buffer
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

	// Protocol listeners
	UDPlistener *net.UDPConn
	TCPlistener *net.TCPListener

	// track current connections so we can close them in Stop()
	conns map[string]*net.TCPConn

	MaxTCPConnections int `toml:"max_tcp_connections"`

	TCPKeepAlive       bool               `toml:"tcp_keep_alive"`
	TCPKeepAlivePeriod *internal.Duration `toml:"tcp_keep_alive_period"`

	graphiteParser *graphite.GraphiteParser

	acc telegraf.Accumulator

	MaxConnections     selfstat.Stat
	CurrentConnections selfstat.Stat
	TotalConnections   selfstat.Stat
	PacketsRecv        selfstat.Stat
	BytesRecv          selfstat.Stat

	// A pool of byte slices to handle parsing
	bufPool sync.Pool
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
	return "Statsd UDP/TCP Server"
}

const sampleConfig = `
  ## Protocol, must be "tcp", "udp", "udp4" or "udp6" (default=udp)
  protocol = "udp"

  ## MaxTCPConnection - applicable when protocol is set to tcp (default=250)
  max_tcp_connections = 250

  ## Enable TCP keep alive probes (default=false)
  tcp_keep_alive = false

  ## Specifies the keep-alive period for an active network connection.
  ## Only applies to TCP sockets and will be ignored if tcp_keep_alive is false.
  ## Defaults to the OS configuration.
  # tcp_keep_alive_period = "2h"

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
  ## https://github.com/influxdata/telegraf/blob/master/docs/TEMPLATE_PATTERN.md
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
			fields[prefix+"sum"] = stats.Sum()
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
		acc.AddGauge(metric.name, metric.fields, metric.tags, now)
	}
	if s.DeleteGauges {
		s.gauges = make(map[string]cachedgauge)
	}

	for _, metric := range s.counters {
		acc.AddCounter(metric.name, metric.fields, metric.tags, now)
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
	s.gauges = make(map[string]cachedgauge)
	s.counters = make(map[string]cachedcounter)
	s.sets = make(map[string]cachedset)
	s.timings = make(map[string]cachedtimings)

	s.Lock()
	defer s.Unlock()
	//
	tags := map[string]string{
		"address": s.ServiceAddress,
	}
	s.MaxConnections = selfstat.Register("statsd", "tcp_max_connections", tags)
	s.MaxConnections.Set(int64(s.MaxTCPConnections))
	s.CurrentConnections = selfstat.Register("statsd", "tcp_current_connections", tags)
	s.TotalConnections = selfstat.Register("statsd", "tcp_total_connections", tags)
	s.PacketsRecv = selfstat.Register("statsd", "tcp_packets_received", tags)
	s.BytesRecv = selfstat.Register("statsd", "tcp_bytes_received", tags)

	s.in = make(chan *bytes.Buffer, s.AllowedPendingMessages)
	s.done = make(chan struct{})
	s.accept = make(chan bool, s.MaxTCPConnections)
	s.conns = make(map[string]*net.TCPConn)
	s.bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	for i := 0; i < s.MaxTCPConnections; i++ {
		s.accept <- true
	}

	if s.ConvertNames {
		log.Printf("I! WARNING statsd: convert_names config option is deprecated," +
			" please use metric_separator instead")
	}

	if s.MetricSeparator == "" {
		s.MetricSeparator = defaultSeparator
	}

	if s.isUDP() {
		address, err := net.ResolveUDPAddr(s.Protocol, s.ServiceAddress)
		if err != nil {
			return err
		}

		conn, err := net.ListenUDP(s.Protocol, address)
		if err != nil {
			return err
		}

		log.Println("I! Statsd UDP listener listening on: ", conn.LocalAddr().String())
		s.UDPlistener = conn

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.udpListen(conn)
		}()
	} else {
		address, err := net.ResolveTCPAddr("tcp", s.ServiceAddress)
		if err != nil {
			return err
		}
		listener, err := net.ListenTCP("tcp", address)
		if err != nil {
			return err
		}

		log.Println("I! TCP Statsd listening on: ", listener.Addr().String())
		s.TCPlistener = listener

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.tcpListen(listener)
		}()
	}

	// Start the line parser
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.parser()
	}()
	log.Printf("I! Started the statsd service on %s\n", s.ServiceAddress)
	return nil
}

// tcpListen() starts listening for udp packets on the configured port.
func (s *Statsd) tcpListen(listener *net.TCPListener) error {
	for {
		select {
		case <-s.done:
			return nil
		default:
			// Accept connection:
			conn, err := listener.AcceptTCP()
			if err != nil {
				return err
			}

			if s.TCPKeepAlive {
				if err = conn.SetKeepAlive(true); err != nil {
					return err
				}

				if s.TCPKeepAlivePeriod != nil {
					if err = conn.SetKeepAlivePeriod(s.TCPKeepAlivePeriod.Duration); err != nil {
						return err
					}
				}
			}

			select {
			case <-s.accept:
				// not over connection limit, handle the connection properly.
				s.wg.Add(1)
				// generate a random id for this TCPConn
				id := internal.RandomString(6)
				s.remember(id, conn)
				go s.handler(conn, id)
			default:
				// We are over the connection limit, refuse & close.
				s.refuser(conn)
			}
		}
	}
}

// udpListen starts listening for udp packets on the configured port.
func (s *Statsd) udpListen(conn *net.UDPConn) error {
	if s.ReadBufferSize > 0 {
		s.UDPlistener.SetReadBuffer(s.ReadBufferSize)
	}

	buf := make([]byte, UDP_MAX_PACKET_SIZE)
	for {
		select {
		case <-s.done:
			return nil
		default:
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil && !strings.Contains(err.Error(), "closed network") {
				log.Printf("E! Error READ: %s\n", err.Error())
				continue
			}
			b := s.bufPool.Get().(*bytes.Buffer)
			b.Reset()
			b.Write(buf[:n])

			select {
			case s.in <- b:
			default:
				s.drops++
				if s.drops == 1 || s.AllowedPendingMessages == 0 || s.drops%s.AllowedPendingMessages == 0 {
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
	for {
		select {
		case <-s.done:
			return nil
		case buf := <-s.in:
			lines := strings.Split(buf.String(), "\n")
			s.bufPool.Put(buf)
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
		log.Printf("E! Error: splitting ':', Unable to parse metric: %s\n", line)
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
			log.Printf("E! Error: splitting '|', Unable to parse metric: %s\n", line)
			return errors.New("Error Parsing statsd line")
		} else if len(pipesplit) > 2 {
			sr := pipesplit[2]
			errmsg := "E! Error: parsing sample rate, %s, it must be in format like: " +
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
			log.Printf("E! Error: Statsd Metric type %s unsupported", pipesplit[1])
			return errors.New("Error Parsing statsd line")
		}

		// Parse the value
		if strings.HasPrefix(pipesplit[0], "-") || strings.HasPrefix(pipesplit[0], "+") {
			if m.mtype != "g" && m.mtype != "c" {
				log.Printf("E! Error: +- values are only supported for gauges & counters: %s\n", line)
				return errors.New("Error Parsing statsd line")
			}
			m.additive = true
		}

		switch m.mtype {
		case "g", "ms", "h":
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				log.Printf("E! Error: parsing value to float64: %s\n", line)
				return errors.New("Error Parsing statsd line")
			}
			m.floatvalue = v
		case "c":
			var v int64
			v, err := strconv.ParseInt(pipesplit[0], 10, 64)
			if err != nil {
				v2, err2 := strconv.ParseFloat(pipesplit[0], 64)
				if err2 != nil {
					log.Printf("E! Error: parsing value to int64: %s\n", line)
					return errors.New("Error Parsing statsd line")
				}
				v = int64(v2)
			}
			// If a sample rate is given with a counter, divide value by the rate
			if m.samplerate != 0 && m.mtype == "c" {
				v = int64(float64(v) / m.samplerate)
			}
			m.intvalue = v
		case "s":
			m.strvalue = pipesplit[0]
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
			tg = append(tg, k+"="+v)
		}
		sort.Strings(tg)
		tg = append(tg, m.name)
		m.hash = strings.Join(tg, "")

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
				fields: make(map[string]map[string]bool),
				tags:   m.tags,
			}
		}
		// check if the field exists
		_, ok = s.sets[m.hash].fields[m.field]
		if !ok {
			s.sets[m.hash].fields[m.field] = make(map[string]bool)
		}
		s.sets[m.hash].fields[m.field][m.strvalue] = true
	}
}

// handler handles a single TCP Connection
func (s *Statsd) handler(conn *net.TCPConn, id string) {
	s.CurrentConnections.Incr(1)
	s.TotalConnections.Incr(1)
	// connection cleanup function
	defer func() {
		s.wg.Done()
		conn.Close()
		// Add one connection potential back to channel when this one closes
		s.accept <- true
		s.forget(id)
		s.CurrentConnections.Incr(-1)
	}()

	var n int
	scanner := bufio.NewScanner(conn)
	for {
		select {
		case <-s.done:
			return
		default:
			if !scanner.Scan() {
				return
			}
			n = len(scanner.Bytes())
			if n == 0 {
				continue
			}
			s.BytesRecv.Incr(int64(n))
			s.PacketsRecv.Incr(1)

			b := s.bufPool.Get().(*bytes.Buffer)
			b.Reset()
			b.Write(scanner.Bytes())
			b.WriteByte('\n')

			select {
			case s.in <- b:
			default:
				s.drops++
				if s.drops == 1 || s.drops%s.AllowedPendingMessages == 0 {
					log.Printf(dropwarn, s.drops)
				}
			}
		}
	}
}

// refuser refuses a TCP connection
func (s *Statsd) refuser(conn *net.TCPConn) {
	conn.Close()
	log.Printf("I! Refused TCP Connection from %s", conn.RemoteAddr())
	log.Printf("I! WARNING: Maximum TCP Connections reached, you may want to" +
		" adjust max_tcp_connections")
}

// forget a TCP connection
func (s *Statsd) forget(id string) {
	s.cleanup.Lock()
	defer s.cleanup.Unlock()
	delete(s.conns, id)
}

// remember a TCP connection
func (s *Statsd) remember(id string, conn *net.TCPConn) {
	s.cleanup.Lock()
	defer s.cleanup.Unlock()
	s.conns[id] = conn
}

func (s *Statsd) Stop() {
	s.Lock()
	log.Println("I! Stopping the statsd service")
	close(s.done)
	if s.isUDP() {
		s.UDPlistener.Close()
	} else {
		s.TCPlistener.Close()
		// Close all open TCP connections
		//  - get all conns from the s.conns map and put into slice
		//  - this is so the forget() function doesnt conflict with looping
		//    over the s.conns map
		var conns []*net.TCPConn
		s.cleanup.Lock()
		for _, conn := range s.conns {
			conns = append(conns, conn)
		}
		s.cleanup.Unlock()
		for _, conn := range conns {
			conn.Close()
		}
	}
	s.Unlock()

	s.wg.Wait()

	s.Lock()
	close(s.in)
	log.Println("I! Stopped Statsd listener service on ", s.ServiceAddress)
	s.Unlock()
}

// IsUDP returns true if the protocol is UDP, false otherwise.
func (s *Statsd) isUDP() bool {
	return strings.HasPrefix(s.Protocol, "udp")
}

func init() {
	inputs.Add("statsd", func() telegraf.Input {
		return &Statsd{
			Protocol:               defaultProtocol,
			ServiceAddress:         ":8125",
			MaxTCPConnections:      250,
			TCPKeepAlive:           false,
			MetricSeparator:        "_",
			AllowedPendingMessages: defaultAllowPendingMessage,
			DeleteCounters:         true,
			DeleteGauges:           true,
			DeleteSets:             true,
			DeleteTimings:          true,
		}
	})
}
