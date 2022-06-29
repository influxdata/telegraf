package starlark

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.starlark.net/starlark"

	"github.com/influxdata/telegraf"
)

type Metric struct {
	metric         telegraf.Metric
	tagIterCount   int
	fieldIterCount int
	frozen         bool
}

// Wrap updates the starlark.Metric to wrap a new telegraf.Metric.
func (m *Metric) Wrap(metric telegraf.Metric) {
	m.metric = metric
	m.tagIterCount = 0
	m.fieldIterCount = 0
	m.frozen = false
}

// Unwrap removes the telegraf.Metric from the startlark.Metric.
func (m *Metric) Unwrap() telegraf.Metric {
	return m.metric
}

// String returns the starlark representation of the Metric.
//
// The String function is called by both the repr() and str() functions, and so
// it behaves more like the repr function would in Python.
func (m *Metric) String() string {
	buf := new(strings.Builder)
	buf.WriteString("Metric(")           //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(m.Name().String())   //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(", tags=")           //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(m.Tags().String())   //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(", fields=")         //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(m.Fields().String()) //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(", time=")           //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(m.Time().String())   //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	buf.WriteString(")")                 //nolint:revive // from builder.go: "It returns the length of r and a nil error."
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

// AttrNames implements the starlark.HasAttrs interface.
func (m *Metric) AttrNames() []string {
	return []string{"name", "tags", "fields", "time"}
}

// Attr implements the starlark.HasAttrs interface.
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

// SetField implements the starlark.HasSetField interface.
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
		return errors.New("cannot set tags")
	case "fields":
		return errors.New("cannot set fields")
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

func (m *Metric) Tags() TagDict {
	return TagDict{m}
}

func (m *Metric) Fields() FieldDict {
	return FieldDict{m}
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
