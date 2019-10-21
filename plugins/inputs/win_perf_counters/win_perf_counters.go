//go:build windows
// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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
  # When running on a localized version of Windows and with UseWildcardsExpansion = true, Windows will
  # localize object and counter names. When LocalizeWildcardsExpansion = false, use the names in object.Counters instead
  # of the localized names. Only Instances can have wildcards in this case. ObjectName and Counters must not have wildcards when this
  # setting is false.
  #LocalizeWildcardsExpansion = true
  # Period after which counters will be reread from configuration and wildcards in counter paths expanded
  CountersRefreshInterval="1m"
  ## Accepts a list of PDH error codes which are defined in pdh.go, if this error is encountered it will be ignored
  ## For example, you can provide "PDH_NO_DATA" to ignore performance counters with no instances
  ## By default no errors are ignored
  ## You can find the list here: https://github.com/influxdata/telegraf/blob/master/plugins/inputs/win_perf_counters/pdh.go
  ## e.g.: IgnoredErrors = ["PDH_NO_DATA"]
  # IgnoredErrors = []
  # Names or ip addresses of remote computers to gather counters from, including local computer.
  # Telegraf's user must be already authenticated to the remote computers.
  # It can be overridden at the object level
  # Sources = ["localhost"]


  [[inputs.win_perf_counters.object]]
    # Processor usage, alternative to native, reports on a per core.
    ObjectName = "Processor"
    Instances = ["*"]
    Counters = [
      "% Idle Time",
      "% Interrupt Time",
      "% Privileged Time",
      "% User Time",
      "% Processor Time",
      "% DPC Time",
    ]
    Measurement = "win_cpu"
    # Set to true to include _Total instance when querying for all (*).
    # IncludeTotal=false
    # Print out when the performance counter is missing from object, counter or instance.
    # WarnOnMissing = false
    # Gather raw values instead of formatted. Raw value is stored in the field name with the "_Raw" suffix, e.g. "Disk_Read_Bytes_sec_Raw".
    # UseRawValues = true
    # Overrides the Sources global parameter for current performance object.
    # Sources = ["localhost", "SQL-server1"]

  [[inputs.win_perf_counters.object]]
    # Disk times and queues
    ObjectName = "LogicalDisk"
    Instances = ["*"]
    Counters = [
      "% Idle Time",
      "% Disk Time",
      "% Disk Read Time",
      "% Disk Write Time",
      "% User Time",
      "% Free Space",
      "Current Disk Queue Length",
      "Free Megabytes",
    ]
    Measurement = "win_disk"

  [[inputs.win_perf_counters.object]]
    ObjectName = "PhysicalDisk"
    Instances = ["*"]
    Counters = [
      "Disk Read Bytes/sec",
      "Disk Write Bytes/sec",
      "Current Disk Queue Length",
      "Disk Reads/sec",
      "Disk Writes/sec",
      "% Disk Time",
      "% Disk Read Time",
      "% Disk Write Time",
    ]
    Measurement = "win_diskio"

  [[inputs.win_perf_counters.object]]
    ObjectName = "Network Interface"
    Instances = ["*"]
    Counters = [
      "Bytes Received/sec",
      "Bytes Sent/sec",
      "Packets Received/sec",
      "Packets Sent/sec",
      "Packets Received Discarded",
      "Packets Outbound Discarded",
      "Packets Received Errors",
      "Packets Outbound Errors",
    ]
    Measurement = "win_net"


  [[inputs.win_perf_counters.object]]
    ObjectName = "System"
    Counters = [
      "Context Switches/sec",
      "System Calls/sec",
      "Processor Queue Length",
      "System Up Time",
    ]
    Instances = ["------"]
    Measurement = "win_system"

  [[inputs.win_perf_counters.object]]
    # Example counterPath where the Instance portion must be removed to get data back,
    # such as from the Memory object.
    ObjectName = "Memory"
    Counters = [
      "Available Bytes",
      "Cache Faults/sec",
      "Demand Zero Faults/sec",
      "Page Faults/sec",
      "Pages/sec",
      "Transition Faults/sec",
      "Pool Nonpaged Bytes",
      "Pool Paged Bytes",
      "Standby Cache Reserve Bytes",
      "Standby Cache Normal Priority Bytes",
      "Standby Cache Core Bytes",
    ]
    Instances = ["------"] # Use 6 x - to remove the Instance bit from the counterPath.
    Measurement = "win_mem"

  [[inputs.win_perf_counters.object]]
    # Example query where the Instance portion must be removed to get data back,
    # such as from the Paging File object.
    ObjectName = "Paging File"
    Counters = [
      "% Usage",
    ]
    Instances = ["_Total"]
    Measurement = "win_swap"
