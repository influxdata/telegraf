//go:generate ../../../tools/readme_config_includer/generator
package ntpq

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"math/bits"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/kballard/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Due to problems with a parsing, we have to use regexp expression in order
// to remove string that starts from '(' and ends with space
// see: https://github.com/influxdata/telegraf/issues/2386
var reBrackets = regexp.MustCompile(`\s+\([\S]*`)

type elementType int64

const (
	None elementType = iota
	Tag
	FieldFloat
	FieldDuration
	FieldIntDecimal
	FieldIntOctal
	FieldIntRatio8
	FieldIntBits
)

type column struct {
	name  string
	etype elementType
}

// Mapping of ntpq header names to tag keys
var tagHeaders = map[string]string{
	"remote": "remote",
	"refid":  "refid",
	"st":     "stratum",
	"t":      "type",
}

// Mapping of fields
var fieldElements = map[string]elementType{
	"delay":  FieldFloat,
	"jitter": FieldFloat,
	"offset": FieldFloat,
	"reach":  FieldIntDecimal,
	"poll":   FieldDuration,
	"when":   FieldDuration,
}

type NTPQ struct {
	DNSLookup   bool     `toml:"dns_lookup" deprecated:"1.24.0;add '-n' to 'options' instead to skip DNS lookup"`
	Options     string   `toml:"options"`
	Servers     []string `toml:"servers"`
	ReachFormat string   `toml:"reach_format"`

	runQ func(string) ([]byte, error)
}

func (*NTPQ) SampleConfig() string {
	return sampleConfig
}

func (n *NTPQ) Init() error {
	if len(n.Servers) == 0 {
		n.Servers = []string{""}
	}

	if n.runQ == nil {
		options, err := shellquote.Split(n.Options)
		if err != nil {
			return fmt.Errorf("splitting options failed: %w", err)
		}
		if !n.DNSLookup {
			if !choice.Contains("-n", options) {
				options = append(options, "-n")
			}
		}

		n.runQ = func(server string) ([]byte, error) {
			bin, err := exec.LookPath("ntpq")
			if err != nil {
				return nil, err
			}

			// Needs to be last argument
			var args []string
			args = append(args, options...)
			if server != "" {
				args = append(args, server)
			}
			cmd := exec.Command(bin, args...)
			return cmd.Output()
		}
	}

	switch n.ReachFormat {
	case "", "octal":
		n.ReachFormat = "octal"
		// Interpret the field as decimal integer returning
		// the raw (octal) representation
		fieldElements["reach"] = FieldIntDecimal
	case "decimal":
		// Interpret the field as octal integer returning
		// decimal number representation
		fieldElements["reach"] = FieldIntOctal
	case "count":
		// Interpret the field as bits set returning
		// the number of bits set
		fieldElements["reach"] = FieldIntBits
	case "ratio":
		// Interpret the field as ratio between the number of
		// bits set and the maximum available bits set (8).
		fieldElements["reach"] = FieldIntRatio8
	default:
		return fmt.Errorf("unknown 'reach_format' %q", n.ReachFormat)
	}

	return nil
}

func (n *NTPQ) Gather(acc telegraf.Accumulator) error {
	for _, server := range n.Servers {
		n.gatherServer(acc, server)
	}
	return nil
}

func (n *NTPQ) gatherServer(acc telegraf.Accumulator, server string) {
	var msgPrefix string
	if server != "" {
		msgPrefix = fmt.Sprintf("[%s] ", server)
	}
	out, err := n.runQ(server)
	if err != nil {
		acc.AddError(fmt.Errorf("%s%w", msgPrefix, err))
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	// Look for the header
	var columns []column
	for scanner.Scan() {
		line := scanner.Text()

		_, elements := processLine(line)
		if len(elements) < 2 {
			continue
		}

		for _, el := range elements {
			// Check if the element is a tag
			if name, isTag := tagHeaders[el]; isTag {
				columns = append(columns, column{
					name:  name,
					etype: Tag,
				})
				continue
			}

			// Add a field
			if etype, isField := fieldElements[el]; isField {
				columns = append(columns, column{
					name:  el,
					etype: etype,
				})
				continue
			}

			// Skip the column if not found
			columns = append(columns, column{etype: None})
		}
		break
	}
	for scanner.Scan() {
		line := scanner.Text()

		prefix, elements := processLine(line)
		if len(elements) != len(columns) {
			continue
		}

		tags := make(map[string]string)
		fields := make(map[string]interface{})

		if prefix != "" {
			tags["state_prefix"] = prefix
		}
		if server != "" {
			tags["source"] = server
		}

		for i, raw := range elements {
			col := columns[i]

			switch col.etype {
			case None:
				continue
			case Tag:
				tags[col.name] = raw
			case FieldFloat:
				value, err := strconv.ParseFloat(raw, 64)
				if err != nil {
					msg := fmt.Sprintf("%sparsing %q (%v) as float failed", msgPrefix, col.name, raw)
					acc.AddError(fmt.Errorf("%s: %w", msg, err))
					continue
				}
				fields[col.name] = value
			case FieldDuration:
				factor := int64(1)
				suffix := raw[len(raw)-1:]
				switch suffix {
				case "d":
					factor = 24 * 3600
				case "h":
					factor = 3600
				case "m":
					factor = 60
				default:
					suffix = ""
				}
				value, err := strconv.ParseInt(strings.TrimSuffix(raw, suffix), 10, 64)
				if err != nil {
					msg := fmt.Sprintf("%sparsing %q (%v) as duration failed", msgPrefix, col.name, raw)
					acc.AddError(fmt.Errorf("%s: %w", msg, err))
					continue
				}
				fields[col.name] = value * factor
			case FieldIntDecimal:
				value, err := strconv.ParseInt(raw, 10, 64)
				if err != nil {
					acc.AddError(fmt.Errorf("parsing %q (%v) as int failed: %w", col.name, raw, err))
					continue
				}
				fields[col.name] = value
			case FieldIntOctal:
				value, err := strconv.ParseInt(raw, 8, 64)
				if err != nil {
					acc.AddError(fmt.Errorf("parsing %q (%v) as int failed: %w", col.name, raw, err))
					continue
				}
				fields[col.name] = value
			case FieldIntBits:
				value, err := strconv.ParseUint(raw, 8, 64)
				if err != nil {
					acc.AddError(fmt.Errorf("parsing %q (%v) as int failed: %w", col.name, raw, err))
					continue
				}
				fields[col.name] = bits.OnesCount64(value)
			case FieldIntRatio8:
				value, err := strconv.ParseUint(raw, 8, 64)
				if err != nil {
					acc.AddError(fmt.Errorf("parsing %q (%v) as int failed: %w", col.name, raw, err))
					continue
				}
				fields[col.name] = float64(bits.OnesCount64(value)) / float64(8)
			}
		}

		acc.AddFields("ntpq", fields, tags)
	}
}

func processLine(line string) (string, []string) {
	// if there is an ntpq state prefix, remove it and make it it's own tag
	// see https://github.com/influxdata/telegraf/issues/1161
	var prefix string
	if strings.ContainsAny(string(line[0]), "*#o+x.-") {
		prefix = string(line[0])
		line = strings.TrimLeft(line, "*#o+x.-")
	}
	line = reBrackets.ReplaceAllString(line, "")

	return prefix, strings.Fields(line)
}

func init() {
	inputs.Add("ntpq", func() telegraf.Input {
		return &NTPQ{
			DNSLookup: true,
			Options:   "-p",
		}
	})
}
