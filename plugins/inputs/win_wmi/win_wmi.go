package win_wmi

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Wmi struct
type Wmi struct {
	Namespace      string   `toml:"namespace"`
	ClassName      string   `toml:"classname"`
	Properties     []string `toml:"properties"`
	Filter         string   `toml:"filter"`
	ExcludeNameKey bool     `toml:"excludenamekey"`
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
	return `
  ## By default, this plugin returns no results.
  ## Uncomment the example below or write your own as you see fit.
  ## The "Name" property of a WMI class is automatically included unless excludenamekey is true.
  ## If the WMI property's value is a string, then it is used as a tag.
  ## If the WMI property's value is a type of int, then it is used as a field.
  ## [[inputs.win_wmi]]
  ##   namespace = "root\\cimv2"
  ##   classname = "Win32_Volume"
  ##   properties = ["Capacity", "FreeSpace"]
  ##   filter = 'NOT Name LIKE "\\\\?\\%"'
  ##   excludenamekey = false
  ##   name_prefix = "win_wmi_"
`
}

func BuildWmiQuery(s *Wmi) (string, error) {
	// build a WMI query
	var query string
	var b bytes.Buffer
	_, err := b.WriteString("SELECT ")
	if err != nil {
		return query, err
	}

	if !s.ExcludeNameKey {
		_, err := b.WriteString("Name, ")
		if err != nil {
			return query, err
		}
	}

	_, err = b.WriteString(strings.Join(s.Properties, ", "))
	if err != nil {
		return query, err
	}

	_, err = b.WriteString(" FROM ")
	if err != nil {
		return query, err
	}

	_, err = b.WriteString(s.ClassName)
	if err != nil {
		return query, err
	}

	if len(s.Filter) > 0 {
		_, err = b.WriteString(" WHERE ")
		if err != nil {
			return query, err
		}

		_, err = b.WriteString(s.Filter)
		if err != nil {
			return query, err
		}
	}
	query = b.String()
	return query, nil
}

// Gather function
func (s *Wmi) Gather(acc telegraf.Accumulator) error {
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
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", nil, s.Namespace)
	if err != nil {
		return err
	}
	service := serviceRaw.ToIDispatch()
	defer func() { _ = serviceRaw.Clear() }()

	query, err := BuildWmiQuery(s)
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

	for i := int64(0); i < count; i++ {
		// item is a SWbemObject
		itemRaw, err := oleutil.CallMethod(result, "ItemIndex", i)
		if err != nil {
			return err
		}

		tags := map[string]string{}
		fields := map[string]interface{}{}

		e1 := func() error {
			item := itemRaw.ToIDispatch()
			defer item.Release()

			if !s.ExcludeNameKey {
				prop, err := oleutil.GetProperty(item, "Name")
				if err != nil {
					return err
				}
				tags["Name"] = prop.ToString()
				defer func() { _ = prop.Clear() }()
			}

			for _, wmiProperty := range s.Properties {
				// Skip Name if it was provided by the user because we already query for it by default.
				if wmiProperty == "Name" && s.ExcludeNameKey {
					continue
				}

				e2 := func() error {
					prop, err := oleutil.GetProperty(item, wmiProperty)
					if err != nil {
						return err
					}
					defer func() { _ = prop.Clear() }()

					// if the property's value is an int, then it is a field.
					// if the property's value is a string, then it is a tag.
					valStr := fmt.Sprintf("%v", prop.Value())
					valInt, err := strconv.ParseInt(valStr, 10, 64)
					if err != nil {
						tags[wmiProperty] = valStr
					} else {
						fields[wmiProperty] = valInt
					}
					return nil
				}()

				if e2 != nil {
					return e2
				}
			}
			return nil
		}()

		if e1 != nil {
			return e1
		}

		acc.AddFields(s.ClassName, fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
