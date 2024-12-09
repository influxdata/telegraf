package solr

import (
	"math"

	"github.com/influxdata/telegraf/internal"
)

// Get float64 from interface
func getFloat(value interface{}) float64 {
	v, err := internal.ToFloat64(value)
	if err != nil || math.IsNaN(v) {
		return 0
	}
	return v
}

// Get int64 from interface
func getInt(value interface{}) int64 {
	v, err := internal.ToInt64(value)
	if err != nil {
		return 0
	}

	return v
}
