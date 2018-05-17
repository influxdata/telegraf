// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"regexp"
	"strings"
	"time"
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
	WarnOnMissing   bool
	FailOnMissing   bool

	configParsed bool
	itemCache    []*item
	queryHandle  PDH_HQUERY
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
	include_total bool
	handle        PDH_HQUERY
	counterHandle PDH_HCOUNTER
}

var sanitizedChars = strings.NewReplacer("/sec", "_persec", "/Sec", "_persec",
	" ", "_", "%", "Percent", `\`, "")

//General Counter path pattern is: \\computer\object(parent/instance#index)\counter
//parent/instance#index part is skipped in single instance objects (e.g. Memory): \\computer\object\counter

var counterPathRE = regexp.MustCompile(`\\\\.*\\(.*)\\(.*)`)
var objectInstanceRE = regexp.MustCompile(`(.*)\((.*)\)`)

//get object name, instance name (if available) and counter name from counter path
func extractObjectInstanceCounterFromQueryRE(query string) (object string, instance string, counter string, err error) {
	pathParts := counterPathRE.FindAllStringSubmatch(query, -1)
	if pathParts == nil || len(pathParts[0]) != 3 {
		err = errors.New("Could not extract counter info from: " + query)
		return
	}
	counter = pathParts[0][2]
	//try to get instance name
	objectInstanceParts := objectInstanceRE.FindAllStringSubmatch(pathParts[0][1], -1)
	if objectInstanceParts == nil || len(objectInstanceParts[0]) != 3 {
		object = pathParts[0][1]
	} else {
		object = objectInstanceParts[0][1]
		instance = objectInstanceParts[0][2]
	}
	return
}


func extractObjectInstanceCounterFromQuery(query string) (object string, instance string, counter string, err error) {
	left_paren_idx := strings.Index(query, "(")
	right_paren_idx := strings.Index(query, ")")

	if left_paren_idx != -1 && right_paren_idx != -1 {
		instance = query[left_paren_idx+1 : right_paren_idx]
	}

	last_slash_idx := strings.LastIndex(query, "\\")

	if last_slash_idx == -1 {
		err = errors.New("Could not extract counter name from: " + query)
		return
	}

	counter = query[last_slash_idx:]

	prelast_slash_idx := strings.LastIndex(query[:last_slash_idx], "\\")

	if prelast_slash_idx == -1 {
		err = errors.New("Could not extract object name from: " + query)
		return
	}

	objNameEndInd := last_slash_idx

	if left_paren_idx != -1 {
		objNameEndInd = left_paren_idx
	}

	object = query[prelast_slash_idx+1 : objNameEndInd]

	return
}

func (m *Win_PerfCounters) AddItem(query string, objectName string, counter string, instance string,
	measurement string, include_total bool) error {

	var counterHandle PDH_HCOUNTER
	var ret uint32
	if !PdhAddEnglishCounterSupported() {
		ret = PdhAddCounter(m.queryHandle, query, 0, &counterHandle)
		if ret != ERROR_SUCCESS {
			return errors.New(PdhFormatError(ret))
		}
	} else {
		ret = PdhAddEnglishCounter(m.queryHandle, query, 0, &counterHandle)
		if ret != ERROR_SUCCESS {
			return errors.New(PdhFormatError(ret))
		}
		ci, err := GetCounterInfo(counterHandle)
		if err != nil {
			return err
		}
		query = UTF16PtrToString(ci.SzFullPath)
	}

	counters, err := ExpandWildCardPath(query)
	if err != nil {
		return err
	}

	for _, counterPath := range counters {
		ret = PdhAddCounter(m.queryHandle, counterPath, 0, &counterHandle)
		if ret != ERROR_SUCCESS {
			return errors.New(PdhFormatError(ret))
		}

		parsedObjectName, parsedInstance, parsedCounter, err := extractObjectInstanceCounterFromQueryRE(counterPath)
		if err != nil {
			return err
		}

		if parsedInstance == "_Total" && !include_total {
			continue
		}

		newItem := &item{counterPath, parsedObjectName, parsedCounter, parsedInstance, measurement,
			include_total, m.queryHandle, counterHandle}
		//fmt.Printf("Added q: %s, o: %s, i: %s, c: %s\n", newItem.query, newItem.objectName, newItem.instance, newItem.counter)
		m.itemCache = append(m.itemCache, newItem)
	}

	return nil
}

func (m *Win_PerfCounters) Description() string {
	return "Input plugin to query Performance Counters on Windows operating systems"
}

func (m *Win_PerfCounters) SampleConfig() string {
	return sampleConfig
}

func (m *Win_PerfCounters) ParseConfig() error {
	var query string

	start := time.Now()

	if len(m.Object) > 0 {
		for _, PerfObject := range m.Object {
			for _, counter := range PerfObject.Counters {
				for _, instance := range PerfObject.Instances {
					objectname := PerfObject.ObjectName

					if instance == "------" {
						query = "\\" + objectname + "\\" + counter
					} else {
						query = "\\" + objectname + "(" + instance + ")\\" + counter
					}

					err := m.AddItem(query, objectname, counter, instance,
						PerfObject.Measurement, PerfObject.IncludeTotal)

					if err == nil {
						if m.PrintValid {
							fmt.Printf("Valid: %s\n", query)
						}
					} else {
						if m.WarnOnMissing || m.FailOnMissing || PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
							fmt.Printf("Invalid query: '%s'. Error: %s", query, err.Error())
						}
						if m.FailOnMissing || PerfObject.FailOnMissing {
							return err
						}
					}
				}
			}
		}
		took := time.Now().Sub(start).Seconds()

		fmt.Printf("ParseConfig: Found %d items, took %.2f\n", len(m.itemCache), took)
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
	var err error
	if !m.configParsed {
		m.queryHandle, err = OpenPerformanceCountersQuery()
		if err != nil {
			return err
		}

		err = m.ParseConfig()
		m.configParsed = true
		if err != nil {
			return err
		}
		//some counters need two data samples before computing a value
		ret := PdhCollectQueryData(m.queryHandle)
		if ret != ERROR_SUCCESS {
			return errors.New(PdhFormatError(ret))
		}
		time.Sleep(time.Second)
	}

	type InstanceGrouping struct {
		name       string
		instance   string
		objectname string
	}

	var collectFields = make(map[InstanceGrouping]map[string]interface{})

	start := time.Now()

	ret := PdhCollectQueryData(m.queryHandle)
	if ret != ERROR_SUCCESS {
		return errors.New(PdhFormatError(ret))
	}
	//some counters
	// For iterate over the known metrics and get the samples.
	for _, metric := range m.itemCache {
		// collect
		//fmt.Printf("Getting values for %s\n", metric.query)

		//time.Sleep(time.Second)
		//ret = PdhCollectQueryData(metric.handle)
		//if ret == ERROR_SUCCESS {
		value, err := GetFormattedCounterValueDouble(metric.counterHandle)
		if err == nil {
			measurement := sanitizedChars.Replace(metric.measurement)
			if measurement == "" {
				measurement = "win_perf_counters"
			}

			var instance = InstanceGrouping{measurement, metric.instance, metric.objectName}
			if collectFields[instance] == nil {
				collectFields[instance] = make(map[string]interface{})
			}
			collectFields[instance][sanitizedChars.Replace(metric.counter)] = float32(value.DoubleValue)
		} else {
			return fmt.Errorf("error while getting value for counter %s: %v", metric.query, err)
		}
		//} else {
		//	return errors.New(PdhFormatError(ret))
		//}
	}

	for instance, fields := range collectFields {
		var tags = map[string]string{
			"instance":   instance.instance,
			"objectname": instance.objectname,
		}
		acc.AddFields(instance.name, fields, tags)
	}
	took := time.Now().Sub(start).Seconds()

	fmt.Printf("Gather: took %.2f\n", took)
	return nil
}

func init() {
	inputs.Add("win_perf_counters", func() telegraf.Input { return &Win_PerfCounters{} })
}
