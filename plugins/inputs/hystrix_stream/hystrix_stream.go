package hystrix_stream

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
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

	entries, errors := latestEntries(s.Hystrix_servlet_url)
	if errors != nil {
		return errors
	}
	for _, entry := range entries {
		acc.AddFields(fieldsFromEntry(entry))
	}
	return nil
}

func fieldsFromEntry(entry HystrixStreamEntry) (string, map[string]interface{}, map[string]string, time.Time) {
	counterName := entry.Group + entry.Name
	fields := getCounterFields(entry)
	tags := getTags(entry)
	entryTime := time.Unix(0, entry.CurrentTime*int64(time.Millisecond))
	return counterName, fields, tags, entryTime
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
	fields["RequestCount"] = entry.RequestCount
	fields["ErrorCount"] = entry.ErrorCount
	fields["LatencyTotal0"] = entry.LatencyTotal.Num0
	fields["LatencyTotal25"] = entry.LatencyTotal.Num25
	fields["LatencyTotal50"] = entry.LatencyTotal.Num50
	fields["LatencyTotal75"] = entry.LatencyTotal.Num75
	fields["LatencyTotal90"] = entry.LatencyTotal.Num90
	fields["LatencyTotal95"] = entry.LatencyTotal.Num95
	fields["LatencyTotal99"] = entry.LatencyTotal.Num99
	fields["LatencyTotal100"] = entry.LatencyTotal.Num100
	fields["LatencyExecute0"] = entry.LatencyExecute.Num0
	fields["LatencyExecute25"] = entry.LatencyExecute.Num25
	fields["LatencyExecute50"] = entry.LatencyExecute.Num50
	fields["LatencyExecute75"] = entry.LatencyExecute.Num75
	fields["LatencyExecute90"] = entry.LatencyExecute.Num90
	fields["LatencyExecute95"] = entry.LatencyExecute.Num95
	fields["LatencyExecute99"] = entry.LatencyExecute.Num99
	fields["LatencyExecute100"] = entry.LatencyExecute.Num100
	fields["ReportingHosts"] = entry.ReportingHosts
	fields["ErrorPercentage"] = entry.ErrorPercentage
	fields["IsCircuitBreakerOpen"] = entry.IsCircuitBreakerOpen
	fields["CurrentConcurrentExecutionCount"] = entry.CurrentConcurrentExecutionCount
	return fields
}

func init() {
	inputs.Add("HystrixStream", func() telegraf.Input { return &HystrixData{} })
}
