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
	defer v.Clear()

	i := int64(v.Val)
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
	defer serviceRaw.Clear()

	// build a WMI query
	var b bytes.Buffer
	b.WriteString("SELECT ")
	if !s.ExcludeNameKey {
		b.WriteString("Name, ")
	}
	b.WriteString(strings.Join(s.Properties, ", "))
	b.WriteString(" FROM ")
	b.WriteString(s.ClassName)
	if len(s.Filter) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(s.Filter)
	}
	query := b.String()

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return err
	}
	result := resultRaw.ToIDispatch()
	defer resultRaw.Clear()

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

		item := itemRaw.ToIDispatch()
		defer item.Release()

		tags := map[string]string{}
		fields := map[string]interface{}{}

		if !s.ExcludeNameKey {
			prop, err := oleutil.GetProperty(item, "Name")
			if err != nil {
				return err
			}
			tags["Name"] = prop.ToString()
			defer prop.Clear()
		}

		for _, wmiProperty := range s.Properties {
			prop, err := oleutil.GetProperty(item, wmiProperty)
			if err != nil {
				return err
			}
			defer prop.Clear()

			// Skip Name if it was provided by the user because we already query for it by default.
			if wmiProperty == "Name" && s.ExcludeNameKey {
				continue
			}

			// if the property's value is an int, then it is a field.
			// if the property's value is a string, then it is a tag.
			valStr := fmt.Sprintf("%v", prop.Value())
			valInt, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				tags[wmiProperty] = valStr
			} else {
				fields[wmiProperty] = valInt
			}
		}

		acc.AddFields(s.ClassName, fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
