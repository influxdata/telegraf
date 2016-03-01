package ltsv

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

type LTSVParser struct {
	MetricName                       string
	TimeLabel                        string
	TimeFormat                       string
	StrFieldLabels                   []string
	IntFieldLabels                   []string
	FloatFieldLabels                 []string
	BoolFieldLabels                  []string
	TagLabels                        []string
	DefaultTags                      map[string]string
	DuplicatePointsModifierMethod    string
	DuplicatePointsIncrementDuration time.Duration
	DuplicatePointsModifierUniqTag   string

	initialized      bool
	fieldLabelSet    map[string]string
	tagLabelSet      map[string]bool
	dupPointModifier DuplicatePointModifier
	buf              bytes.Buffer
}

func (p *LTSVParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	if buf == nil {
		if p.buf.Len() > 0 {
			metric, err := p.ParseLine(p.buf.String())
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, metric)
		}
	} else {
		for {
			i := bytes.IndexByte(buf, byte('\n'))
			if i == -1 {
				p.buf.Write(buf)
				break
			}

			p.buf.Write(buf[:i])
			if p.buf.Len() > 0 {
				metric, err := p.ParseLine(p.buf.String())
				if err != nil {
					return nil, err
				}
				metrics = append(metrics, metric)
				p.buf.Reset()
			}
			buf = buf[i+1:]
		}
	}
	return metrics, nil
}

func (p *LTSVParser) ParseLine(line string) (telegraf.Metric, error) {
	if !p.initialized {
		err := p.initialize()
		if err != nil {
			return nil, err
		}
	}

	var t time.Time
	timeLabelFound := false
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	for k, v := range p.DefaultTags {
		tags[k] = v
	}
	terms := strings.Split(line, "\t")
	for _, term := range terms {
		kv := strings.SplitN(term, ":", 2)
		k := kv[0]
		if k == p.TimeLabel {
			timeLabelFound = true
			var err error
			t, err = time.Parse(p.TimeFormat, kv[1])
			if err != nil {
				return nil, err
			}
		} else if typ, ok := p.fieldLabelSet[k]; ok {
			switch typ {
			case "string":
				fields[k] = kv[1]
			case "int":
				val, err := strconv.ParseInt(kv[1], 10, 64)
				if err != nil {
					return nil, err
				}
				fields[k] = val
			case "float":
				val, err := strconv.ParseFloat(kv[1], 64)
				if err != nil {
					return nil, err
				}
				fields[k] = val
			case "boolean":
				val, err := strconv.ParseBool(kv[1])
				if err != nil {
					return nil, err
				}
				fields[k] = val
			}
		} else if _, ok := p.tagLabelSet[k]; ok {
			tags[k] = kv[1]
		}
	}
	if !timeLabelFound {
		t = time.Now().UTC()
	}
	p.dupPointModifier.Modify(&t, tags)
	return telegraf.NewMetric(p.MetricName, tags, fields, t)
}

func (p *LTSVParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *LTSVParser) initialize() error {
	p.fieldLabelSet = newFieldLabelSet(p.StrFieldLabels, p.IntFieldLabels, p.FloatFieldLabels, p.BoolFieldLabels)
	p.tagLabelSet = newTagLabelSet(p.TagLabels)
	dupPointModifier, err := newDupPointModifier(
		p.DuplicatePointsModifierMethod,
		p.DuplicatePointsIncrementDuration,
		p.DuplicatePointsModifierUniqTag)
	if err != nil {
		return err
	}
	p.dupPointModifier = dupPointModifier
	p.initialized = true
	return nil
}

func newFieldLabelSet(strFieldLabels, intFieldLabels, floatFieldLabels, boolFieldLabels []string) map[string]string {
	s := make(map[string]string)
	for _, label := range strFieldLabels {
		s[label] = "string"
	}
	for _, label := range intFieldLabels {
		s[label] = "int"
	}
	for _, label := range floatFieldLabels {
		s[label] = "float"
	}
	for _, label := range boolFieldLabels {
		s[label] = "boolean"
	}
	return s
}

func newTagLabelSet(labels []string) map[string]bool {
	s := make(map[string]bool)
	for _, label := range labels {
		s[label] = true
	}
	return s
}

type DuplicatePointModifier interface {
	Modify(t *time.Time, tags map[string]string)
}

func newDupPointModifier(method string, incrementDuration time.Duration, uniqTagName string) (DuplicatePointModifier, error) {
	switch method {
	case "add_uniq_tag":
		return &AddTagDupPointModifier{UniqTagName: uniqTagName}, nil
	case "increment_time":
		return &IncTimeDupPointModifier{IncrementDuration: incrementDuration}, nil
	case "no_op":
		return &NoOpDupPointModifier{}, nil
	default:
		return nil, fmt.Errorf("invalid duplicate_points_modifier_method: %s", method)
	}
}

type AddTagDupPointModifier struct {
	UniqTagName string
	prevTime    time.Time
	dupCount    int64
}

func (m *AddTagDupPointModifier) Modify(t *time.Time, tags map[string]string) {
	if t.Equal(m.prevTime) {
		m.dupCount++
		tags[m.UniqTagName] = strconv.FormatInt(m.dupCount, 10)
	} else {
		m.dupCount = 0
		m.prevTime = *t
	}
}

type IncTimeDupPointModifier struct {
	IncrementDuration time.Duration
	prevTime          time.Time
}

func (m *IncTimeDupPointModifier) Modify(t *time.Time, _ map[string]string) {
	if !t.After(m.prevTime) {
		*t = m.prevTime.Add(m.IncrementDuration)
	}
	m.prevTime = *t
}

type NoOpDupPointModifier struct{}

func (n *NoOpDupPointModifier) Modify(_ *time.Time, _ map[string]string) {}
