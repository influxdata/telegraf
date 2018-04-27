// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/metric"
)

var sampleConfig = `
  ## By default this plugin returns basic CPU and Disk statistics.
  ## See the README file for more examples.
  ## Uncomment examples below or write your own as you see fit. If the system
  ## being polled for data does not have the Object at startup of the Telegraf
  ## agent, it will not be gathered.
  ## Settings:
  # PrintValid = false # Print All matching performance counters

  [[inputs.win_perf_counters.object]]
    # Processor usage, alternative to native, reports on a per core.
    ObjectName = "Processor"
    Instances = ["*"]
    Counters = [
      "%% Idle Time", "%% Interrupt Time",
      "%% Privileged Time", "%% User Time",
      "%% Processor Time"
    ]
    Measurement = "win_cpu"
    # Set to true to include _Total instance when querying for all (*).
    # IncludeTotal=false
    # Print out when the performance counter is missing from object, counter or instance.
    # WarnOnMissing = false

  [[inputs.win_perf_counters.object]]
    # Disk times and queues
    ObjectName = "LogicalDisk"
    Instances = ["*"]
    Counters = [
      "%% Idle Time", "%% Disk Time","%% Disk Read Time",
      "%% Disk Write Time", "%% User Time", "Current Disk Queue Length"
    ]
    Measurement = "win_disk"

  [[inputs.win_perf_counters.object]]
    ObjectName = "System"
    Counters = ["Context Switches/sec","System Calls/sec"]
    Instances = ["------"]
    Measurement = "win_system"

  [[inputs.win_perf_counters.object]]
    # Example query where the Instance portion must be removed to get data back,
    # such as from the Memory object.
    ObjectName = "Memory"
    Counters = [
      "Available Bytes", "Cache Faults/sec", "Demand Zero Faults/sec",
      "Page Faults/sec", "Pages/sec", "Transition Faults/sec",
      "Pool Nonpaged Bytes", "Pool Paged Bytes"
    ]
    Instances = ["------"] # Use 6 x - to remove the Instance bit from the query.
    Measurement = "win_mem"
`

type Win_PerfCounters struct {
	PrintValid      bool
	PreVistaSupport bool
	Object          []perfobject

	configParsed bool
	itemCache    []*item
}

type perfobject struct {
	ObjectName    string
	Counters      []string
	Instances     []string
	Measurement   string
	WarnOnMissing bool
	FailOnMissing bool
	IncludeTotal  bool
}

type item struct {
	query         string
	objectName    string
	counter       string
	instance      string
	measurement   string
	handle        PDH_HQUERY
	counterHandle PDH_HCOUNTER
}

