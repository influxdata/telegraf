package strings

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
		lowercase      converter
        uppercase      converter
        trim           converter
        trimleft       converter
        trimright      converter
        trimprefix     converter
        trimsuffix     converter
		expectedFields map[string]interface{}
	}{
		{
			message: "Should change existing field to lowercase",
			lowercase: converter{
				Field:         "request",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/case/path/?from=-1d&to=now",
			},
		},
		{
			message: "Should change existing field to uppercase",
			uppercase: converter{
				Field:         "request",
			},
			expectedFields: map[string]interface{}{
                "request": "/MIXED/CASE/PATH/?FROM=-1D&TO=NOW",
			},
		},
		{
			message: "Should add new lowercase field",
			lowercase: converter{
				Field:         "request",
				ResultKey:   "lowercase_request",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/CASE/paTH/?from=-1D&to=now",
                "lowercase_request": "/mixed/case/path/?from=-1d&to=now",
			},
		},
		{
			message: "Should trim from both sides",
			trim: converter{
				Field:         "request",
                Argument:      "/w",
			},
			expectedFields: map[string]interface{}{
                "request": "mixed/CASE/paTH/?from=-1D&to=no",
			},
		},
		{
			message: "Should trim from both sides and make lowercase",
			trim: converter{
				Field:         "request",
                Argument:      "/w",
			},
            lowercase: converter{
                Field:         "request",
            },
			expectedFields: map[string]interface{}{
                "request": "mixed/case/path/?from=-1d&to=no",
			},
		},
		{
			message: "Should trim from left side",
			trimleft: converter{
				Field:         "request",
                Argument:      "/w",
			},
			expectedFields: map[string]interface{}{
                "request": "mixed/CASE/paTH/?from=-1D&to=now",
			},
		},
		{
			message: "Should trim from right side",
			trimright: converter{
				Field:         "request",
                Argument:      "/w",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/CASE/paTH/?from=-1D&to=no",
			},
		},
		{
			message: "Should trim prefix '/mixed'",
			trimprefix: converter{
				Field:         "request",
                Argument:      "/mixed",
			},
			expectedFields: map[string]interface{}{
                "request": "/CASE/paTH/?from=-1D&to=now",
			},
		},
		{
			message: "Should trim suffix '-1D&to=now'",
			trimprefix: converter{
				Field:         "request",
                Argument:      "-1D&to=now",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/CASE/paTH/?from=-1D&to=now",
			},
		},
	}

	for _, test := range tests {
		strings := &Strings{}
		strings.Lowercase = []converter{
			test.lowercase,
		}
        strings.Uppercase = []converter{
            test.uppercase,
        }
        strings.Trim = []converter{
            test.trim,
        }
        strings.TrimLeft = []converter{
            test.trimleft,
        }
        strings.TrimRight = []converter{
            test.trimright,
        }
        strings.TrimPrefix = []converter{
            test.trimprefix,
        }
        strings.TrimSuffix = []converter{
            test.trimsuffix,
        }

		processed := strings.Apply(newM1())

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
		lowercase    converter
        uppercase    converter
		expectedTags map[string]string
	}{
		{
			message: "Should change existing tag to lowercase",
			lowercase: converter{
				Tag:         "s-computername",
			},
			expectedTags: map[string]string{
				"verb":    "GET",
                "s-computername": "mixedcase_hostname",
			},
		},
		{
			message: "Should add new lowercase tag",
			lowercase: converter{
				Tag:         "s-computername",
				ResultKey:   "s-computername_lowercase",
			},
			expectedTags: map[string]string{
				"verb":       "GET",
                "s-computername": "MIXEDCASE_hostname",
                "s-computername_lowercase": "mixedcase_hostname",
			},
		},
		{
			message: "Should add new uppercase tag",
			uppercase: converter{
				Tag:         "s-computername",
				ResultKey:   "s-computername_uppercase",
			},
			expectedTags: map[string]string{
				"verb":       "GET",
                "s-computername": "MIXEDCASE_hostname",
                "s-computername_uppercase": "MIXEDCASE_HOSTNAME",
			},
		},
	}

	for _, test := range tests {
		strings := &Strings{}
		strings.Lowercase = []converter{
			test.lowercase,
		}
		strings.Uppercase = []converter{
			test.uppercase,
		}

		processed := strings.Apply(newM1())

		expectedFields := map[string]interface{}{
            "request": "/mixed/CASE/paTH/?from=-1D&to=now",
		}

		assert.Equal(t, expectedFields, processed[0].Fields(), test.message, "Should not change fields")
		assert.Equal(t, test.expectedTags, processed[0].Tags(), test.message)
		assert.Equal(t, "IIS_log", processed[0].Name(), "Should not change name")
	}
}

func TestMultipleConversions(t *testing.T) {
	strings := &Strings{}
	strings.Lowercase = []converter{
		{
			Tag:         "s-computername",
		},
		{
			Field:       "request",
		},
		{
			Field:       "cs-host",
			ResultKey:   "cs-host_lowercase",
		},
	}
    strings.Uppercase = []converter{
        {
            Tag:        "verb",
        },
    }

	processed := strings.Apply(newM2())

	expectedFields := map[string]interface{}{
        "request":           "/mixed/case/path/?from=-1d&to=now",
		"ignore_number":     int64(200),
		"ignore_bool":       true,
        "cs-host":           "AAAbbb",
        "cs-host_lowercase": "aaabbb",
	}
	expectedTags := map[string]string{
		"verb":           "GET",
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
				Field:         "not_exists",
			},
			expectedFields: map[string]interface{}{
                "request": "/mixed/CASE/paTH/?from=-1D&to=now",
			},
		},
	}

	for _, test := range tests {
		strings := &Strings{}
		strings.Lowercase = []converter{
			test.converter,
		}

		processed := strings.Apply(newM1())

		assert.Equal(t, test.expectedFields, processed[0].Fields(), test.message)
	}
}
