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
	UseSudo  bool            `toml:"use_sudo"`
	Path     string          `toml:"path"`
	Interval config.Duration `toml:"interval"`
	Log      telegraf.Logger `toml:"-"`

	command []string
	process *process.Process
}

type column struct {
	name   string
	isTag  bool
	isTime bool
}

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
	cmd := make([]string, 0, 10)
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
	columns := make([]column, 0, 50)
	for _, s := range strings.Fields(header) {
		columns = append(columns, createColumn(s))
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == header {
			continue
		}
		values := strings.Fields(line)
		err := processValues(acc, columns, values)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

// Creates a metric from columns and values, and adds it to an accumulator.
func processValues(acc telegraf.Accumulator, columns []column, values []string) error {
	if len(values) > len(columns) {
		return fmt.Errorf("too many values: %d columns, %d values", len(columns), len(values))
	}
	tags := make(map[string]string, len(values))
	fields := make(map[string]any, len(values))
	timestamps := make([]time.Time, 0, 1)
	for i, value := range values {
		column := columns[i]
		switch {
		case column.isTime:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("unable to parse time: %w", err)
			}
			s, ns := math.Modf(f)
			timestamps = append(timestamps, time.Unix(int64(s), int64(ns*(1e9))))
		case column.isTag:
			if !isValidTagValue(value) {
				return fmt.Errorf("invalid tag: %s", value)
			}
			tags[column.name] = value
		default:
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
	// Add underscores around words known not to have delimiters.
	words := []string{"Watt", "MHz", "Tmp", "Thr", "GFX", "SAM"}
	params := make([]string, 0, len(words)<<1)
	for _, token := range words {
		params = append(params, token, "_"+token+"_")
	}
	s = strings.NewReplacer(params...).Replace(s)

	// Split the column name into lower case tokens.
	re := regexp.MustCompile(`[0-9a-zA-Z]+|\%|\+|\-`)
	tokens := re.FindAllString(strings.ToLower(s), -1)

	// Replace abbreviations.
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
	for i, token := range tokens {
		if replacement, found := abbreviations[token]; found {
			tokens[i] = replacement
		}
	}

	// Create and return the column.
	name := strings.Join(tokens, "_")
	switch name {
	case "time_of_day_seconds":
		return column{name: name, isTime: true}
	case "package", "node", "die", "core", "cpu", "apic", "x2apic":
		return column{name: name, isTag: true}
	default:
		return column{name: name}
	}
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
