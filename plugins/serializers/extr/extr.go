package extr

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"math"
	"strings"
	"time"
	"unicode"
)

// Fixed map keys
var deviceStr string = "device"
var itemsStr string = "items"
var tsStr string = "ts"
var keysStr string = "keys"
var nameStr string = "name"

var keyStr string = "key"
var minStr string = "min"
var maxStr string = "max"
var avgStr string = "avg"

type serializer struct {
	TimestampUnits  time.Duration
	JsonBatchFields bool
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

// We SerializeBatch by default, so this function will never be called
func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	m := s.createObject(metric)
	serialized, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

// - Combine sequential metrics whose Name, Tags, and TimeStamp match into a single batched metric
//   where the fields are grouped into an array.
// - Group these batched metrics into a single toplevel map
// - Group min/max/avg into new fields
// - Change InFluxDB naming:
//     tags ------> device
//     timestamp -> ts
//     fields ----> items
//
func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized []byte
	var err error
	var toplevelObjName string

	// Top level map of grouped objects
	// i.e. "cpuStats": []
	//      toplevel["cpuStats"] = metricARRAY

	toplevel := make(map[string]interface{})

	// Metric object
	object := make(map[string]interface{}, 4)

	for _, metric := range metrics {

		if object[itemsStr] == nil {
			// First batched metric object
			object = s.createObject(metric)

		} else if s.metricMatch(object, metric) {
			// This metric's name, tags, and timestamp matches first metric.

			item := createItem(metric)

			itemsArray := object[itemsStr].([]interface{})

			// Append field to itemsArray and re-assign new array to current metric object
			// "items" (itemsStr) are batched LIKE metrics
			object[itemsStr] = append(itemsArray, item)

		} else {
			// This metric does not match first metric parameters.
			// Done with the last batched metric object

			// Get object "name" (nameStr) and convert to first character to lowercase.
			toplevelObjName = firstCharToLower(object[nameStr].(string))
			if _, found := toplevel[toplevelObjName]; !found {

				// Allocate slice for toplevel object name
				toplevel[toplevelObjName] = make([]interface{}, 0)
			}

			// Append this objects to toplevel[toplevelObjName] slice
			toplevel[toplevelObjName] = append(toplevel[toplevelObjName].([]interface{}), object)

			// Create a new batched metric object with current metric
			object = s.createObject(metric)
		}
	}

	// Append this last object to toplevel[toplevelObjName] slice
	if _, found := object[nameStr]; found {

		toplevelObjName = firstCharToLower(object[nameStr].(string))
		if _, found := toplevel[toplevelObjName]; !found {

			// Allocate slice for toplevel object name
			toplevel[toplevelObjName] = make([]interface{}, 0)
		}

		toplevel[toplevelObjName] = append(toplevel[toplevelObjName].([]interface{}), object)
	}

	serialized, err = json.Marshal(toplevel)
	if err != nil {
		return []byte{}, err
	}

	return serialized, nil
}

// Match on metric name, timestmp, and tags.
func (s *serializer) metricMatch(newMetric map[string]interface{}, metric telegraf.Metric) bool {

	tags := newMetric[deviceStr].(map[string]string)
	name := newMetric[nameStr]
	timestamp := newMetric[tsStr]

	if name != metric.Name() {
		return false
	}

	if timestamp != metric.Time().UnixNano()/int64(s.TimestampUnits) {
		return false
	}

	for _, tag := range metric.TagList() {
		if tags[tag.Key] != tag.Value {
			return false
		}
	}

	return true
}

func (s *serializer) createObject(metric telegraf.Metric) map[string]interface{} {
	m := make(map[string]interface{}, 4)
	itemsArray := make([]interface{}, 0)

	tags := make(map[string]string, len(metric.TagList()))
	for _, tag := range metric.TagList() {
		tags[tag.Key] = tag.Value
	}
	m[deviceStr] = tags

	// Create a items array
	f := createItem(metric)
	itemsArray = append(itemsArray, f)
	m[itemsStr] = itemsArray

	m[nameStr] = metric.Name()
	m[tsStr] = metric.Time().UnixNano() / int64(s.TimestampUnits)
	return m
}

// Create a metric item for append into "items" slice, grouping certain fields.
//   Group _key fields into "keys" (keysStr) map
//   Group xyz_min, xyz_max, xyz_avg fields into "xyz" map
func createItem(metric telegraf.Metric) map[string]interface{} {

	item := make(map[string]interface{}, len(metric.FieldList()))

	for _, field := range metric.FieldList() {

		switch fv := field.Value.(type) {
		case float64:
			// JSON does not support these special values
			if math.IsNaN(fv) || math.IsInf(fv, 0) {
				continue
			}
		}

		// Check if we need to do a _key or _min,_max,_avg grouping.
		// ex. {"ifIndex_key":1} {"name_key":"2:1"}
		//      --> "keys":{"ifIndex":1,"name":"2:1"}
		// ex. {"pwm_avg":26.7} {"pwm_max":99.9} {"pwm_min":21.1}
		//      --> "pwm":{"avg":26.7,"max":99.9,"min":21.1}

		id, key, value := splitMetricFieldId(field.Key, field.Value)

		if id == "key" {

			// Found a "_key" field.  Group this metric field under "keys" map.

			var mType map[string]interface{}

			// Check if keys[key] exists
			if _, found := item[keysStr]; !found {
				item[keysStr] = make(map[string]interface{})
			}

			mType = item[keysStr].(map[string]interface{})

			mType[key] = value

		} else if id == minStr || id == maxStr || id == avgStr {

			// Found _min,_max,_avg field. Do grouping.

			var mType map[string]interface{}

			// Check if name[key] exists
			if _, found := item[key]; !found {
				item[key] = make(map[string]interface{})
			}

			mType = item[key].(map[string]interface{})
			mType[id] = value

		} else {

			// Not a key or min/max/avg field.
			item[key] = value
		}
	}

	return item
}

func truncateDuration(units time.Duration) time.Duration {
	// Default precision is 1s
	if units <= 0 {
		return time.Second
	}

	// Search for the power of ten less than the duration.
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}

// Takes a Field entry and looks for special identifiers, min,max,avg or key as
// last _ seperated element
// i.e. usage_max, core_key
//
// Parameter:
//     fKey  (i.e. usage_max | core_key | etc.)
//     fValue
// Returns:
//     id    - special field identifier or nil. (i.e. "max", "key", nil)
//     key   - string before special field identifier. (i.e. "usage" | "core" | "linkCnt")
//     value - value
func splitMetricFieldId(fKey string, fValue interface{}) (id string, key string, value interface{}) {

	var mcut string

	s := strings.Split(fKey, "_")

	// If last element matches "min","max","avg" or "key" string
	lastElem := s[len(s)-1]

	if lastElem == minStr || lastElem == maxStr || lastElem == avgStr || lastElem == keyStr {
		id = lastElem
		mcut = "_" + lastElem
		key = strings.Replace(fKey, mcut, "", -1)
	} else {
		key = fKey
	}

	value = fValue

	return
}

func firstCharToLower(str string) string {

	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}

	return ""
}
