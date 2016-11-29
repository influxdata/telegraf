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

	// TODO remove
	"github.com/influxdata/influxdb/client/v2"
)

const MaxInt = int(^uint(0) >> 1)

var (
	// escaper is for escaping:
	//   - tag keys
	//   - tag values
	//   - field keys
	// see https://docs.influxdata.com/influxdb/v1.0/write_protocols/line_protocol_tutorial/#special-characters-and-keywords
	escaper = strings.NewReplacer(`,`, `\,`, `"`, `\"`, ` `, `\ `, `=`, `\=`)

	// nameEscaper is for escaping measurement names only.
	// see https://docs.influxdata.com/influxdb/v1.0/write_protocols/line_protocol_tutorial/#special-characters-and-keywords
	nameEscaper = strings.NewReplacer(`,`, `\,`, ` `, `\ `)

	// stringFieldEscaper is for escaping string field values only.
	// see https://docs.influxdata.com/influxdb/v1.0/write_protocols/line_protocol_tutorial/#special-characters-and-keywords
	stringFieldEscaper = strings.NewReplacer(`"`, `\"`)
)

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

	var thisType telegraf.ValueType
	if len(mType) > 0 {
		thisType = mType[0]
	} else {
		thisType = telegraf.Untyped
	}

	m := &metric{
		name:  []byte(nameEscaper.Replace(name)),
		t:     []byte(fmt.Sprint(t.UnixNano())),
		nsec:  t.UnixNano(),
		mType: thisType,
	}

	// pre-allocate exact size of the tags slice
	taglen := 0
	for k, v := range tags {
		taglen += 2 + len(escaper.Replace(k)) + len(escaper.Replace(v))
	}
	m.tags = make([]byte, taglen)

	i := 0
	for k, v := range tags {
		m.tags[i] = ','
		i++
		i += copy(m.tags[i:], escaper.Replace(k))
		m.tags[i] = '='
		i++
		i += copy(m.tags[i:], escaper.Replace(v))
	}

	// pre-allocate capacity of the fields slice
	fieldlen := 0
	for k, _ := range fields {
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
		if countBackslashes(buf, keyi-1)%2 == 0 {
			break
		} else {
			keyi++
		}
	}
	return keyi
}

// countBackslashes counts the number of preceding backslashes starting at
// the 'start' index.
func countBackslashes(buf []byte, index int) int {
	var count int
	for {
		if index < 0 {
			return count
		}
		if buf[index] == '\\' {
			count++
			index--
		} else {
			break
		}
	}
	return count
}

type metric struct {
	name   []byte
	tags   []byte
	fields []byte
	t      []byte

	mType     telegraf.ValueType
	aggregate bool

	// cached values for reuse in "get" functions
	hashID   uint64
	nsec     int64
	fieldMap map[string]interface{}
	tagMap   map[string]string
}

func (m *metric) Point() *client.Point {
	c, _ := client.NewPoint(m.Name(), m.Tags(), m.Fields(), m.Time())
	return c
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

func (m *metric) Fields() map[string]interface{} {
	if m.fieldMap != nil {
		// TODO should we return a copy?
		return m.fieldMap
	}

	m.fieldMap = map[string]interface{}{}
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
			m.fieldMap[string(m.fields[i:][0:i1])] = string(m.fields[i:][i2+1 : i3-1])
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// number field
			switch m.fields[i:][i3-1] {
			case 'i':
				// integer field
				n, err := strconv.ParseInt(string(m.fields[i:][i2:i3-1]), 10, 64)
				if err == nil {
					m.fieldMap[string(m.fields[i:][0:i1])] = n
				} else {
					// TODO handle error or just ignore field silently?
				}
			default:
				// float field
				n, err := strconv.ParseFloat(string(m.fields[i:][i2:i3]), 64)
				if err == nil {
					m.fieldMap[string(m.fields[i:][0:i1])] = n
				} else {
					// TODO handle error or just ignore field silently?
				}
			}
		case 'T', 't':
			m.fieldMap[string(m.fields[i:][0:i1])] = true
		case 'F', 'f':
			m.fieldMap[string(m.fields[i:][0:i1])] = false
		default:
			// TODO handle unsupported field type
		}

		i += i3 + 1
	}

	return m.fieldMap
}

