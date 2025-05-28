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
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		var oleCode *ole.OleError
		if !errors.As(err, &oleCode) || (oleCode.Code() != ole.S_OK && oleCode.Code() != sFalse) {
			return fmt.Errorf("initialization of COM object failed: %w", err)
		}
	}
	defer ole.CoUninitialize()

	// Initialize the WMI service
	locator, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return fmt.Errorf("creation of OLE object failed: %w", err)
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

	serviceRaw, err := wmi.CallMethod("ConnectServer", m.connectionParams...)
	if err != nil {
		return fmt.Errorf("failed calling method ConnectServer: %w", err)
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	// Get the specified class-method
	classRaw, err := service.CallMethod("Get", m.ClassName)
	if err != nil {
		return fmt.Errorf("failed to get class %s: %w", m.ClassName, err)
	}
	class := classRaw.ToIDispatch()
	defer class.Release()

	classMethodsRaw, err := class.GetProperty("Methods_")
	if err != nil {
		return fmt.Errorf("failed to call method %s: %w", m.Method, err)
	}
	classMethods := classMethodsRaw.ToIDispatch()
	defer classMethods.Release()

	methodRaw, err := classMethods.CallMethod("Item", m.Method)
	if err != nil {
		return fmt.Errorf("failed to call method %s: %w", m.Method, err)
	}
	method := methodRaw.ToIDispatch()
	defer method.Release()

	// Fill the input parameters of the method
	inputParamsRaw, err := method.GetProperty("InParameters")
	if err != nil {
		return fmt.Errorf("failed to get input parameters for %s: %w", m.Method, err)
	}
	inputParams := inputParamsRaw.ToIDispatch()
	defer inputParams.Release()
	inputProps := make([]*ole.VARIANT, 0, len(m.Arguments))
	defer func() {
		for _, p := range inputProps {
			p.Clear()
		}
	}()
	for k, v := range m.Arguments {
		p, err := inputParams.PutProperty(k, v)
		if err != nil {
			return fmt.Errorf("setting param %q for method %q failed: %w", k, m.Method, err)
		}
		inputProps = append(inputProps, p)
	}

	// Get the output parameters of the method
	outputParamsRaw, err := method.GetProperty("OutParameters")
	if err != nil {
		return fmt.Errorf("failed to get output parameters for %s: %w", m.Method, err)
	}
	outputParams := outputParamsRaw.ToIDispatch()
	defer outputParams.Release()

	// Execute the method
	outputRaw, err := service.CallMethod("ExecMethod", m.ClassName, m.Method, inputParamsRaw)
	if err != nil {
		return fmt.Errorf("failed to execute method %s: %w", m.Method, err)
	}
	output := outputRaw.ToIDispatch()
	defer output.Release()

	outputPropertiesRaw, err := outputParams.GetProperty("Properties_")
	if err != nil {
		return fmt.Errorf("failed to get output properties for method %s: %w", m.Method, err)
	}
	outputProperties := outputPropertiesRaw.ToIDispatch()
	defer outputProperties.Release()

	// Convert the results to fields and tags
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	// Add a source tag if we use remote queries
	if m.host != "" {
		tags["source"] = m.host
	}

	if err := oleutil.ForEach(outputProperties, func(p *ole.VARIANT) error {
		propTags, propFields, err := m.extractData(p, output)
		if err != nil {
			return err
		}
		for k, v := range propTags {
			tags[k] = v
		}
		for k, v := range propFields {
			fields[k] = v
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cannot iterate the output properties: %w", err)
	}

	acc.AddFields(m.ClassName, fields, tags)

	return nil
}

func (m *method) extractData(prop *ole.VARIANT, output *ole.IDispatch) (map[string]string, map[string]interface{}, error) {
	// Name of the returned result item
	namePropertyRaw := prop.ToIDispatch()
	defer namePropertyRaw.Release()
	nameProperty, err := namePropertyRaw.GetProperty("Name")
	if err != nil {
		return nil, nil, errors.New("cannot get output property name")
	}
	defer nameProperty.Clear()
	raw := nameProperty.ToString()

	// Map the fieldname if provided
	name := raw
	if n, found := m.FieldMapping[name]; found {
		name = n
	}

	// Value of the returned result item
	property, err := output.GetProperty(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get value for output property %s: %w", raw, err)
	}
	defer property.Clear()

	// We might get either scalar values or an array of values...
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	if value := property.Value(); value != nil {
		if m.tagFilter != nil && m.tagFilter.Match(name) {
			if s, err := internal.ToString(value); err == nil && s != "" {
				tags[name] = s
			}
		} else {
			fields[name] = value
		}
		return tags, fields, nil
	}
	if array := property.ToArray(); array != nil {
		defer array.Release()
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
		return tags, fields, nil
	}
	return nil, nil, fmt.Errorf("cannot handle property %q with value %v", name, property)
}
