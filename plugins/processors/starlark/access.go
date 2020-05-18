package starlark

import (
	"github.com/influxdata/telegraf"
)

type AccessibleEntry struct {
	Key   string
	Value interface{}
}

type Accessible interface {
	Add(key string, value interface{})
	Remove(key string)
	Clear()
	Get(key string) (interface{}, bool)
	List() []AccessibleEntry
	Len() int
}

type AccessibleField struct {
	metric telegraf.Metric
	frozen bool
}

func (m *AccessibleField) Add(key string, value interface{}) {
	m.metric.AddField(key, value)
}

func (m *AccessibleField) Remove(key string) {
	m.metric.RemoveField(key)
}

func (m *AccessibleField) Clear() {
	for _, field := range m.metric.FieldList() {
		m.metric.RemoveField(field.Key)
	}
}

func (m *AccessibleField) Get(key string) (interface{}, bool) {
	return m.metric.GetField(key)
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

func (m *AccessibleTag) Add(key string, value interface{}) {
	m.metric.AddTag(key, value.(string))
}

func (m *AccessibleTag) Remove(key string) {
	m.metric.RemoveTag(key)
}

func (m *AccessibleTag) Clear() {
	for _, field := range m.metric.TagList() {
		m.metric.RemoveTag(field.Key)
	}
}

func (m *AccessibleTag) Get(key string) (interface{}, bool) {
	return m.metric.GetTag(key)
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
