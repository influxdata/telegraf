package parsers_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
)

func TestRegistry_BackwardCompatibility(t *testing.T) {
	cfg := &parsers.Config{
		MetricName:        "parser_compatibility_test",
		CSVHeaderRowCount: 42,
		XPathProtobufFile: "xpath/testcases/protos/addressbook.proto",
		XPathProtobufType: "addressbook.AddressBook",
		JSONStrict:        true,
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
		"xpath_protobuf": {
			param: map[string]interface{}{
				"ProtobufMessageDef":  cfg.XPathProtobufFile,
				"ProtobufMessageType": cfg.XPathProtobufType,
			},
		},
	}

	// Define parsers that do not have an old-school init
	newStyleOnly := []string{"binary"}

	for name, creator := range parsers.Parsers {
		if choice.Contains(name, newStyleOnly) {
			t.Logf("skipping new-style-only %q...", name)
			continue
		}
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

		// Determine the underlying type of the parser
		stype := reflect.Indirect(reflect.ValueOf(expected)).Interface()
		// Ignore all unexported fields and fields not relevant for functionality
		options := []cmp.Option{
			cmpopts.IgnoreUnexported(stype),
			cmpopts.IgnoreTypes(sync.Mutex{}),
			cmpopts.IgnoreInterfaces(struct{ telegraf.Logger }{}),
		}

		// Add overrides and masks to compare options
		if settings, found := override[name]; found {
			options = append(options, cmpopts.IgnoreFields(stype, settings.mask...))
		}

		// Do a manual comparison as require.EqualValues will also work on unexported fields
		// that cannot be cleared or ignored.
		diff := cmp.Diff(expected, actual, options...)
		require.Emptyf(t, diff, "Difference for %q", name)
	}
}
