package testutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

type LineParser interface {
	ParseLine(line string) (telegraf.Metric, error)
}

//ParseRawLinesFrom returns the raw lines between the given header and a trailing blank line
func ParseRawLinesFrom(lines []string, header string) ([]string, error) {
	if len(lines) < 2 {
		// We need a line for HEADER and EMPTY TRAILING LINE
		return nil, fmt.Errorf("expected at least two lines to parse from")
	}
	start := -1
	for i := range lines {
		if strings.TrimLeft(lines[i], "# ") == header {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("header %q does not exist", header)
	}

	output := make([]string, 0)
	for _, line := range lines[start:] {
		if !strings.HasPrefix(strings.TrimLeft(line, "\t "), "#") {
			return nil, fmt.Errorf("section does not end with trailing empty line")
		}

		// Stop at empty line
		content := strings.TrimLeft(line, "# \t")
		if content == "" || content == "'''" {
			break
		}

		output = append(output, content)
	}
	return output, nil
}

//ParseMetricsFrom parses metrics from the given lines in line-protocol following a header, with a trailing blank line
func ParseMetricsFrom(lines []string, header string, parser LineParser) ([]telegraf.Metric, error) {
	if len(lines) < 2 {
		// We need a line for HEADER and EMPTY TRAILING LINE
		return nil, fmt.Errorf("expected at least two lines to parse from")
	}
	start := -1
	for i := range lines {
		if strings.TrimLeft(lines[i], "# ") == header {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("header %q does not exist", header)
	}

	metrics := make([]telegraf.Metric, 0)
	for _, line := range lines[start:] {
		if !strings.HasPrefix(strings.TrimLeft(line, "\t "), "#") {
			return nil, fmt.Errorf("section does not end with trailing empty line")
		}

		// Stop at empty line
		content := strings.TrimLeft(line, "# \t")
		if content == "" || content == "'''" {
			break
		}

		m, err := parser.ParseLine(content)
		if err != nil {
			return nil, fmt.Errorf("unable to parse metric in %q failed: %v", content, err)
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func ReadMetricFile(path string) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	expectedFile, err := os.Open(path)
	if err != nil {
		return metrics, err
	}
	defer expectedFile.Close()

	parser := influx.NewParser(influx.NewMetricHandler())
	scanner := bufio.NewScanner(expectedFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			m, err := parser.ParseLine(line)
			// The timezone needs to be UTC to match the timestamp test results
			m.SetTime(m.Time().UTC())
			if err != nil {
				return nil, fmt.Errorf("unable to parse metric in %q failed: %v", line, err)
			}
			metrics = append(metrics, m)
		}
	}
	err = expectedFile.Close()
	if err != nil {
		return metrics, err
	}

	return metrics, nil
}
