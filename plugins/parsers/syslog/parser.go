package syslog

import (
	"fmt"
	"time"

	"github.com/goller/telegraf/metric"
	"github.com/influxdata/go-syslog/rfc5424"
	"github.com/influxdata/telegraf"
)

// Parser wraps rfc5424 syslog parser in an interface for telegraf
type Parser struct {
	DefaultTags map[string]string
	p           *rfc5424.Parser
}

// NewParser returns a parser conforming to the telegraf Parser interface
func NewParser() *Parser {
	return &Parser{
		p: rfc5424.NewParser(),
	}
}

// Parse what does this do?
func (s *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	msg, err := s.p.Parse(buf)
	if err != nil {
		return nil, err
	}
	// TODO: translate syslog message to telegraf metrics

	// name

	// tm
	tm := time.Now()
	if msg.Timestamp != nil {
		tm := *msg.Timestamp
	}

	// TAG
	tags := map[string]string{
		"level":    msg.SeverityLevel(),
		"facility": msg.FacilityMessage(),
	}

	if msg.Hostname != nil {
		tags["hostname"] = *msg.Hostname
	}

	if msg.Appname != nil {
		tags["appname"] = *msg.Appname
	}

	// Fields
	fields := map[string]interface{}{
		"version": msg.Version,
	}

	if msg.ProcID != nil {
		fields["procid"] = *msg.ProcID
	}

	if msg.MsgID != nil {
		fields["msgid"] = *msg.MsgID
	}

	if msg.Message != nil {
		fields["message"] = *msg.Message
	}

	if msg.StructuredData != nil {
		for sdid, sdparams := range *msg.StructuredData {
			if len(sdparams) == 0 {
				// TODO: should this just be a bool?
				fields[sdid] = false
				continue
			}
			for name, value := range sdparams {
				// space is not allowed by the grammar within SDID
				fields[sdid+" "+name] = value
			}
		}
	}

	// TODO: what should this name be?
	name := ""
	// TODO: what should we do with the default tags?
	// TODO: should these tags be after parsing? before?
	for k, v := range s.DefaultTags {
		tags[k] = v
	}
	return []telegraf.Metric{
		metric.New(name, tags, fields, tm),
	}, nil
}

// ParseLine will translate a single syslog line into a telegraf.Metric
func (s *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := s.Parse([]byte(line))
	if err != nil {
		return nil, err
	}
	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: value", line)
	}
	return metrics[0], nil
}

// SetDefaultTags adds extra tags to every line parsed
func (s *Parser) SetDefaultTags(tags map[string]string) {
	s.DefaultTags = tags
}
