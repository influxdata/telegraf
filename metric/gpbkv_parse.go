package metric

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"hash/maphash"
	"log"
	"strings"
	"time"
)

const (
	GBPVALUE  = "ValueByType"
	GBPFIELDS = "fields"
	GBPNAME = "name"
)


type gbpRow struct {
	Timestamp float64
	// 直接解析成 content, keys 可能会出问题，
	// 可以将 content，keys 包含在 Rows 中比较保险
	//Rows interface{}
	Content interface{}
	Keys interface{}
}

type SDNMetric struct {
	Row []gbpRow
	Telemetry map[string]interface{} `json:Telemetry`
	Source string `json:Source`
}

type SDNGrouper struct {
	metrics map[uint64]telegraf.Metric
	ordered []telegraf.Metric

	hashSeed maphash.Seed
}

func NewSDNGrouper() *SDNGrouper {
	return &SDNGrouper{
		metrics:  make(map[uint64]telegraf.Metric),
		ordered:  []telegraf.Metric{},
		hashSeed: maphash.MakeSeed(),
	}
}

func (g *SDNGrouper) Add(
	measurement string,
	tm time.Time,
	field string,
	fieldValue interface{},
) error {
	id := groupID(g.hashSeed, measurement, nil, tm)
	m := g.metrics[id]
	if m == nil {
		m = New(measurement, nil, map[string]interface{}{field: fieldValue}, tm)
		g.metrics[id] = m
		g.ordered = append(g.ordered, m)
		m.AddField(field, fieldValue)
	} else {
		m.AddField(field, fieldValue)
	}
	return nil
}

func (g *SDNGrouper) AddMetric(
	metric telegraf.Metric,
) {
	id := groupID(g.hashSeed, metric.Name(), metric.TagList(), metric.Time())
	m := g.metrics[id]
	if m == nil {
		m = metric.Copy()
		g.metrics[id] = m
		g.ordered = append(g.ordered, m)
	} else {
		for _, f := range metric.FieldList() {
			m.AddField(f.Key, f.Value)
		}
	}
}

func (g *SDNGrouper) Metrics() []telegraf.Metric {
	return g.ordered
}

func GbpkvParse(data []byte, sourceIP string) SDNMetric {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Println("unmarshal json err: ", err)
	}

	gpbkv := v.(map[string]interface{})

	var rows []gbpRow
	telemetry := make(map[string]interface{}, 0)

	for k, v := range gpbkv {
		if k != "data_gpbkv" {
			parseTelemetry(telemetry, k, v)
		}else {
			rows = parseRow(v)
		}
	}

	m := SDNMetric{
		Telemetry: telemetry,
		Row: rows,
		Source: sourceIP,
	}

	//mtericByte, err := json.Marshal(m)
	//if err != nil {
	//	log.Println(err)
	//}

	return m
}

func parseTelemetry(telemetry map[string]interface{},key string, value interface{})  {
	if e, ok := value.(map[string]interface{}); ok {
		for k, v := range e {
			k = CamelCaseToUnderscore(k)
			telemetry[k] = decodeValue(v)
		}
	}else {
		key = CamelCaseToUnderscore(key)
		telemetry[key] = decodeValue(value)
	}
}

func parseRow(value interface{}) []gbpRow {
	var rowArr []gbpRow
	for _, arr := range value.([]interface{}){
		var row gbpRow
		data := arr.(map[string]interface{})
		if data[GBPVALUE] == nil && data[GBPFIELDS] != nil {
			row.Timestamp = data["timestamp"].(float64)
			field := parseFields(data[GBPFIELDS])
			if _, ok := field["content"]; ok {
				row.Content = field["content"]
			}else {
				log.Println("No field named content")
			}
			if _, ok := field["keys"]; ok {
				row.Keys = field["keys"]
			}else {
				log.Println("No field named keys")
			}
			//row.Rows = field
			rowArr = append(rowArr, row)
		}
	}
	return rowArr
}

func parseFields(v interface{}) map[string]interface{} {
	s := make(map[string]interface{})
	placeInArrayMap := map[string]bool{}
	for _, arr := range v.([]interface{}) {
		field := arr.(map[string]interface{})
		key := field[GBPNAME].(string)
		var fieldVal interface{}
		var hint int
		// 判断值是否已经存在 field 中，如果已经存在，替换值为数组类型，并标识为已经替换
		existingEntry, exists := s[key]
		_, placeInArray := placeInArrayMap[key]
		_, children := field[GBPFIELDS]
		if !children {
			fieldVal = field[GBPVALUE]
			if fieldVal != nil {
				for _, value := range fieldVal.(map[string]interface{}) {
					fieldVal = value
				}
			}
			hint = 10
		}else {
			fieldVal = parseFields(field[GBPFIELDS])
			hint = len(field[GBPFIELDS].([]interface{}))
		}

		if !placeInArray && !exists {
			// this is the common case by far!
			s[key] = fieldVal
		}else {
			newName := key + GPB_XINQXION_EDIT_SUFFIX
			if exists {
				if !placeInArray {
					// Create list
					s[newName] = make([]interface{}, 0, hint)
					// Remember that this field name is arrayified(!?)
					placeInArrayMap[key] = true
					// Add existing entry to new array)
					s[newName] = append(s[newName].([]interface{}),existingEntry)
					// Delete existing entry from old
					delete(s, key)
					placeInArray = true
				} else {
					log.Println("GPB KV inconsistency, processing repeated field names")
				}
			}
			if placeInArray && fieldVal != nil {
				s[newName] = append(s[newName].([]interface{}), fieldVal)
			}
		}
	}
	return s
}

func decodeValue(v interface{}) interface{} {
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

// 驼峰转下划线
func CamelCaseToUnderscore(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		// or通过ASCII码进行大小写的转化
		// 65-90（A-Z），97-122（a-z）
		//判断如果字母为大写的A-Z就在前面拼接一个_
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	//ToLower把大写字母统一转小写
	return strings.ToLower(string(data[:]))
}
