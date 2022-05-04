//go:build windows
// +build windows

package win_perf_counters

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Win_PerfCounters struct {
	PrintValid                 bool `toml:"PrintValid"`
	PreVistaSupport            bool `toml:"PreVistaSupport" deprecated:"1.7.0;determined dynamically"`
	UsePerfCounterTime         bool
	Object                     []perfobject
	CountersRefreshInterval    config.Duration
	UseWildcardsExpansion      bool
	LocalizeWildcardsExpansion bool
	IgnoredErrors              []string `toml:"IgnoredErrors"`

	Log telegraf.Logger

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
	UseRawValues  bool
}

type counter struct {
	counterPath   string
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

func newCounter(counterHandle PDH_HCOUNTER, counterPath string, objectName string, instance string, counterName string, measurement string, includeTotal bool, useRawValue bool) *counter {
	measurementName := sanitizedChars.Replace(measurement)
	if measurementName == "" {
		measurementName = "win_perf_counters"
	}
	newCounterName := sanitizedChars.Replace(counterName)
	if useRawValue {
		newCounterName += "_Raw"
	}
	return &counter{counterPath, objectName, newCounterName, instance, measurementName,
		includeTotal, useRawValue, counterHandle}
}

func (m *Win_PerfCounters) AddItem(counterPath string, objectName string, instance string, counterName string, measurement string, includeTotal bool, useRawValue bool) error {
	origCounterPath := counterPath
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

		origObjectName, _, origCounterName, err := extractCounterInfoFromCounterPath(origCounterPath)
		if err != nil {
			return err
		}

		for _, counterPath := range counters {
			var err error

			objectName, instance, counterName, err = extractCounterInfoFromCounterPath(counterPath)
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
				counterPath = formatPath(origObjectName, newInstance, origCounterName)
				counterHandle, err = m.query.AddEnglishCounterToQuery(counterPath)
				newItem = newCounter(
					counterHandle,
					counterPath,
					origObjectName, instance,
					origCounterName,
					measurement,
					includeTotal,
					useRawValue,
				)
			} else {
				counterHandle, err = m.query.AddCounterToQuery(counterPath)
				newItem = newCounter(
					counterHandle,
					counterPath,
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

			m.counters = append(m.counters, newItem)

			if m.PrintValid {
				m.Log.Infof("Valid: %s", counterPath)
			}
		}
	} else {
		newItem := newCounter(
			counterHandle,
			counterPath,
			objectName,
			instance,
			counterName,
			measurement,
			includeTotal,
			useRawValue,
		)
		m.counters = append(m.counters, newItem)
		if m.PrintValid {
			m.Log.Infof("Valid: %s", counterPath)
		}
	}

	return nil
}

const emptyInstance = "------"

func formatPath(objectname string, instance string, counter string) string {
	if instance == emptyInstance {
		return "\\" + objectname + "\\" + counter
	} else {
		return "\\" + objectname + "(" + instance + ")\\" + counter
	}
}

func (m *Win_PerfCounters) ParseConfig() error {
	var counterPath string

	if len(m.Object) > 0 {
		for _, PerfObject := range m.Object {
			for _, counter := range PerfObject.Counters {
				if len(PerfObject.Instances) == 0 {
					m.Log.Warnf("Missing 'Instances' param for object '%s'\n", PerfObject.ObjectName)
				}
				for _, instance := range PerfObject.Instances {
					objectname := PerfObject.ObjectName

					counterPath = formatPath(objectname, instance, counter)

					err := m.AddItem(counterPath, objectname, instance, counter, PerfObject.Measurement, PerfObject.IncludeTotal, PerfObject.UseRawValues)

					if err != nil {
						if PerfObject.FailOnMissing || PerfObject.WarnOnMissing {
							m.Log.Errorf("Invalid counterPath: '%s'. Error: %s\n", counterPath, err.Error())
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

func (m *Win_PerfCounters) checkError(err error) error {
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

func (m *Win_PerfCounters) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	var err error

	if m.lastRefreshed.IsZero() || (m.CountersRefreshInterval > 0 && m.lastRefreshed.Add(time.Duration(m.CountersRefreshInterval)).Before(time.Now())) {
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
			return m.checkError(err)
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
	var value interface{}
	// For iterate over the known metrics and get the samples.
	for _, metric := range m.counters {
		// collect
		if m.UseWildcardsExpansion {
			if metric.useRawValue {
				value, err = m.query.GetRawCounterValue(metric.counterHandle)
			} else {
				value, err = m.query.GetFormattedCounterValueDouble(metric.counterHandle)
			}
			if err != nil {
				//ignore invalid data  as some counters from process instances returns this sometimes
				if !isKnownCounterDataError(err) {
					return fmt.Errorf("error while getting value for counter %s: %v", metric.counterPath, err)
				}
				m.Log.Warnf("error while getting value for counter %q, will skip metric: %v", metric.counterPath, err)
				continue
			}
			addCounterMeasurement(metric, metric.instance, value, collectFields)
		} else {
			var counterValues []CounterValue
			if metric.useRawValue {
				counterValues, err = m.query.GetRawCounterArray(metric.counterHandle)
			} else {
				counterValues, err = m.query.GetFormattedCounterArrayDouble(metric.counterHandle)
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
					addCounterMeasurement(metric, cValue.InstanceName, cValue.Value, collectFields)
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

func addCounterMeasurement(metric *counter, instanceName string, value interface{}, collectFields map[instanceGrouping]map[string]interface{}) {
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

func (m *Win_PerfCounters) Init() error {
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
		return &Win_PerfCounters{
			query:                      &PerformanceQueryImpl{},
			CountersRefreshInterval:    config.Duration(time.Second * 60),
			LocalizeWildcardsExpansion: true,
		}
	})
}
