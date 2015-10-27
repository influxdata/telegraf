package testutil

import (
	"net"
	"net/url"
	"os"

	"github.com/influxdb/influxdb/client/v2"
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
func MockBatchPoints() client.BatchPoints {
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{})

	// Create a point and add to batch
	tags := map[string]string{"tag1": "value1"}
	fields := map[string]interface{}{"value": 1.0}
	pt := client.NewPoint("test_point", tags, fields)
	bp.AddPoint(pt)

	return bp
}
