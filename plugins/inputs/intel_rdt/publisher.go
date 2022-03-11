//go:build !windows
// +build !windows

package intel_rdt

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

type parsedCoresMeasurement struct {
	cores  string
	values []float64
	time   time.Time
}

type parsedProcessMeasurement struct {
	process string
	cores   string
	values  []float64
	time    time.Time
}

// Publisher for publish new RDT metrics to telegraf accumulator
type Publisher struct {
	acc               telegraf.Accumulator
	Log               telegraf.Logger
	shortenedMetrics  bool
	BufferChanProcess chan processMeasurement
	BufferChanCores   chan string
	errChan           chan error
}

func NewPublisher(acc telegraf.Accumulator, log telegraf.Logger, shortenedMetrics bool) Publisher {
	return Publisher{
		acc:               acc,
		Log:               log,
		shortenedMetrics:  shortenedMetrics,
		BufferChanProcess: make(chan processMeasurement),
		BufferChanCores:   make(chan string),
		errChan:           make(chan error),
	}
}

func (p *Publisher) publish(ctx context.Context) {
	go func() {
		for {
			select {
			case newMeasurements := <-p.BufferChanCores:
				p.publishCores(newMeasurements)
			case newMeasurements := <-p.BufferChanProcess:
				p.publishProcess(newMeasurements)
			case err := <-p.errChan:
				p.Log.Error(err)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *Publisher) publishCores(measurement string) {
	parsedCoresMeasurement, err := parseCoresMeasurement(measurement)
	if err != nil {
		p.errChan <- err
	}
	p.addToAccumulatorCores(parsedCoresMeasurement)
}

func (p *Publisher) publishProcess(measurement processMeasurement) {
	parsedProcessMeasurement, err := parseProcessesMeasurement(measurement)
	if err != nil {
		p.errChan <- err
	}
	p.addToAccumulatorProcesses(parsedProcessMeasurement)
}

func parseCoresMeasurement(measurements string) (parsedCoresMeasurement, error) {
	var values []float64
	splitCSV, err := splitCSVLineIntoValues(measurements)
	if err != nil {
		return parsedCoresMeasurement{}, err
	}
	timestamp, err := parseTime(splitCSV.timeValue)
	if err != nil {
		return parsedCoresMeasurement{}, err
	}
	// change string slice to one string and separate it by coma
	coresString := strings.Join(splitCSV.coreOrPIDsValues, ",")
	// trim unwanted quotes
	coresString = strings.Trim(coresString, "\"")

	for _, metric := range splitCSV.metricsValues {
		parsedValue, err := parseFloat(metric)
		if err != nil {
			return parsedCoresMeasurement{}, err
		}
		values = append(values, parsedValue)
	}
	return parsedCoresMeasurement{coresString, values, timestamp}, nil
}

func (p *Publisher) addToAccumulatorCores(measurement parsedCoresMeasurement) {
	for i, value := range measurement.values {
		if p.shortenedMetrics {
			//0: "IPC"
			//1: "LLC_Misses"
			if i == 0 || i == 1 {
				continue
			}
		}
		tags := map[string]string{}
		fields := make(map[string]interface{})

		tags["cores"] = measurement.cores
		tags["name"] = pqosMetricOrder[i]
		fields["value"] = value

		p.acc.AddFields("rdt_metric", fields, tags, measurement.time)
	}
}

func parseProcessesMeasurement(measurement processMeasurement) (parsedProcessMeasurement, error) {
	splitCSV, err := splitCSVLineIntoValues(measurement.measurement)
	if err != nil {
		return parsedProcessMeasurement{}, err
	}
	pids, err := findPIDsInMeasurement(measurement.measurement)
	if err != nil {
		return parsedProcessMeasurement{}, err
	}
	lenOfPIDs := len(strings.Split(pids, ","))
	if lenOfPIDs > len(splitCSV.coreOrPIDsValues) {
		return parsedProcessMeasurement{}, errors.New("detected more pids (quoted) than actual number of pids in csv line")
	}
	timestamp, err := parseTime(splitCSV.timeValue)
	if err != nil {
		return parsedProcessMeasurement{}, err
	}
	actualProcess := measurement.name
	cores := strings.Trim(strings.Join(splitCSV.coreOrPIDsValues[lenOfPIDs:], ","), `"`)

	var values []float64
	for _, metric := range splitCSV.metricsValues {
		parsedValue, err := parseFloat(metric)
		if err != nil {
			return parsedProcessMeasurement{}, err
		}
		values = append(values, parsedValue)
	}
	return parsedProcessMeasurement{actualProcess, cores, values, timestamp}, nil
}

func (p *Publisher) addToAccumulatorProcesses(measurement parsedProcessMeasurement) {
	for i, value := range measurement.values {
		if p.shortenedMetrics {
			//0: "IPC"
			//1: "LLC_Misses"
			if i == 0 || i == 1 {
				continue
			}
		}
		tags := map[string]string{}
		fields := make(map[string]interface{})

		tags["process"] = measurement.process
		tags["cores"] = measurement.cores
		tags["name"] = pqosMetricOrder[i]
		fields["value"] = value

		p.acc.AddFields("rdt_metric", fields, tags, measurement.time)
	}
}
