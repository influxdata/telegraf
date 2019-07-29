package graphite

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal/templating"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Minimum and maximum supported dates for timestamps.
var (
	MinDate = time.Date(1901, 12, 13, 0, 0, 0, 0, time.UTC)
	MaxDate = time.Date(2038, 1, 19, 0, 0, 0, 0, time.UTC)
)

// Parser encapsulates a Graphite Parser.
type GraphiteParser struct {
	Separator      string
	Templates      []string
	DefaultTags    map[string]string
	templateEngine *templating.Engine
}

func (p *GraphiteParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func NewGraphiteParser(
	separator string,
	templates []string,
	defaultTags map[string]string,
) (*GraphiteParser, error) {
	var err error

	if separator == "" {
		separator = DefaultSeparator
	}
	p := &GraphiteParser{
		Separator: separator,
		Templates: templates,
	}

	if defaultTags != nil {
		p.DefaultTags = defaultTags
	}
	defaultTemplate, _ := templating.NewDefaultTemplateWithPattern("measurement*")
	p.templateEngine, err = templating.NewEngine(p.Separator, defaultTemplate, p.Templates)

	if err != nil {
		return p, fmt.Errorf("exec input parser config is error: %s ", err.Error())
	}
	return p, nil
}

func (p *GraphiteParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	// parse even if the buffer begins with a newline
	if len(buf) != 0 && buf[0] == '\n' {
		buf = buf[1:]
	}

	var metrics []telegraf.Metric
	var errs []string

	for {
		n := bytes.IndexByte(buf, '\n')
		var line []byte
		if n >= 0 {
			line = bytes.TrimSpace(buf[:n:n])
		} else {
			line = bytes.TrimSpace(buf) // last line
		}
		if len(line) != 0 {
			metric, err := p.ParseLine(string(line))
			if err == nil {
				metrics = append(metrics, metric)
			} else {
				errs = append(errs, err.Error())
			}
		}
		if n < 0 {
			break
		}
		buf = buf[n+1:]
	}
	if len(errs) != 0 {
		return metrics, errors.New(strings.Join(errs, "\n"))
	}
	return metrics, nil
}

// Parse performs Graphite parsing of a single line.
func (p *GraphiteParser) ParseLine(line string) (telegraf.Metric, error) {
	// Break into 3 fields (name, value, timestamp).
	fields := strings.Fields(line)
	if len(fields) != 2 && len(fields) != 3 {
		return nil, fmt.Errorf("received %q which doesn't have required fields", line)
	}

	// decode the name and tags
	measurement, tags, field, err := p.templateEngine.Apply(fields[0])
	if err != nil {
		return nil, err
	}

	// Could not extract measurement, use the raw value
	if measurement == "" {
		measurement = fields[0]
	}

	// Parse value.
	v, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf(`field "%s" value: %s`, fields[0], err)
	}

	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil, &UnsupposedValueError{Field: fields[0], Value: v}
	}

	fieldValues := map[string]interface{}{}
	if field != "" {
		fieldValues[field] = v
	} else {
		fieldValues["value"] = v
	}

	// If no 3rd field, use now as timestamp
	timestamp := time.Now().UTC()

	if len(fields) == 3 {
		// Parse timestamp.
		unixTime, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return nil, fmt.Errorf(`field "%s" time: %s`, fields[0], err)
		}

		// -1 is a special value that gets converted to current UTC time
		// See https://github.com/graphite-project/carbon/issues/54
		if unixTime != float64(-1) {
			// Check if we have fractional seconds
			timestamp = time.Unix(int64(unixTime), int64((unixTime-math.Floor(unixTime))*float64(time.Second)))
			if timestamp.Before(MinDate) || timestamp.After(MaxDate) {
				return nil, fmt.Errorf("timestamp out of range")
			}
		}
	}
	// Set the default tags on the point if they are not already set
	for k, v := range p.DefaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	return metric.New(measurement, tags, fieldValues, timestamp)
}

// ApplyTemplate extracts the template fields from the given line and
// returns the measurement name and tags.
func (p *GraphiteParser) ApplyTemplate(line string) (string, map[string]string, string, error) {
	// Break line into fields (name, value, timestamp), only name is used
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", make(map[string]string), "", nil
	}
	// decode the name and tags
	name, tags, field, err := p.templateEngine.Apply(fields[0])

	// Set the default tags on the point if they are not already set
	for k, v := range p.DefaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	return name, tags, field, err
}
