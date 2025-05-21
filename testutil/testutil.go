package testutil

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

const (
	DefaultDelta   = 0.001
	DefaultEpsilon = 0.1
)

// GetLocalHost returns the DOCKER_HOST environment variable, parsing
// out any scheme or ports so that only the IP address is returned.
func GetLocalHost() string {
	dockerHostVar := os.Getenv("DOCKER_HOST")
	if dockerHostVar == "" {
		return "localhost"
	}

	u, err := url.Parse(dockerHostVar)
	if err != nil {
		return dockerHostVar
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return dockerHostVar
	}

	return host
}

// GetRandomString returns a random alphanumerical string of the given length.
// Please note, this function is different to `internal.RandomString` as it will
// not use `crypto.Rand` and will therefore not rely on the entropy-pool of the
// host which might be drained e.g. in CI pipelines. This is useful to e.g.
// create random passwords for tests where security is not a concern.
func GetRandomString(chars int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buffer := make([]byte, chars)
	for i := range buffer {
		//nolint:gosec // Using a weak random number generator on purpose to not drain entropy
		buffer[i] = charset[rand.Intn(len(charset))]
	}
	return string(buffer)
}

// MockMetrics returns a mock []telegraf.Metric object for using in unit tests
// of telegraf output sinks.
func MockMetrics() []telegraf.Metric {
	return []telegraf.Metric{TestMetric(1.0)}
}

func MockMetricsWithValue(value float64) []telegraf.Metric {
	return []telegraf.Metric{TestMetric(value)}
}

// TestMetric Returns a simple test point:
//
//	measurement -> "test1" or name
//	tags -> "tag1":"value1"
//	value -> value
//	time -> time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
func TestMetric(value interface{}, name ...string) telegraf.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	measurement := "test1"
	if len(name) > 0 {
		measurement = name[0]
	}

	return metric.New(
		measurement,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": value},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
}

// OnlyTags returns an option for keeping only "Tags" for a given Metric
func OnlyTags() cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		path := p.String()
		return path != "Tags" && path != ""
	}, cmp.Ignore())
}

func PrintMetrics(m []telegraf.Metric) {
	s := &influx.Serializer{
		SortFields:  true,
		UintSupport: true,
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	buf, err := s.SerializeBatch(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
}

// DefaultSampleConfig returns the sample config with the default parameters
// uncommented to also be able to test the validity of default setting.
func DefaultSampleConfig(sampleConfig string) []byte {
	re := regexp.MustCompile(`(?m)(^\s+)#\s*`)
	return []byte(re.ReplaceAllString(sampleConfig, "$1"))
}

func WithinDefaultDelta(dt float64) bool {
	return dt >= -DefaultDelta && dt <= DefaultDelta
}
