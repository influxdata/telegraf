// +build windows

package wpc

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	lxn "github.com/lxn/win"
)

var sampleConfig string = `
  ## A plugin to collect stats from Windows Performance Counters.
  ## If the system being polled for data does not have a particular Counter at startup 
  ## of the Telegraf agent, it will not be gathered.
  # Prints all matching performance counters (useful for debugging)
  # PrintValid = false

  [[inputs.wpc.template]]
    # Processor usage, alternative to native.
    Counters = [
      [ "usage_idle", "\\Processor(*)\\%% Idle Time" ],
      [ "usage_user", "\\Processor(*)\\%% User Time" ],
      [ "usage_system", "\\Processor(*)\\%% Processor Time" ]
    ]
    Measurement = "win_cpu"
    # Print out when the performance counter is missing from object, counter or instance.
    # WarnOnMissing = false


  [[inputs.wpc.template]]
    # Disk times and queues
    Counters = [
      [ "usage_idle", "\\LogicalDisk(*)\\%% Idle Time" ],
      [ "usage_used", "\\LogicalDisk(*)\\%% Disk Time" ],
      [ "usage_read", "\\LogicalDisk(*)\\%% Disk Read Time" ],
      [ "usage_write", "\\LogicalDisk(*)\\%% Disk Write Time" ], 
      [ "usage_user", "\\LogicalDisk(*)\\%% User Time" ],
      [ "qcur", "\\LogicalDisk(*)\\Current Disk Queue Length" ]
    ]
    Measurement = "win_diskio"
    # Print out when the performance counter is missing from object, counter or instance.
    # WarnOnMissing = false

  [[inputs.wpc.template]]
    # System and memory details
    Counters = [
      [ "cs_rate", "\\System\\Context Switches/sec" ],
      [ "syscall_rate", "\\System\\System Calls/sec" ],
      [ "mem_available", "\\Memory\\Available Bytes" ]
    ]
    Measurement = "win_system"
    # Print out when the performance counter is missing from object, counter or instance.
    # WarnOnMissing = false
`

type WindowsPerformanceCounter struct {
	PrintValid      bool
	TestName        string
	PreVistaSupport bool
	Template        []template
}

type template struct {
	Counters      [][]string
	Measurement   string
	WarnOnMissing bool
	FailOnMissing bool
}

type task struct {
	measurement string
	fields      map[string]*counter
}

type counter struct {
	query         string
	handle        lxn.PDH_HQUERY
	counterHandle lxn.PDH_HCOUNTER
	current       map[string]float32
}

