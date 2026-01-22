//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && amd64

package turbostat

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/process"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	// Mapping from abbreviations to full words.
	abbreviations = map[string]string{
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

	// Replacer to add underscores around certain words.
	replacer = strings.NewReplacer(
		"Watt", "_Watt_",
		"MHz", "_MHz_",
		"Tmp", "_Tmp_",
		"Thr", "_Thr_",
		"GFX", "_GFX_",
		"SAM", "_SAM_",
	)

	// Regex to split a string into alphanumeric tokens, percent (%), plus (+), and minus (-).
	splitter = regexp.MustCompile(`[0-9a-zA-Z]+|\%|\+|\-`)
)

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

func (t *Turbostat) Init() error {
	if t.Path == "" {
		t.Path = "turbostat"
	}

	if t.Interval <= 0 {
		t.Interval = config.Duration(10 * time.Second)
	}

	// Build the command line to spawn Turbostat.
	t.command = make([]string, 0, 10)
	if t.UseSudo {
		t.command = append(t.command, "sudo")
	}
	interval := strconv.FormatFloat(time.Duration(t.Interval).Seconds(), 'f', -1, 64)
	t.command = append(t.command, t.Path, "--quiet", "--interval", interval, "--show", "all")

	return nil
}

func (t *Turbostat) Start(acc telegraf.Accumulator) error {
	var err error
	t.process, err = process.New(t.command, nil)
	if err != nil {
		return fmt.Errorf("failed to create process %s: %w", t.command, err)
	}

	// Read error lines from Turbostat stderr and add them to the accumulator.
	t.process.ReadStderrFn = func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			acc.AddError(errors.New(scanner.Text()))
		}
		acc.AddError(scanner.Err())
	}

	// Read metrics from Turbostat stdout and add them to the accumulator.
	// Add any error to the accumulator and continue processing the next lines.
	t.process.ReadStdoutFn = func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		if !scanner.Scan() {
			acc.AddError(scanner.Err())
			return
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
			acc.AddError(processValues(acc, columns, values))
		}
		acc.AddError(scanner.Err())
	}

	t.process.StopOnError = false
	t.process.Log = t.Log

	if err = t.process.Start(); err != nil {
		return fmt.Errorf("failed to start process %s: %w", t.command, err)
	}
	return nil
}

func (t *Turbostat) Stop() {
	t.process.Stop()
}

func (*Turbostat) Gather(telegraf.Accumulator) error {
	return nil
}

// Creates a metric from columns and values, and adds it to an accumulator.
func processValues(acc telegraf.Accumulator, columns []column, values []string) error {
	if len(values) > len(columns) {
		return fmt.Errorf("too many values: %d columns, %d values", len(columns), len(values))
	}
	tags := make(map[string]string, len(values))
	fields := make(map[string]any, len(values))
	timestamp := time.Now()
	for i, value := range values {
		column := columns[i]
		switch {
		case column.isTime:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("unable to parse time: %w", err)
			}
			timestamp = time.Unix(0, int64(f*1e9))
		case column.isTag:
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
	acc.AddFields("turbostat", fields, tags, timestamp)
	return nil
}

// Creates a column struct from a Turbostat column name.
func createColumn(s string) column {
	// Add underscores around words known not to have delimiters.
	s = replacer.Replace(s)
	// Split the column name into lower case tokens.
	tokens := splitter.FindAllString(strings.ToLower(s), -1)
	// Expand abbreviations into full words.
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

func init() {
	inputs.Add("turbostat", func() telegraf.Input {
		return &Turbostat{}
	})
}
