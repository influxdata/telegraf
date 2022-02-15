package kinesis_consumer

import (
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	consumer "github.com/harlow/kinesis-consumer"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestKinesisConsumer_onMessage(t *testing.T) {
	zlibBytpes, _ := base64.StdEncoding.DecodeString("eF5FjlFrgzAUhf9KuM+2aNB2zdsQ2xe3whQGW8qIeqdhaiSJK0P874u1Y4+Hc/jON0GHxoga858BgUF8fs5fzunHU5Jlj6cEPFDXHvXStGqsrsKWTapq44pW1SetxsF1a8qsRtGt0YyFKbUcrFT9UbYWtQH2frntkm/s7RInkNU6t9JpWNE5WBAFPo3CcHeg+9D703OziUOhCg6MQ/yakrspuZsyEjdYfsm+Jg2K1jZEfZLKQWUvFglylBobZXDLwSP8//EGpD4NNj7dUJpT6hQY3W33h/AhCt84zDBf5l/MDl08")
	gzippedBytes, _ := base64.StdEncoding.DecodeString("H4sIAAFXNGAAA0WOUWuDMBSF/0q4z7Zo0HbN2xDbF7fCFAZbyoh6p2FqJIkrQ/zvi7Vjj4dz+M43QYfGiBrznwGBQXx+zl/O6cdTkmWPpwQ8UNce9dK0aqyuwpZNqmrjilbVJ63GwXVryqxG0a3RjIUptRysVP1Rtha1AfZ+ue2Sb+ztEieQ1Tq30mlY0TlYEAU+jcJwd6D70PvTc7OJQ6EKDoxD/JqSuym5mzISN1h+yb4mDYrWNkR9kspBZS8WCXKUGhtlcMvBI/z/8QakPg02Pt1QmlPqFBjdbfeH8CEK3zjMMF/mX0TaxZUpAQAA")
	notZippedBytes := []byte(`{"messageType":"CONTROL_MESSAGE","owner":"CloudwatchLogs","logGroup":"","logStream":"",
"subscriptionFilters":[],"logEvents":[
	{"id":"","timestamp":1510254469274,"message":"{\"bob\":\"CWL CONTROL MESSAGE: Checking health of destination Firehose.\", \"timestamp\":\"2021-02-22T22:15:26.794854Z\"},"},
	{"id":"","timestamp":1510254469274,"message":"{\"bob\":\"CWL CONTROL MESSAGE: Checking health of destination Firehose.\", \"timestamp\":\"2021-02-22T22:15:26.794854Z\"}"}
]}`)
	parser, _ := json.New(&json.Config{
		MetricName:   "json_test",
		Query:        "logEvents",
		StringFields: []string{"message"},
	})

	type fields struct {
		ContentEncoding string
		parser          parsers.Parser
		records         map[telegraf.TrackingID]string
	}
	type args struct {
		r *consumer.Record
	}
	type expected struct {
		numberOfMetrics int
		messageContains string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expected expected
	}{
		{
			name: "test no compression",
			fields: fields{
				ContentEncoding: "none",
				parser:          parser,
				records:         make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{
					Record: types.Record{
						Data:           notZippedBytes,
						SequenceNumber: aws.String("anything"),
					},
				},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 2,
			},
		},
		{
			name: "test no compression via empty string for ContentEncoding",
			fields: fields{
				ContentEncoding: "",
				parser:          parser,
				records:         make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{
					Record: types.Record{
						Data:           notZippedBytes,
						SequenceNumber: aws.String("anything"),
					},
				},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 2,
			},
		},
		{
			name: "test no compression via identity ContentEncoding",
			fields: fields{
				ContentEncoding: "identity",
				parser:          parser,
				records:         make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{
					Record: types.Record{
						Data:           notZippedBytes,
						SequenceNumber: aws.String("anything"),
					},
				},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 2,
			},
		},
		{
			name: "test no compression via no ContentEncoding",
			fields: fields{
				parser:  parser,
				records: make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{
					Record: types.Record{
						Data:           notZippedBytes,
						SequenceNumber: aws.String("anything"),
					},
				},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 2,
			},
		},
		{
			name: "test gzip compression",
			fields: fields{
				ContentEncoding: "gzip",
				parser:          parser,
				records:         make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{
					Record: types.Record{
						Data:           gzippedBytes,
						SequenceNumber: aws.String("anything"),
					},
				},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 1,
			},
		},
		{
			name: "test zlib compression",
			fields: fields{
				ContentEncoding: "zlib",
				parser:          parser,
				records:         make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{
					Record: types.Record{
						Data:           zlibBytpes,
						SequenceNumber: aws.String("anything"),
					},
				},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 1,
			},
		},
	}

	k := &KinesisConsumer{
		ContentEncoding: "notsupported",
	}
	err := k.Init()
	require.NotNil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KinesisConsumer{
				ContentEncoding: tt.fields.ContentEncoding,
				parser:          tt.fields.parser,
				records:         tt.fields.records,
			}
			err := k.Init()
			require.Nil(t, err)

			acc := testutil.Accumulator{}
			if err := k.onMessage(acc.WithTracking(tt.expected.numberOfMetrics), tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("onMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			require.Equal(t, tt.expected.numberOfMetrics, len(acc.Metrics))

			for _, metric := range acc.Metrics {
				if logEventMessage, ok := metric.Fields["message"]; ok {
					require.Contains(t, logEventMessage.(string), tt.expected.messageContains)
				} else {
					t.Errorf("Expect logEvents to be present")
				}
			}
		})
	}
}
