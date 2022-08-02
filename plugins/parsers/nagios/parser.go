package nagios

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// unknownExitCode is the nagios unknown status code
// the exit code should be used if an error occurs or something unexpected happens
const unknownExitCode = 3

// getExitCode get the exit code from an error value which is the result
// of running a command through exec package api.
func getExitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	ee, ok := err.(*exec.ExitError)
	if !ok {
		return unknownExitCode, err
	}

	ws, ok := ee.Sys().(syscall.WaitStatus)
	if !ok {
		return 0, errors.New("expected syscall.WaitStatus")
	}

	return ws.ExitStatus(), nil
}

// AddState adds a state derived from the runErr. Unknown state will be set as fallback.
// If any error occurs, it is guaranteed to be added to the service output.
// An updated slice of metrics will be returned.
func AddState(runErr error, errMessage []byte, metrics []telegraf.Metric) []telegraf.Metric {
	state, exitErr := getExitCode(runErr)
	// This will ensure that in every error case the valid nagios state 'unknown' will be returned.
	// No error needs to be thrown because the output will contain the error information.
	// Description found at 'Plugin Return Codes' https://nagios-plugins.org/doc/guidelines.html
	if exitErr != nil || state < 0 || state > unknownExitCode {
		state = unknownExitCode
	}

	for _, m := range metrics {
		if m.Name() == "nagios_state" {
			m.AddField("state", state)

			if state == unknownExitCode {
				errorMessage := string(errMessage)
				if exitErr != nil && exitErr.Error() != "" {
					errorMessage = exitErr.Error()
				}
				value, ok := m.GetField("service_output")
				if !ok || value == "" {
					// By adding the error message as output, the metric contains all needed information to understand
					// the problem and fix it
					m.AddField("service_output", errorMessage)
				}
			}
			return metrics
		}
	}

	var ts time.Time
	if len(metrics) != 0 {
		ts = metrics[0].Time()
	} else {
		ts = time.Now().UTC()
	}
	f := map[string]interface{}{
		"state": state,
	}
	m := metric.New("nagios_state", nil, f, ts)

	return append(metrics, m)
}

type Parser struct {
	DefaultTags map[string]string `toml:"-"`
	Log         telegraf.Logger   `toml:"-"`

	metricName string
}

// Got from Alignak
// https://github.com/Alignak-monitoring/alignak/blob/develop/alignak/misc/perfdata.py
var (
	perfSplitRegExp = regexp.MustCompile(`([^=]+=\S+)`)
	nagiosRegExp    = regexp.MustCompile(`^([^=]+)=([\d\.\-\+eE]+)([\w\/%]*);?([\d\.\-\+eE:~@]+)?;?([\d\.\-\+eE:~@]+)?;?([\d\.\-\+eE]+)?;?([\d\.\-\+eE]+)?;?\s*`)
)

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	return metrics[0], err
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	ts := time.Now().UTC()

	s := bufio.NewScanner(bytes.NewReader(buf))

	var msg bytes.Buffer
	var longmsg bytes.Buffer

	metrics := make([]telegraf.Metric, 0)

	// Scan the first line.
	if !s.Scan() && s.Err() != nil {
		return nil, s.Err()
	}
	parts := bytes.Split(s.Bytes(), []byte{'|'})
	switch len(parts) {
	case 2:
		ms, err := parsePerfData(string(parts[1]), ts)
		if err != nil {
			p.Log.Errorf("Failed to parse performance data: %s\n", err.Error())
		}
		metrics = append(metrics, ms...)
		fallthrough
	case 1:
		msg.Write(bytes.TrimSpace(parts[0])) //nolint:revive // from buffer.go: "err is always nil"
	default:
		return nil, errors.New("illegal output format")
	}

	// Read long output.
	for s.Scan() {
		if bytes.Contains(s.Bytes(), []byte{'|'}) {
			parts := bytes.Split(s.Bytes(), []byte{'|'})
			if longmsg.Len() != 0 {
				longmsg.WriteByte('\n') //nolint:revive // from buffer.go: "err is always nil"
			}
			longmsg.Write(bytes.TrimSpace(parts[0])) //nolint:revive // from buffer.go: "err is always nil"

			ms, err := parsePerfData(string(parts[1]), ts)
			if err != nil {
				p.Log.Errorf("Failed to parse performance data: %s\n", err.Error())
			}
			metrics = append(metrics, ms...)
			break
		}
		if longmsg.Len() != 0 {
			longmsg.WriteByte('\n') //nolint:revive // from buffer.go: "err is always nil"
		}
		longmsg.Write(bytes.TrimSpace(s.Bytes())) //nolint:revive // from buffer.go: "err is always nil"
	}

	// Parse extra performance data.
	for s.Scan() {
		ms, err := parsePerfData(s.Text(), ts)
		if err != nil {
			p.Log.Errorf("Failed to parse performance data: %s\n", err.Error())
		}
		metrics = append(metrics, ms...)
	}

	if s.Err() != nil {
		p.Log.Debugf("Unexpected io error: %s\n", s.Err())
	}

	// Create nagios state.
	fields := map[string]interface{}{
		"service_output": msg.String(),
	}
	if longmsg.Len() != 0 {
		fields["long_service_output"] = longmsg.String()
	}

	m := metric.New("nagios_state", nil, fields, ts)
	metrics = append(metrics, m)

	return metrics, nil
}

func parsePerfData(perfdatas string, timestamp time.Time) ([]telegraf.Metric, error) {
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
			str := perf[3]
			if str != "" {
				tags["unit"] = str
			}
		}

		fields := make(map[string]interface{})
		if perf[2] == "U" {
			return nil, errors.New("value undetermined")
		}

		f, err := strconv.ParseFloat(perf[2], 64)
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
		m := metric.New("nagios", tags, fields, timestamp)

		// Add Metric
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// from math
const (
	MaxFloat64 = 1.797693134862315708145274237317043567981e+308 // 2**1023 * (2**53 - 1) / 2**52
	MinFloat64 = 4.940656458412465441765687928682213723651e-324 // 1 / 2**(1023 - 1 + 52)
)

var ErrBadThresholdFormat = errors.New("bad threshold format")

// Handles all cases from https://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
func parseThreshold(threshold string) (min float64, max float64, err error) {
	thresh := strings.Split(threshold, ":")
	switch len(thresh) {
	case 1:
		max, err = strconv.ParseFloat(thresh[0], 64)
		if err != nil {
			return 0, 0, ErrBadThresholdFormat
		}

		return 0, max, nil
	case 2:
		if thresh[0] == "~" {
			min = MinFloat64
		} else {
			min, err = strconv.ParseFloat(thresh[0], 64)
			if err != nil {
				min = 0
			}
		}

		if thresh[1] == "" {
			max = MaxFloat64
		} else {
			max, err = strconv.ParseFloat(thresh[1], 64)
			if err != nil {
				return 0, 0, ErrBadThresholdFormat
			}
		}
	default:
		return 0, 0, ErrBadThresholdFormat
	}

	return min, max, err
}

func init() {
	// Register parser
	parsers.Add("nagios",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{metricName: defaultMetricName}
		},
	)
}

// InitFromConfig is a compatibility function to construct the parser the old way
func (p *Parser) InitFromConfig(config *parsers.Config) error {
	p.metricName = config.MetricName
	p.DefaultTags = config.DefaultTags
	return nil
}
