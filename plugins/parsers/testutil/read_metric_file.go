package testutil

import (
	"bufio"
	"fmt"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

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
