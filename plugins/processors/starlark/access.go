package starlark

import (
	"errors"

	"github.com/influxdata/telegraf"
)

type AccessibleEntry struct {
	Key   string
	Value interface{}
}

type Accessible interface {
	Add(key string, value interface{}) error
	Remove(key string)
	Clear()
	Get(key string) (interface{}, bool)
	GetIndex(index int) string
	List() []AccessibleEntry
	Len() int
}

type AccessibleField struct {
	metric telegraf.Metric
	frozen bool
}

func (m *AccessibleField) Add(key string, value interface{}) error {
	m.metric.AddField(key, value)
	return nil
}

func (m *AccessibleField) Remove(key string) {
	m.metric.RemoveField(key)
}

func (m *AccessibleField) Clear() {
	keys := make([]string, 0, len(m.metric.FieldList()))
	for _, field := range m.metric.FieldList() {
		keys = append(keys, field.Key)
	}

	for _, key := range keys {
		m.metric.RemoveField(key)
	}
}

func (m *AccessibleField) Get(key string) (interface{}, bool) {
	return m.metric.GetField(key)
}

func (m *AccessibleField) GetIndex(index int) string {
	return m.metric.FieldList()[index].Key
}

func (m *AccessibleField) List() []AccessibleEntry {
	fields := m.metric.FieldList()
	entries := make([]AccessibleEntry, len(fields))
	for i, field := range fields {
		entries[i].Key = field.Key
		entries[i].Value = field.Value
	}

	return entries
}

func (m *AccessibleField) Len() int {
	return len(m.metric.FieldList())
}

type AccessibleTag struct {
	metric telegraf.Metric
	frozen bool
}

func (m *AccessibleTag) Add(key string, value interface{}) error {
	switch str := value.(type) {
	case string:
		m.metric.AddTag(key, str)
		return nil
	default:
		return errors.New("value must be of type 'str'")
	}
}

func (m *AccessibleTag) Remove(key string) {
	m.metric.RemoveTag(key)
}

func (m *AccessibleTag) Clear() {
	keys := make([]string, 0, len(m.metric.TagList()))
	for _, tag := range m.metric.TagList() {
		keys = append(keys, tag.Key)
	}

	for _, key := range keys {
		m.metric.RemoveTag(key)
	}
}

func (m *AccessibleTag) Get(key string) (interface{}, bool) {
	return m.metric.GetTag(key)
}

func (m *AccessibleTag) GetIndex(index int) string {
	return m.metric.TagList()[index].Key
}

func (m *AccessibleTag) List() []AccessibleEntry {
	tags := m.metric.TagList()
	entries := make([]AccessibleEntry, len(tags))
	for i, tag := range tags {
		entries[i].Key = tag.Key
		entries[i].Value = tag.Value
	}

	return entries
}

func (m *AccessibleTag) Len() int {
	return len(m.metric.TagList())
}
