package nagios

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type NagiosParser struct {
	MetricName  string
	DefaultTags map[string]string
}

// Got from Alignak
// https://github.com/Alignak-monitoring/alignak/blob/develop/alignak/misc/perfdata.py
var (
	perfSplitRegExp = regexp.MustCompile(`([^=]+=\S+)`)
	nagiosRegExp    = regexp.MustCompile(`^([^=]+)=([\d\.\-\+eE]+)([\w\/%]*);?([\d\.\-\+eE:~@]+)?;?([\d\.\-\+eE:~@]+)?;?([\d\.\-\+eE]+)?;?([\d\.\-\+eE]+)?;?\s*`)
)

func (p *NagiosParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	return metrics[0], err
}

func (p *NagiosParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *NagiosParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	lines := strings.Split(strings.TrimSpace(string(buf)), "\n")

	for _, line := range lines {
		data_splitted := strings.Split(line, "|")

		if len(data_splitted) != 2 {
			// got human readable output only or bad line
			continue
		}
		m, err := parsePerfData(data_splitted[1])
		if err != nil {
			log.Printf("E! [parser.nagios] failed to parse performance data: %s\n", err.Error())
			continue
		}
		metrics = append(metrics, m...)
	}
	return metrics, nil
}

func parsePerfData(perfdatas string) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	for _, unParsedPerf := range perfSplitRegExp.FindAllString(perfdatas, -1) {
		trimedPerf := strings.TrimSpace(unParsedPerf)
		perf := nagiosRegExp.FindStringSubmatch(trimedPerf)

		// verify at least `'label'=value[UOM];` existed
		if len(perf) < 3 {
			continue
		}
		if perf[1] == "" || perf[2] == "" {
			continue
		}

		fieldName := strings.Trim(perf[1], "'")
		tags := map[string]string{"perfdata": fieldName}
		if perf[3] != "" {
			str := string(perf[3])
			if str != "" {
				tags["unit"] = str
			}
		}

		fields := make(map[string]interface{})
		if perf[2] == "U" {
			return nil, errors.New("Value undetermined")
		}

		f, err := strconv.ParseFloat(string(perf[2]), 64)
		if err == nil {
			fields["value"] = f
		}
		if perf[4] != "" {
			low, high, err := parseThreshold(perf[4])
			if err == nil {
				if strings.Contains(perf[4], "@") {
					fields["warning_le"] = low
					fields["warning_ge"] = high
				} else {
					fields["warning_lt"] = low
					fields["warning_gt"] = high
				}
			}
		}
		if perf[5] != "" {
			low, high, err := parseThreshold(perf[5])
			if err == nil {
				if strings.Contains(perf[5], "@") {
					fields["critical_le"] = low
					fields["critical_ge"] = high
				} else {
					fields["critical_lt"] = low
					fields["critical_gt"] = high
				}
			}
		}
		if perf[6] != "" {
			f, err := strconv.ParseFloat(perf[6], 64)
			if err == nil {
				fields["min"] = f
			}
		}
		if perf[7] != "" {
			f, err := strconv.ParseFloat(perf[7], 64)
			if err == nil {
				fields["max"] = f
			}
		}

		// Create metric
		metric, err := metric.New("nagios", tags, fields, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		// Add Metric
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// from math
const (
	MaxFloat64 = 1.797693134862315708145274237317043567981e+308 // 2**1023 * (2**53 - 1) / 2**52
	MinFloat64 = 4.940656458412465441765687928682213723651e-324 // 1 / 2**(1023 - 1 + 52)
)

var ErrBadThresholdFormat = errors.New("Bad threshold format")

// Handles all cases from https://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
func parseThreshold(threshold string) (min float64, max float64, err error) {
	thresh := strings.Split(threshold, ":")
	switch len(thresh) {
	case 1:
		max, err = strconv.ParseFloat(string(thresh[0]), 64)
		if err != nil {
			return 0, 0, ErrBadThresholdFormat
		}

		return 0, max, nil
	case 2:
		if thresh[0] == "~" {
			min = MinFloat64
		} else {
			min, err = strconv.ParseFloat(string(thresh[0]), 64)
			if err != nil {
				min = 0
			}
		}

		if thresh[1] == "" {
			max = MaxFloat64
		} else {
			max, err = strconv.ParseFloat(string(thresh[1]), 64)
			if err != nil {
				return 0, 0, ErrBadThresholdFormat
			}
		}
	default:
		return 0, 0, ErrBadThresholdFormat
	}

	return
}
