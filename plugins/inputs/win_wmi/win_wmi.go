//go:build windows
// +build windows

package win_wmi

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## By default, this plugin returns no results.
  ## Uncomment the example below or write your own as you see fit.
  ## [[inputs.win_wmi]]
  ##   name_prefix = "win_wmi_"
  ##   [[inputs.win_wmi.query]]
  ##     Namespace = "root\\cimv2"
  ##     ClassName = "Win32_Volume"
  ##     Properties = ["Name", "Capacity", "FreeSpace"]
  ##     Filter = 'NOT Name LIKE "\\\\?\\%"'
  ##     TagPropertiesInclude = ["Name"]
`

// Query struct
type Query struct {
	Query                string   `toml:"query"`
	Namespace            string   `toml:"Namespace"`
	ClassName            string   `toml:"ClassName"`
	Properties           []string `toml:"Properties"`
	Filter               string   `toml:"Filter"`
	TagPropertiesInclude []string `toml:"TagPropertiesInclude"`
	tagFilter            filter.Filter
}

var lock sync.Mutex

// Wmi struct
type Wmi struct {
	Queries []Query `toml:"query"`
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

// Description function
func (s *Wmi) Description() string {
	return "returns WMI query results as metrics"
}

// SampleConfig function
func (s *Wmi) SampleConfig() string {
	return sampleConfig
}

// build a WMI query
func BuildWmiQuery(q Query) (string, error) {
	var wql string
	var b bytes.Buffer
	_, err := b.WriteString("SELECT ")
	if err != nil {
		return wql, err
	}

	_, err = b.WriteString(strings.Join(q.Properties, ", "))
	if err != nil {
		return wql, err
	}

	_, err = b.WriteString(" FROM ")
	if err != nil {
		return wql, err
	}

	_, err = b.WriteString(q.ClassName)
	if err != nil {
		return wql, err
	}

	if len(q.Filter) > 0 {
		_, err = b.WriteString(" WHERE ")
		if err != nil {
			return wql, err
		}

		_, err = b.WriteString(q.Filter)
		if err != nil {
			return wql, err
		}
	}
	wql = b.String()

	return wql, nil
}

func DoQuery(q Query, acc telegraf.Accumulator) error {
	lock.Lock()
	defer lock.Unlock()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tags := map[string]string{}
	fields := map[string]interface{}{}

	// init COM
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
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
		panic("CreateObject returned nil")
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, q.Namespace)
	if err != nil {
		return err
	}
	service := serviceRaw.ToIDispatch()
	defer func() { _ = serviceRaw.Clear() }()

	query, err := BuildWmiQuery(q)
	if err != nil {
		return err
	}

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return err
	}
	result := resultRaw.ToIDispatch()
	defer func() { _ = resultRaw.Clear() }()

	count, err := oleInt64(result, "Count")
	if err != nil {
		return err
	}

	// Compile the tag filter
	tagfilter, err := filter.NewIncludeExcludeFilterDefaults(q.TagPropertiesInclude, nil, false, false)
	if err != nil {
		return fmt.Errorf("creating tag filter failed: %v", err)
	}
	q.tagFilter = tagfilter

	for i := int64(0); i < count; i++ {
		// item is a SWbemObject
		itemRaw, err := oleutil.CallMethod(result, "ItemIndex", i)
		if err != nil {
			return err
		}

		outerErr := func() error {
			item := itemRaw.ToIDispatch()
			defer item.Release()

			for _, wmiProperty := range q.Properties {
				innerErr := func() error {
					prop, err := oleutil.GetProperty(item, wmiProperty)
					if err != nil {
						return err
					}
					defer func() { _ = prop.Clear() }()

					// if an empty property is returned from WMI, then move on
					if prop.Value() == nil {
						return nil
					}

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
			err := DoQuery(q, acc)
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