var sanitizedChars = strings.NewReplacer("/sec", "_persec", "/Sec", "_persec",
	" ", "_", "%", "Percent", `\`, "")

func (m *Win_PerfCounters) AddItem(query string, objectName string, counter string, instance string,
	measurement string) error {

	var handle PDH_HQUERY
	var counterHandle PDH_HCOUNTER
	ret := PdhOpenQuery(0, 0, &handle)
	if m.PreVistaSupport {
		ret = PdhAddCounter(handle, query, 0, &counterHandle)
	} else {
		ret = PdhAddEnglishCounter(handle, query, 0, &counterHandle)
	}

	// Call PdhCollectQueryData one time to check existence of the counter
	ret = PdhCollectQueryData(handle)
	if ret != ERROR_SUCCESS {
		PdhCloseQuery(handle)
		return errors.New(PdhFormatError(ret))
	}

	sanitized_measurement := sanitizedChars.Replace(measurement)
	if sanitized_measurement == "" {
		sanitized_measurement = "win_perf_counters"
	}

	sanitized_counter := sanitizedChars.Replace(counter)

	newItem := &item{query, objectName, sanitized_counter, instance, sanitized_measurement, handle, counterHandle}
	m.itemCache = append(m.itemCache, newItem)

	return nil
}

func (m *Win_PerfCounters) Description() string {
	return "Input plugin to query Performance Counters on Windows operating systems"
}

func (m *Win_PerfCounters) SampleConfig() string {
	return sampleConfig
}

func expandCounterQuery(query string) ([]string, error) {
	var bufSize uint32
	var buf []uint16
	ret := PdhExpandWildCardPath(query, nil, &bufSize)
	for ret == PDH_MORE_DATA {
		buf = make([]uint16, bufSize)
		ret = PdhExpandWildCardPath(query, &buf[0], &bufSize)
	}
	if ret == ERROR_SUCCESS {
		return UTF16ToStringArray(buf), nil
	}
	return nil, fmt.Errorf("Failed to expand query: '%s', err(%d)", query, ret)
}

func formatCounterQuery(objectname string, instance string, counter string) ([]string, error) {
	if instance == "------" {
		return []string{"\\" + objectname + "\\" + counter}, nil
	}
	if counter == "*" || strings.Contains(instance, "*"){
		return expandCounterQuery("\\" + objectname + "(" + instance + ")\\" + counter)
	}

	return []string{"\\" + objectname + "(" + instance + ")\\" + counter}, nil
}

func extractInstanceFromQuery(query string) (string, error) {
	left_paren_idx := strings.Index(query, "(")
	right_paren_idx := strings.Index(query, ")")

	if left_paren_idx == -1 || right_paren_idx == -1 {
		return "", errors.New("Could not extract instance name from: " + query)
	}

	return query[left_paren_idx+1:right_paren_idx], nil
}

func extractCounterFromQuery(query string) (string, error) {
	last_slash_idx := strings.LastIndex(query, "\\")

	if last_slash_idx == -1 {
		return "", errors.New("Could not extract counter name from: " + query)
	}

	return query[last_slash_idx:], nil
}

func (m *Win_PerfCounters) ParseConfig() error {
	if len(m.Object) > 0 {
		for _, PerfObject := range m.Object {
			for _, counter := range PerfObject.Counters {
				for _, instance := range PerfObject.Instances {
					var err error
					var expandedQueries []string

					objectname := PerfObject.ObjectName
					expandedQueries, err = formatCounterQuery(objectname, instance, counter)

					for _, expandedQuery := range expandedQueries {
						var extracted_counter string
						extracted_counter, err = extractCounterFromQuery(expandedQuery)
						if err != nil {
							fmt.Printf(err.Error())
							continue
						}

						if instance == "------" {
							err = m.AddItem(expandedQuery, objectname, extracted_counter, instance, PerfObject.Measurement)
						} else {
							var extracted_instance string
							extracted_instance, err = extractInstanceFromQuery(expandedQuery)
							if err != nil {
								fmt.Printf(err.Error())
								continue
							}

							if extracted_instance == "_Total" && !PerfObject.IncludeTotal {
								continue
							}

							err = m.AddItem(expandedQuery, objectname, extracted_counter, extracted_instance, PerfObject.Measurement)
						}

						if err == nil {
							if m.PrintValid {
								fmt.Printf("Valid: %s\n", expandedQuery)
							}
						} else {
							if PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
								fmt.Printf("Invalid query: '%s'. Error: %s", expandedQuery, err.Error())
							}
						}
					}
				}
			}
		}

		return nil
	} else {
		err := errors.New("No performance objects configured!")
		return err
	}
}

func (m *Win_PerfCounters) GetParsedItemsForTesting() []*item {
	return m.itemCache
}

func (m *Win_PerfCounters) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	if !m.configParsed {
		err := m.ParseConfig()
		m.configParsed = true
		if err != nil {
			return err
		}
	}

	var emptyBuf [1]uint32 // need at least 1 addressable null ptr.
	var counterValue PDH_FMT_COUNTERVALUE_DOUBLE

	// For iterate over the known metrics and get the samples.
	for _, metric := range m.itemCache {
		// collect
		ret := PdhCollectQueryData(metric.handle)
		if ret == ERROR_SUCCESS {
			ret = PdhGetFormattedCounterValueDouble(metric.counterHandle, &emptyBuf[0], &counterValue)
			if ret == ERROR_SUCCESS {
				fields := make(map[string]interface{})
				tags := make(map[string]string)
				tags["instance"] = metric.instance
				tags["objectname"] = metric.objectName
				fields[metric.counter] = float32(counterValue.DoubleValue)
				acc.AddFields(metric.measurement, fields, tags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("win_perf_counters", func() telegraf.Input { return &Win_PerfCounters{} })
}
