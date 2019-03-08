// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## By default this plugin returns basic CPU and Disk statistics.
  ## See the README file for more examples.
  ## Uncomment examples below or write your own as you see fit. If the system
  ## being polled for data does not have the Object at startup of the Telegraf
  ## agent, it will not be gathered.
  ## Settings:
  # PrintValid = false # Print All matching performance counters
  # Whether request a timestamp along with the PerfCounter data or just use current time
  # UsePerfCounterTime=true
  # If UseWildcardsExpansion params is set to true, wildcards (partial wildcards in instance names and wildcards in counters names) in configured counter paths will be expanded
  # and in case of localized Windows, counter paths will be also localized. It also returns instance indexes in instance names.
  # If false, wildcards (not partial) in instance names will still be expanded, but instance indexes will not be returned in instance names.
  #UseWildcardsExpansion = false
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
	UsePerfCounterTime      bool
	Object                  []perfobject
	CountersRefreshInterval internal.Duration
	UseWildcardsExpansion   bool

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

type instanceGrouping struct {
	name       string
	instance   string
	objectname string
}

var sanitizedChars = strings.NewReplacer("/sec", "_persec", "/Sec", "_persec",
	" ", "_", "%", "Percent", `\`, "")

// extractCounterInfoFromCounterPath gets object name, instance name (if available) and counter name from counter path
// General Counter path pattern is: \\computer\object(parent/instance#index)\counter
// parent/instance#index part is skipped in single instance objects (e.g. Memory): \\computer\object\counter
func extractCounterInfoFromCounterPath(counterPath string) (object string, instance string, counter string, err error) {

	rightObjectBorderIndex := -1
	leftObjectBorderIndex := -1
	leftCounterBorderIndex := -1
	rightInstanceBorderIndex := -1
	leftInstanceBorderIndex := -1
	bracketLevel := 0

	for i := len(counterPath) - 1; i >= 0; i-- {
		switch counterPath[i] {
		case '\\':
			if bracketLevel == 0 {
				if leftCounterBorderIndex == -1 {
					leftCounterBorderIndex = i
				} else if leftObjectBorderIndex == -1 {
					leftObjectBorderIndex = i
				}
			}
		case '(':
			bracketLevel--
			if leftInstanceBorderIndex == -1 && bracketLevel == 0 && leftObjectBorderIndex == -1 && leftCounterBorderIndex > -1 {
				leftInstanceBorderIndex = i
				rightObjectBorderIndex = i
			}
		case ')':
			if rightInstanceBorderIndex == -1 && bracketLevel == 0 && leftCounterBorderIndex > -1 {
				rightInstanceBorderIndex = i
			}
			bracketLevel++
		}
	}
	if rightObjectBorderIndex == -1 {
		rightObjectBorderIndex = leftCounterBorderIndex
	}
	if rightObjectBorderIndex == -1 || leftObjectBorderIndex == -1 {
		err = errors.New("cannot parse object from: " + counterPath)
		return
	}

	if leftInstanceBorderIndex > -1 && rightInstanceBorderIndex > -1 {
		instance = counterPath[leftInstanceBorderIndex+1 : rightInstanceBorderIndex]
	} else if (leftInstanceBorderIndex == -1 && rightInstanceBorderIndex > -1) || (leftInstanceBorderIndex > -1 && rightInstanceBorderIndex == -1) {
		err = errors.New("cannot parse instance from: " + counterPath)
		return
	}
	object = counterPath[leftObjectBorderIndex+1 : rightObjectBorderIndex]
	counter = counterPath[leftCounterBorderIndex+1:]
	return
}

func (m *Win_PerfCounters) Description() string {
	return "Input plugin to counterPath Performance Counters on Windows operating systems"
}

func (m *Win_PerfCounters) SampleConfig() string {
	return sampleConfig
}

//objectName string, counter string, instance string, measurement string, include_total bool
func (m *Win_PerfCounters) AddItem(counterPath string, objectName string, instance string, counterName string, measurement string, includeTotal bool) error {
	var err error
	var counterHandle PDH_HCOUNTER
	if !m.query.IsVistaOrNewer() {
		counterHandle, err = m.query.AddCounterToQuery(counterPath)
		if err != nil {
			return err
		}
	} else {
		counterHandle, err = m.query.AddEnglishCounterToQuery(counterPath)
		if err != nil {
			return err
		}

	}

	if m.UseWildcardsExpansion {
		origInstance := instance
		counterPath, err = m.query.GetCounterPath(counterHandle)
		if err != nil {
			return err
		}
		counters, err := m.query.ExpandWildCardPath(counterPath)
		if err != nil {
			return err
		}

		for _, counterPath := range counters {
			var err error
			counterHandle, err := m.query.AddCounterToQuery(counterPath)

			objectName, instance, counterName, err = extractCounterInfoFromCounterPath(counterPath)
			if err != nil {
				return err
			}

			if instance == "_Total" && origInstance == "*" && !includeTotal {
				continue
			}

			newItem := &counter{counterPath, objectName, counterName, instance, measurement,
				includeTotal, counterHandle}
			m.counters = append(m.counters, newItem)

			if m.PrintValid {
				log.Printf("Valid: %s\n", counterPath)
			}
		}
	} else {
		newItem := &counter{counterPath, objectName, counterName, instance, measurement,
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

					err := m.AddItem(counterPath, objectname, instance, counter, PerfObject.Measurement, PerfObject.IncludeTotal)

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
		if m.counters != nil {
			m.counters = m.counters[:0]
		}

		if err = m.query.Open(); err != nil {
			return err
		}

		if err = m.ParseConfig(); err != nil {
			return err
		}
		//some counters need two data samples before computing a value
		if err = m.query.CollectData(); err != nil {
			return err
		}
		m.lastRefreshed = time.Now()

		time.Sleep(time.Second)
	}

	var collectFields = make(map[instanceGrouping]map[string]interface{})

	var timestamp time.Time
	if m.UsePerfCounterTime && m.query.IsVistaOrNewer() {
		timestamp, err = m.query.CollectDataWithTime()
		if err != nil {
			return err
		}
	} else {
		timestamp = time.Now()
		if err = m.query.CollectData(); err != nil {
			return err
		}
	}

	// For iterate over the known metrics and get the samples.
	for _, metric := range m.counters {
		// collect
		if m.UseWildcardsExpansion {
			value, err := m.query.GetFormattedCounterValueDouble(metric.counterHandle)
			if err == nil {
				addCounterMeasurement(metric, metric.instance, value, collectFields)
			} else {
				//ignore invalid data  as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
				}
			}
		} else {
			counterValues, err := m.query.GetFormattedCounterArrayDouble(metric.counterHandle)
			if err == nil {
				for _, cValue := range counterValues {
					var add bool
					if metric.includeTotal {
						// If IncludeTotal is set, include all.
						add = true
					} else if metric.instance == "*" && !strings.Contains(cValue.InstanceName, "_Total") {
						// Catch if set to * and that it is not a '*_Total*' instance.
						add = true
					} else if metric.instance == cValue.InstanceName {
						// Catch if we set it to total or some form of it
						add = true
					} else if strings.Contains(metric.instance, "#") && strings.HasPrefix(metric.instance, cValue.InstanceName) {
						// If you are using a multiple instance identifier such as "w3wp#1"
						// phd.dll returns only the first 2 characters of the identifier.
						add = true
						cValue.InstanceName = metric.instance
					} else if metric.instance == "------" {
						add = true
					}

					if add {
						addCounterMeasurement(metric, cValue.InstanceName, cValue.Value, collectFields)
					}
				}
			} else {
				//ignore invalid data as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
				}
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
		acc.AddFields(instance.name, fields, tags, timestamp)
	}

	return nil
}

func addCounterMeasurement(metric *counter, instanceName string, value float64, collectFields map[instanceGrouping]map[string]interface{}) {
	measurement := sanitizedChars.Replace(metric.measurement)
	if measurement == "" {
		measurement = "win_perf_counters"
	}
	var instance = instanceGrouping{measurement, instanceName, metric.objectName}
	if collectFields[instance] == nil {
		collectFields[instance] = make(map[string]interface{})
	}
	collectFields[instance][sanitizedChars.Replace(metric.counter)] = float32(value)
}

func isKnownCounterDataError(err error) bool {
	if pdhErr, ok := err.(*PdhError); ok && (pdhErr.ErrorCode == PDH_INVALID_DATA ||
		pdhErr.ErrorCode == PDH_CALC_NEGATIVE_VALUE ||
		pdhErr.ErrorCode == PDH_CSTATUS_INVALID_DATA ||
		pdhErr.ErrorCode == PDH_NO_DATA) {
		return true
	}
	return false
}

func init() {
	inputs.Add("win_perf_counters", func() telegraf.Input {
		return &Win_PerfCounters{query: &PerformanceQueryImpl{}, CountersRefreshInterval: internal.Duration{Duration: time.Second * 60}}
	})
}
