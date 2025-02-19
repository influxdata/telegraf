//go:build windows

package win_wmi

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
)

type method struct {
	Namespace            string                 `toml:"namespace"`
	ClassName            string                 `toml:"class_name"`
	Method               string                 `toml:"method"`
	Arguments            map[string]interface{} `toml:"arguments"`
	FieldMapping         map[string]string      `toml:"fields"`
	Filter               string                 `toml:"filter"`
	TagPropertiesInclude []string               `toml:"tag_properties"`

	host             string
	connectionParams []interface{}
	tagFilter        filter.Filter
}

func (m *method) prepare(host string, username, password config.Secret) error {
	// Compile the filter
	f, err := filter.Compile(m.TagPropertiesInclude)
	if err != nil {
		return fmt.Errorf("compiling tag-filter failed: %w", err)
	}
	m.tagFilter = f

	// Setup the connection parameters
	m.host = host
	if m.host != "" {
		m.connectionParams = append(m.connectionParams, m.host)
	} else {
		m.connectionParams = append(m.connectionParams, nil)
	}
	m.connectionParams = append(m.connectionParams, m.Namespace)
	if !username.Empty() {
		u, err := username.Get()
		if err != nil {
			return fmt.Errorf("getting username secret failed: %w", err)
		}
		m.connectionParams = append(m.connectionParams, u.String())
		username.Destroy()
	}
	if !password.Empty() {
		p, err := password.Get()
		if err != nil {
			return fmt.Errorf("getting password secret failed: %w", err)
		}
		m.connectionParams = append(m.connectionParams, p.String())
		password.Destroy()
	}

	return nil
}

func (m *method) execute(acc telegraf.Accumulator) error {
	// The only way to run WMI queries in parallel while being thread-safe is to
	// ensure the CoInitialize[Ex]() call is bound to its current OS thread.
	// Otherwise, attempting to initialize and run parallel queries across
	// goroutines will result in protected memory errors.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Init the COM client
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		var oleCode *ole.OleError
		if errors.As(err, &oleCode) && oleCode.Code() != ole.S_OK && oleCode.Code() != sFalse {
			return err
		}
	}
	defer ole.CoUninitialize()

	// Initialize the WMI service
	locator, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return err
	}
	if locator == nil {
		return errors.New("failed to create WbemScripting.SWbemLocator, maybe WMI is broken")
	}
	defer locator.Release()
	wmi, err := locator.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("failed to query interface: %w", err)
	}
	defer wmi.Release()

	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", m.connectionParams...)
	if err != nil {
		return fmt.Errorf("failed calling method ConnectServer: %w", err)
	}
	service := serviceRaw.ToIDispatch()
	defer serviceRaw.Clear()

	// Get the specified class-method
	classRaw, err := oleutil.CallMethod(service, "Get", m.ClassName)
	if err != nil {
		return fmt.Errorf("failed to get class %s: %w", m.ClassName, err)
	}
	class := classRaw.ToIDispatch()
	defer classRaw.Clear()

	classMethodsRaw, err := class.GetProperty("Methods_")
	if err != nil {
		return fmt.Errorf("failed to call method %s: %w", m.Method, err)
	}
	classMethods := classMethodsRaw.ToIDispatch()
	defer classMethodsRaw.Clear()

	methodRaw, err := classMethods.CallMethod("Item", m.Method)
	if err != nil {
		return fmt.Errorf("failed to call method %s: %w", m.Method, err)
	}
	method := methodRaw.ToIDispatch()
	defer methodRaw.Clear()

	// Fill the input parameters of the method
	inputParamsRaw, err := oleutil.GetProperty(method, "InParameters")
	if err != nil {
		return fmt.Errorf("failed to get input parameters for %s: %w", m.Method, err)
	}
	inputParams := inputParamsRaw.ToIDispatch()
	defer inputParamsRaw.Clear()
	for k, v := range m.Arguments {
		if _, err := inputParams.PutProperty(k, v); err != nil {
			return fmt.Errorf("setting param %q for method %q failed: %w", k, m.Method, err)
		}
	}

	// Get the output parameters of the method
	outputParamsRaw, err := oleutil.GetProperty(method, "OutParameters")
	if err != nil {
		return fmt.Errorf("failed to get output parameters for %s: %w", m.Method, err)
	}
	outputParams := outputParamsRaw.ToIDispatch()
	defer outputParamsRaw.Clear()

	// Execute the method
	outputRaw, err := service.CallMethod("ExecMethod", "StdRegProv", m.Method, inputParamsRaw)
	if err != nil {
		return fmt.Errorf("failed to execute method %s: %w", m.Method, err)
	}
	output := outputRaw.ToIDispatch()
	defer outputRaw.Clear()

	outputPropertiesRaw, err := oleutil.GetProperty(outputParams, "Properties_")
	if err != nil {
		return fmt.Errorf("failed to get output properties for method %s: %w", m.Method, err)
	}
	outputProperties := outputPropertiesRaw.ToIDispatch()
	defer outputPropertiesRaw.Clear()

	// Convert the results to fields and tags
	tags, fields := make(map[string]string), make(map[string]interface{})

	// Add a source tag if we use remote queries
	if m.host != "" {
		tags["source"] = m.host
	}

	err = oleutil.ForEach(outputProperties, func(p *ole.VARIANT) error {
		// Name of the returned result item
		nameProperty, err := p.ToIDispatch().GetProperty("Name")
		if err != nil {
			return errors.New("cannot get output property name")
		}
		name := nameProperty.ToString()
		defer nameProperty.Clear()

		// Value of the returned result item
		property, err := output.GetProperty(name)
		if err != nil {
			return fmt.Errorf("failed to get value for output property %s: %w", name, err)
		}

		// Map the fieldname if provided
		if n, found := m.FieldMapping[name]; found {
			name = n
		}

		// We might get either scalar values or an array of values...
		if value := property.Value(); value != nil {
			if m.tagFilter != nil && m.tagFilter.Match(name) {
				if s, err := internal.ToString(value); err == nil && s != "" {
					tags[name] = s
				}
			} else {
				fields[name] = value
			}
			return nil
		}
		if array := property.ToArray(); array != nil {
			if m.tagFilter != nil && m.tagFilter.Match(name) {
				for i, v := range array.ToValueArray() {
					if s, err := internal.ToString(v); err == nil && s != "" {
						tags[fmt.Sprintf("%s_%d", name, i)] = s
					}
				}
			} else {
				for i, v := range array.ToValueArray() {
					fields[fmt.Sprintf("%s_%d", name, i)] = v
				}
			}
			return nil
		}
		return fmt.Errorf("cannot handle property %q with value %v", name, property)
	})
	if err != nil {
		return fmt.Errorf("cannot iterate the output properties: %w", err)
	}

	acc.AddFields(m.ClassName, fields, tags)

	return nil
}
