package starlark

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"go.starlark.net/starlark"
)

type Metric struct {
	metric telegraf.Metric
	frozen bool
}

func (m *Metric) Unwrap() telegraf.Metric {
	return m.metric
}

// The String function is called by both the repr() and str() functions.  There
// is a slight difference in how strings are quoted between the two functions,
// but this type isn't a string so it is irrelevent.
func (m *Metric) String() string {
	buf := new(strings.Builder)
	buf.WriteString("Metric(")
	buf.WriteString(m.Name().String())
	buf.WriteString(", tags=")
	buf.WriteString(m.Tags().String())
	buf.WriteString(", fields=")
	buf.WriteString(m.Fields().String())
	buf.WriteString(", time=")
	buf.WriteString(m.Time().String())
	buf.WriteString(")")
	return buf.String()
}

func (m *Metric) Type() string {
	return "Metric"
}

func (m *Metric) Freeze() {
	m.frozen = true
}

func (m *Metric) Truth() starlark.Bool {
	return true
}

func (m *Metric) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

func (m *Metric) AttrNames() []string {
	return []string{"name", "tags", "fields", "time"}
}

// Attr implements the HasAttrs interface
func (m *Metric) Attr(name string) (starlark.Value, error) {
	switch name {
	case "name":
		return m.Name(), nil
	case "tags":
		return m.Tags(), nil
	case "fields":
		return m.Fields(), nil
	case "time":
		return m.Time(), nil
	default:
		// Returning nil, nil indicates "no such field or method"
		return nil, nil
	}
}

func (m *Metric) SetField(name string, value starlark.Value) error {
	if m.frozen {
		return fmt.Errorf("cannot modify frozen metric")
	}

	switch name {
	case "name":
		return m.SetName(value)
	case "time":
		return m.SetTime(value)
	case "tags":
		return errors.New("AttributeError: can't set tags")
	case "fields":
		return errors.New("AttributeError: can't set fields")
	default:
		return starlark.NoSuchAttrError(
			fmt.Sprintf("cannot assign to field '%s'", name))
	}
}

func (m *Metric) Name() starlark.String {
	return starlark.String(m.metric.Name())
}

func (m *Metric) SetName(value starlark.Value) error {
	if str, ok := value.(starlark.String); ok {
		m.metric.SetName(str.GoString())
		return nil
	}

	return errors.New("type error")
}

func (m *Metric) Tags() *MetricDataDict {
	tagsaccessor := AccessibleTag(*m)
	tags := MetricDataDict{data: &tagsaccessor, typename: "tags"}
	return &tags
}

func (m *Metric) Fields() *MetricDataDict {
	fieldaccessor := AccessibleField(*m)
	fields := MetricDataDict{data: &fieldaccessor, typename: "fields"}
	return &fields
}

func (m *Metric) Time() starlark.Int {
	return starlark.MakeInt64(m.metric.Time().UnixNano())
}

func (m *Metric) SetTime(value starlark.Value) error {
	switch v := value.(type) {
	case starlark.Int:
		ns, ok := v.Int64()
		if !ok {
			return errors.New("type error: unrepresentable time")
		}
		tm := time.Unix(0, ns)
		m.metric.SetTime(tm)
		return nil
	default:
		return errors.New("type error")
	}
}
