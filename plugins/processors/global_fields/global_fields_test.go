package global_fields

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGlobalFieldsWhenNotAlreadyDefined(t *testing.T) {

	var m = testutil.TestMetric(1.0)

	globalFields := globalFieldsConfig{
		Fields: []*globalField{
			&globalField{
				Name:  "host",
				Value: "computer",
			},
		},
	}

	globalFields.Apply(m)

	assert.Equal(t, 2, len(m.Fields()))
	assert.Equal(t, 1.0, m.Fields()["value"])
	assert.Equal(t, "computer", m.Fields()["host"])
}

func TestGlobalFieldsWhenAlreadyDefined(t *testing.T) {

	var m = testutil.TestMetric(1.0)
	m.AddField("host", "abc")

	globalFields := globalFieldsConfig{
		Fields: []*globalField{
			&globalField{
				Name:  "host",
				Value: "computer",
			},
		},
	}

	globalFields.Apply(m)

	assert.Equal(t, 2, len(m.Fields()))
	assert.Equal(t, 1.0, m.Fields()["value"])
	assert.Equal(t, "abc", m.Fields()["host"])
}
