// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
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
  # Period after which counters will be reread from configuration and wildcards in counter paths expanded
  CountersRefreshInterval="1m"

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
	PrintValid bool
	//deprecated: determined dynamically
	PreVistaSupport         bool
	Object                  []perfobject
	CountersRefreshInterval internal.Duration

	lastRefreshed time.Time
	counters      []*counter
	query         PerformanceQuery
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

var counterPathRE = regexp.MustCompile(`.*\\(.*)\\(.*)`)
var objectInstanceRE = regexp.MustCompile(`(.*)\((.*)\)`)

//extractObjectInstanceCounterFromQuery gets object name, instance name (if available) and counter name from counter path
func extractObjectInstanceCounterFromQuery(query string) (object string, instance string, counter string, err error) {
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

func (m *Win_PerfCounters) Description() string {
	return "Input plugin to counterPath Performance Counters on Windows operating systems"
}

func (m *Win_PerfCounters) SampleConfig() string {
	return sampleConfig
}

func (m *Win_PerfCounters) AddItem(counterPath string, instance string, measurement string, includeTotal bool) error {
	if !m.query.AddEnglishCounterSupported() {
		_, err := m.query.AddCounterToQuery(counterPath)
		if err != nil {
			return err
		}
	} else {
		counterHandle, err := m.query.AddEnglishCounterToQuery(counterPath)
		if err != nil {
			return err
		}
		counterPath, err = m.query.GetCounterPath(counterHandle)
		if err != nil {
			return err
		}
	}

	counters, err := m.query.ExpandWildCardPath(counterPath)
	if err != nil {
		return err
	}

	for _, counterPath := range counters {
		var err error
		counterHandle, err := m.query.AddCounterToQuery(counterPath)

		parsedObjectName, parsedInstance, parsedCounter, err := extractObjectInstanceCounterFromQuery(counterPath)
		if err != nil {
			return err
		}

		if parsedInstance == "_Total" && instance == "*" && !includeTotal {
			continue
		}

		newItem := &counter{counterPath, parsedObjectName, parsedCounter, parsedInstance, measurement,
			includeTotal, counterHandle}
		m.counters = append(m.counters, newItem)

		if m.PrintValid {
			log.Printf("Valid: %s\n", counterPath)
		}
	}

	return nil
}

func (m *Win_PerfCounters) ParseConfig() error {
	var counterPath string

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

					err := m.AddItem(counterPath, instance, PerfObject.Measurement, PerfObject.IncludeTotal)

					if err != nil {
						if PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
							log.Printf("Invalid counterPath: '%s'. Error: %s\n", counterPath, err.Error())
						}
						if PerfObject.FailOnMissing {
							return err
						}
					}
				}
			}
		}
		return nil
	} else {
		err := errors.New("no performance objects configured")
		return err
	}

}

func (m *Win_PerfCounters) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	var err error

	if m.lastRefreshed.IsZero() || (m.CountersRefreshInterval.Duration.Nanoseconds() > 0 && m.lastRefreshed.Add(m.CountersRefreshInterval.Duration).Before(time.Now())) {
		m.counters = m.counters[:0]

		err = m.query.Open()
		if err != nil {
			return err
		}

		err = m.ParseConfig()
		if err != nil {
			return err
		}
		//some counters need two data samples before computing a value
		err = m.query.CollectData()
		if err != nil {
			return err
		}
		m.lastRefreshed = time.Now()

		time.Sleep(time.Second)
	}

	type InstanceGrouping struct {
		name       string
		instance   string
		objectname string
	}

	var collectFields = make(map[InstanceGrouping]map[string]interface{})

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
			//ignore invalid data from as some counters from process instances returns this sometimes
			if phderr, ok := err.(*PdhError); ok && phderr.ErrorCode != PDH_INVALID_DATA && phderr.ErrorCode != PDH_CALC_NEGATIVE_VALUE {
				return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
			}
		}
	}

	for instance, fields := range collectFields {
		var tags = map[string]string{
			"objectname": instance.objectname,
		}
		if len(instance.instance) > 0 {
			tags["instance"] = instance.instance
		}
		acc.AddFields(instance.name, fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("win_perf_counters", func() telegraf.Input {
		return &Win_PerfCounters{query: &PerformanceQueryImpl{}, CountersRefreshInterval: internal.Duration{Duration: time.Second * 60}}
	})
}
