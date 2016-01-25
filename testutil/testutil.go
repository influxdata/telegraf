package testutil

import (
	"net"
	"net/url"
	"os"
	"time"

	"github.com/influxdata/telegraf/models"
)

var localhost = "localhost"

// GetLocalHost returns the DOCKER_HOST environment variable, parsing
// out any scheme or ports so that only the IP address is returned.
func GetLocalHost() string {
	if dockerHostVar := os.Getenv("DOCKER_HOST"); dockerHostVar != "" {
		u, err := url.Parse(dockerHostVar)
		if err != nil {
			return dockerHostVar
		}

		// split out the ip addr from the port
		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return dockerHostVar
		}

		return host
	}
	return localhost
}

// MockBatchPoints returns a mock BatchPoints object for using in unit tests
// of telegraf output sinks.
func MockBatchPoints() []models.Metric {
	// Create a new point batch
	ret := []models.Metric{}
	ret = append(ret, TestPoint(1.0))
	return ret
}

// TestPoint Returns a simple test point:
//     measurement -> "test1" or name
//     tags -> "tag1":"value1"
//     value -> value
//     time -> time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
func TestPoint(value interface{}, name ...string) models.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	measurement := "test1"
	if len(name) > 0 {
		measurement = name[0]
	}
	tags := map[string]string{"tag1": "value1"}
	pt, _ := models.NewMetric(
		measurement,
		tags,
		map[string]interface{}{"value": value},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	return pt
}
