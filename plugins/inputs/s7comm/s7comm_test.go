package s7comm

import (
	_ "embed"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/robinson/gos7"
	"github.com/stretchr/testify/require"
)

func TestSampleConfig(t *testing.T) {
	plugin := &S7comm{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name          string
		server        string
		rack          int
		slot          int
		configs       []metricDefinition
		expectedError string
	}{
		{
			name:          "empty settings",
			rack:          -1, // This is the default in `init()`
			slot:          -1, // This is the default in `init()`
			expectedError: "'server' has to be specified",
		},
		{
			name:          "missing rack",
			server:        "127.0.0.1:102",
			rack:          -1, // This is the default in `init()`
			slot:          -1, // This is the default in `init()`
			expectedError: "'rack' has to be specified",
		},
		{
			name:          "missing slot",
			server:        "127.0.0.1:102",
			rack:          0,
			slot:          -1, // This is the default in `init()`
			expectedError: "'slot' has to be specified",
		},
		{
			name:          "missing configs",
			server:        "127.0.0.1:102",
			expectedError: "no metric defined",
		},
		{
			name:          "single empty metric",
			server:        "127.0.0.1:102",
			configs:       []metricDefinition{{}},
			expectedError: "no fields defined for metric",
		},
		{
			name:   "single empty metric field",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{{}},
				},
			},
			expectedError: "unnamed field in metric",
		},
		{
			name:   "no address",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name: "foo",
						},
					},
				},
			},
			expectedError: "invalid address",
		},
		{
			name:   "invalid address pattern",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "FOO",
						},
					},
				},
			},
			expectedError: "invalid address",
		},
		{
			name:   "invalid address area",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "FOO1.W2",
						},
					},
				},
			},
			expectedError: "invalid area",
		},
		{
			name:   "invalid address area index",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB.W2",
						},
					},
				},
			},
			expectedError: "invalid address",
		},
		{
			name:   "invalid address type",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.A2",
						},
					},
				},
			},
			expectedError: "unknown data type",
		},
		{
			name:   "invalid address start",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.A",
						},
					},
				},
			},
			expectedError: "invalid address",
		},
		{
			name:   "missing extra parameter bit",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.X1",
						},
					},
				},
			},
			expectedError: "extra parameter required",
		},
		{
			name:   "missing extra parameter string",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.S1",
						},
					},
				},
			},
			expectedError: "extra parameter required",
		},
		{
			name:   "invalid address extra parameter",
			server: "127.0.0.1:102",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.W1.23",
						},
					},
				},
			},
			expectedError: "extra parameter specified but not used",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &S7comm{
				Server:  tt.server,
				Rack:    tt.rack,
				Slot:    tt.slot,
				Configs: tt.configs,
				Log:     &testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), tt.expectedError)
		})
	}
}

func TestInit(t *testing.T) {
	plugin := &S7comm{
		Server: "127.0.0.1:102",
		Rack:   0,
		Slot:   0,
		Configs: []metricDefinition{
			{
				Fields: []metricFieldDefinition{
					{
						Name:    "foo",
						Address: "DB1.W2",
					},
				},
			},
		},
		Log: &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
}

func TestFieldMappings(t *testing.T) {
	tests := []struct {
		name     string
		configs  []metricDefinition
		expected []batch
	}{
		{
			name: "single field bit",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.X3.2",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x01,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 1),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return false },
						},
					},
				},
			},
		},
		{
			name: "single field byte",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.B3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x02,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 1),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return byte(0) },
						},
					},
				},
			},
		},
		{
			name: "single field char",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.C3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x03,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 1),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return string([]byte{0}) },
						},
					},
				},
			},
		},
		{
			name: "single field string",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.S3.10",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x03,
							DBNumber: 5,
							Start:    3,
							Amount:   10,
							Data:     make([]byte, 12),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return "" },
						},
					},
				},
			},
		},
		{
			name: "single field word",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.W3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x04,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 2),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return uint16(0) },
						},
					},
				},
			},
		},
		{
			name: "single field integer",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.I3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x05,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 2),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return int16(0) },
						},
					},
				},
			},
		},
		{
			name: "single field double word",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.DW3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x06,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 4),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return uint32(0) },
						},
					},
				},
			},
		},
		{
			name: "single field double integer",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.DI3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x07,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 4),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return int32(0) },
						},
					},
				},
			},
		},
		{
			name: "single field float",
			configs: []metricDefinition{
				{
					Name: "test",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB5.R3",
						},
					},
				},
			},
			expected: []batch{
				{
					items: []gos7.S7DataItem{
						{
							Area:     0x84,
							WordLen:  0x08,
							DBNumber: 5,
							Start:    3,
							Amount:   1,
							Data:     make([]byte, 4),
						},
					},
					mappings: []fieldMapping{
						{
							measurement: "test",
							field:       "foo",
							convert:     func(b []byte) interface{} { return float32(0) },
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &S7comm{
				Server:  "127.0.0.1:102",
				Rack:    0,
				Slot:    2,
				Configs: tt.configs,
				Log:     &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Check the length
			require.Len(t, plugin.batches, len(tt.expected))
			// Check the actual content
			for i, eb := range tt.expected {
				ab := plugin.batches[i]
				require.Len(t, ab.items, len(eb.items))
				require.Len(t, ab.mappings, len(eb.mappings))
				require.EqualValues(t, eb.items, plugin.batches[i].items, "different items")
				for j, em := range eb.mappings {
					am := ab.mappings[j]
					require.Equal(t, em.measurement, am.measurement)
					require.Equal(t, em.field, am.field)
					buf := ab.items[j].Data
					require.Equal(t, em.convert(buf), am.convert(buf))
				}
			}
		})
	}
}

func TestMetricCollisions(t *testing.T) {
	tests := []struct {
		name          string
		configs       []metricDefinition
		expectedError string
	}{
		{
			name: "duplicate fields same config",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.W1",
						},
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
				},
			},
			expectedError: "duplicate field definition",
		},
		{
			name: "duplicate fields different config",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
				},
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
				},
			},
			expectedError: "duplicate field definition",
		},
		{
			name: "same fields different name",
			configs: []metricDefinition{
				{
					Name: "foo",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
				},
				{
					Name: "bar",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
				},
			},
		},
		{
			name: "same fields different tags",
			configs: []metricDefinition{
				{
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
					Tags: map[string]string{"device": "foo"},
				},
				{
					Name: "bar",
					Fields: []metricFieldDefinition{
						{
							Name:    "foo",
							Address: "DB1.B1",
						},
					},
					Tags: map[string]string{"device": "bar"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &S7comm{
				Server:  "127.0.0.1:102",
				Rack:    0,
				Slot:    2,
				Configs: tt.configs,
				Log:     &testutil.Logger{},
			}
			err := plugin.Init()
			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
