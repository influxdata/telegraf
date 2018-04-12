package syslog

import (
	"fmt"
	"time"

	"github.com/influxdata/go-syslog/rfc5424"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Parser wraps rfc5424 syslog parser in an interface for telegraf
type Parser struct {
	DefaultTags map[string]string
	Name        string
	p           *rfc5424.Parser
	now         func() time.Time
}

// NewParser returns a parser conforming to the telegraf Parser interface
func NewParser(opts ...ParserOpt) *Parser {
	return &Parser{
		Name: "syslog",
		p:    rfc5424.NewParser(),
		now:  time.Now,
	}
}

// ParserOpt sets options for the syslog parser
type ParserOpt func(p *Parser) *Parser

// WithName sets the metric output name to name
func WithName(name string) ParserOpt {
	return func(p *Parser) *Parser {
		p.Name = name
		return p
	}
}

func (s *Parser) tags(msg *rfc5424.SyslogMessage) map[string]string {
	ts := map[string]string{
		"severity": msg.SeverityLevel(),
		"facility": msg.FacilityMessage(),
	}

	if msg.Hostname != nil {
		ts["hostname"] = *msg.Hostname
	}

	if msg.Appname != nil {
		ts["appname"] = *msg.Appname
	}

	for k, v := range s.DefaultTags {
		ts[k] = v
	}
	return ts
}

func (s *Parser) fields(msg *rfc5424.SyslogMessage) map[string]interface{} {
	flds := map[string]interface{}{
		"version": int(msg.Version),
	}

	if msg.ProcID != nil {
		flds["procid"] = *msg.ProcID
	}

	if msg.MsgID != nil {
		flds["msgid"] = *msg.MsgID
	}

	if msg.Message != nil {
		flds["message"] = *msg.Message
	}

	if msg.StructuredData != nil {
		for sdid, sdparams := range *msg.StructuredData {
			if len(sdparams) == 0 {
				// TODO: should this just be a bool?
				flds[sdid] = false
				continue
			}
			for name, value := range sdparams {
				// space is not allowed by the grammar within SDID
				flds[sdid+" "+name] = value
			}
		}
	}

	return flds
}

func (s *Parser) tm(msg *rfc5424.SyslogMessage) time.Time {
	t := s.now()
	if msg.Timestamp != nil {
		t = *msg.Timestamp
	}
	return t
}

// Parse converts a single syslog message of bytes into a single telegraf metric
func (s *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	msg, err := s.p.Parse(buf)
	if err != nil {
		return nil, err
	}

	m, err := metric.New(
		s.Name,
		s.tags(msg),
		s.fields(msg),
		s.tm(msg),
	)
	if err != nil {
		return nil, err
	}

	return []telegraf.Metric{
		m,
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