// Globals
var (
	gConfigParsed bool

	// Parsed configuration ends up here after it has been validated
	gTaskList []*task

	// Counter cache to avoid gathering the same counter more than once per Gather
	gCounterCache = make(map[string]*counter)

	// Various error messages
	errBadConfig        error = errors.New("inputs.wpc.series contains invalid configuration")
	errObjectNotExist   error = errors.New("Performance object does not exist")
	errCounterNotExist  error = errors.New("Counter in Performance object does not exist")
	errInstanceNotExist error = errors.New("Instance in Performance object does not exist")
	errBadQuery         error = errors.New("Invalid query for Performance Counters")

	// Used to cleanup instance names
	sanitizedChars = strings.NewReplacer("/sec", "_persec", "/Sec", "_persec", " ", "_", "%", "Percent", `\`, "", ",", "_")
)

func (m *WindowsPerformanceCounter) Description() string {
	return "Input plugin to query Performance Counters on Windows operating systems"
}

func (m *WindowsPerformanceCounter) SampleConfig() string {
	return sampleConfig
}

func (m *WindowsPerformanceCounter) Gather(acc telegraf.Accumulator) error {
	// We only need to parse the config during the init, it uses the global variable after.
	if gConfigParsed == false {
		err := m.parseConfig()
		gConfigParsed = true
		if err != nil {
			return err
		}
	}

	// Sample counters
	for _, c := range gCounterCache {
		if ok := c.queryPerformanceCounter(); !ok {
			continue
		}
	}

	type grouping struct {
		fields map[string]interface{}
		tags   map[string]string
	}

	for _, t := range gTaskList {
		groups := make(map[string]*grouping)

		// Regroup samples by (measurement, instance) to minimize points generated.
		for field, c := range t.fields {
			for instance, f32 := range c.current {
				g, ok := groups[instance]
				if !ok {
					g = &grouping{
						tags:   make(map[string]string),
						fields: make(map[string]interface{})}
					g.tags["instance"] = sanitizedChars.Replace(instance)
					groups[instance] = g
				}

				g.fields[field] = f32
			}
		}

		for _, g := range groups {
			acc.AddFields(t.measurement, g.fields, g.tags)
		}
	}

	return nil
}

// Samples (instance, value) tuples from the performance counter
func (c *counter) queryPerformanceCounter() (ok bool) {
	if ret := lxn.PdhCollectQueryData(c.handle); ret != lxn.ERROR_SUCCESS {
		return false
	}

	var bufSize uint32
	var bufCount uint32
	var size uint32 = uint32(unsafe.Sizeof(lxn.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE{}))
	var emptyBuf [1]lxn.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE // need at least 1 addressable null ptr.

	// uses null ptr here according to MSDN.
	if ret := lxn.PdhGetFormattedCounterArrayDouble(c.counterHandle, &bufSize, &bufCount, &emptyBuf[0]); ret != lxn.PDH_MORE_DATA || bufCount == 0 {
		return false
	}

	coll := make(map[string]float32)
	data := make([]lxn.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE, bufCount*size)

	_ = lxn.PdhGetFormattedCounterArrayDouble(c.counterHandle, &bufSize, &bufCount, &data[0])
	for i := 0; i < int(bufCount); i++ {
		res := data[i]
		instance := lxn.UTF16PtrToString(res.SzName)
		value := float32(res.FmtValue.DoubleValue)
		coll[instance] = value
	}

	c.current = coll
	return true
}

func (m *WindowsPerformanceCounter) validatePerformanceCounterPath(query string, onMissingWarn, onMissingFail bool) (ok bool, err error) {
	const (
		lxn_PDH_CSTATUS_NO_OBJECT   uint32 = 3221228472
		lxn_PDH_CSTATUS_NO_COUNTER  uint32 = 3221228473
		lxn_PDH_CSTATUS_NO_INSTANCE uint32 = 2147485649
	)

	exists := lxn.PdhValidatePath(query)
	if exists == lxn.ERROR_SUCCESS {
		if m.PrintValid {
			fmt.Printf("Valid: %s\n", query)
		}

		return true, nil
	} else if !onMissingWarn && !onMissingFail {
		return false, nil
	}

	switch exists {
	case lxn_PDH_CSTATUS_NO_OBJECT:
		if onMissingFail {
			return false, errObjectNotExist
		}

		fmt.Printf("Performance Object does not exist in query: %s\n", query)
		break

	case lxn_PDH_CSTATUS_NO_COUNTER:
		if onMissingFail {
			return false, errCounterNotExist
		}

		fmt.Printf("Counter does not exist in query: %s\n", query)
		break

	case lxn_PDH_CSTATUS_NO_INSTANCE:
		if onMissingFail {
			return false, errInstanceNotExist
		}

		fmt.Printf("Instance does not exist in query: %s\n", query)
		break

	default:
		fmt.Printf("Invalid result: %v, query: %s\n", exists, query)
		if onMissingFail {
			return false, errBadQuery
		}
		break
	}

	return false, nil
}

func (m *WindowsPerformanceCounter) openPerformanceCounter(query string) *counter {
	var handle lxn.PDH_HQUERY
	var counterHandle lxn.PDH_HCOUNTER

	ret := lxn.PdhOpenQuery(0, 0, &handle)
	if m.PreVistaSupport {
		ret = lxn.PdhAddCounter(handle, query, 0, &counterHandle)
	} else {
		ret = lxn.PdhAddEnglishCounter(handle, query, 0, &counterHandle)
	}
	_ = ret

	return &counter{query, handle, counterHandle, nil}
}

// Populates the global counter cache and task list.
func (m *WindowsPerformanceCounter) parseConfig() error {
	if len(m.Template) == 0 {
		err := errors.New("Nothing to do!")
		return err
	}

	for _, tmpl := range m.Template {
		t := &task{
			measurement: tmpl.Measurement,
			fields:      make(map[string]*counter),
		}

		for _, pair := range tmpl.Counters {
			if len(pair) != 2 {
				return errBadConfig
			}

			field := pair[0]
			query := pair[1]

			if _, ok := gCounterCache[query]; !ok {
				if ok, err := m.validatePerformanceCounterPath(query, tmpl.WarnOnMissing, tmpl.FailOnMissing); !ok && err != nil {
					return err
				} else if !ok {
					continue
				}

				gCounterCache[query] = m.openPerformanceCounter(query)
			}

			t.fields[field] = gCounterCache[query]
		}

		gTaskList = append(gTaskList, t)
	}

	return nil
}

// func (m *WindowsPerformanceCounter) cleanup(metrics *itemList) {
// 	// Cleanup

// 	for _, metric := range metrics.items {
// 		ret := lxn.PdhCloseQuery(metric.handle)
// 		_ = ret
// 	}
// }

func init() {
	inputs.Add("wpc", func() telegraf.Input { return &WindowsPerformanceCounter{} })
}
