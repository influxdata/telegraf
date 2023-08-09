//go:generate ../../../tools/readme_config_includer/generator
package s7comm

import (
	_ "embed"
	"errors"
	"fmt"
	"hash/maphash"
	"log" //nolint:depguard // Required for tracing connection issues
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robinson/gos7"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const maxRequestsPerBatch = 20
const addressRegexp = `^(?P<area>[A-Z]+)(?P<no>[0-9]+)\.(?P<type>[A-Z]+)(?P<start>[0-9]+)(?:\.(?P<extra>.*))?$`

var (
	regexAddr = regexp.MustCompile(addressRegexp)
	// Area mapping taken from https://github.com/robinson/gos7/blob/master/client.go
	areaMap = map[string]int{
		"PE": 0x81, // process inputs
		"PA": 0x82, // process outputs
		"MK": 0x83, // Merkers
		"DB": 0x84, // DB
		"C":  0x1C, // counters
		"T":  0x1D, // timers
	}
	// Word-length mapping taken from https://github.com/robinson/gos7/blob/master/client.go
	wordLenMap = map[string]int{
		"X":  0x01, // Bit
		"B":  0x02, // Byte (8 bit)
		"C":  0x03, // Char (8 bit)
		"S":  0x03, // String (8 bit)
		"W":  0x04, // Word (16 bit)
		"I":  0x05, // Integer (16 bit)
		"DW": 0x06, // Double Word (32 bit)
		"DI": 0x07, // Double integer (32 bit)
		"R":  0x08, // IEEE 754 real (32 bit)
		// see https://support.industry.siemens.com/cs/document/36479/date_and_time-format-for-s7-?dti=0&lc=en-DE
		"DT": 0x0F, // Date and time (7 byte)
	}
)

type metricFieldDefinition struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`
}

type metricDefinition struct {
	Name   string                  `toml:"name"`
	Fields []metricFieldDefinition `toml:"fields"`
	Tags   map[string]string       `toml:"tags"`
}

type converterFunc func([]byte) interface{}

type batch struct {
	items    []gos7.S7DataItem
	mappings []fieldMapping
}

type fieldMapping struct {
	measurement string
	field       string
	tags        map[string]string
	convert     converterFunc
}

// S7comm represents the plugin
type S7comm struct {
	Server          string             `toml:"server"`
	Rack            int                `toml:"rack"`
	Slot            int                `toml:"slot"`
	Timeout         config.Duration    `toml:"timeout"`
	DebugConnection bool               `toml:"debug_connection"`
	Configs         []metricDefinition `toml:"metric"`
	Log             telegraf.Logger    `toml:"-"`

	handler *gos7.TCPClientHandler
	client  gos7.Client
	batches []batch
}

// SampleConfig returns a basic configuration for the plugin
func (*S7comm) SampleConfig() string {
	return sampleConfig
}

// Init checks the config settings and prepares the plugin. It's called
// once by the Telegraf agent after parsing the config settings.
func (s *S7comm) Init() error {
	// Check settings
	if s.Server == "" {
		return errors.New("'server' has to be specified")
	}
	if s.Rack < 0 {
		return errors.New("'rack' has to be specified")
	}
	if s.Slot < 0 {
		return errors.New("'slot' has to be specified")
	}
	if len(s.Configs) == 0 {
		return errors.New("no metric defined")
	}

	// Set default port to 102 if none is given
	var nerr *net.AddrError
	if _, _, err := net.SplitHostPort(s.Server); errors.As(err, &nerr) {
		if !strings.Contains(nerr.Err, "missing port") {
			return errors.New("invalid 'server' address")
		}
		s.Server += ":102"
	}

	// Create the requests
	return s.createRequests()
}

// Start initializes the connection to the remote endpoint
func (s *S7comm) Start(_ telegraf.Accumulator) error {
	// Create handler for the connection
	s.handler = gos7.NewTCPClientHandler(s.Server, s.Rack, s.Slot)
	s.handler.Timeout = time.Duration(s.Timeout)
	if s.DebugConnection {
		s.handler.Logger = log.New(os.Stderr, "D! [inputs.s7comm]", log.LstdFlags)
	}
	if err := s.handler.Connect(); err != nil {
		return fmt.Errorf("connecting to %q failed: %w", s.Server, err)
	}
	s.client = gos7.NewClient(s.handler)

	return nil
}

// Stop disconnects from the remote endpoint and cleans up
func (s *S7comm) Stop() {
	if s.handler != nil {
		s.handler.Close()
	}
}

// Gather collects the data from the device
func (s *S7comm) Gather(acc telegraf.Accumulator) error {
	timestamp := time.Now()
	grouper := metric.NewSeriesGrouper()

	for i, b := range s.batches {
		// Read the batch
		s.Log.Debugf("Reading batch %d...", i+1)
		if err := s.client.AGReadMulti(b.items, len(b.items)); err != nil {
			return fmt.Errorf("reading batch %d failed: %w", i+1, err)
		}

		// Dissect the received data into fields
		for j, m := range b.mappings {
			// Convert the data
			buf := b.items[j].Data
			value := m.convert(buf)
			s.Log.Debugf("  got %v for field %q @ %d --> %v (%T)", buf, m.field, b.items[j].Start, value, value)

			// Group the data by series
			grouper.Add(m.measurement, m.tags, timestamp, m.field, value)
		}
	}

	// Add the metrics grouped by series to the accumulator
	for _, x := range grouper.Metrics() {
		acc.AddMetric(x)
	}

	return nil
}

// Internal functions
func (s *S7comm) createRequests() error {
	seed := maphash.MakeSeed()
	seenFields := make(map[uint64]bool)
	s.batches = make([]batch, 0)

	current := batch{}
	for i, cfg := range s.Configs {
		// Set the defaults
		if cfg.Name == "" {
			cfg.Name = "s7comm"
		}

		// Check the metric definitions
		if len(cfg.Fields) == 0 {
			return fmt.Errorf("no fields defined for metric %q", cfg.Name)
		}

		// Create requests for all fields  and add it to the current slot
		for _, f := range cfg.Fields {
			if f.Name == "" {
				return fmt.Errorf("unnamed field in metric %q", cfg.Name)
			}

			item, cfunc, err := handleFieldAddress(f.Address)
			if err != nil {
				return fmt.Errorf("field %q of metric %q: %w", f.Name, cfg.Name, err)
			}
			m := fieldMapping{
				measurement: cfg.Name,
				field:       f.Name,
				tags:        s.Configs[i].Tags,
				convert:     cfunc,
			}
			current.items = append(current.items, *item)
			current.mappings = append(current.mappings, m)

			// If the batch is full, start a new one
			if len(current.items) == maxRequestsPerBatch {
				s.batches = append(s.batches, current)
				current = batch{}
			}

			// Check for duplicate field definitions
			id, err := fieldID(seed, cfg, f)
			if err != nil {
				return fmt.Errorf("cannot determine field id for %q: %w", f.Name, err)
			}
			if seenFields[id] {
				return fmt.Errorf("duplicate field definition field %q in metric %q", f.Name, cfg.Name)
			}
			seenFields[id] = true
		}

		// Update the configuration if changed
		s.Configs[i] = cfg
	}

	// Add the last batch if any
	if len(current.items) > 0 {
		s.batches = append(s.batches, current)
	}

	return nil
}

func handleFieldAddress(address string) (*gos7.S7DataItem, converterFunc, error) {
	// Parse the address into the different parts
	if !regexAddr.MatchString(address) {
		return nil, nil, fmt.Errorf("invalid address %q", address)
	}
	names := regexAddr.SubexpNames()[1:]
	parts := regexAddr.FindStringSubmatch(address)[1:]
	if len(names) != len(parts) {
		return nil, nil, fmt.Errorf("names %v do not match parts %v", names, parts)
	}
	groups := make(map[string]string, len(names))
	for i, n := range names {
		groups[n] = parts[i]
	}

	// Check that we do have the required entries in the address
	if _, found := groups["area"]; !found {
		return nil, nil, errors.New("area is missing from address")
	}

	if _, found := groups["no"]; !found {
		return nil, nil, errors.New("area index is missing from address")
	}
	if _, found := groups["type"]; !found {
		return nil, nil, errors.New("type is missing from address")
	}
	if _, found := groups["start"]; !found {
		return nil, nil, errors.New("start address is missing from address")
	}
	dtype := groups["type"]

	// Lookup the item values from names and check the params
	area, found := areaMap[groups["area"]]
	if !found {
		return nil, nil, errors.New("invalid area")
	}
	wordlen, found := wordLenMap[dtype]
	if !found {
		return nil, nil, errors.New("unknown data type")
	}
	areaidx, err := strconv.Atoi(groups["no"])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid area index: %w", err)
	}
	start, err := strconv.Atoi(groups["start"])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid start address: %w", err)
	}

	// Check the amount parameter if any
	var extra int
	switch dtype {
	case "X", "S":
		// We require an extra parameter
		x := groups["extra"]
		if x == "" {
			return nil, nil, errors.New("extra parameter required")
		}

		extra, err = strconv.Atoi(x)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid extra parameter: %w", err)
		}
		if extra < 1 {
			return nil, nil, fmt.Errorf("invalid extra parameter %d", extra)
		}
	default:
		if groups["extra"] != "" {
			return nil, nil, errors.New("extra parameter specified but not used")
		}
	}

	// Get the required buffer size
	amount := 1
	var buflen int
	switch dtype {
	case "X", "B", "C": // 8-bit types
		buflen = 1
	case "W", "I": // 16-bit types
		buflen = 2
	case "DW", "DI", "R": // 32-bit types
		buflen = 4
	case "DT": // 7-byte
		buflen = 7
	case "S":
		amount = extra
		// Extra bytes as the first byte is the max-length of the string and
		// the second byte is the actual length of the string.
		buflen = extra + 2
	default:
		return nil, nil, errors.New("invalid data type")
	}

	// Setup the data item
	item := &gos7.S7DataItem{
		Area:     area,
		WordLen:  wordlen,
		DBNumber: areaidx,
		Start:    start,
		Amount:   amount,
		Data:     make([]byte, buflen),
	}

	// Determine the type converter function
	f := determineConversion(dtype, extra)
	return item, f, nil
}

func fieldID(seed maphash.Seed, def metricDefinition, field metricFieldDefinition) (uint64, error) {
	var mh maphash.Hash
	mh.SetSeed(seed)

	if _, err := mh.WriteString(def.Name); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(field.Name); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}

	// Tags
	for k, v := range def.Tags {
		if _, err := mh.WriteString(k); err != nil {
			return 0, err
		}
		if err := mh.WriteByte('='); err != nil {
			return 0, err
		}
		if _, err := mh.WriteString(v); err != nil {
			return 0, err
		}
		if err := mh.WriteByte(':'); err != nil {
			return 0, err
		}
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}

	return mh.Sum64(), nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("s7comm", func() telegraf.Input {
		return &S7comm{
			Rack:    -1,
			Slot:    -1,
			Timeout: config.Duration(10 * time.Second),
		}
	})
}
