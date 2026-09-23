package cisco_telemetry_mdt

import (
	"testing"
	"time"

	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// nest wraps the field in the given number of single-child levels, so the
// returned field has to be descended that many times to reach it again.
func nest(depth int, field *telemetry.TelemetryField) *telemetry.TelemetryField {
	for range depth {
		field = &telemetry.TelemetryField{Fields: []*telemetry.TelemetryField{field}}
	}
	return field
}

func TestParseEventsShortAttributes(t *testing.T) {
	subscription := &telemetry.TelemetryField{Name: "subscriptionId"}
	attributes := &telemetry.TelemetryField{
		Fields: []*telemetry.TelemetryField{
			{Name: "rn", ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "eth1/1"}},
			{Name: "speed", ValueByType: &telemetry.TelemetryField_Uint32Value{Uint32Value: 100}},
		},
	}

	tests := []struct {
		name   string
		fields []*telemetry.TelemetryField
	}{
		{
			name:   "subscription identifier first without attributes",
			fields: []*telemetry.TelemetryField{subscription, {}},
		},
		{
			name:   "subscription identifier second without attributes",
			fields: []*telemetry.TelemetryField{{}, subscription},
		},
		{
			name:   "attributes one level deep",
			fields: []*telemetry.TelemetryField{subscription, nest(1, attributes)},
		},
		{
			name:   "attributes two levels deep",
			fields: []*telemetry.TelemetryField{subscription, nest(2, attributes)},
		},
		{
			name:   "attributes three levels deep",
			fields: []*telemetry.TelemetryField{subscription, nest(3, attributes)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state{
				measurement: "dme",
				isNXOS:      true,
				isEvent:     true,
				extraTags:   make(map[string]map[string]bool),
				propMap:     make(map[string]func(*telemetry.TelemetryField) any),
				nxpathMap:   make(map[string]string),
				grouper:     metric.NewSeriesGrouper(),
			}

			events := []*telemetry.TelemetryField{{Fields: tt.fields}}
			require.Empty(t, s.parseEvents(events, "interface", make(map[string]string), time.Unix(0, 0)))
			require.Empty(t, s.grouper.Metrics())
		})
	}
}

func TestParseEventsAttributesFound(t *testing.T) {
	// The attributes have to sit five levels below the sibling of the
	// subscription identifier for the parser to pick them up.
	attributes := &telemetry.TelemetryField{
		Fields: []*telemetry.TelemetryField{
			{Name: "rn", ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "eth1/1"}},
			{Name: "speed", ValueByType: &telemetry.TelemetryField_Uint32Value{Uint32Value: 100}},
		},
	}
	events := []*telemetry.TelemetryField{
		{
			Fields: []*telemetry.TelemetryField{
				{Name: "subscriptionId"},
				nest(5, attributes),
			},
		},
	}

	s := &state{
		measurement: "dme",
		isNXOS:      true,
		isEvent:     true,
		extraTags:   make(map[string]map[string]bool),
		propMap:     make(map[string]func(*telemetry.TelemetryField) any),
		nxpathMap:   make(map[string]string),
		grouper:     metric.NewSeriesGrouper(),
	}

	require.Empty(t, s.parseEvents(events, "interface", make(map[string]string), time.Unix(0, 0)))

	expected := []telegraf.Metric{
		metric.New(
			"dme",
			map[string]string{"interface": "eth1/1"},
			map[string]any{"speed": uint32(100)},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, s.grouper.Metrics())
}
