//go:build windows
// +build windows

package win_wmi

import (
	_ "embed"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/go-ole/go-ole"
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
	defer v.Clear()

	return v.Val, nil
}

// Init function
func (s *Wmi) Init() error {
	return compileInputs(s)
}

// SampleConfig function
func (s *Wmi) SampleConfig() string {
	return sampleConfig
}

func compileInputs(s *Wmi) error {
	buildWqlStatements(s)
	return compileTagFilters(s)
}

func compileTagFilters(s *Wmi) error {
	for i, q := range s.Queries {
		var err error
		s.Queries[i].tagFilter, err = compileTagFilter(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func compileTagFilter(q Query) (filter.Filter, error) {
	tagFilter, err := filter.NewIncludeExcludeFilterDefaults(q.TagPropertiesInclude, nil, false, false)
	if err != nil {
		return nil, fmt.Errorf("creating tag filter failed: %w", err)
	}
	return tagFilter, nil
}

// build a WMI query from input configuration
func buildWqlStatements(s *Wmi) {
	for i, q := range s.Queries {
		wql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.Properties, ", "), q.ClassName)
		if len(q.Filter) > 0 {
			wql = fmt.Sprintf("%s WHERE %s", wql, q.Filter)
		}
		s.Queries[i].query = wql
	}
}

func (q *Query) doQuery(acc telegraf.Accumulator) error {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// init COM
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != sFalse {
			return err
		}
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return err
	}
	if unknown == nil {
		return errors.New("failed to create WbemScripting.SWbemLocator, maybe WMI is broken")
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("failed to QueryInterface: %w", err)
	}
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, q.Namespace)
	if err != nil {
		return fmt.Errorf("failed calling method ConnectServer: %w", err)
	}
	service := serviceRaw.ToIDispatch()
	defer serviceRaw.Clear()

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", q.query)
	if err != nil {
		return fmt.Errorf("failed calling method ExecQuery for query %s: %w", q.query, err)
	}
	result := resultRaw.ToIDispatch()
	defer resultRaw.Clear()

	count, err := oleInt64(result, "Count")
	if err != nil {
		return fmt.Errorf("failed getting Count: %w", err)
	}

	for i := int64(0); i < count; i++ {
		// item is a SWbemObject
		itemRaw, err := oleutil.CallMethod(result, "ItemIndex", i)
		if err != nil {
			return fmt.Errorf("failed calling method ItemIndex: %w", err)
		}

		err = q.extractProperties(itemRaw, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Query) extractProperties(itemRaw *ole.VARIANT, acc telegraf.Accumulator) error {
	tags, fields := map[string]string{}, map[string]interface{}{}

	item := itemRaw.ToIDispatch()
	defer item.Release()

	for _, wmiProperty := range q.Properties {
		propertyValue, err := getPropertyValue(item, wmiProperty)
		if err != nil {
			return err
		}

		if q.tagFilter.Match(wmiProperty) {
			valStr, err := internal.ToString(propertyValue)
			if err != nil {
				return fmt.Errorf("converting property %q failed: %w", wmiProperty, err)
			}
			tags[wmiProperty] = valStr
		} else {
			var fieldValue interface{}
			switch v := propertyValue.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				fieldValue = v
			case string:
				fieldValue = v
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
	}
	acc.AddFields(q.ClassName, fields, tags)
	return nil
}

func getPropertyValue(item *ole.IDispatch, wmiProperty string) (interface{}, error) {
	prop, err := oleutil.GetProperty(item, wmiProperty)
	if err != nil {
		return nil, fmt.Errorf("failed GetProperty: %w", err)
	}
	defer prop.Clear()

	return prop.Value(), nil
}

// Gather function
func (s *Wmi) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, query := range s.Queries {
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			err := q.doQuery(acc)
			if err != nil {
				acc.AddError(err)
			}
		}(query)
	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
