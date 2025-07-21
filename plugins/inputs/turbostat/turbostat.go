//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && amd64

package turbostat

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/process"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Turbostat struct {
	UseSudo  bool            `toml:"use_sudo,omitempty"`
	Path     string          `toml:"path,omitempty"`
	Interval config.Duration `toml:"interval,omitempty"`
	Log      telegraf.Logger `toml:"-"`

	command []string
	process *process.Process
}

type column struct {
	name   string
	isTag  bool
	isTime bool
}

type tagMap = map[string]string
type fieldMap = map[string]any

func (*Turbostat) SampleConfig() string {
	return sampleConfig
}

// Init initialises the plugin.
func (t *Turbostat) Init() error {
	if t.Path == "" {
		t.Path = "turbostat"
	}

	if t.Interval <= 0 {
		t.Interval = config.Duration(10 * time.Second)
	}

	t.command = t.buildCmd()
	return nil
}

// Start starts the plugin.
func (t *Turbostat) Start(acc telegraf.Accumulator) error {
	var err error
	t.process, err = process.New(t.command, make([]string, 0))
	if err != nil {
		return fmt.Errorf("failed to create process %s: %w", t.command, err)
	}

	t.process.ReadStderrFn = func(r io.Reader) {
		if err := processStderr(r, acc); err != nil {
			acc.AddError(err)
		}
	}
	t.process.ReadStdoutFn = func(r io.Reader) {
		if err := processStdout(r, acc); err != nil {
			acc.AddError(err)
		}
	}
	t.process.StopOnError = false
	t.process.Log = t.Log

	if err = t.process.Start(); err != nil {
		return fmt.Errorf("failed to start process %s: %w", t.command, err)
	}
	return nil
}

// Stop stops the plugin.
func (t *Turbostat) Stop() {
	t.process.Stop()
}

func (*Turbostat) Gather(telegraf.Accumulator) error {
	return nil
}

// buildCmd builds the command line to start Turbostat.
func (t *Turbostat) buildCmd() []string {
	cmd := make([]string, 0)
	if t.UseSudo {
		cmd = append(cmd, "sudo")
	}
	s := int(time.Duration(t.Interval).Seconds())
	cmd = append(cmd, t.Path, "--quiet", "--interval", strconv.Itoa(s), "--show", "all")
	return cmd
}

// processStderr reads error lines from a stream (such as Turbostat stderr)
// and adds them to an accumulator.
func processStderr(r io.Reader, acc telegraf.Accumulator) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		acc.AddError(errors.New(scanner.Text()))
	}
	return scanner.Err()
}

// processStdout reads metrics from a stream (such as Turbostat stdout)
// and adds them to an accumulator. If an error is encountered, the function
// returns it and stops further processing.
func processStdout(r io.Reader, acc telegraf.Accumulator) error {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return scanner.Err()
	}
	header := scanner.Text()
	columns := make([]column, 0)
	for _, s := range strings.Fields(header) {
		columns = append(columns, createColumn(s))
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == header {
			continue
		}
		values := strings.Fields(line)
		err := processValues(columns, values, acc)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

// Creates a metric from columns and values, and adds it to an accumulator.
func processValues(columns []column, values []string, acc telegraf.Accumulator) error {
	if len(values) > len(columns) {
		return fmt.Errorf("too many values: %d columns, %d values", len(columns), len(values))
	}
	tags := make(tagMap, 0)
	fields := make(fieldMap, 0)
	timestamps := make([]time.Time, 0)
	for i, value := range values {
		column := columns[i]
		if column.isTime {
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("unable to parse time: %w", err)
			}
			s, ns := math.Modf(f)
			timestamps = append(timestamps, time.Unix(int64(s), int64(ns*(1e9))))
			continue
		}
		if column.isTag {
			if !isValidTagValue(value) {
				return fmt.Errorf("invalid tag: %s", value)
			}
			tags[column.name] = value
		} else {
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("unable to parse column %q: %w", column.name, err)
			}
			fields[column.name] = v
		}
	}
	if len(fields) == 0 {
		return errors.New("no value for any field")
	}
	acc.AddFields("turbostat", fields, tags, timestamps...)
	return nil
}

// Creates a column struct from a Turbostat column name.
func createColumn(s string) column {
	abbreviations := map[string]string{
		"%":    "percent",
		"+":    "plus",
		"-":    "minus",
		"watt": "power_watt",
		"mhz":  "frequency_mhz",
		"tmp":  "temperature_celsius",
		"thr":  "throttle",
		"avg":  "average",
		"cor":  "core",
		"bzy":  "busy",
		"pkg":  "package",
		"sys":  "system",
		"unc":  "uncore",
		"u":    "uncore",
		"a":    "actual",
		"j":    "energy_joule",
	}
	tags := map[string]struct{}{
		"package": {},
		"node":    {},
		"die":     {},
		"core":    {},
		"cpu":     {},
		"apic":    {},
		"x2apic":  {},
	}
	c := column{}
	// Split the Turbostat column name into tokens.
	tokens := make([]string, 0)
	for _, element := range splitTokens(s) {
		tokens = append(tokens, splitKnownTokens(element)...)
	}
	for i, token := range tokens {
		token = strings.ToLower(token)
		// If a token is an abbreviation, replace it.
		if replacement, ok := abbreviations[token]; ok {
			token = replacement
		}
		tokens[i] = token
	}
	// Build the camel case column name.
	c.name = strings.Join(tokens, "_")
	// Mark the time column	as such.
	if c.name == "time_of_day_seconds" {
		c.isTime = true
	}
	// If the name matches a tag, mark the column as such.
	if _, ok := tags[c.name]; ok {
		c.isTag = true
	}
	return c
}

// Splits a string into tokens. Each token is a contiguous series
// of alphanumeric characters, or the special characters %, +, and -.
func splitTokens(s string) []string {
	re := regexp.MustCompile(`[0-9a-zA-Z]+|\%|\+|\-`)
	return re.FindAllString(s, -1)
}

// Splits an alphanumeric string into tokens, using a list
// of known tokens to determine boundaries.
func splitKnownTokens(s string) []string {
	knownTokens := []string{
		"Watt",
		"MHz",
		"Tmp",
		"Thr",
		"GFX",
		"SAM",
	}
	tokens := make([]string, 0)
	i := 0
	start := i
	for i < len(s) {
		match := false
		for _, hint := range knownTokens {
			if strings.HasPrefix(s[i:], hint) {
				match = true
				if start < i {
					tokens = append(tokens, s[start:i])
				}
				tokens = append(tokens, hint)
				i += len(hint)
				start = i
				break
			}
		}
		if !match {
			i++
		}
	}
	if start < i {
		tokens = append(tokens, s[start:i])
	}
	return tokens
}

// Returns whether a string represents a tag value or not.
// Turbostat only uses integers and "-".
func isValidTagValue(s string) bool {
	if s == "-" {
		return true
	}
	for _, c := range s {
		if !unicode.IsNumber(c) {
			return false
		}
	}
	return true
}

func init() {
	inputs.Add("turbostat", func() telegraf.Input {
		return &Turbostat{}
	})
}
