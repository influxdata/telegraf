package metric

import (
	"fmt"
	"hash/fnv"
	"sort"
	"time"

	"github.com/influxdata/telegraf"
)

type metric struct {
	MetricName   string
	MetricTags   []*telegraf.Tag
	MetricFields []*telegraf.Field
	MetricTime   time.Time

	MetricType telegraf.ValueType
}

func New(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	tm time.Time,
	tp ...telegraf.ValueType,
) telegraf.Metric {
	var vtype telegraf.ValueType
	if len(tp) > 0 {
		vtype = tp[0]
	} else {
		vtype = telegraf.Untyped
	}

	m := &metric{
		MetricName:   name,
		MetricTags:   nil,
		MetricFields: nil,
		MetricTime:   tm,
		MetricType:   vtype,
	}

	if len(tags) > 0 {
		m.MetricTags = make([]*telegraf.Tag, 0, len(tags))
		for k, v := range tags {
			m.MetricTags = append(m.MetricTags,
				&telegraf.Tag{Key: k, Value: v})
		}
		sort.Slice(m.MetricTags, func(i, j int) bool { return m.MetricTags[i].Key < m.MetricTags[j].Key })
	}

	if len(fields) > 0 {
		m.MetricFields = make([]*telegraf.Field, 0, len(fields))
		for k, v := range fields {
			v := convertField(v)
			if v == nil {
				continue
			}
			m.AddField(k, v)
		}
	}

	return m
}

// FromMetric returns a deep copy of the metric with any tracking information
// removed.
func FromMetric(other telegraf.Metric) telegraf.Metric {
	m := &metric{
		MetricName:   other.Name(),
		MetricTags:   make([]*telegraf.Tag, len(other.TagList())),
		MetricFields: make([]*telegraf.Field, len(other.FieldList())),
		MetricTime:   other.Time(),
		MetricType:   other.Type(),
	}

	for i, tag := range other.TagList() {
		m.MetricTags[i] = &telegraf.Tag{Key: tag.Key, Value: tag.Value}
	}

	for i, field := range other.FieldList() {
		m.MetricFields[i] = &telegraf.Field{Key: field.Key, Value: field.Value}
	}
	return m
}

func (m *metric) String() string {
	return fmt.Sprintf("%s %v %v %d", m.MetricName, m.Tags(), m.Fields(), m.MetricTime.UnixNano())
}

func (m *metric) Name() string {
	return m.MetricName
}

func (m *metric) Tags() map[string]string {
	tags := make(map[string]string, len(m.MetricTags))
	for _, tag := range m.MetricTags {
		tags[tag.Key] = tag.Value
	}
	return tags
}

func (m *metric) TagList() []*telegraf.Tag {
	return m.MetricTags
}

func (m *metric) Fields() map[string]interface{} {
	fields := make(map[string]interface{}, len(m.MetricFields))
	for _, field := range m.MetricFields {
		fields[field.Key] = field.Value
	}

	return fields
}

func (m *metric) FieldList() []*telegraf.Field {
	return m.MetricFields
}

func (m *metric) Time() time.Time {
	return m.MetricTime
}

func (m *metric) Type() telegraf.ValueType {
	return m.MetricType
}

func (m *metric) SetName(name string) {
	m.MetricName = name
}

func (m *metric) AddPrefix(prefix string) {
	m.MetricName = prefix + m.MetricName
}

func (m *metric) AddSuffix(suffix string) {
	m.MetricName = m.MetricName + suffix
}

func (m *metric) AddTag(key, value string) {
	for i, tag := range m.MetricTags {
		if key > tag.Key {
			continue
		}

		if key == tag.Key {
			tag.Value = value
			return
		}

		m.MetricTags = append(m.MetricTags, nil)
		copy(m.MetricTags[i+1:], m.MetricTags[i:])
		m.MetricTags[i] = &telegraf.Tag{Key: key, Value: value}
		return
	}

	m.MetricTags = append(m.MetricTags, &telegraf.Tag{Key: key, Value: value})
}

func (m *metric) HasTag(key string) bool {
	for _, tag := range m.MetricTags {
		if tag.Key == key {
			return true
		}
	}
	return false
}

