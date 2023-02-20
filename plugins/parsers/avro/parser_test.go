package avro

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/require"
)

func JSONToAvroMessage(schemaID int32, schema string, message []byte) ([]byte, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}

	// We could use json.Unmarshal, but our codec can do it directly too.
	native, _, err := codec.NativeFromTextual(message)
	if err != nil {
		return nil, err
	}

	binaryMsg, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, err
	}

	// Put schemaID into []byte
	schemaBytes := new(bytes.Buffer)
	err = binary.Write(schemaBytes, binary.BigEndian, schemaID)
	if err != nil {
		return nil, err
	}
	schemaBin := schemaBytes.Bytes()

	// Create serialized Avro binary message
	magicByte := []byte{0x01}
	binaryMsg = append(schemaBin, binaryMsg...)
	binaryMsg = append(magicByte, binaryMsg...)

	return binaryMsg, nil
}

type AvroCfg struct {
	Inputs struct {
		File []struct {
			Parser
			DataFormat string
		}
	}
}

func BuildParser(buf []byte) (*Parser, error) {
	var cfg AvroCfg

	err := toml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}

	pinput := cfg.Inputs.File[0].Parser
	p := Parser{
		MetricName:      pinput.MetricName,
		SchemaRegistry:  pinput.SchemaRegistry,
		Schema:          pinput.Schema,
		Measurement:     pinput.Measurement,
		Tags:            pinput.Tags,
		Fields:          pinput.Fields,
		Timestamp:       pinput.Timestamp,
		TimestampFormat: pinput.TimestampFormat,
		FieldSeparator:  pinput.FieldSeparator,
		DefaultTags:     pinput.DefaultTags,
	}
	return &p, nil
}

func TestMultipleConfigs(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testdata")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.NotEmpty(t, folders)

	for _, f := range folders {
		fname := f.Name()
		testdataPath := filepath.Join("testdata", fname)
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
		testSchema := filepath.Join(testdataPath, "schema.json")
		testJSON := filepath.Join(testdataPath, "message.json")
		expectedFilename := filepath.Join(testdataPath, "expected.out")
		expectedErrorFilename := filepath.Join(testdataPath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			inputs.Add("file", func() telegraf.Input {
				return &file.File{}
			})
			// Read the expected output
			stdParser := &influx.Parser{}
			require.NoError(t, stdParser.Init())
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, stdParser)
			require.NoError(t, err)

			// Read the expected errors if any
			var expectedErrors []string

			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Configure the plugin
			rawConfig, err := os.ReadFile(configFilename)
			require.NoError(t, err)
			cfg := config.NewConfig()
			err = cfg.LoadConfigData(rawConfig)
			require.NoError(t, err)
			// Gather the metrics from the input file configure
			var actualErrorMsgs []string

			// Load the schema and message; produce the Avro
			// format message.
			var schema string
			var message []byte
			schemaBytes, err := os.ReadFile(testSchema)
			if err != nil {
				actualErrorMsgs = append(actualErrorMsgs, err.Error())
			} else {
				schema = string(schemaBytes)
			}

			message, err = os.ReadFile(testJSON)
			if err != nil {
				actualErrorMsgs = append(actualErrorMsgs, err.Error())
			}
			avroMessage, err := JSONToAvroMessage(1, schema, message)
			if err != nil {
				actualErrorMsgs = append(actualErrorMsgs, err.Error())
			}

			// Get a new parser each time, because it may need
			// reconfiguration.

			// This is the bit where I should use the file plugin
			// and Gather, but I don't understand how to do that.

			parser, err := BuildParser(rawConfig)
			require.NoError(t, err)
			err = parser.Init()
			var actual []telegraf.Metric
			if err != nil {
				actualErrorMsgs = append(actualErrorMsgs, err.Error())
			} else {
				// Only parse if the parser actually loaded.
				actual, err = parser.Parse(avroMessage)
				if err != nil {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
			}

			// If the test has expected error(s) then compare them
			if len(expectedErrors) > 0 {
				sort.Strings(actualErrorMsgs)
				sort.Strings(expectedErrors)
				for i, msg := range expectedErrors {
					require.Contains(t, actualErrorMsgs[i], msg)
				}
			} else {
				require.Empty(t, actualErrorMsgs)
			}

			// Process expected metrics and compare with resulting metrics
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}
