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
    # Example counterPath where the Instance portion must be removed to get data back,
    # such as from the Memory object.
    ObjectName = "Memory"
    Counters = [
      "Available Bytes", "Cache Faults/sec", "Demand Zero Faults/sec",
      "Page Faults/sec", "Pages/sec", "Transition Faults/sec",
      "Pool Nonpaged Bytes", "Pool Paged Bytes"
    ]
    Instances = ["------"] # Use 6 x - to remove the Instance bit from the counterPath.
    Measurement = "win_mem"
`

type Win_PerfCounters struct {
	PrintValid      bool
	PreVistaSupport bool
	Object          []perfobject
	WarnOnMissing   bool
	FailOnMissing   bool

	configParsed bool
	counters     []*counter
	query        PerformanceQuery
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

type counter struct {
	counterPath   string
	objectName    string
	counter       string
	instance      string
	measurement   string
	includeTotal  bool
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


func (m *Win_PerfCounters) AddItem(counterPath string,	measurement string, includeTotal bool) error {

	if !m.query.AddEnglishCounterSupported() {
		_, err := m.query.AddCounterToQuery (counterPath)
		if err != nil {
			return err
		}
	} else {
		counterHandle, err := m.query.AddEnglishCounterToQuery (counterPath)
		if err != nil {
			return err
		}
		counterPath, err =  m.query.GetCounterPath(counterHandle)
		if err != nil {
			return err
		}
	}

	counters, err :=  m.query.ExpandWildCardPath(counterPath)
	if err != nil {
		return err
	}

	for _, counterPath := range counters {
		counterHandle, err := m.query.AddCounterToQuery (counterPath)

		parsedObjectName, parsedInstance, parsedCounter, err := extractObjectInstanceCounterFromQueryRE(counterPath)
		if err != nil {
			return err
		}

		if parsedInstance == "_Total" && !includeTotal {
			continue
		}

		newItem := &counter{counterPath, parsedObjectName, parsedCounter, parsedInstance, measurement,
			includeTotal, counterHandle}
		//fmt.Printf("Added q: %s, o: %s, i: %s, c: %s\n", newItem.counterPath, newItem.objectName, newItem.instance, newItem.counter)
		m.counters = append(m.counters, newItem)
	}

	return nil
}

func (m *Win_PerfCounters) Description() string {
	return "Input plugin to counterPath Performance Counters on Windows operating systems"
}

func (m *Win_PerfCounters) SampleConfig() string {
	return sampleConfig
}

func (m *Win_PerfCounters) ParseConfig() error {
	var counterPath string

	start := time.Now()

	if len(m.Object) > 0 {
		for _, PerfObject := range m.Object {
			for _, counter := range PerfObject.Counters {
				for _, instance := range PerfObject.Instances {
					objectname := PerfObject.ObjectName

					if instance == "------" {
						counterPath = "\\" + objectname + "\\" + counter
					} else {
						counterPath = "\\" + objectname + "(" + instance + ")\\" + counter
					}

					err := m.AddItem(counterPath, PerfObject.Measurement, PerfObject.IncludeTotal)

					if err == nil {
						if m.PrintValid {
							fmt.Printf("Valid: %s\n", counterPath)
						}
					} else {
						if m.WarnOnMissing || m.FailOnMissing || PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
							fmt.Printf("Invalid counterPath: '%s'. Error: %s", counterPath, err.Error())
						}
						if m.FailOnMissing || PerfObject.FailOnMissing {
							return err
						}
					}
				}
			}
		}
		took := time.Now().Sub(start).Seconds()

		fmt.Printf("ParseConfig: Found %d items, took %.2f\n", len(m.counters), took)
		return nil
	} else {
		err := errors.New("no performance objects configured")
		return err
	}

}

func (m *Win_PerfCounters) GetParsedItemsForTesting() []*counter {
	return m.counters
}

func (m *Win_PerfCounters) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	var err error
	if !m.configParsed {
		err = m.query.Open()
		if err != nil {
			return err
		}

		err = m.ParseConfig()
		m.configParsed = true
		if err != nil {
			return err
		}
		//some counters need two data samples before computing a value
		err = m.query.CollectData()
		if err != nil {
			return err
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

	err = m.query.CollectData()
	if err != nil {
		return err
	}
	// For iterate over the known metrics and get the samples.
	for _, metric := range m.counters {
		// collect
		value, err := m.query.GetFormattedCounterValueDouble(metric.counterHandle)
		if err == nil {
			measurement := sanitizedChars.Replace(metric.measurement)
			if measurement == "" {
				measurement = "win_perf_counters"
			}

			var instance = InstanceGrouping{measurement, metric.instance, metric.objectName}
			if collectFields[instance] == nil {
				collectFields[instance] = make(map[string]interface{})
			}
			collectFields[instance][sanitizedChars.Replace(metric.counter)] = float32(value)
		} else {
			return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
		}
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
	inputs.Add("win_perf_counters", func() telegraf.Input { return &Win_PerfCounters{query: &PerformanceQueryImpl{} }})
}
