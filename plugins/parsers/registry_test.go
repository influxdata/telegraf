package parsers_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
)

func TestRegistry_BackwardCompatibility(t *testing.T) {
	cfg := &parsers.Config{
		MetricName:        "parser_compatibility_test",
		CSVHeaderRowCount: 42,
	}

	// Some parsers need certain settings to not error. Furthermore, we
	// might need to clear some (pointer) fields for comparison...
	override := map[string]struct {
		param map[string]interface{}
		mask  []string
	}{
		"csv": {
			param: map[string]interface{}{
				"HeaderRowCount": cfg.CSVHeaderRowCount,
			},
			mask: []string{"TimeFunc"},
		},
	}

	for name, creator := range parsers.Parsers {
		t.Logf("testing %q...", name)
		cfg.DataFormat = name

		// Create parser the new way
		expected := creator(cfg.MetricName)
		if settings, found := override[name]; found {
			s := reflect.Indirect(reflect.ValueOf(expected))
			for key, value := range settings.param {
				v := reflect.ValueOf(value)
				s.FieldByName(key).Set(v)
			}
		}
		if p, ok := expected.(telegraf.Initializer); ok {
			require.NoError(t, p.Init())
		}

		// Create parser the old way
		actual, err := parsers.NewParser(cfg)
		require.NoError(t, err)

		// Compare with mask
		if settings, found := override[name]; found {
			a := reflect.Indirect(reflect.ValueOf(actual))
			e := reflect.Indirect(reflect.ValueOf(expected))
			for _, key := range settings.mask {
				af := a.FieldByName(key)
				ef := e.FieldByName(key)

				v := reflect.Zero(ef.Type())
				af.Set(v)
				ef.Set(v)
			}
		}
		require.EqualValuesf(t, expected, actual, "format %q", name)
	}
}
