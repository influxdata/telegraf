package regex_parser

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type REGEXParser struct {
	MetricName    string
	RegexEXPRList map[string][]string
	DefaultTags   map[string]string
}

func (p *REGEXParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	for mName, list := range p.RegexEXPRList {
		matches := make([]*regexp.Regexp, 0)
		fields := make(map[string]interface{}, 0)
		for _, value := range list {
			item, err := regexp.Compile(value)
			if err != nil {
				return nil, err
			}
			matches = append(matches, item)
		}

		for _, s_for := range matches {
			list := s_for.FindAllStringSubmatch(string(buf), -1)
			if list != nil {
				for _, item := range list {
					if len(item) == 3 {
						value, err := strconv.ParseFloat(item[2], 64)
						fName := strings.TrimSpace(item[1])
						if err != nil {
							return nil, fmt.Errorf("Can't parse secound match of regex as Float, %s", err.Error())
						}
						if fields[fName] == nil {
							fields[fName] = value
							continue
						}
						if val, ok := fields[fName].(float64); ok {
							val += value
							fields[fName] = val
						}
					} else {
						fName := strings.TrimSpace(item[0])
						if fields[fName] == nil {
							var v float64 = 1.0
							fields[fName] = v
							continue
						}
						if val, ok := fields[item[0]].(float64); ok {
							val++
							fields[fName] = val
						}
					}
				}
			}
		}
		if len(fields) == 0 {
			return metrics, nil
		}
		metric, err := telegraf.NewMetric(mName, p.DefaultTags, fields, time.Now().UTC())

		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (p *REGEXParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, nil
	}

	return metrics[0], nil
}

func (p *REGEXParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
