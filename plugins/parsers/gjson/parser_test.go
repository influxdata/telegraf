package gjson

import (
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJsonPath(t *testing.T) {
	testString := `{
		"total_devices": 5,
		"total_threads": 10,
		"shares": {
			"total": 5,
			"accepted": 5,
			"rejected": 0,
			"avg_find_time": 4,
			"tester": "work",
			"tester2": true,
			"tester3": {
				"hello":"sup",
				"fun":"money",
				"break":9
			}
		}
	}`

	jsonParser := JSONPath{
		MetricName: "jsonpather",
		TagPath:    map[string]string{"hello": "shares.tester3.hello"},
		BoolPath:   map[string]string{"bool": "shares.tester2"},
	}

	metrics, err := jsonParser.Parse([]byte(testString))
	assert.NoError(t, err)
	log.Printf("m[0] name: %v, tags: %v, fields: %v", metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields())

}

func TestTagTypes(t *testing.T) {
	testString := `{
		"total_devices": 5,
		"total_threads": 10,
		"shares": {
			"total": 5,
			"accepted": 5,
			"rejected": 0,
			"my_bool": true,
			"tester": "work",
			"tester2": {
				"hello":"sup",
				"fun":true,
				"break":9.97
			}
		}
	}`

	r := JSONPath{
		TagPath:   map[string]string{"int1": "total_devices", "my_bool": "shares.my_bool"},
		FloatPath: map[string]string{"total": "shares.total"},
		BoolPath:  map[string]string{"fun": "shares.tester2.fun"},
		StrPath:   map[string]string{"hello": "shares.tester2"},
		IntPath:   map[string]string{"accepted": "shares.accepted"},
	}

	metrics, err := r.Parse([]byte(testString))
	log.Printf("m[0] name: %v, tags: %v, fields: %v", metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields())
	assert.NoError(t, err)
	assert.Equal(t, true, reflect.DeepEqual(map[string]interface{}{"total": 5.0, "fun": true, "hello": "sup", "accepted": int64(5)}, metrics[0].Fields()))
}
