package lowercase

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func newM1() telegraf.Metric {
	m1, _ := metric.New("IIS_log",
		map[string]string{
			"verb":           "GET",
            "s-computername": "MIXEDCASE_hostname",
		},
		map[string]interface{}{
            "request": "/mixed/CASE/paTH/?from=-1D&to=now",
		},
		time.Now(),
	)
	return m1
}

func newM2() telegraf.Metric {
	m2, _ := metric.New("IIS_log",
		map[string]string{
			"verb":           "GET",
			"resp_code":      "200",
            "s-computername": "MIXEDCASE_hostname",
		},
		map[string]interface{}{
            "request":       "/mixed/CASE/paTH/?from=-1D&to=now",
            "cs-host":       "AAAbbb",
			"ignore_number": int64(200),
			"ignore_bool":   true,
		},
		time.Now(),
	)
	return m2
}

func TestFieldConversions(t *testing.T) {
	tests := []struct {
		message        string
		converter      converter
		expectedFields map[string]interface{}
	}{
		{
			message: "Should change existing field",
			converter: converter{
				Key:         "request",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/case/path/?from=-1d&to=now",
			},
		},
		{
			message: "Should add new field",
			converter: converter{
				Key:         "request",
				ResultKey:   "lowercase_request",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/CASE/paTH/?from=-1D&to=now",
                "lowercase_request": "/mixed/case/path/?from=-1d&to=now",
			},
		},
	}

	for _, test := range tests {
		lowercase := &Lowercase{}
		lowercase.Fields = []converter{
			test.converter,
		}

		processed := lowercase.Apply(newM1())

		expectedTags := map[string]string{
			"verb":           "GET",
            "s-computername": "MIXEDCASE_hostname",
		}

		assert.Equal(t, test.expectedFields, processed[0].Fields(), test.message)
		assert.Equal(t, expectedTags, processed[0].Tags(), "Should not change tags")
		assert.Equal(t, "IIS_log", processed[0].Name(), "Should not change name")
	}
}

func TestTagConversions(t *testing.T) {
	tests := []struct {
		message      string
		converter    converter
		expectedTags map[string]string
	}{
		{
			message: "Should change existing tag",
			converter: converter{
				Key:         "s-computername",
			},
			expectedTags: map[string]string{
				"verb":    "GET",
                "s-computername": "mixedcase_hostname",
			},
		},
		{
			message: "Should add new tag",
			converter: converter{
				Key:         "s-computername",
				ResultKey:   "s-computername_lowercase",
			},
			expectedTags: map[string]string{
				"verb":       "GET",
                "s-computername": "MIXEDCASE_hostname",
                "s-computername_lowercase": "mixedcase_hostname",
			},
		},
	}

	for _, test := range tests {
		lowercase := &Lowercase{}
		lowercase.Tags = []converter{
			test.converter,
		}

		processed := lowercase.Apply(newM1())

		expectedFields := map[string]interface{}{
            "request": "/mixed/CASE/paTH/?from=-1D&to=now",
		}

		assert.Equal(t, expectedFields, processed[0].Fields(), test.message, "Should not change fields")
		assert.Equal(t, test.expectedTags, processed[0].Tags(), test.message)
		assert.Equal(t, "IIS_log", processed[0].Name(), "Should not change name")
	}
}

func TestMultipleConversions(t *testing.T) {
	lowercase := &Lowercase{}
	lowercase.Tags = []converter{
		{
			Key:         "verb",
		},
		{
			Key:         "s-computername",
		},
	}
	lowercase.Fields = []converter{
		{
			Key:         "request",
		},
		{
			Key:         "cs-host",
			ResultKey:   "cs-host_lowercase",
		},
	}

	processed := lowercase.Apply(newM2())

	expectedFields := map[string]interface{}{
        "request":           "/mixed/case/path/?from=-1d&to=now",
		"ignore_number":     int64(200),
		"ignore_bool":       true,
        "cs-host":           "AAAbbb",
        "cs-host_lowercase": "aaabbb",
	}
	expectedTags := map[string]string{
		"verb":           "get",
        "resp_code":      "200",
        "s-computername": "mixedcase_hostname",
	}

	assert.Equal(t, expectedFields, processed[0].Fields())
	assert.Equal(t, expectedTags, processed[0].Tags())
}

func TestNoKey(t *testing.T) {
	tests := []struct {
		message        string
		converter      converter
		expectedFields map[string]interface{}
	}{
		{
			message: "Should not change anything if there is no field with given key",
			converter: converter{
				Key:         "not_exists",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/CASE/paTH/?from=-1D&to=now",
			},
		},
	}

	for _, test := range tests {
		lowercase := &Lowercase{}
		lowercase.Fields = []converter{
			test.converter,
		}

		processed := lowercase.Apply(newM1())

		assert.Equal(t, test.expectedFields, processed[0].Fields(), test.message)
	}
}
