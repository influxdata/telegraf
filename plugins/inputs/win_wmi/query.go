//go:build windows

package win_wmi

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
)

type query struct {
	Namespace            string   `toml:"namespace"`
	ClassName            string   `toml:"class_name"`
	Properties           []string `toml:"properties"`
	Filter               string   `toml:"filter"`
	TagPropertiesInclude []string `toml:"tag_properties"`

	host             string
	query            string
	connectionParams []interface{}
	tagFilter        filter.Filter
}

func (q *query) prepare(host string, username, password config.Secret) error {
	// Compile the filter
	f, err := filter.Compile(q.TagPropertiesInclude)
	if err != nil {
		return fmt.Errorf("compiling tag-filter failed: %w", err)
	}
	q.tagFilter = f

	// Setup the connection parameters
	q.host = host
	if q.host != "" {
		q.connectionParams = append(q.connectionParams, q.host)
	} else {
		q.connectionParams = append(q.connectionParams, nil)
	}
	q.connectionParams = append(q.connectionParams, q.Namespace)
	if !username.Empty() {
		u, err := username.Get()
		if err != nil {
			return fmt.Errorf("getting username secret failed: %w", err)
		}
		q.connectionParams = append(q.connectionParams, u.String())
		username.Destroy()
	}
	if !password.Empty() {
		p, err := password.Get()
		if err != nil {
			return fmt.Errorf("getting password secret failed: %w", err)
		}
		q.connectionParams = append(q.connectionParams, p.String())
		password.Destroy()
	}

	// Construct the overall query from the given parts
	wql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.Properties, ", "), q.ClassName)
	if len(q.Filter) > 0 {
		wql += " WHERE " + q.Filter
	}
	q.query = wql

	return nil
}

func (q *query) execute(acc telegraf.Accumulator) error {
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
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", q.connectionParams...)

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

	countRaw, err := oleutil.GetProperty(result, "Count")
	if err != nil {
		return fmt.Errorf("failed getting Count: %w", err)
	}
	count := countRaw.Val
	defer countRaw.Clear()

	for i := int64(0); i < count; i++ {
		itemRaw, err := oleutil.CallMethod(result, "ItemIndex", i)
		if err != nil {
			return fmt.Errorf("failed calling method ItemIndex: %w", err)
		}

		if err := q.extractProperties(acc, itemRaw); err != nil {
			return err
		}
	}
	return nil
}

func (q *query) extractProperties(acc telegraf.Accumulator, itemRaw *ole.VARIANT) error {
	tags, fields := make(map[string]string), make(map[string]interface{})

	if q.host != "" {
		tags["source"] = q.host
	}

	item := itemRaw.ToIDispatch()
	defer item.Release()

	for _, name := range q.Properties {
		propertyRaw, err := oleutil.GetProperty(item, name)
		if err != nil {
			return fmt.Errorf("getting property %q failed: %w", name, err)
		}
		value := propertyRaw.Value()
		propertyRaw.Clear()

		if q.tagFilter != nil && q.tagFilter.Match(name) {
			s, err := internal.ToString(value)
			if err != nil {
				return fmt.Errorf("converting property %q failed: %w", s, err)
			}
			tags[name] = s
			continue
		}

		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			fields[name] = v
		case string:
			fields[name] = v
		case bool:
			fields[name] = v
		case []byte:
			fields[name] = string(v)
		case fmt.Stringer:
			fields[name] = v.String()
		case nil:
			fields[name] = nil
		default:
			return fmt.Errorf("property %q of type \"%T\" unsupported", name, v)
		}
	}
	acc.AddFields(q.ClassName, fields, tags)
	return nil
}
