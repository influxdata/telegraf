package breaker

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestBreaker(t *testing.T) {
	tests := map[string]struct {
		Enabled      bool
		Name         string
		Field        string
		ValueEnable  interface{}
		ValueDisable interface{}
		metricsIn    []telegraf.Metric
		metricsOut   []telegraf.Metric
	}{
		"breaker disabled by default, first metric pass": {
			Enabled: false,
			metricsIn: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			metricsOut: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
		},
		"breaker enabled by default, first is dropped": {
			Enabled: true,
			metricsIn: []telegraf.Metric{
				testutil.MustMetric("name",
					map[string]string{
						"host": "hostname",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			metricsOut: []telegraf.Metric{},
		},
		"breaker enabled by default, disable with the flag metric and see how next metric pass": {
			Enabled:      true,
			Name:         "flag",
			Field:        "value",
			ValueDisable: 0,
			metricsIn: []telegraf.Metric{
				testutil.MustMetric("flag",
					map[string]string{},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				testutil.MustMetric("foo",
					map[string]string{
						"bar": "bar_value",
					},
					map[string]interface{}{
						"foo": int64(100),
					},
					time.Unix(1522082244, 0),
				),
			},
			metricsOut: []telegraf.Metric{
				testutil.MustMetric("flag",
					map[string]string{},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				testutil.MustMetric("foo",
					map[string]string{
						"bar": "bar_value",
					},
					map[string]interface{}{
						"foo": int64(100),
					},
					time.Unix(1522082244, 0),
				),
			},
		},
		"matching the flag metric ignores tags": {
			Enabled:      true,
			Name:         "flag",
			Field:        "value",
			ValueDisable: 0,
			metricsIn: []telegraf.Metric{
				testutil.MustMetric("flag",
					map[string]string{
						"some_tag": "ignored",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				testutil.MustMetric("foo",
					map[string]string{
						"bar": "bar_value",
					},
					map[string]interface{}{
						"foo": int64(100),
					},
					time.Unix(1522082244, 0),
				),
			},
			metricsOut: []telegraf.Metric{
				testutil.MustMetric("flag",
					map[string]string{
						"some_tag": "ignored",
					},
					map[string]interface{}{
						"value": int64(0),
					},
					time.Unix(1522082244, 0),
				),
				testutil.MustMetric("foo",
					map[string]string{
						"bar": "bar_value",
					},
					map[string]interface{}{
						"foo": int64(100),
					},
					time.Unix(1522082244, 0),
				),
			},
		},
		"metrics with same name as flag metric but with different fields should be treated as a regular metric": {
			Enabled:      false,
			Name:         "flag",
			Field:        "value",
			ValueDisable: 0,
			metricsIn: []telegraf.Metric{
				testutil.MustMetric("flag",
					map[string]string{},
					map[string]interface{}{
						"foo": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
			metricsOut: []telegraf.Metric{
				testutil.MustMetric("flag",
					map[string]string{},
					map[string]interface{}{
						"foo": int64(0),
					},
					time.Unix(1522082244, 0),
				),
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			b := &Breaker{
				Enabled:      test.Enabled,
				Name:         test.Name,
				Field:        test.Field,
				ValueEnable:  test.ValueEnable,
				ValueDisable: test.ValueDisable,
			}

			out := b.Apply(test.metricsIn...)
			assert.Equal(t, test.metricsOut, out)
		})
	}
}
