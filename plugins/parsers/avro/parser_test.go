package avro

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
                    "name":"measurement",
                    "type":"string"
                },
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

	schema = strings.ReplaceAll(schema, "\"", "\\\"")
	schema = strings.ReplaceAll(schema, "\n", "\\n")
	schema = fmt.Sprintf("{\"schema\": \"%s\"}", schema)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(schema))
	}))
	defer ts.Close()

	p := Parser{
		SchemaRegistry:  ts.URL,
		Measurement:     "measurement",
		Tags:            []string{"tag"},
		Fields:          []string{"field"},
		Timestamp:       "timestamp",
		TimestampFormat: "unix",
		TimeFunc:        DefaultTime,
	}

	msg := []byte{0x00, 0x00, 0x00, 0x00, 0x17, 0x20, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x6d, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x10, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x74, 0x61, 0x67, 0x26, 0xf0, 0xb6, 0x97, 0xd4, 0xb0, 0x5b}

	msgString := "[measurement map[tag:test_tag] map[field:19] 1925572734688112640]"

	m, err := p.Parse(msg)

	require.NoError(t, err)

	require.Equal(t, fmt.Sprintf("%v", m), msgString, "The message should be decoded correctly.")
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

	schema = strings.ReplaceAll(schema, "\"", "\\\"")
	schema = strings.ReplaceAll(schema, "\n", "\\n")
	schema = fmt.Sprintf("{\"schema\": \"%s\"}", schema)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(schema))
	}))
	defer ts.Close()

	p := Parser{
		SchemaRegistry:  ts.URL,
		Measurement:     "rating",
		Tags:            []string{"user_id", "route_id", "channel"},
		Fields:          []string{"stars", "message"},
		Timestamp:       "rating_time",
		TimestampFormat: "unix_ms",
		TimeFunc:        DefaultTime,
	}

	msg := []byte{0, 0, 0, 0, 1, 2, 144, 16, 2, 14, 2, 4, 2, 244, 42, 2, 226, 196, 231, 151, 148, 92, 2, 6, 105, 111, 115, 2, 104, 119, 104, 121, 32, 105, 115, 32, 105, 116, 32, 115, 111, 32, 100, 105, 102, 102, 105, 99, 117, 108, 116, 32, 116, 111, 32, 107, 101, 101, 112, 32, 116, 104, 101, 32, 98, 97, 116, 104, 114, 111, 111, 109, 115, 32, 99, 108, 101, 97, 110, 32, 63}

	msgString := "[rating map[channel:ios route_id:2746 user_id:7] map[message:why is it so difficult to keep the bathrooms clean ? stars:2] 1583257284913000000]"

	m, err := p.Parse(msg)

	require.NoError(t, err)

	require.Equal(t, fmt.Sprintf("%v", m), msgString, "The message should be decoded correctly.")
}
