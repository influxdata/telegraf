package kinesis_consumer

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	consumer "github.com/harlow/kinesis-consumer"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestTrackingAccumulator struct {
	telegraf.TrackingAccumulator
	Metrics *[]telegraf.Metric
}

func (t TestTrackingAccumulator) AddTrackingMetricGroup(group []telegraf.Metric) telegraf.TrackingID {
	*t.Metrics = append(*t.Metrics, group...)
	return 1
}

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
		DecompressionType string
		parser            parsers.Parser
		records           map[telegraf.TrackingID]string
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
				DecompressionType: "none",
				parser:            parser,
				records:           make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{Data: notZippedBytes, SequenceNumber: aws.String("anything")},
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
				DecompressionType: "gzip",
				parser:            parser,
				records:           make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{Data: gzippedBytes, SequenceNumber: aws.String("anything")},
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
				DecompressionType: "zlib",
				parser:            parser,
				records:           make(map[telegraf.TrackingID]string),
			},
			args: args{
				r: &consumer.Record{Data: zlibBytpes, SequenceNumber: aws.String("anything")},
			},
			wantErr: false,
			expected: expected{
				messageContains: "bob",
				numberOfMetrics: 1,
			},
		},
	}

	k := &KinesisConsumer{
		DecompressionType: "notsupported",
	}
	err := k.Init()
	assert.NotNil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KinesisConsumer{
				DecompressionType: tt.fields.DecompressionType,
				parser:            tt.fields.parser,
				records:           tt.fields.records,
			}
			err := k.Init()
			assert.Nil(t, err)

			var metrics []telegraf.Metric
			if err := k.onMessage(TestTrackingAccumulator{Metrics: &metrics}, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("onMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.expected.numberOfMetrics, len(metrics))

			for _, metric := range metrics {
				if logEventMessage, ok := metric.Fields()["message"]; ok {
					assert.Contains(t, logEventMessage.(string), tt.expected.messageContains)
				} else {
					t.Errorf("Expect logEvents to be present")
				}
			}
		})
	}

}