`

type WinPerfCounters struct {
	PrintValid                 bool `toml:"PrintValid"`
	PreVistaSupport            bool `toml:"PreVistaSupport" deprecated:"1.7.0;determined dynamically"`
	UsePerfCounterTime         bool
	Object                     []perfobject
	CountersRefreshInterval    config.Duration
	UseWildcardsExpansion      bool
	LocalizeWildcardsExpansion bool
	IgnoredErrors              []string `toml:"IgnoredErrors"`
	Sources                    []string

	Log telegraf.Logger

	lastRefreshed time.Time
	queryCreator  PerformanceQueryCreator
	hostCounters  map[string]*hostCountersInfo
	// cached os.Hostname()
	cachedHostname string
}

type hostCountersInfo struct {
	// computer name used as key and for printing
	computer string
	// computer name used in tag
	tag       string
	counters  []*counter
	query     PerformanceQuery
	timestamp time.Time
}

type perfobject struct {
	Sources       []string
	ObjectName    string
	Counters      []string
	Instances     []string
	Measurement   string
	WarnOnMissing bool
	FailOnMissing bool
	IncludeTotal  bool
	UseRawValues  bool
}

type counter struct {
	counterPath   string
	computer      string
	objectName    string
	counter       string
	instance      string
	measurement   string
	includeTotal  bool
	useRawValue   bool
	counterHandle PDH_HCOUNTER
}

type instanceGrouping struct {
	name       string
	instance   string
	objectname string
}

type fieldGrouping map[instanceGrouping]map[string]interface{}

var sanitizedChars = strings.NewReplacer("/sec", "_persec", "/Sec", "_persec",
	" ", "_", "%", "Percent", `\`, "")

// extractCounterInfoFromCounterPath gets object name, instance name (if available) and counter name from counter path
// General Counter path pattern is: \\computer\object(parent/instance#index)\counter
// parent/instance#index part is skipped in single instance objects (e.g. Memory): \\computer\object\counter
func extractCounterInfoFromCounterPath(counterPath string) (computer, object, instance, counter string, err error) {
	computer = ""
	leftComputerBorderIndex := -1
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
				} else if leftComputerBorderIndex == -1 {
					leftComputerBorderIndex = i
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
	if leftComputerBorderIndex > -1 {
		// validate there is leading \\ and not empty computer (\\\O)
		if leftComputerBorderIndex != 1 || leftComputerBorderIndex == leftObjectBorderIndex-1 {
			err = errors.New("cannot parse computer from: " + counterPath)
			return
		} else {
			computer = counterPath[leftComputerBorderIndex+1 : leftObjectBorderIndex]
		}
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

func (m *WinPerfCounters) Description() string {
	return "Input plugin to counterPath Performance Counters on Windows operating systems"
}

func (m *WinPerfCounters) SampleConfig() string {
	return sampleConfig
}

func (m *WinPerfCounters) hostname() string {
	if m.cachedHostname == "" {
		hostname, err := os.Hostname()
		if err == nil {
			m.cachedHostname = hostname
		} else {
			m.cachedHostname = "localhost"
		}
	}
	return m.cachedHostname
}

func newCounter(counterHandle PDH_HCOUNTER, counterPath string, computer, objectName string, instance string, counterName string, measurement string, includeTotal bool, useRawValue bool) *counter {
	measurementName := sanitizedChars.Replace(measurement)
	if measurementName == "" {
		measurementName = "win_perf_counters"
	}
	newCounterName := sanitizedChars.Replace(counterName)
	if useRawValue {
		newCounterName += "_Raw"
	}
	return &counter{counterPath, computer, objectName, newCounterName, instance, measurementName,
		includeTotal, useRawValue, counterHandle}
}

func (m *WinPerfCounters) AddItem(counterPath, computer, objectName, instance, counterName, measurement string, includeTotal bool, useRawValue bool) error {
	origCounterPath := counterPath
	var err error
	var counterHandle PDH_HCOUNTER
	var hostCounter *hostCountersInfo
	var ok bool
	sourceTag := computer
	if computer == "localhost" {
		sourceTag = m.hostname()
	}
	if m.hostCounters == nil {
		m.hostCounters = make(map[string]*hostCountersInfo)
	}
	if hostCounter, ok = m.hostCounters[computer]; !ok {
		hostCounter = &hostCountersInfo{computer: computer, tag: sourceTag}
		m.hostCounters[computer] = hostCounter
		hostCounter.query = m.queryCreator.NewPerformanceQuery(computer)
		if err = hostCounter.query.Open(); err != nil {
			return err
		}
		hostCounter.counters = make([]*counter, 0, 0)
	}

	if !hostCounter.query.IsVistaOrNewer() {
		counterHandle, err = hostCounter.query.AddCounterToQuery(counterPath)
		if err != nil {
			return err
		}
	} else {
		counterHandle, err = hostCounter.query.AddEnglishCounterToQuery(counterPath)
		if err != nil {
			return err
		}

	}

	if m.UseWildcardsExpansion {
		origInstance := instance
		counterPath, err = hostCounter.query.GetCounterPath(counterHandle)
		if err != nil {
			return err
		}
		counters, err := hostCounter.query.ExpandWildCardPath(counterPath)
		if err != nil {
			return err
		}

		_, origObjectName, _, origCounterName, err := extractCounterInfoFromCounterPath(origCounterPath)
		if err != nil {
			return err
		}

		for _, counterPath := range counters {
			var err error
			counterHandle, err := hostCounter.query.AddCounterToQuery(counterPath)

			computer, objectName, instance, counterName, err = extractCounterInfoFromCounterPath(counterPath)
			if err != nil {
				return err
			}

			var newItem *counter
			if !m.LocalizeWildcardsExpansion {
				// On localized installations of Windows, Telegraf
				// should return English metrics, but
				// ExpandWildCardPath returns localized counters. Undo
				// that by using the original object and counter
				// names, along with the expanded instance.

				var newInstance string
				if instance == "" {
					newInstance = emptyInstance
				} else {
					newInstance = instance
				}
				counterPath = formatPath(computer, origObjectName, newInstance, origCounterName)
				counterHandle, err = hostCounter.query.AddEnglishCounterToQuery(counterPath)
				if err != nil {
					return err
				}
				newItem = newCounter(
					counterHandle,
					counterPath,
					computer,
					origObjectName, instance,
					origCounterName,
					measurement,
					includeTotal,
					useRawValue,
				)
			} else {
				counterHandle, err = hostCounter.query.AddCounterToQuery(counterPath)
				if err != nil {
					return err
				}
				newItem = newCounter(
					counterHandle,
					counterPath,
					computer,
					objectName,
					instance,
					counterName,
					measurement,
					includeTotal,
					useRawValue,
				)
			}

			if instance == "_Total" && origInstance == "*" && !includeTotal {
				continue
			}

			hostCounter.counters = append(hostCounter.counters, newItem)

			if m.PrintValid {
				m.Log.Infof("Valid: %s", counterPath)
			}
		}
	} else {
		newItem := newCounter(
			counterHandle,
			counterPath,
			computer,
			objectName,
			instance,
			counterName,
			measurement,
			includeTotal,
			useRawValue,
		)
		hostCounter.counters = append(hostCounter.counters, newItem)
		if m.PrintValid {
			m.Log.Infof("Valid: %s", counterPath)
		}
	}

	return nil
}

const emptyInstance = "------"

func formatPath(computer, objectname, instance, counter string) string {
	path := ""
	if instance == emptyInstance {
		path = fmt.Sprintf(`\%s\%s`, objectname, counter)
	} else {
		path = fmt.Sprintf(`\%s(%s)\%s`, objectname, instance, counter)
	}
	if computer != "" && computer != "localhost" {
		path = fmt.Sprintf(`\\%s%s`, computer, path)
	}
	return path
}

func (m *WinPerfCounters) ParseConfig() error {
	var counterPath string

	if len(m.Sources) == 0 {
		m.Sources = []string{"localhost"}
	}

	if len(m.Object) > 0 {
		for _, PerfObject := range m.Object {
			computers := PerfObject.Sources
			if len(computers) == 0 {
				computers = m.Sources
			}
			for _, computer := range computers {
				if computer == "" {
					// localhost as a computer name in counter path doesn't work
					computer = "localhost"
				}
				for _, counter := range PerfObject.Counters {
					if len(PerfObject.Instances) == 0 {
						m.Log.Warnf("Missing 'Instances' param for object '%s'\n", PerfObject.ObjectName)
					}
					for _, instance := range PerfObject.Instances {
						objectname := PerfObject.ObjectName

						counterPath = formatPath(computer, objectname, instance, counter)

						err := m.AddItem(counterPath, computer, objectname, instance, counter, PerfObject.Measurement, PerfObject.IncludeTotal, PerfObject.UseRawValues)
						if err != nil {
							if PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
								m.Log.Errorf("invalid counterPath: '%s'. Error: %s\n", counterPath, err.Error())
							}
							if PerfObject.FailOnMissing {
								return err
							}
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

func (m *WinPerfCounters) checkError(err error) error {
	if pdhErr, ok := err.(*PdhError); ok {
		for _, ignoredErrors := range m.IgnoredErrors {
			if PDHErrors[pdhErr.ErrorCode] == ignoredErrors {
				return nil
			}
		}

		return err
	}
	return err
}

func (m *WinPerfCounters) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	var err error

	if m.lastRefreshed.IsZero() || (m.CountersRefreshInterval > 0 && m.lastRefreshed.Add(time.Duration(m.CountersRefreshInterval)).Before(time.Now())) {
		if err = m.cleanQueries(); err != nil {
			return err
		}

		if err = m.ParseConfig(); err != nil {
			return err
		}
		for _, hostCounterSet := range m.hostCounters {
			//some counters need two data samples before computing a value
			if err = hostCounterSet.query.CollectData(); err != nil {
				return m.checkError(err)
			}
		}
		m.lastRefreshed = time.Now()
		// minimum time between collecting two samples
		time.Sleep(time.Second)
	}

	for _, hostCounterSet := range m.hostCounters {
		if m.UsePerfCounterTime && hostCounterSet.query.IsVistaOrNewer() {
			hostCounterSet.timestamp, err = hostCounterSet.query.CollectDataWithTime()
			if err != nil {
				return err
			}
		} else {
			hostCounterSet.timestamp = time.Now()
			if err = hostCounterSet.query.CollectData(); err != nil {
				return err
			}
		}
	}
	var wg sync.WaitGroup
	//iterate over computers
	for _, hostCounterInfo := range m.hostCounters {
		wg.Add(1)
		go func(hostInfo *hostCountersInfo) {
			m.Log.Debugf("gathering from %s", hostInfo.computer)
			start := time.Now()
			err := m.gatherComputerCounters(hostInfo, acc)
			m.Log.Debugf("gathering from %s finished in %.3fs", hostInfo.computer, time.Now().Sub(start).Seconds())
			if err != nil {
				acc.AddError(fmt.Errorf("error during collecting data on host '%s': %s", hostInfo.computer, err.Error()))
			}
			wg.Done()
		}(hostCounterInfo)
	}

	wg.Wait()
	return nil
}

func (m *WinPerfCounters) gatherComputerCounters(hostCounterInfo *hostCountersInfo, acc telegraf.Accumulator) error {
	var value interface{}
	var err error
	collectedFields := make(fieldGrouping)
	// For iterate over the known metrics and get the samples.
	for _, metric := range hostCounterInfo.counters {
		// collect
		if m.UseWildcardsExpansion {
			if metric.useRawValue {
				value, err = hostCounterInfo.query.GetRawCounterValue(metric.counterHandle)
			} else {
				value, err = hostCounterInfo.query.GetFormattedCounterValueDouble(metric.counterHandle)
			}
			if err != nil {
				//ignore invalid data  as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
				}
				m.Log.Warnf("error while getting value for counter %q, will skip metric: %v", metric.counterPath, err)
				continue
			}
			addCounterMeasurement(metric, metric.instance, value, collectedFields)
		} else {
			var counterValues []CounterValue
			if metric.useRawValue {
				counterValues, err = hostCounterInfo.query.GetRawCounterArray(metric.counterHandle)
			} else {
				counterValues, err = hostCounterInfo.query.GetFormattedCounterArrayDouble(metric.counterHandle)
			}
			if err != nil {
				//ignore invalid data  as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
				}
				m.Log.Warnf("error while getting value for counter %q, will skip metric: %v", metric.counterPath, err)
				continue
			}
			for _, cValue := range counterValues {

				if strings.Contains(metric.instance, "#") && strings.HasPrefix(metric.instance, cValue.InstanceName) {
					// If you are using a multiple instance identifier such as "w3wp#1"
					// phd.dll returns only the first 2 characters of the identifier.
					cValue.InstanceName = metric.instance
				}

				if shouldIncludeMetric(metric, cValue) {
					addCounterMeasurement(metric, cValue.InstanceName, cValue.Value, collectedFields)
				}
			}
		}
	}
	for instance, fields := range collectedFields {
		var tags = map[string]string{
			"objectname": instance.objectname,
		}
		if len(instance.instance) > 0 {
			tags["instance"] = instance.instance
		}
		if len(hostCounterInfo.tag) > 0 {
			tags["source"] = hostCounterInfo.tag
		}
		acc.AddFields(instance.name, fields, tags, hostCounterInfo.timestamp)
	}
	return nil

}

func (m *WinPerfCounters) cleanQueries() error {
	for _, hostCounterInfo := range m.hostCounters {
		if err := hostCounterInfo.query.Close(); err != nil {
			return err
		}
	}
	m.hostCounters = nil
	return nil
}

func shouldIncludeMetric(metric *counter, cValue CounterValue) bool {
	if metric.includeTotal {
		// If IncludeTotal is set, include all.
		return true
	}
	if metric.instance == "*" && !strings.Contains(cValue.InstanceName, "_Total") {
		// Catch if set to * and that it is not a '*_Total*' instance.
		return true
	}
	if metric.instance == cValue.InstanceName {
		// Catch if we set it to total or some form of it
		return true
	}
	if metric.instance == emptyInstance {
		return true
	}
	return false
}

func addCounterMeasurement(metric *counter, instanceName string, value interface{}, collectFields fieldGrouping) {
	var instance = instanceGrouping{metric.measurement, instanceName, metric.objectName}
	if collectFields[instance] == nil {
		collectFields[instance] = make(map[string]interface{})
	}
	collectFields[instance][sanitizedChars.Replace(metric.counter)] = value
}

func isKnownCounterDataError(err error) bool {
	if pdhErr, ok := err.(*PdhError); ok && (pdhErr.ErrorCode == PDH_INVALID_DATA ||
		pdhErr.ErrorCode == PDH_CALC_NEGATIVE_DENOMINATOR ||
		pdhErr.ErrorCode == PDH_CALC_NEGATIVE_VALUE ||
		pdhErr.ErrorCode == PDH_CSTATUS_INVALID_DATA ||
		pdhErr.ErrorCode == PDH_NO_DATA) {
		return true
	}
	return false
}

func (m *WinPerfCounters) Init() error {
	if m.UseWildcardsExpansion && !m.LocalizeWildcardsExpansion {
		// Counters must not have wildcards with this option

		found := false
		wildcards := []string{"*", "?"}

		for _, object := range m.Object {
			for _, wildcard := range wildcards {
				if strings.Contains(object.ObjectName, wildcard) {
					found = true
					m.Log.Errorf("object: %s, contains wildcard %s", object.ObjectName, wildcard)
				}
			}
			for _, counter := range object.Counters {
				for _, wildcard := range wildcards {
					if strings.Contains(counter, wildcard) {
						found = true
						m.Log.Errorf("object: %s, counter: %s contains wildcard %s", object.ObjectName, counter, wildcard)
					}
				}
			}
		}

		if found {
			return fmt.Errorf("wildcards can't be used with LocalizeWildcardsExpansion=false")
		}
	}
	return nil
}

func init() {
	inputs.Add("win_perf_counters", func() telegraf.Input {
		return &WinPerfCounters{
			CountersRefreshInterval:    config.Duration(time.Second * 60),
			LocalizeWildcardsExpansion: true,
			queryCreator:               &PerformanceQueryCreatorImpl{},
		}
	})
}
