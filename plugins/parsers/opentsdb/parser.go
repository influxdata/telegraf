package opentsdb

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Parser encapsulates a OpenTSDB Parser.
type OpenTSDBParser struct {
	DefaultTags map[string]string
}

func NewOpenTSDBParser() (*OpenTSDBParser, error) {
	p := &OpenTSDBParser{}

	return p, nil
}

func (p *OpenTSDBParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		line := scanner.Text()

		// delete LF and CR
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		m, err := p.ParseLine(line)
		if err != nil {
			return nil, err
		}

		if m != nil {
			metrics = append(metrics, m)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

// ParseLine performs OpenTSDB parsing of a single line.
func (p *OpenTSDBParser) ParseLine(line string) (telegraf.Metric, error) {
	// Break into fields ("put", name, timestamp, value, tag1, tag2, ..., tagN).
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, nil
	}
	if len(fields) < 4 || fields[0] != "put" {
		return nil, fmt.Errorf("received %q which doesn't have required fields", line)
	}

	// decode the name and tags
	measurement := fields[1]
	tsStr := fields[2]
	valueStr := fields[3]
	tagStrs := fields[4:]

	// Parse value.
	v, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf(`field "%s" value: %s`, measurement, err)
	}

	fieldValues := map[string]interface{}{}
	fieldValues["value"] = v

	// Parse timestamp.
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`field "%s" time: %s`, measurement, err)
	}

	var timestamp time.Time
	switch len(tsStr) {
	case 10:
		// second resolution
		timestamp = time.Unix(ts, 0)
	case 13:
		// millisecond resolution
		timestamp = time.Unix(ts/1000, (ts%1000)*1000000)
	default:
		return nil, fmt.Errorf(`field "%s" time: "%s" time must be 10 or 13 chars`, measurement, tsStr)
	}

	// Split name and tags
	tags := make(map[string]string)
	for _, tag := range tagStrs {
		tagValue := strings.Split(tag, "=")
		if len(tagValue) != 2 || len(tagValue[0]) == 0 || len(tagValue[1]) == 0 {
			continue
		}
		tags[tagValue[0]] = tagValue[1]
	}

	// Set the default tags on the point if they are not already set
	for k, v := range p.DefaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	return metric.New(measurement, tags, fieldValues, timestamp), nil
}

func (p *OpenTSDBParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
