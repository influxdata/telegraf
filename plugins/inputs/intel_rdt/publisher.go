//go:build !windows

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

// publisher for publish new RDT metrics to telegraf accumulator
type publisher struct {
	acc               telegraf.Accumulator
	log               telegraf.Logger
	shortenedMetrics  bool
	bufferChanProcess chan processMeasurement
	bufferChanCores   chan string
	errChan           chan error
}

func newPublisher(acc telegraf.Accumulator, log telegraf.Logger, shortenedMetrics bool) publisher {
	return publisher{
		acc:               acc,
		log:               log,
		shortenedMetrics:  shortenedMetrics,
		bufferChanProcess: make(chan processMeasurement),
		bufferChanCores:   make(chan string),
		errChan:           make(chan error),
	}
}

func (p *publisher) publish(ctx context.Context) {
	go func() {
		for {
			select {
			case newMeasurements := <-p.bufferChanCores:
				p.publishCores(newMeasurements)
			case newMeasurements := <-p.bufferChanProcess:
				p.publishProcess(newMeasurements)
			case err := <-p.errChan:
				p.log.Error(err)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *publisher) publishCores(measurement string) {
	parsedCoresMeasurement, err := parseCoresMeasurement(measurement)
	if err != nil {
		p.errChan <- err
	}
	p.addToAccumulatorCores(parsedCoresMeasurement)
}

func (p *publisher) publishProcess(measurement processMeasurement) {
	parsedProcessMeasurement, err := parseProcessesMeasurement(measurement)
	if err != nil {
		p.errChan <- err
	}
	p.addToAccumulatorProcesses(parsedProcessMeasurement)
}

func parseCoresMeasurement(measurements string) (parsedCoresMeasurement, error) {
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

	values := make([]float64, 0, len(splitCSV.metricsValues))
	for _, metric := range splitCSV.metricsValues {
		parsedValue, err := parseFloat(metric)
		if err != nil {
			return parsedCoresMeasurement{}, err
		}
		values = append(values, parsedValue)
	}
	return parsedCoresMeasurement{coresString, values, timestamp}, nil
}

func (p *publisher) addToAccumulatorCores(measurement parsedCoresMeasurement) {
	for i, value := range measurement.values {
		if p.shortenedMetrics {
			// 0: "IPC"
			// 1: "LLC_Misses"
			if i == 0 || i == 1 {
				continue
			}
		}
		tags := make(map[string]string, 2)
		fields := make(map[string]interface{}, 1)

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

	values := make([]float64, 0, len(splitCSV.metricsValues))
	for _, metric := range splitCSV.metricsValues {
		parsedValue, err := parseFloat(metric)
		if err != nil {
			return parsedProcessMeasurement{}, err
		}
		values = append(values, parsedValue)
	}
	return parsedProcessMeasurement{actualProcess, cores, values, timestamp}, nil
}

func (p *publisher) addToAccumulatorProcesses(measurement parsedProcessMeasurement) {
	for i, value := range measurement.values {
		if p.shortenedMetrics {
			// 0: "IPC"
			// 1: "LLC_Misses"
			if i == 0 || i == 1 {
				continue
			}
		}
		tags := make(map[string]string, 3)
		fields := make(map[string]interface{}, 1)

		tags["process"] = measurement.process
		tags["cores"] = measurement.cores
		tags["name"] = pqosMetricOrder[i]
		fields["value"] = value

		p.acc.AddFields("rdt_metric", fields, tags, measurement.time)
	}
}
