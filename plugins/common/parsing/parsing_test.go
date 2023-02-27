package parsing

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		sut         func() *Config
		expectedErr error
	}{
		{
			name: "no error with nil entries",
			sut: func() *Config {
				return NewConfig(nil, "", "/", "+", testutil.Logger{})
			},
		},
		{
			name: "no error with empty entries",
			sut: func() *Config {
				return NewConfig(make([]ConfigEntry, 0), "", "/", "+", testutil.Logger{})
			},
		},
		{
			name: "empty delimiter",
			sut: func() *Config {
				return NewConfig(make([]ConfigEntry, 0), "", "", "+", testutil.Logger{})
			},
			expectedErr: ErrEmptyDelimiter,
		},
		{
			name: "empty delimiter",
			sut: func() *Config {
				return NewConfig(make([]ConfigEntry, 0), "", "/", "", testutil.Logger{})
			},
			expectedErr: ErrEmptyWildcard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := tt.sut()
			err := sut.Init()
			require.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestParsing(t *testing.T) {
	tests := []struct {
		name          string
		topic         string
		expectedError error
		topicParsing  []ConfigEntry
		expected      telegraf.Metric
	}{
		{
			name:  "topic parsing configured",
			topic: "telegraf/123/test",
			topicParsing: []ConfigEntry{
				{
					Base:        "telegraf/123/test",
					Measurement: "_/_/measurement",
					Tags:        "testTag/_/_",
					Fields:      "_/testNumber/_",
					FieldTypes: map[string]string{
						"testNumber": "int",
					},
				},
			},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"testTag": "telegraf",
				},
				map[string]interface{}{
					"testNumber": 123,
				},
				time.Unix(0, 0),
			),
		},
		{
			name:  "topic parsing configured with a mqtt wild card `+`",
			topic: "telegraf/123/test/hello",
			topicParsing: []ConfigEntry{
				{
					Base:        "telegraf/+/test/hello",
					Measurement: "_/_/measurement/_",
					Tags:        "testTag/_/_/_",
					Fields:      "_/testNumber/_/testString",
					FieldTypes: map[string]string{
						"testNumber": "int",
					},
				},
			},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"testTag": "telegraf",
				},
				map[string]interface{}{
					"testNumber": 123,
					"testString": "hello",
				},
				time.Unix(0, 0),
			),
		},
		{
			name:          "topic parsing configured incorrectly",
			topic:         "telegraf/123/test/hello",
			expectedError: fmt.Errorf("config error topic parsing: fields length does not equal topic length"),
			topicParsing: []ConfigEntry{
				{
					Measurement: "_/_/measurement/_",
					Tags:        "testTag/_/_/_",
					Fields:      "_/_/testNumber:int/_/testString:string",
					FieldTypes: map[string]string{
						"testNumber": "int",
					},
					Base: "telegraf/+/test/hello",
				},
			},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"testTag": "telegraf",
				},
				map[string]interface{}{
					"testNumber": 123,
					"testString": "hello",
				},
				time.Unix(0, 0),
			),
		},
		{
			name:  "topic parsing configured without fields",
			topic: "telegraf/123/test/hello",
			topicParsing: []ConfigEntry{
				{
					Measurement: "_/_/measurement/_",
					Tags:        "testTag/_/_/_",
					FieldTypes: map[string]string{
						"testNumber": "int",
					},
					Base: "telegraf/+/test/hello",
				},
			},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"testTag": "telegraf",
				},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
		},
		{
			name:  "topic parsing configured without measurement",
			topic: "telegraf/123/test/hello",
			topicParsing: []ConfigEntry{
				{
					Tags:   "testTag/_/_/_",
					Fields: "_/testNumber/_/testString",
					FieldTypes: map[string]string{
						"testNumber": "int",
					},
					Base: "telegraf/+/test/hello",
				},
			},
			expected: testutil.MustMetric(
				"default",
				map[string]string{
					"testTag": "telegraf",
				},
				map[string]interface{}{
					"testNumber": 123,
					"testString": "hello",
				},
				time.Unix(0, 0),
			),
		},
		{
			name:  "topic parsing configured topic with a prefix `/`",
			topic: "/telegraf/123/test/hello",
			topicParsing: []ConfigEntry{
				{
					Measurement: "/_/_/measurement/_",
					Tags:        "/testTag/_/_/_",
					Fields:      "/_/testNumber/_/testString",
					FieldTypes: map[string]string{
						"testNumber": "int",
					},
					Base: "/telegraf/+/test/hello",
				},
			},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"testTag": "telegraf",
				},
				map[string]interface{}{
					"testNumber": 123,
					"testString": "hello",
				},
				time.Unix(0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewConfig(tt.topicParsing, "topic", "/", "+", testutil.Logger{})
			err := parser.Init()
			require.Equal(t, tt.expectedError, err)
			if tt.expectedError != nil {
				return
			}

			m := metric.New("default", make(map[string]string), make(map[string]any), time.Unix(0, 0))
			err = parser.Parse(tt.topic, m)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, m,
				testutil.IgnoreTime())
		})
	}
}
