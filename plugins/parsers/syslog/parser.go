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
	BestEffort  bool
	p           *rfc5424.Parser
	now         func() time.Time
}

// NewParser returns a parser conforming to the telegraf Parser interface
func NewParser(opts ...ParserOpt) *Parser {
	p := &Parser{
		Name: "syslog",
		p:    rfc5424.NewParser(),
		now:  time.Now,
	}
	for _, opt := range opts {
		p = opt(p)
	}
	return p
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

// WithBestEffort tries to recover as much of the syslog message as possible
// when the message has been truncated.
func WithBestEffort() ParserOpt {
	return func(p *Parser) *Parser {
		p.BestEffort = true
		return p
	}
}

func (s *Parser) tags(msg *rfc5424.SyslogMessage) map[string]string {
	ts := map[string]string{}
	if lvl := msg.SeverityLevel(); lvl != nil {
		ts["severity"] = *lvl
	}

	if f := msg.FacilityMessage(); f != nil {
		ts["facility"] = *f
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
				// When SD-ID does not have params we indicate its presence with a bool
				flds[sdid] = true
				continue
			}
			for name, value := range sdparams {
				// Using whitespace as separator since it is not allowed by the grammar within SDID
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
	msg, err := s.p.Parse(buf, &s.BestEffort)
	// In best effort mode the parser returns
	// minimally and partially valid messages
	// also when it detects an error.
	// So there is an error only when parser does not return any message.
	// In standard mode, the parser returns a message without error or a nil message with error.
	if (err != nil && s.BestEffort == false) || (s.BestEffort == true && msg == nil) {
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

// ParseLine will translate a single syslog line into a single telegraf.Metric
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
