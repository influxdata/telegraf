package testutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/influxdata/telegraf"
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

//ParseMetricsFromFile parses metrics from the given file in line-protocol
func ParseMetricsFromFile(filename string, parser telegraf.Parser) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 || strings.HasPrefix(string(line), "#") {
			continue
		}

		nonutc, err := parser.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("unable to parse metric in %q failed: %v", line, err)
		}
		for _, m := range nonutc {
			// The timezone needs to be UTC to match the timestamp test results
			m.SetTime(m.Time().UTC())
			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}

//ParseLinesFromFile returns the lines of the file as strings
func ParseLinesFromFile(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}

	return lines, nil
}
