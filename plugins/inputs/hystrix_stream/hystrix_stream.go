package hystrix_stream

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io"
	"net/http"
	"time"
)

const sampleConfig = `
  ## Hystrix stream servlet to connect to (with port and full path)
  hystrix_servlet_url = "http://localhost:8090/hystrix"
 `

type HystrixData struct {
	Hystrix_servlet_url string
}

func (s *HystrixData) Description() string {
	return "Plugin to parse a hystrix-stream-servlet."
}

func (s *HystrixData) SampleConfig() string {
	return sampleConfig
}

func (s *HystrixData) Gather(acc telegraf.Accumulator) error {

	for {
		resp, err := http.Get(s.Hystrix_servlet_url)
		if err != nil {
			return err
		} else {
			entries, errors := entryStream(resp.Body, 10)
			for {
				select {
				case entry := <-entries:
					entryToAccumulator(entry, acc)
				case err := <-errors:
					if err == io.EOF {
						return nil
					} else {
						return err
					}
				}
			}
		}
	}

	return nil
}

func entryToAccumulator(entry HystrixStreamEntry, accumulator telegraf.Accumulator) {
	tags := getTags(entry)
	counterName := entry.Group + entry.Name
	accumulator.AddCounter(counterName, getCounterFields(entry), tags, time.Unix(entry.CurrentTime/1000, 0))
}

func getTags(entry HystrixStreamEntry) map[string]string {
	tags := make(map[string]string)
	tags["name"] = entry.Name
	tags["type"] = entry.Type
	tags["group"] = entry.Group
	tags["threadpool"] = entry.ThreadPool
	return tags
}

func getCounterFields(entry HystrixStreamEntry) map[string]interface{} {
	fields := make(map[string]interface{})
	fields["ErrorCount"] = entry.ErrorCount
	fields["LatencyTotal"] = entry.LatencyTotal
	fields["RequestCount"] = entry.RequestCount
	fields["LatencyExecute"] = entry.LatencyExecute
	fields["ReportingHosts"] = entry.ReportingHosts
	fields["ErrorPercentage"] = entry.ErrorPercentage
	fields["IsCircuitBreakerOpen"] = entry.IsCircuitBreakerOpen
	fields["CurrentConcurrentExecutionCount"] = entry.CurrentConcurrentExecutionCount
	return fields
}

func init() {
	inputs.Add("HystrixStream", func() telegraf.Input { return &HystrixData{} })
}
