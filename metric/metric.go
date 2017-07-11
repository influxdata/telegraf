package metric

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

const MaxInt = int(^uint(0) >> 1)

func New(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
	mType ...telegraf.ValueType,
) (telegraf.Metric, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("Metric cannot be made without any fields")
	}
	if len(name) == 0 {
		return nil, fmt.Errorf("Metric cannot be made with an empty name")
	}
	if strings.HasSuffix(name, `\`) {
		return nil, fmt.Errorf("Metric cannot have measurement name ending with a backslash")
	}

	var thisType telegraf.ValueType
	if len(mType) > 0 {
		thisType = mType[0]
	} else {
		thisType = telegraf.Untyped
	}

	m := &metric{
		name:  []byte(escape(name, "name")),
		t:     []byte(fmt.Sprint(t.UnixNano())),
		nsec:  t.UnixNano(),
		mType: thisType,
	}

	// pre-allocate exact size of the tags slice
	taglen := 0
	for k, v := range tags {
		if strings.HasSuffix(k, `\`) {
			return nil, fmt.Errorf("Metric cannot have tag key ending with a backslash")
		}
		if strings.HasSuffix(v, `\`) {
			return nil, fmt.Errorf("Metric cannot have tag value ending with a backslash")
		}

		if len(k) == 0 || len(v) == 0 {
			continue
		}
		taglen += 2 + len(escape(k, "tagkey")) + len(escape(v, "tagval"))
	}
	m.tags = make([]byte, taglen)

	i := 0
	for k, v := range tags {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		m.tags[i] = ','
		i++
		i += copy(m.tags[i:], escape(k, "tagkey"))
		m.tags[i] = '='
		i++
		i += copy(m.tags[i:], escape(v, "tagval"))
	}

	// pre-allocate capacity of the fields slice
	fieldlen := 0
	for k, v := range fields {
		if strings.HasSuffix(k, `\`) {
			return nil, fmt.Errorf("Metric cannot have field key ending with a backslash")
		}
		switch val := v.(type) {
		case string:
			if strings.HasSuffix(val, `\`) {
				return nil, fmt.Errorf("Metric cannot have field value ending with a backslash")
			}
		}

		// 10 bytes is completely arbitrary, but will at least prevent some
		// amount of allocations. There's a small possibility this will create
		// slightly more allocations for a metric that has many short fields.
		fieldlen += len(k) + 10
	}
	m.fields = make([]byte, 0, fieldlen)

	i = 0
	for k, v := range fields {
		if i != 0 {
			m.fields = append(m.fields, ',')
		}
		m.fields = appendField(m.fields, k, v)
		i++
	}

	return m, nil
}

// indexUnescapedByte finds the index of the first byte equal to b in buf that
// is not escaped. Returns -1 if not found.
func indexUnescapedByte(buf []byte, b byte) int {
	var keyi int
	for {
		i := bytes.IndexByte(buf[keyi:], b)
		if i == -1 {
			return -1
		} else if i == 0 {
			break
		}
		keyi += i
		if buf[keyi-1] != '\\' {
			break
		} else {
			keyi++
		}
	}
	return keyi
}

type metric struct {
	name   []byte
	tags   []byte
	fields []byte
	t      []byte

	mType     telegraf.ValueType
	aggregate bool

	// cached values for reuse in "get" functions
	hashID uint64
	nsec   int64
}

func (m *metric) String() string {
	return string(m.name) + string(m.tags) + " " + string(m.fields) + " " + string(m.t) + "\n"
}

func (m *metric) SetAggregate(b bool) {
	m.aggregate = b
}

func (m *metric) IsAggregate() bool {
	return m.aggregate
}

func (m *metric) Type() telegraf.ValueType {
	return m.mType
}

func (m *metric) Len() int {
	// 3 is for 2 spaces surrounding the fields array + newline at the end.
	return len(m.name) + len(m.tags) + len(m.fields) + len(m.t) + 3
}

func (m *metric) Serialize() []byte {
	tmp := make([]byte, m.Len())
	i := 0
	i += copy(tmp[i:], m.name)
	i += copy(tmp[i:], m.tags)
	tmp[i] = ' '
	i++
	i += copy(tmp[i:], m.fields)
	tmp[i] = ' '
	i++
	i += copy(tmp[i:], m.t)
	tmp[i] = '\n'
	return tmp
}

func (m *metric) SerializeTo(dst []byte) int {
	i := 0
	if i >= len(dst) {
		return i
	}

	i += copy(dst[i:], m.name)
	if i >= len(dst) {
		return i
	}

	i += copy(dst[i:], m.tags)
	if i >= len(dst) {
		return i
	}

	dst[i] = ' '
	i++
	if i >= len(dst) {
		return i
	}

	i += copy(dst[i:], m.fields)
	if i >= len(dst) {
		return i
	}

	dst[i] = ' '
	i++
	if i >= len(dst) {
		return i
	}

	i += copy(dst[i:], m.t)
	if i >= len(dst) {
		return i
	}
	dst[i] = '\n'

	return i + 1
}

func (m *metric) Split(maxSize int) []telegraf.Metric {
	if m.Len() <= maxSize {
		return []telegraf.Metric{m}
	}
	var out []telegraf.Metric

	// constant number of bytes for each metric (in addition to field bytes)
	constant := len(m.name) + len(m.tags) + len(m.t) + 3
	// currently selected fields
	fields := make([]byte, 0, maxSize)

	i := 0
	for {
		if i >= len(m.fields) {
			// hit the end of the field byte slice
			if len(fields) > 0 {
				out = append(out, copyWith(m.name, m.tags, fields, m.t))
			}
			break
		}

		// find the end of the next field
		j := indexUnescapedByte(m.fields[i:], ',')
		if j == -1 {
			j = len(m.fields)
		} else {
			j += i
		}

		// if true, then we need to create a metric _not_ including the currently
		// selected field
		if len(m.fields[i:j])+len(fields)+constant >= maxSize {
			// if false, then we'll create a metric including the currently
			// selected field anyways. This means that the given maxSize is too
			// small for a single field to fit.
			if len(fields) > 0 {
				out = append(out, copyWith(m.name, m.tags, fields, m.t))
			}

			fields = make([]byte, 0, maxSize)
		}
		if len(fields) > 0 {
			fields = append(fields, ',')
		}
		fields = append(fields, m.fields[i:j]...)

		i = j + 1
	}
	return out
}

func (m *metric) Fields() map[string]interface{} {
	fieldMap := map[string]interface{}{}
	i := 0
	for {
		if i >= len(m.fields) {
			break
		}
		// end index of field key
		i1 := indexUnescapedByte(m.fields[i:], '=')
		if i1 == -1 {
			break
		}
		// start index of field value
		i2 := i1 + 1

		// end index of field value
		var i3 int
		if m.fields[i:][i2] == '"' {
			i3 = indexUnescapedByte(m.fields[i:][i2+1:], '"')
			if i3 == -1 {
				i3 = len(m.fields[i:])
			}
			i3 += i2 + 2 // increment index to the comma
		} else {
			i3 = indexUnescapedByte(m.fields[i:], ',')
			if i3 == -1 {
				i3 = len(m.fields[i:])
			}
		}

		switch m.fields[i:][i2] {
		case '"':
			// string field
			fieldMap[unescape(string(m.fields[i:][0:i1]), "fieldkey")] = unescape(string(m.fields[i:][i2+1:i3-1]), "fieldval")
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// number field
			switch m.fields[i:][i3-1] {
			case 'i':
				// integer field
				n, err := parseIntBytes(m.fields[i:][i2:i3-1], 10, 64)
				if err == nil {
					fieldMap[unescape(string(m.fields[i:][0:i1]), "fieldkey")] = n
				} else {
					// TODO handle error or just ignore field silently?
				}
			default:
				// float field
				n, err := parseFloatBytes(m.fields[i:][i2:i3], 64)
				if err == nil {
					fieldMap[unescape(string(m.fields[i:][0:i1]), "fieldkey")] = n
				} else {
					// TODO handle error or just ignore field silently?
				}
			}
		case 'T', 't':
			fieldMap[unescape(string(m.fields[i:][0:i1]), "fieldkey")] = true
		case 'F', 'f':
			fieldMap[unescape(string(m.fields[i:][0:i1]), "fieldkey")] = false
		default:
			// TODO handle unsupported field type
		}

		i += i3 + 1
	}

	return fieldMap
}

func (m *metric) Tags() map[string]string {
	tagMap := map[string]string{}
	if len(m.tags) == 0 {
		return tagMap
	}

	i := 0
	for {
		// start index of tag key
		i0 := indexUnescapedByte(m.tags[i:], ',') + 1
		if i0 == 0 {
			// didn't find a tag start
			break
		}
		// end index of tag key
		i1 := indexUnescapedByte(m.tags[i:], '=')
		// start index of tag value
		i2 := i1 + 1
		// end index of tag value (starting from i2)
		i3 := indexUnescapedByte(m.tags[i+i2:], ',')
		if i3 == -1 {
			tagMap[unescape(string(m.tags[i:][i0:i1]), "tagkey")] = unescape(string(m.tags[i:][i2:]), "tagval")
			break
		}
		tagMap[unescape(string(m.tags[i:][i0:i1]), "tagkey")] = unescape(string(m.tags[i:][i2:i2+i3]), "tagval")
		// increment start index for the next tag
		i += i2 + i3
	}

	return tagMap
}

func (m *metric) Name() string {
	return unescape(string(m.name), "name")
}

func (m *metric) Time() time.Time {
	// assume metric has been verified already and ignore error:
	if m.nsec == 0 {
		m.nsec, _ = parseIntBytes(m.t, 10, 64)
	}
	return time.Unix(0, m.nsec)
}

func (m *metric) UnixNano() int64 {
	// assume metric has been verified already and ignore error:
	if m.nsec == 0 {
		m.nsec, _ = parseIntBytes(m.t, 10, 64)
	}
	return m.nsec
}

func (m *metric) SetName(name string) {
	m.hashID = 0
	m.name = []byte(nameEscaper.Replace(name))
}

func (m *metric) SetPrefix(prefix string) {
	m.hashID = 0
	m.name = append([]byte(nameEscaper.Replace(prefix)), m.name...)
}

func (m *metric) SetSuffix(suffix string) {
	m.hashID = 0
	m.name = append(m.name, []byte(nameEscaper.Replace(suffix))...)
}

func (m *metric) AddTag(key, value string) {
	m.RemoveTag(key)
	m.tags = append(m.tags, []byte(","+escape(key, "tagkey")+"="+escape(value, "tagval"))...)
}

func (m *metric) HasTag(key string) bool {
	i := bytes.Index(m.tags, []byte(escape(key, "tagkey")+"="))
	if i == -1 {
		return false
	}
	return true
}

func (m *metric) RemoveTag(key string) {
	m.hashID = 0

	i := bytes.Index(m.tags, []byte(escape(key, "tagkey")+"="))
	if i == -1 {
		return
	}

	tmp := m.tags[0 : i-1]
	j := indexUnescapedByte(m.tags[i:], ',')
	if j != -1 {
		tmp = append(tmp, m.tags[i+j:]...)
	}
	m.tags = tmp
	return
}

func (m *metric) AddField(key string, value interface{}) {
	m.fields = append(m.fields, ',')
	m.fields = appendField(m.fields, key, value)
}

func (m *metric) HasField(key string) bool {
	i := bytes.Index(m.fields, []byte(escape(key, "tagkey")+"="))
	if i == -1 {
		return false
	}
	return true
}

func (m *metric) RemoveField(key string) error {
	i := bytes.Index(m.fields, []byte(escape(key, "tagkey")+"="))
	if i == -1 {
		return nil
	}

	var tmp []byte
	if i != 0 {
		tmp = m.fields[0 : i-1]
	}
	j := indexUnescapedByte(m.fields[i:], ',')
	if j != -1 {
		tmp = append(tmp, m.fields[i+j:]...)
	}

	if len(tmp) == 0 {
		return fmt.Errorf("Metric cannot remove final field: %s", m.fields)
	}

	m.fields = tmp
	return nil
}

func (m *metric) Copy() telegraf.Metric {
	return copyWith(m.name, m.tags, m.fields, m.t)
}

func copyWith(name, tags, fields, t []byte) telegraf.Metric {
	out := metric{
		name:   make([]byte, len(name)),
		tags:   make([]byte, len(tags)),
		fields: make([]byte, len(fields)),
		t:      make([]byte, len(t)),
	}
	copy(out.name, name)
	copy(out.tags, tags)
	copy(out.fields, fields)
	copy(out.t, t)
	return &out
}

func (m *metric) HashID() uint64 {
	if m.hashID == 0 {
		h := fnv.New64a()
		h.Write(m.name)

		tags := m.Tags()
		tmp := make([]string, len(tags))
		i := 0
		for k, v := range tags {
			tmp[i] = k + v
			i++
		}
		sort.Strings(tmp)

		for _, s := range tmp {
			h.Write([]byte(s))
		}

		m.hashID = h.Sum64()
	}
	return m.hashID
}

func appendField(b []byte, k string, v interface{}) []byte {
	if v == nil {
		return b
	}
	b = append(b, []byte(escape(k, "tagkey")+"=")...)

	// check popular types first
	switch v := v.(type) {
	case float64:
		b = strconv.AppendFloat(b, v, 'f', -1, 64)
	case int64:
		b = strconv.AppendInt(b, v, 10)
		b = append(b, 'i')
	case string:
		b = append(b, '"')
		b = append(b, []byte(escape(v, "fieldval"))...)
		b = append(b, '"')
	case bool:
		b = strconv.AppendBool(b, v)
	case int32:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case int16:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case int8:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case int:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint64:
		// Cap uints above the maximum int value
		var intv int64
		if v <= uint64(MaxInt) {
			intv = int64(v)
		} else {
			intv = int64(MaxInt)
		}
		b = strconv.AppendInt(b, intv, 10)
		b = append(b, 'i')
	case uint32:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint16:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint8:
		b = strconv.AppendInt(b, int64(v), 10)
		b = append(b, 'i')
	case uint:
		// Cap uints above the maximum int value
		var intv int64
		if v <= uint(MaxInt) {
			intv = int64(v)
		} else {
			intv = int64(MaxInt)
		}
		b = strconv.AppendInt(b, intv, 10)
		b = append(b, 'i')
	case float32:
		b = strconv.AppendFloat(b, float64(v), 'f', -1, 32)
	case []byte:
		b = append(b, v...)
	default:
		// Can't determine the type, so convert to string
		b = append(b, '"')
		b = append(b, []byte(escape(fmt.Sprintf("%v", v), "fieldval"))...)
		b = append(b, '"')
	}

	return b
}
