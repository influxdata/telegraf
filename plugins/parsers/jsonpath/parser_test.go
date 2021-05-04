package jsonpath

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// *** Notes:***
// import cycle is caused trying to put influx line protocol expected output in a separate file, need influx line parser
// How to print telegraf.Metric to stdout? currently only get: file map[] map[name:John] 3600000000000
// Integration tests idea, completely separate test file that uses docker to run telegraf with a toml config
// Use: https://github.com/testcontainers/testcontainers-go
// Trying to load TOML config in unit tests a bit too complicated trying to get parser data

var DefaultTime = func() time.Time {
	return time.Unix(3600, 0)
}

const singleMetricValuesJSON = `
{
	"Name": "Device TestDevice1",
	"State": "ok"
}
`

type TestData struct {
	JSONData           []byte
	InfluxLineProtocol []byte
}

func ParseTestData(jsonFilepath string) (TestData, error) {
	var t TestData
	var err error
	t.JSONData, err = ioutil.ReadFile(jsonFilepath)
	if err != nil {
		return TestData{}, err
	}

	return t, nil
}

func TestJSONPath(t *testing.T) {

}

func TestParseLine(t *testing.T) {
	var tests = []struct {
		name           string
		JSONDataPath   string
		influxDataPath string
		configs        []Config
		expected       telegraf.Metric
	}{
		{
			name:         "Parse simple JSON data",
			JSONDataPath: "testdata/simple.json",
			configs: []Config{
				{
					MetricSelection: "name",
					MetricName:      "file",
					Fields: FieldKeys{
						FieldName: "name",
					},
				},
			},
			expected: testutil.MustMetric(
				"file",
				map[string]string{},
				map[string]interface{}{
					"name": "John",
				},
				DefaultTime(),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				Configs:  tt.configs,
				Log:      testutil.Logger{Name: "parsers.jsonpath"},
				TimeFunc: DefaultTime,
			}

			testData, err := ParseTestData(tt.JSONDataPath)
			require.NoError(t, err)

			actual, err := parser.ParseLine(string(testData.JSONData))
			require.NoError(t, err)

			fmt.Println(actual)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}
