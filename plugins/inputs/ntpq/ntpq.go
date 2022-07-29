//go:generate ../../../tools/readme_config_includer/generator
package ntpq

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type elementType int64

const (
	None elementType = iota
	Tag
	FieldInt
	FieldFloat
	FieldDuration
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
	"reach":  FieldInt,
	"poll":   FieldDuration,
	"when":   FieldDuration,
}

type NTPQ struct {
	DNSLookup bool `toml:"dns_lookup"`

	runQ   func() ([]byte, error)
	tagI   map[string]int
	floatI map[string]int
	intI   map[string]int
}

func (*NTPQ) SampleConfig() string {
	return sampleConfig
}

func (n *NTPQ) Gather(acc telegraf.Accumulator) error {
	out, err := n.runQ()
	if err != nil {
		return err
	}

	// Due to problems with a parsing, we have to use regexp expression in order
	// to remove string that starts from '(' and ends with space
	// see: https://github.com/influxdata/telegraf/issues/2386
	reg, err := regexp.Compile(`\s+\([\S]*`)
	if err != nil {
		return err
	}

	var foundHeader bool
	var columns []column
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()

		// if there is an ntpq state prefix, remove it and make it it's own tag
		// see https://github.com/influxdata/telegraf/issues/1161
		var prefix string
		if strings.ContainsAny(string(line[0]), "*#o+x.-") {
			prefix = string(line[0])
			line = strings.TrimLeft(line, "*#o+x.-")
		}
		line = reg.ReplaceAllString(line, "")

		elements := strings.Fields(line)
		if len(elements) < 2 {
			continue
		}

		// If lineCounter == 0, then this is the header line
		if !foundHeader {
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
			foundHeader = true
			continue
		}

		if len(elements) != len(columns) {
			continue
		}

		tags := make(map[string]string)
		fields := make(map[string]interface{})

		if prefix != "" {
			tags["state_prefix"] = prefix
		}

		for i, raw := range elements {
			col := columns[i]

			switch col.etype {
			case None:
				continue
			case Tag:
				tags[col.name] = raw
			case FieldInt:
				value, err := strconv.ParseInt(raw, 10, 64)
				if err != nil {
					acc.AddError(fmt.Errorf("parsing %q (%v) as int failed: %w", col.name, raw, err))
					continue
				}
				fields[col.name] = value
			case FieldFloat:
				value, err := strconv.ParseFloat(raw, 64)
				if err != nil {
					acc.AddError(fmt.Errorf("parsing %q (%v) as float failed: %w", col.name, raw, err))
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
					acc.AddError(fmt.Errorf("parsing %q (%v) as duration failed: %w", col.name, raw, err))
					continue
				}
				fields[col.name] = value * factor
			}
		}

		acc.AddFields("ntpq", fields, tags)
	}

	return nil
}

func (n *NTPQ) runq() ([]byte, error) {
	bin, err := exec.LookPath("ntpq")
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if n.DNSLookup {
		cmd = exec.Command(bin, "-p")
	} else {
		cmd = exec.Command(bin, "-p", "-n")
	}

	return cmd.Output()
}

func newNTPQ() *NTPQ {
	// Mapping of the ntpq tag key to the index in the command output
	tagI := map[string]int{
		"remote":  -1,
		"refid":   -1,
		"stratum": -1,
		"type":    -1,
	}

	// Mapping of float metrics to their index in the command output
	floatI := map[string]int{
		"delay":  -1,
		"offset": -1,
		"jitter": -1,
	}

	// Mapping of int metrics to their index in the command output
	intI := map[string]int{
		"when":  -1,
		"poll":  -1,
		"reach": -1,
	}

	n := &NTPQ{
		tagI:   tagI,
		floatI: floatI,
		intI:   intI,
	}
	n.runQ = n.runq
	return n
}

func init() {
	inputs.Add("ntpq", func() telegraf.Input {
		return newNTPQ()
	})
}
