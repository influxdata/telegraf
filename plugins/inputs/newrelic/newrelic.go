package newrelic

import (
	"fmt"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/rmanocha/go_newrelic_api"
)

// Metric struct
type Metric struct {
	Values []string
}

// NewRelic Plugin Struct
type NewRelic struct {
	Name                    string
	APIKey                  string
	APPID                   int
	Metrics                 map[string]Metric
	PollServerAverages      bool
	PollApplicationAverages bool
}

// Description of Plugin
func (nr *NewRelic) Description() string {
	return "Newrelic Metrics Query Plugin"
}

var sampleConfig = `
# Name
Name = ""

#APIKey
APIKey = ""

#Application ID
APPID =

# Poll for Averages
PollServerAverages=false
PollApplicationAverages=false

# Poll for individual metrics
[inputs.newrelic.Metrics]
  [inputs.newrelic.Metrics.Apdex]
  Values = ["f", "count", "value", "threshold", "threshold_min", "score", "s"]

  [inputs.newrelic.Metrics."Errors/all"]
  Values = ["errors_per_minute", "error_count"]

  [inputs.newrelic.Metrics."EndUser"]
  Values = ["average_response_time", "calls_per_minute", "call_count"]

  [inputs.newrelic.Metrics."HttpDispatcher"]
  Values = ["average_response_time", "calls_per_minute", "call_count"]
`

// SampleConfig of Plugin
func (nr *NewRelic) SampleConfig() string {
	return sampleConfig
}

func (nr *NewRelic) GatherServerAverages(acc telegraf.Accumulator, nrapi *go_newrelic_api.Newrelic) error {
	dataserver := nrapi.GetServers()

	for v := range dataserver.Servers {
		tags := map[string]string{
			"ServerName": fmt.Sprintf("%s", dataserver.Servers[v].Name),
		}

		fieldsServers := map[string]interface{}{
			"CPU":             dataserver.Servers[v].ServerSummary.CPU,
			"CPUStolen":       dataserver.Servers[v].ServerSummary.CPUStolen,
			"DiskIOPercent":   dataserver.Servers[v].ServerSummary.DiskIO,
			"MemoryPercent":   dataserver.Servers[v].ServerSummary.Memory,
			"MemoryUsed":      dataserver.Servers[v].ServerSummary.MemoryUsed,
			"MemoryTotal":     dataserver.Servers[v].ServerSummary.MemoryTotal,
			"FullestDisk":     dataserver.Servers[v].ServerSummary.FullestDisk,
			"FullestDiskFree": dataserver.Servers[v].ServerSummary.FullestDiskFree,
		}

		acc.AddFields("Servers", fieldsServers, tags)
	}

	return nil
}

func (nr *NewRelic) GatherApplicationAverages(acc telegraf.Accumulator, nrapi *go_newrelic_api.Newrelic) error {
	dataapplication := nrapi.GetApplication(nr.APPID)

	fieldsApplication := map[string]interface{}{
		"ResponceTime": dataapplication.Application.ApplicationSummary.ResponseTime,
		"Throughput":   dataapplication.Application.ApplicationSummary.Throughput,
		"ErrorRate":    dataapplication.Application.ApplicationSummary.ErrorRate,
		"ApdexTarget":  dataapplication.Application.ApplicationSummary.ApdexTarget,
		"ApdexScore":   dataapplication.Application.ApplicationSummary.ApdexScore,
	}

	fieldsEndUser := map[string]interface{}{
		"ResponceTime": dataapplication.Application.EndUserSummary.ResponseTime,
		"Throughput":   dataapplication.Application.EndUserSummary.Throughput,
		"ApdexTarget":  dataapplication.Application.EndUserSummary.ApdexTarget,
		"ApdexScore":   dataapplication.Application.EndUserSummary.ApdexScore,
	}

	acc.AddFields("Application", fieldsApplication, nil)
	acc.AddFields("EndUser", fieldsEndUser, nil)

	return nil
}

func (nr *NewRelic) GatherMetrics(acc telegraf.Accumulator, nrapi *go_newrelic_api.Newrelic) error {
	return nil
}

// Gather requested metrics
func (nr *NewRelic) Gather(acc telegraf.Accumulator) error {
	conn := go_newrelic_api.NewNewrelic(nr.APIKey)

	if nr.PollServerAverages {
		nr.GatherServerAverages(acc, conn)
	}

	if nr.PollApplicationAverages {
		nr.GatherApplicationAverages(acc, conn)
	}

	if len(nr.Metrics) > 0 {
		tNow := time.Now()
		tFrom := tNow.Add(-1 * time.Minute)

		tFromStr := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
			tFrom.Year(), tFrom.Month(), tFrom.Day(),
			tFrom.Hour(), tFrom.Minute(), 00)

		tToStr := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
			tNow.Year(), tNow.Month(), tNow.Day(),
			tNow.Hour(), tNow.Minute(), 00)

		vals := url.Values{}

		vals.Add("from", tFromStr)
		vals.Add("to", tToStr)

		for k := range nr.Metrics {
			vals.Add("names[]", k)
		}

		result := conn.GetMetricData(nr.APPID, vals)

		var fieldsMetrics map[string]interface{}
		fieldsMetrics = make(map[string]interface{})

		for metricRequest := range nr.Metrics {
			for metricResult := range result.MetricData.Metrics {
				if metricRequest == result.MetricData.Metrics[metricResult].Name {
					// If new relic returned a metric we are after
					for valueRequest := range nr.Metrics[metricRequest].Values {
						for valueResult := range result.MetricData.Metrics[metricResult].Timeslices[0].Values {
							if valueResult == nr.Metrics[metricRequest].Values[valueRequest] {
								// If we matched a returned metric value that was requested
								fieldsMetrics[metricRequest+"_"+valueResult] = result.MetricData.Metrics[metricResult].Timeslices[0].Values[valueResult]
							}
						}
					}
				}
			}
		}
		acc.AddFields("Metrics", fieldsMetrics, nil)
	}

	return nil
}

func init() {
	inputs.Add("newrelic", func() telegraf.Input { return &NewRelic{} })
}