func (m *metric) Tags() map[string]string {
	if m.tagMap != nil {
		// TODO should we return a copy?
		return m.tagMap
	}

	m.tagMap = map[string]string{}
	if len(m.tags) == 0 {
		return m.tagMap
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
			m.tagMap[string(m.tags[i:][i0:i1])] = string(m.tags[i:][i2:])
			break
		}
		m.tagMap[string(m.tags[i:][i0:i1])] = string(m.tags[i:][i2 : i2+i3])
		// increment start index for the next tag
		i += i2 + i3
	}

	return m.tagMap
}

func (m *metric) Name() string {
	return string(m.name)
}

func (m *metric) Time() time.Time {
	// assume metric has been verified already and ignore error:
	if m.nsec == 0 {
		m.nsec, _ = strconv.ParseInt(string(m.t), 10, 64)
	}
	return time.Unix(0, m.nsec)
}

func (m *metric) UnixNano() int64 {
	// assume metric has been verified already and ignore error:
	if m.nsec == 0 {
		m.nsec, _ = strconv.ParseInt(string(m.t), 10, 64)
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
	m.tags = append(m.tags, []byte(","+escaper.Replace(key)+"="+escaper.Replace(value))...)
}

func (m *metric) HasTag(key string) bool {
	i := bytes.Index(m.tags, []byte(escaper.Replace(key)+"="))
	if i == -1 {
		return false
	}
	return true
}

func (m *metric) RemoveTag(key string) bool {
	m.tagMap = nil
	m.hashID = 0

	i := bytes.Index(m.tags, []byte(escaper.Replace(key)+"="))
	if i == -1 {
		return false
	}

	tmp := m.tags[0 : i-1]
	j := indexUnescapedByte(m.tags[i:], ',')
	if j != -1 {
		tmp = append(tmp, m.tags[i+j:]...)
	}
	m.tags = tmp
	return true
}

func (m *metric) AddField(key string, value interface{}) {
	m.fieldMap = nil
	m.fields = append(m.fields, ',')
	appendField(m.fields, key, value)
}

func (m *metric) HasField(key string) bool {
	i := bytes.Index(m.fields, []byte(escaper.Replace(key)+"="))
	if i == -1 {
		return false
	}
	return true
}

func (m *metric) RemoveField(key string) bool {
	m.fieldMap = nil
	m.hashID = 0
	i := bytes.Index(m.fields, []byte(escaper.Replace(key)+"="))
	if i == -1 {
		return false
	}

	tmp := m.fields[0 : i-1]
	j := indexUnescapedByte(m.fields[i:], ',')
	if j != -1 {
		tmp = append(tmp, m.fields[i+j:]...)
	}
	m.fields = tmp
	return true
}

func (m *metric) Copy() telegraf.Metric {
	mOut := metric{
		name:   make([]byte, len(m.name)),
		tags:   make([]byte, len(m.tags)),
		fields: make([]byte, len(m.fields)),
		t:      make([]byte, len(m.t)),
	}
	copy(mOut.name, m.name)
	copy(mOut.tags, m.tags)
	copy(mOut.fields, m.fields)
	copy(mOut.t, m.t)
	return &mOut
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
	b = append(b, []byte(escaper.Replace(k)+"=")...)

	// check popular types first
	switch v := v.(type) {
	case float64:
		b = strconv.AppendFloat(b, v, 'f', -1, 64)
	case int64:
		b = strconv.AppendInt(b, v, 10)
		b = append(b, 'i')
	case string:
		b = append(b, '"')
		b = append(b, []byte(stringFieldEscaper.Replace(v))...)
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
	case nil:
		// skip
	default:
		// Can't determine the type, so convert to string
		b = append(b, '"')
		b = append(b, []byte(stringFieldEscaper.Replace(fmt.Sprintf("%v", v)))...)
		b = append(b, '"')
	}

	return b
}
