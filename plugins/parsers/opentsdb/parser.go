package opentsdb

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// Parser encapsulates a OpenTSDB Parser.
type Parser struct {
	DefaultTags map[string]string `toml:"-"`
	Log         telegraf.Logger   `toml:"-"`
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric

	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		line := scanner.Text()

		// delete LF and CR
		line = strings.TrimRight(line, "\r\n")

		m, err := p.ParseLine(line)
		if err != nil {
			p.Log.Errorf("Error parsing %q as opentsdb: %s", line, err)

			// Don't let one bad line spoil a whole batch. In particular, it may
			// be a valid opentsdb telnet protocol command, like "version", that
			// we don't support.
			continue
		}

		metrics = append(metrics, m)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

// ParseLine performs OpenTSDB parsing of a single line.
func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	// Break into fields ("put", name, timestamp, value, tag1, tag2, ..., tagN).
	fields := strings.Fields(line)
	if len(fields) < 4 || fields[0] != "put" {
		return nil, errors.New("doesn't have required fields")
	}

	// decode the name and tags
	measurement := fields[1]
	tsStr := fields[2]
	valueStr := fields[3]
	tagStrs := fields[4:]

	// Parse value.
	v, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing field %q value failed: %w", measurement, err)
	}

	fieldValues := map[string]interface{}{"value": v}

	// Parse timestamp.
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing field %q time failed: %w", measurement, err)
	}

	var timestamp time.Time
	if ts < 1e12 {
		// second resolution
		timestamp = time.Unix(ts, 0)
	} else {
		// millisecond resolution
		timestamp = time.UnixMilli(ts)
	}

	tags := make(map[string]string, len(p.DefaultTags)+len(tagStrs))
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	for _, tag := range tagStrs {
		tagValue := strings.Split(tag, "=")
		if len(tagValue) != 2 {
			continue
		}

		name := tagValue[0]
		value := tagValue[1]
		if name == "" || value == "" {
			continue
		}
		tags[name] = value
	}

	return metric.New(measurement, tags, fieldValues, timestamp), nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func init() {
	parsers.Add("opentsdb",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{}
		})
}