func (m *metric) GetTag(key string) (string, bool) {
	for _, tag := range m.MetricTags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

func (m *metric) Tag(key string) string {
	v, _ := m.GetTag(key)
	return v
}

func (m *metric) RemoveTag(key string) {
	for i, tag := range m.MetricTags {
		if tag.Key == key {
			copy(m.MetricTags[i:], m.MetricTags[i+1:])
			m.MetricTags[len(m.MetricTags)-1] = nil
			m.MetricTags = m.MetricTags[:len(m.MetricTags)-1]
			return
		}
	}
}

func (m *metric) AddField(key string, value interface{}) {
	for i, field := range m.MetricFields {
		if key == field.Key {
			m.MetricFields[i] = &telegraf.Field{Key: key, Value: convertField(value)}
			return
		}
	}
	m.MetricFields = append(m.MetricFields, &telegraf.Field{Key: key, Value: convertField(value)})
}

func (m *metric) HasField(key string) bool {
	for _, field := range m.MetricFields {
		if field.Key == key {
			return true
		}
	}
	return false
}

func (m *metric) GetField(key string) (interface{}, bool) {
	for _, field := range m.MetricFields {
		if field.Key == key {
			return field.Value, true
		}
	}
	return nil, false
}

func (m *metric) Field(key string) interface{} {
	if v, found := m.GetField(key); found {
		return v
	}
	return nil
}

func (m *metric) RemoveField(key string) {
	for i, field := range m.MetricFields {
		if field.Key == key {
			copy(m.MetricFields[i:], m.MetricFields[i+1:])
			m.MetricFields[len(m.MetricFields)-1] = nil
			m.MetricFields = m.MetricFields[:len(m.MetricFields)-1]
			return
		}
	}
}

func (m *metric) SetTime(t time.Time) {
	m.MetricTime = t
}

func (m *metric) SetType(t telegraf.ValueType) {
	m.MetricType = t
}

func (m *metric) Copy() telegraf.Metric {
	m2 := &metric{
		MetricName:   m.MetricName,
		MetricTags:   make([]*telegraf.Tag, len(m.MetricTags)),
		MetricFields: make([]*telegraf.Field, len(m.MetricFields)),
		MetricTime:   m.MetricTime,
		MetricType:   m.MetricType,
	}

	for i, tag := range m.MetricTags {
		m2.MetricTags[i] = &telegraf.Tag{Key: tag.Key, Value: tag.Value}
	}

	for i, field := range m.MetricFields {
		m2.MetricFields[i] = &telegraf.Field{Key: field.Key, Value: field.Value}
	}
	return m2
}

func (m *metric) HashID() uint64 {
	h := fnv.New64a()
	h.Write([]byte(m.MetricName))
	h.Write([]byte("\n"))
	for _, tag := range m.MetricTags {
		h.Write([]byte(tag.Key))
		h.Write([]byte("\n"))
		h.Write([]byte(tag.Value))
		h.Write([]byte("\n"))
	}
	return h.Sum64()
}

func (m *metric) Accept() {
}

func (m *metric) Reject() {
}

func (m *metric) Drop() {
}

// Convert field to a supported type or nil if inconvertible
func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return v
	case int64:
		return v
	case string:
		return v
	case bool:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
		return v
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	case *float64:
		if v != nil {
			return *v
		}
	case *int64:
		if v != nil {
			return *v
		}
	case *string:
		if v != nil {
			return *v
		}
	case *bool:
		if v != nil {
			return *v
		}
	case *int:
		if v != nil {
			return int64(*v)
		}
	case *uint:
		if v != nil {
			return uint64(*v)
		}
	case *uint64:
		if v != nil {
			return *v
		}
	case *[]byte:
		if v != nil {
			return string(*v)
		}
	case *int32:
		if v != nil {
			return int64(*v)
		}
	case *int16:
		if v != nil {
			return int64(*v)
		}
	case *int8:
		if v != nil {
			return int64(*v)
		}
	case *uint32:
		if v != nil {
			return uint64(*v)
		}
	case *uint16:
		if v != nil {
			return uint64(*v)
		}
	case *uint8:
		if v != nil {
			return uint64(*v)
		}
	case *float32:
		if v != nil {
			return float64(*v)
		}
	default:
		return nil
	}
	return nil
}
