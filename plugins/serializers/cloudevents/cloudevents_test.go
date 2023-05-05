package cloudevents

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	outputs.Add("dummy", func() telegraf.Output {
		return &OutputDummy{}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		t.Run(fname, func(t *testing.T) {
			testdataPath := filepath.Join("testcases", fname)
			configFilename := filepath.Join(testdataPath, "telegraf.conf")
			inputFilename := filepath.Join(testdataPath, "input.influx")
			expectedFilename := filepath.Join(testdataPath, "expected.json")

			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			input, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)

			var expected []map[string]interface{}
			ebuf, err := os.ReadFile(expectedFilename)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(ebuf, &expected))

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Outputs, 1, "wrong number of outputs")
			plugin, ok := cfg.Outputs[0].Output.(*OutputDummy)
			require.True(t, ok)
			serializer, ok := plugin.serializer.(*models.RunningSerializer).Serializer.(*Serializer)
			require.True(t, ok)
			serializer.idgen = &dummygen{}

			// Write input and compare with expected metrics
			require.NoError(t, plugin.Write(input))
			require.NoError(t, checkEvents(plugin.output))

			var joined string
			switch len(plugin.output) {
			case 0:
				require.Emptyf(t, expected, "empty output but %d expected", len(expected))
			case 1:
				joined = string(plugin.output[0])
				if !strings.HasPrefix(joined, "[") {
					joined = "[" + joined + "]"
				}
			default:
				joined = "[" + string(bytes.Join(plugin.output, []byte(","))) + "]"
			}
			var actual []map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(joined), &actual))
			require.Len(t, actual, len(expected))
			require.ElementsMatch(t, expected, actual)
		})
	}
}

/* Internal testing functions */
func unmarshalEvents(messages [][]byte) ([]cloudevents.Event, error) {
	var events []cloudevents.Event

	for i, msg := range messages {
		// Check for batch settings
		var es []cloudevents.Event
		if err := json.Unmarshal(msg, &es); err != nil {
			if errors.Is(err, &json.UnmarshalTypeError{}) {
				return nil, fmt.Errorf("message %d: %w", i, err)
			}
			var e cloudevents.Event
			if err := json.Unmarshal(msg, &e); err != nil {
				return nil, fmt.Errorf("message %d: %w", i, err)
			}
			events = append(events, e)
		} else {
			events = append(events, es...)
		}
	}

	return events, nil
}

func checkEvents(messages [][]byte) error {
	events, err := unmarshalEvents(messages)
	if err != nil {
		return err
	}

	for i, e := range events {
		if err := e.Validate(); err != nil {
			return fmt.Errorf("event %d: %w", i, err)
		}

		// Do an additional schema validation
		var schema *jsonschema.Schema
		switch e.SpecVersion() {
		case "0.3":
			schema = jsonschema.MustCompile("testcases/cloudevents-v0.3-schema.json")
		case "1.0":
			schema = jsonschema.MustCompile("testcases/cloudevents-v1.0-schema.json")
		default:
			return fmt.Errorf("unhandled spec version %q in event %d", e.SpecVersion(), i)
		}
		serializedEvent, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("serializing raw event %d: %w", i, err)
		}
		var rawEvent interface{}
		if err := json.Unmarshal(serializedEvent, &rawEvent); err != nil {
			return fmt.Errorf("deserializing raw event %d: %w", i, err)
		}
		if err := schema.Validate(rawEvent); err != nil {
			return fmt.Errorf("validation of event %d: %w", i, err)
		}
	}
	return nil
}

/* Dummy output to allow full config parsing loop */
type OutputDummy struct {
	Batch      bool `toml:"batch"`
	serializer telegraf.Serializer
	output     [][]byte
}

func (*OutputDummy) SampleConfig() string {
	return "dummy"
}

func (o *OutputDummy) Connect() error {
	o.output = make([][]byte, 0)
	return nil
}

func (*OutputDummy) Close() error {
	return nil
}

func (o *OutputDummy) Write(metrics []telegraf.Metric) error {
	if o.Batch {
		buf, err := o.serializer.SerializeBatch(metrics)
		if err != nil {
			return err
		}
		o.output = append(o.output, buf)
	} else {
		for _, m := range metrics {
			buf, err := o.serializer.Serialize(m)
			if err != nil {
				return err
			}
			o.output = append(o.output, buf)
		}
	}

	return nil
}

func (o *OutputDummy) SetSerializer(s telegraf.Serializer) {
	o.serializer = s
}

/* Dummy UUID generator to get predictable UUIDs for testing */
const testid = "845f6acae52a11ed9976d8bbc1a4a0c6"

type dummygen struct{}

func (*dummygen) NewV1() (uuid.UUID, error) {
	id, err := hex.DecodeString(testid)
	if err != nil {
		return uuid.UUID([16]byte{}), err
	}
	return uuid.UUID(id), nil
}

func (*dummygen) NewV3(_ uuid.UUID, _ string) uuid.UUID {
	return uuid.UUID([16]byte{})
}

func (*dummygen) NewV4() (uuid.UUID, error) {
	return uuid.UUID([16]byte{}), errors.New("wrong type")
}

func (*dummygen) NewV5(_ uuid.UUID, _ string) uuid.UUID {
	return uuid.UUID([16]byte{})
}

func (*dummygen) NewV6() (uuid.UUID, error) {
	return uuid.UUID([16]byte{}), errors.New("wrong type")
}

func (*dummygen) NewV7() (uuid.UUID, error) {
	return uuid.UUID([16]byte{}), errors.New("wrong type")
}

/* Benchmarks */
func BenchmarkSerializer(b *testing.B) {
	m := metric.New(
		"test",
		map[string]string{
			"source": "somehost.company.com",
			"host":   "localhost",
			"status": "healthy",
		},
		map[string]interface{}{
			"temperature":     23.5,
			"operating_hours": 4242,
			"connections":     123,
			"standby":         true,
			"SN":              "DC5423DE4CE/2",
		},
		time.Now(),
	)

	serializer := &Serializer{}
	for n := 0; n < b.N; n++ {
		_, err := serializer.Serialize(m)
		require.NoError(b, err)
	}
}
