//go:build windows
// +build windows

package win_wmi

import (
	_ "embed"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Query struct
type Query struct {
	query                string
	Namespace            string   `toml:"namespace"`
	ClassName            string   `toml:"class_name"`
	Properties           []string `toml:"properties"`
	Filter               string   `toml:"filter"`
	TagPropertiesInclude []string `toml:"tag_properties"`
	tagFilter            filter.Filter
}

// Wmi struct
type Wmi struct {
	Queries []Query `toml:"query"`
	Log     telegraf.Logger
}

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
const sFalse = 0x00000001

func oleInt64(item *ole.IDispatch, prop string) (int64, error) {
	v, err := oleutil.GetProperty(item, prop)
	if err != nil {
		return 0, err
	}
	defer func() { _ = v.Clear() }()

	i := v.Val
	return i, nil
}

// Init function
func (s *Wmi) Init() error {
	if err := CompileInputs(s); err != nil {
		return err
	}
	return nil
}

// Description function
func (s *Wmi) Description() string {
	return "returns WMI query results as metrics"
}

// SampleConfig function
func (s *Wmi) SampleConfig() string {
	return sampleConfig
}

func CompileInputs(s *Wmi) error {
	BuildWqlStatements(s)
	if err := CompileTagFilters(s); err != nil {
		return err
	}
	return nil
}

func CompileTagFilters(s *Wmi) error {
	var err error
	for i, q := range s.Queries {
		s.Queries[i].tagFilter, err = CompileTagFilter(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func CompileTagFilter(q Query) (filter.Filter, error) {
	tagFilter, err := filter.NewIncludeExcludeFilterDefaults(q.TagPropertiesInclude, nil, false, false)
	if err != nil {
		return nil, fmt.Errorf("creating tag filter failed: %v", err)
	}
	return tagFilter, nil
}

func BuildWqlStatements(s *Wmi) {
	for i, q := range s.Queries {
		s.Queries[i].query = BuildWqlStatement(q)
	}
}

// build a WMI query from input configuration
func BuildWqlStatement(q Query) string {
	wql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.Properties, ", "), q.ClassName)
	if len(q.Filter) > 0 {
		wql = fmt.Sprintf("%s WHERE %s", wql, q.Filter)
	}
	return wql
}

func (q *Query) DoQuery(acc telegraf.Accumulator) error {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tags := map[string]string{}
	fields := map[string]interface{}{}

	// init COM
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != sFalse {
			return err
		}
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return err
	} else if unknown == nil {
		return fmt.Errorf("failed to create WbemScripting.SWbemLocator, is WMI broken?")
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("failed to QueryInterface: %v", err)
	}
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, q.Namespace)
	if err != nil {
		return fmt.Errorf("failed calling method ConnectServer: %v", err)
	}
	service := serviceRaw.ToIDispatch()
	defer func() { _ = serviceRaw.Clear() }()

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", q.query)
	if err != nil {
		return fmt.Errorf("failed calling method ExecQuery for query %s: %v", q.query, err)
	}
	result := resultRaw.ToIDispatch()
	defer func() { _ = resultRaw.Clear() }()

	count, err := oleInt64(result, "Count")
	if err != nil {
		return fmt.Errorf("failed getting Count: %v", err)
	}

	for i := int64(0); i < count; i++ {
		// item is a SWbemObject
		itemRaw, err := oleutil.CallMethod(result, "ItemIndex", i)
		if err != nil {
			return fmt.Errorf("failed calling method ItemIndex: %v", err)
		}

		outerErr := func() error {
			item := itemRaw.ToIDispatch()
			defer item.Release()

			for _, wmiProperty := range q.Properties {
				innerErr := func() error {
					prop, err := oleutil.GetProperty(item, wmiProperty)
					if err != nil {
						return fmt.Errorf("failed GetProperty: %v", err)
					}
					defer func() { _ = prop.Clear() }()

					if q.tagFilter.Match(wmiProperty) {
						valStr, err := internal.ToString(prop.Value())
						if err != nil {
							return fmt.Errorf("converting property %q failed: %v", wmiProperty, err)
						}
						tags[wmiProperty] = valStr
					} else {
						var fieldValue interface{}
						switch v := prop.Value().(type) {
						case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
							fieldValue = v
						case string:
							// still might be an int because WMI
							valInt, err := strconv.ParseInt(v, 10, 64)
							if err == nil {
								fieldValue = valInt
							} else {
								fieldValue = v
							}
						case bool:
							fieldValue = v
						case []byte:
							fieldValue = string(v)
						case fmt.Stringer:
							fieldValue = v.String()
						case nil:
							fieldValue = nil
						default:
							return fmt.Errorf("property %q of type \"%T\" unsupported", wmiProperty, v)
						}
						fields[wmiProperty] = fieldValue
					}

					return nil
				}()

				if innerErr != nil {
					return innerErr
				}
			}
			acc.AddFields(q.ClassName, fields, tags)
			return nil
		}()

		if outerErr != nil {
			return outerErr
		}
	}
	return nil
}

// Gather function
func (s *Wmi) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, query := range s.Queries {
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			start := time.Now()
			err := q.DoQuery(acc)
			if err != nil {
				acc.AddError(err)
			}
			elapsed := time.Since(start)
			s.Log.Debugf("Query \"%s\" took %s", q.query, elapsed)
		}(query)
	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
