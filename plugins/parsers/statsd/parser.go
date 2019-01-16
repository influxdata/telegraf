package statsd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/graphite"
)

const (
	defaultFieldName = "value"
	defaultSeparator = "_"
	defaultTagKey    = "metric_type"
)

type Parser struct {
	DefaultTags map[string]string

	Separator      string
	Templates      []string
	graphiteParser *graphite.GraphiteParser
}

func NewParser(separator string, templates []string, defaultTags map[string]string) (*Parser, error) {
	if separator == "" {
		separator = defaultSeparator
	}
	p := &Parser{
		Separator: separator,
		Templates: templates,
	}

	if defaultTags != nil {
		p.DefaultTags = defaultTags
	}

	return p, nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// add newline to end if not exists:
	if len(buf) > 0 && !bytes.HasSuffix(buf, []byte("\n")) {
		buf = append(buf, []byte("\n")...)
	}

	metrics := make([]telegraf.Metric, 0)

	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)
	for {
		buf, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line := strings.TrimSpace(string(buf))
		if line == "" {
			log.Printf("I! Statsd line is empty.\n")
			continue
		}

		newMetrics, err := p.parseLine(line)
		if err != nil {
			log.Printf("E! Error parse statsd line: %s.\n", err)
			continue
		}
		metrics = append(metrics, newMetrics...)
	}

	return metrics, nil
}

// statsd allow merge individual stats into a single line, so we just return first metric
func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.parseLine(line)
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, errors.New("No metric in line")
	}
	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) parseLine(line string) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	// Validate splitting the line on ":"
	bits := strings.Split(line, ":")
	if len(bits) < 2 {
		log.Printf("E! Error: splitting ':', Unable to parse metric: %s\n", line)
		return nil, errors.New("Error Parsing statsd line")
	}

	// Extract bucket name from individual metric bits
	bucketName, bits := bits[0], bits[1:]
	// Parse the name & tags from bucket
	name, _, tags := p.parseName(bucketName)

	// Add a metric for each bit available
	for _, bit := range bits {
		// Validate splitting the bit on "|"
		pipesplit := strings.Split(bit, "|")

		var samplerate float64
		var err error
		if len(pipesplit) < 2 {
			log.Printf("E! Error: splitting '|', Unable to parse metric: %s\n", line)
			continue
		} else if len(pipesplit) > 2 {
			sr := pipesplit[2]
			errmsg := "E! Error: parsing sample rate, %s, it must be in format like: " +
				"@0.1, @0.5, etc. Ignoring sample rate for line: %s\n"
			if strings.Contains(sr, "@") && len(sr) > 1 {
				samplerate, err = strconv.ParseFloat(sr[1:], 64)
				if err != nil {
					log.Printf(errmsg, err.Error(), line)
					continue
				}
			} else {
				log.Printf(errmsg, "", line)
				continue
			}
		}

		var mtype string
		// Validate metric type
		switch pipesplit[1] {
		case "g", "c", "s", "ms", "h":
			mtype = pipesplit[1]
		default:
			log.Printf("E! Error: Statsd Metric type %s unsupported\n", pipesplit[1])
			return nil, errors.New("Error Parsing statsd line")
		}

		v, newTags, err := parseField(mtype, pipesplit[0])
		if err != nil {
			return nil, err
		}
		for k, v := range newTags {
			tags[k] = v
		}

		// samplerate with counter
		if samplerate > 0 && mtype == "c" {
			v = int64(float64(v.(int64)) / samplerate)
		}

		fields := map[string]interface{}{
			defaultFieldName: v,
		}
		switch mtype {
		case "c":
			tags[defaultTagKey] = "counter"
		case "g":
			tags[defaultTagKey] = "gauge"
		case "s":
			tags[defaultTagKey] = "set"
		case "ms":
			tags[defaultTagKey] = "timing"
		case "h":
			tags[defaultTagKey] = "histogram"
		}

		m, err := metric.New(name, tags, fields, time.Now())
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)

		// samplerate with timing
		if samplerate > 0 && (mtype == "ms" || mtype == "h") {
			size := int(1.0/samplerate) - 1
			if size > 0 {
				sampleMetrics := make([]telegraf.Metric, size)
				for i := 0; i < size; i++ {
					sampleMetrics[i], err = metric.New(name, tags, fields, time.Now())
					if err != nil {
						return nil, err
					}
				}
				metrics = append(metrics, sampleMetrics...)
			}
		}
	}
	return metrics, nil
}

func (p *Parser) parseName(bucket string) (string, string, map[string]string) {
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

	graphiteParser := p.graphiteParser
	var err error

	if graphiteParser == nil {
		graphiteParser, err = graphite.NewGraphiteParser(
			p.Separator, p.Templates, nil)
		p.graphiteParser = graphiteParser
	}

	if err == nil {
		graphiteParser.DefaultTags = tags
		name, tags, field, _ = graphiteParser.ApplyTemplate(name)
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
		// why?
		val = split[0]
	}

	return key, val
}

func parseField(mtype string, value string) (interface{}, map[string]string, error) {
	switch mtype {
	case "ms", "h":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, nil, err
		}
		return v, nil, nil
	case "c":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			v2, err2 := strconv.ParseFloat(value, 64)
			if err2 != nil {
				return nil, nil, errors.New("Error Parsing statsd line")
			}
			v = int64(v2)
		}
		return v, nil, nil
	case "g":
		newTags := map[string]string{}
		if strings.HasPrefix(value, "-") || strings.HasPrefix(value, "+") {
			newTags["operation"] = "additive"
		}

		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, nil, errors.New("Error Parsing statsd line")
		}
		return v, newTags, nil
	case "s":
		// s & g should be dealed by [aggregator statsd]
		v := value
		return v, nil, nil
	default:
		return nil, nil, errors.New("Unexpected type of statsd line")
	}
}
