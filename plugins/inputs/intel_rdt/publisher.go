// +build !windows

package intel_rdt

import (
	"context"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

// Publisher for publish new RDT metrics to telegraf accumulator
type Publisher struct {
	acc               telegraf.Accumulator
	Log               telegraf.Logger
	shortenedMetrics  bool
	BufferChanProcess chan processMeasurement
	BufferChanCores   chan string
	errChan           chan error
	stopChan          chan bool
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
	coresString, values, timestamp, err := parseCoresMeasurement(measurement)
	if err != nil {
		p.errChan <- err
	}
	p.addToAccumulatorCores(coresString, values, timestamp)
	return
}

func (p *Publisher) publishProcess(measurement processMeasurement) {
	process, coresString, values, timestamp, err := parseProcessesMeasurement(measurement)
	if err != nil {
		p.errChan <- err
	}
	p.addToAccumulatorProcesses(process, coresString, values, timestamp)
	return
}

func parseCoresMeasurement(measurements string) (string, []float64, time.Time, error) {
	var values []float64
	timeValue, metricsValues, cores, err := splitCSVLineIntoValues(measurements)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	timestamp, err := parseTime(timeValue)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	// change string slice to one string and separate it by coma
	coresString := strings.Join(cores, ",")
	// trim unwanted quotes
	coresString = strings.Trim(coresString, "\"")

	for _, metric := range metricsValues {
		parsedValue, err := parseFloat(metric)
		if err != nil {
			return "", nil, time.Time{}, err
		}
		values = append(values, parsedValue)
	}
	return coresString, values, timestamp, nil
}

func (p *Publisher) addToAccumulatorCores(cores string, metricsValues []float64, timestamp time.Time) {
	for i, value := range metricsValues {
		if p.shortenedMetrics {
			//0: "IPC"
			//1: "LLC_Misses"
			if i == 0 || i == 1 {
				continue
			}
		}
		tags := map[string]string{}
		fields := make(map[string]interface{})

		tags["cores"] = cores
		tags["name"] = pqosMetricOrder[i]
		fields["value"] = value

		p.acc.AddFields("rdt_metric", fields, tags, timestamp)
	}
}

func parseProcessesMeasurement(measurement processMeasurement) (string, string, []float64, time.Time, error) {
	var values []float64
	timeValue, metricsValues, coreOrPidsValues, pids, err := parseProcessMeasurement(measurement.measurement)
	if err != nil {
		return "", "", nil, time.Time{}, err
	}
	timestamp, err := parseTime(timeValue)
	if err != nil {
		return "", "", nil, time.Time{}, err
	}
	actualProcess := measurement.name
	lenOfPids := len(strings.Split(pids, ","))
	cores := coreOrPidsValues[lenOfPids:]
	coresString := strings.Trim(strings.Join(cores, ","), `"`)

	for _, metric := range metricsValues {
		parsedValue, err := parseFloat(metric)
		if err != nil {
			return "", "", nil, time.Time{}, err
		}
		values = append(values, parsedValue)
	}
	return actualProcess, coresString, values, timestamp, nil
}

func (p *Publisher) addToAccumulatorProcesses(process string, cores string, metricsValues []float64, timestamp time.Time) {
	for i, value := range metricsValues {
		if p.shortenedMetrics {
			//0: "IPC"
			//1: "LLC_Misses"
			if i == 0 || i == 1 {
				continue
			}
		}
		tags := map[string]string{}
		fields := make(map[string]interface{})

		tags["process"] = process
		tags["cores"] = cores
		tags["name"] = pqosMetricOrder[i]
		fields["value"] = value

		p.acc.AddFields("rdt_metric", fields, tags, timestamp)
	}
}

func parseProcessMeasurement(measurements string) (string, []string, []string, string, error) {
	timeValue, metricsValues, coreOrPidsValues, err := splitCSVLineIntoValues(measurements)
	if err != nil {
		return "", nil, nil, "", err
	}
	pids, err := findPIDsInMeasurement(measurements)
	if err != nil {
		return "", nil, nil, "", err
	}
	return timeValue, metricsValues, coreOrPidsValues, pids, nil
}
