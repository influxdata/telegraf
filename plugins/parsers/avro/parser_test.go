package avro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/linkedin/goavro"
	"github.com/stretchr/testify/require"
)

var DefaultTime = func() time.Time {
	return time.Unix(3600, 0)
}

func TestBasicAvroMessage(t *testing.T) {

	schema := `
        {
            "type":"record",
            "name":"Value",
            "namespace":"com.example",
            "fields":[
                {
                    "name":"tag",
                    "type":"string"
                },
                {
                    "name":"field",
                    "type":"long"
                },
                {
                    "name":"timestamp",
                    "type":"long"
                }
            ]
        }
    `

	message := `
        {
            "tag":"test_tag",
            "field": 19,
            "timestamp": 1593002503937
        }
    `

	schemaRegistry := LocalSchemaRegistry{schema: schema}
	schemaRegistry.start()
	defer schemaRegistry.stop()

	p := Parser{
		SchemaRegistry:  schemaRegistry.url(),
		Measurement:     "measurement",
		Tags:            []string{"tag"},
		Fields:          []string{"field"},
		Timestamp:       "timestamp",
		TimestampFormat: "unix_ms",
		TimeFunc:        DefaultTime,
	}

	msg, err := makeAvroMessage(schema, message)
	require.NoError(t, err)

	metrics, err := p.Parse(msg)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"measurement",
			map[string]string{
				"tag": "test_tag",
			},
			map[string]interface{}{
				"field": 19,
			},
			time.Unix(1593002503, 937000000),
		),
	}

	testutil.RequireMetricsEqual(t, expected, metrics)
}

func TestKafkaDemoAvroMessage(t *testing.T) {

	schema := `
        {
            "type":"record",
            "name":"KsqlDataSourceSchema",
            "namespace":"io.confluent.ksql.avro_schemas",
            "fields":[
                {
                    "name":"rating_id",
                    "type":[
                        "null",
                        "long"
                    ],
                    "default":null
                },
                {
                    "name":"user_id",
                    "type":[
                        "null",
                        "int"
                    ],
                    "default":null
                },
                {
                    "name":"stars",
                    "type":[
                        "null",
                        "int"
                    ],
                    "default":null
                },
                {
                    "name":"route_id",
                    "type":[
                        "null",
                        "int"
                    ],
                    "default":null
                },
                {
                    "name":"rating_time",
                    "type":[
                        "null",
                        "long"
                    ],
                    "default":null
                },
                {
                    "name":"channel",
                    "type":[
                        "null",
                        "string"
                    ],
                    "default":null
                },
                {
                    "name":"message",
                    "type":[
                        "null",
                        "string"
                    ],
                    "default":null
                }
            ]
        }
    `

	message := `
        {
            "rating_id":{
                "long":1175
            },
            "user_id":{
                "int":14
            },
            "stars":{
                "int":1
            },
            "route_id":{
                "int":3229
            },
            "rating_time":{
                "long":1593009409638
            },
            "channel":{
                "string":"ios"
            },
            "message":{
                "string":"thank you for the most friendly, helpful experience today at your new lounge"
            }
        }
    `

	schemaRegistry := LocalSchemaRegistry{schema: schema}
	schemaRegistry.start()
	defer schemaRegistry.stop()

	p := Parser{
		SchemaRegistry:  schemaRegistry.url(),
		Measurement:     "ratings",
		Tags:            []string{"user_id", "route_id", "channel"},
		Fields:          []string{"stars", "message"},
		Timestamp:       "rating_time",
		TimestampFormat: "unix_ms",
		TimeFunc:        DefaultTime,
	}

	msg, err := makeAvroMessage(schema, message)
	require.NoError(t, err)

	metrics, err := p.Parse(msg)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"ratings",
			map[string]string{
				"user_id":  "14",
				"route_id": "3229",
				"channel":  "ios",
			},
			map[string]interface{}{
				"stars":   1,
				"message": "thank you for the most friendly, helpful experience today at your new lounge",
			},
			time.Unix(1593009409, 638000000),
		),
	}

	testutil.RequireMetricsEqual(t, expected, metrics)
}

type LocalSchemaRegistry struct {
	schema string
	ts     *httptest.Server
}

func (sr *LocalSchemaRegistry) start() {
	schema := strings.ReplaceAll(sr.schema, "\"", "\\\"")
	schema = strings.ReplaceAll(schema, "\n", "\\n")
	schema = fmt.Sprintf("{\"schema\": \"%s\"}", schema)
	sr.ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(schema))
	}))
}

func (sr *LocalSchemaRegistry) stop() {
	sr.ts.Close()
}

func (sr *LocalSchemaRegistry) url() string {
	return sr.ts.URL
}

func makeAvroMessage(schema string, message string) ([]byte, error) {

	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}

	bytes := []byte(message)

	native, _, err := codec.NativeFromTextual(bytes)
	if err != nil {
		return nil, err
	}

	binary, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, err
	}

	magicByte := []byte{0x01}
	schemaID := []byte{0x00, 0x00, 0x00, 0x01}
	binary = append(schemaID, binary...)
	binary = append(magicByte, binary...)

	return binary, nil
}
