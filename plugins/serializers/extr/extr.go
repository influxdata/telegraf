package extr

import (
	"encoding/json"
	"errors"
	"github.com/influxdata/telegraf"
	"math"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// Fixed map keys
var deviceStr string = "device"
var itemsStr string = "items"
var tsStr string = "ts"
var nameStr string = "name"
var keysStr string = "keys"
var tagsStr string = "tags"

// Special appended grouping identifiers
var keyStr string = "key"
var tagStr string = "tag"
var minStr string = "min"
var maxStr string = "max"
var avgStr string = "avg"
var newStr string = "new"
var oldStr string = "old"

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
		// ex. {"ifAdminStatus_old":0} {"ifAdminStatus_new":1}
		//      --> "ifAdminStatus":{"old":0,"new":1}

		id, fKey, fValue := splitMetricFieldId(field.Key, field.Value)

		if id == keyStr || id == tagStr || id == "" {

			var itemKey string

			if id == keyStr {
				// Found a "_key" field. Group this metric field under "keys" map.
				itemKey = keysStr
			} else if id == tagStr {
				// Found a "_tag" field. Group this metric field under "tags" map.
				itemKey = tagsStr
			} else {
				// No id field. Make this a top level item
				itemKey = ""
			}

			myMap, myIndex, isArray, err := splitKey(item, itemKey, fKey)
			if err != nil {
				continue
			}

			if isArray == true {

				rt := reflect.TypeOf(myMap[myIndex])
				if rt.Kind() != reflect.Slice {
					// Make it a slice
					myMap[myIndex] = make([]interface{}, 0)
				}
				
				myMap[myIndex] = append(myMap[myIndex].([]interface{}), fValue)

			} else {
				myMap[myIndex] = fValue
			}

		} else if id == minStr || id == maxStr || id == avgStr || id == oldStr || id == newStr {

			// Found _min,_max,_avg or _old,_new field.
			itemKey := ""

			myMap, myIndex, _, err := splitKey(item, itemKey, fKey)
			if err != nil {
				continue
			}

			mMap := myMap[myIndex].(map[string]interface{})
			mMap[id] = fValue
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

// Takes a Field entry and looks for special identifiers, min,max,avg or key or old,new as
// last _ seperated element
// i.e. usage_max, core_key
//
// Parameter:
//     fKey  (i.e. usage_max | core_key | ifAdminStatus_old etc.)
//     fValue
// Returns:
//     id    - special field identifier or nil. (i.e. "max", "key", nil)
//     key   - string before special field identifier. (i.e. "usage" | "core" | "linkCnt")
//     value - value
func splitMetricFieldId(fKey string, fValue interface{}) (id string, key string, value interface{}) {

	var mcut string

	s := strings.Split(fKey, "_")

	// If last element matches "min","max","avg" or "key" or "old","new" string
	lastElem := s[len(s)-1]

	if lastElem == minStr || lastElem == maxStr || lastElem == avgStr ||
		lastElem == keyStr || lastElem == oldStr || lastElem == newStr || lastElem == tagStr {

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

func buildMap(startMap map[string]interface{}, startKey string, s []string, numKeys int) (
	map[string]interface{}, string, error) {

	if numKeys == 0 {
		return startMap, startKey, errors.New("numKeys 0.")

	} else {

		if len(startKey) == 0 {
			return nil, "", errors.New("No start key string provided")
		}

		sIndex := numKeys - 1

		_, ok := startMap[startKey].(map[string]interface{})
		if !ok {
			// This occurs if bad syntax is used for influx field
			// i.e. car_tag=Ford and engine_car_tag=F150.
			// Once we create..
			//    {tags:{car: Ford}}
			//"ab" We cannot create a
			//    {tags:{car:{engine:F150}}}
			// because the top level "car" element is already assigned the value "Ford" and thus
			// cannot contain any subvalues.
			// Expecting and map value for startMap[startKey]
			return nil, "", errors.New("Expecting map value. Invalid metric name syntax.")
		}

		myMap := startMap[startKey].(map[string]interface{})
		if _, found := myMap[s[sIndex]]; !found {
			myMap[s[sIndex]] = make(map[string]interface{})
		}

		myMap, myKey, _ := buildMap(myMap, s[sIndex], s, sIndex)
		return myMap, myKey, nil
	}
}

// Split key into tokens and build nested map based on elements of fKey
// Top level map is passed along with that maps name. i.e. input["keys"] or input["tags"]
func splitKey(topMap map[string]interface{}, topMapName string, fKey string) (map[string]interface{}, string, bool, error) {

	// topMap - Current items array items[]
	// topMapName - Top map name
	// fKey - Current key string i.e. j_foo_bar, k_foo_bar
	isArray := false

	// Split fKey into tokens.  Already removed type, _key, _min, _max, etc.
	// i.e.           rtrId   // 1 level
	//             name_vrd   // 2-level
	//      value1_opt1_vrf   // 3-level

	// Accounts for keeping _ in name vs nesting.
	// i.e. foo/_bar:100 --> foo_bar:100
	//      foo_bar:100 --> {bar:{foo:100}}
	fKey = strings.ReplaceAll(fKey, "/_", "/")

	// Returns an array of strings
	// i.e. j_foo_bar --> [j foo bar]
	s := strings.Split(fKey, "_")

	// Check if this is an array @j_foo_bar.  First letter of first
	// element is @ sign.
	if strings.HasPrefix(s[0], "@") {
		isArray = true
		// Get rid of array identifier @xxx_
		s = s[1:]
	}

	// Convert / to _
	for i, _ := range s {
		s[i] = strings.ReplaceAll(s[i], "/", "_")
	}

	// If no top level map name is passed, use the first element of array.
	// This is the case for _min,_max,_avg and _old,_new fields.
	if len(topMapName) == 0 {
		// use last element
		topMapName = s[len(s)-1]
		// Remove last element from array
		s = s[:len(s)-1]
	}

	// Allocate map if not there.
	// i.e. topMap["keys"] or topMap["tags"]
	if _, found := topMap[topMapName]; !found {
		topMap[topMapName] = make(map[string]interface{})
	}

	numKeys := len(s)

	if len(s) <= 0 {
		// This is okay for a top level data element.. i.e.psuTemp_min=100
		return topMap, topMapName, isArray, nil
	}

	myMap, myIndex, err := buildMap(topMap, topMapName, s, numKeys)

	return myMap, myIndex, isArray, err
}
