package starlark

import (
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"go.starlark.net/starlark"
)

type Metric struct {
	metric telegraf.Metric
}

func (m *Metric) Unwrap() telegraf.Metric {
	return m.metric
}

func (m *Metric) String() string {
	return "metric"
}

func (m *Metric) Type() string {
	return "Metric"
}

func (m *Metric) Freeze() {
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
	switch name {
	case "name":
		m.SetName(value)
		return nil
	case "time":
		m.SetTime(value)
		return nil
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
