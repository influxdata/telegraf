package cloudwatch_logs

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudwatchlogsV2 "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/testutil"
)

type mockCloudWatchLogs struct {
	logStreamName   string
	pushedLogEvents []types.InputLogEvent
}

func (c *mockCloudWatchLogs) Init(lsName string) {
	c.logStreamName = lsName
	c.pushedLogEvents = make([]types.InputLogEvent, 0)
}

func (c *mockCloudWatchLogs) DescribeLogGroups(context.Context, *cloudwatchlogsV2.DescribeLogGroupsInput, ...func(options *cloudwatchlogsV2.Options)) (*cloudwatchlogsV2.DescribeLogGroupsOutput, error) {
	return nil, nil
}

func (c *mockCloudWatchLogs) DescribeLogStreams(context.Context, *cloudwatchlogsV2.DescribeLogStreamsInput, ...func(options *cloudwatchlogsV2.Options)) (*cloudwatchlogsV2.DescribeLogStreamsOutput, error) {
	arn := "arn"
	creationTime := time.Now().Unix()
	sequenceToken := "arbitraryToken"
	output := &cloudwatchlogsV2.DescribeLogStreamsOutput{
		LogStreams: []types.LogStream{
			{
				Arn:                 &arn,
				CreationTime:        &creationTime,
				FirstEventTimestamp: &creationTime,
				LastEventTimestamp:  &creationTime,
				LastIngestionTime:   &creationTime,
				LogStreamName:       &c.logStreamName,
				UploadSequenceToken: &sequenceToken,
			}},
		NextToken: &sequenceToken,
	}
	return output, nil
}
func (c *mockCloudWatchLogs) CreateLogStream(context.Context, *cloudwatchlogsV2.CreateLogStreamInput, ...func(options *cloudwatchlogsV2.Options)) (*cloudwatchlogsV2.CreateLogStreamOutput, error) {
	return nil, nil
}
func (c *mockCloudWatchLogs) PutLogEvents(_ context.Context, input *cloudwatchlogsV2.PutLogEventsInput, _ ...func(options *cloudwatchlogsV2.Options)) (*cloudwatchlogsV2.PutLogEventsOutput, error) {
	sequenceToken := "arbitraryToken"
	output := &cloudwatchlogsV2.PutLogEventsOutput{NextSequenceToken: &sequenceToken}
	//Saving messages
	c.pushedLogEvents = append(c.pushedLogEvents, input.LogEvents...)

	return output, nil
}

//Ensure mockCloudWatchLogs implement cloudWatchLogs interface
var _ cloudWatchLogs = (*mockCloudWatchLogs)(nil)

func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func TestInit(t *testing.T) {
	tests := []struct {
		name                string
		expectedErrorString string
		plugin              *CloudWatchLogs
	}{
		{
			name:                "log group is not set",
			expectedErrorString: "log group is not set",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "",
				LogStream:    "tag:source",
				LDMetricName: "docker_log",
				LDSource:     "field:message",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
		{
			name:                "log stream is not set",
			expectedErrorString: "log stream is not set",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "TestLogGroup",
				LogStream:    "",
				LDMetricName: "docker_log",
				LDSource:     "field:message",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
		{
			name:                "log data metrics name is not set",
			expectedErrorString: "log data metrics name is not set",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "TestLogGroup",
				LogStream:    "tag:source",
				LDMetricName: "",
				LDSource:     "field:message",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
		{
			name:                "log data source is not set",
			expectedErrorString: "log data source is not set",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "TestLogGroup",
				LogStream:    "tag:source",
				LDMetricName: "docker_log",
				LDSource:     "",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
		{
			name: "log data source is not properly formatted (no divider)",
			expectedErrorString: "log data source is not properly formatted, ':' is missed.\n" +
				"Should be 'tag:<tag_mame>' or 'field:<field_name>'",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "TestLogGroup",
				LogStream:    "tag:source",
				LDMetricName: "docker_log",
				LDSource:     "field_message",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
		{
			name: "log data source is not properly formatted (inappropriate fields)",
			expectedErrorString: "log data source is not properly formatted.\n" +
				"Should be 'tag:<tag_mame>' or 'field:<field_name>'",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "TestLogGroup",
				LogStream:    "tag:source",
				LDMetricName: "docker_log",
				LDSource:     "bla:bla",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
		{
			name: "valid config",
			plugin: &CloudWatchLogs{
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "eu-central-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
				LogGroup:     "TestLogGroup",
				LogStream:    "tag:source",
				LDMetricName: "docker_log",
				LDSource:     "tag:location",
				Log: testutil.Logger{
					Name: "outputs.cloudwatch_logs",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedErrorString != "" {
				require.EqualError(t, tt.plugin.Init(), tt.expectedErrorString)
			} else {
				require.Nil(t, tt.plugin.Init())
			}
		})
	}
}

func TestConnect(t *testing.T) {
	//mock cloudwatch logs endpoint that is used only in plugin.Connect
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w,
			`{
				   "logGroups": [ 
					  { 
						 "arn": "string",
						 "creationTime": 123456789,
						 "kmsKeyId": "string",
						 "logGroupName": "TestLogGroup",
						 "metricFilterCount": 1,
						 "retentionInDays": 10,
						 "storedBytes": 0
					  }
				   ]
				}`)
	}))
	defer ts.Close()

	plugin := &CloudWatchLogs{
		CredentialConfig: internalaws.CredentialConfig{
			Region:      "eu-central-1",
			AccessKey:   "dummy",
			SecretKey:   "dummy",
			EndpointURL: ts.URL,
		},
		LogGroup:     "TestLogGroup",
		LogStream:    "tag:source",
		LDMetricName: "docker_log",
		LDSource:     "field:message",
		Log: testutil.Logger{
			Name: "outputs.cloudwatch_logs",
		},
	}

	require.Nil(t, plugin.Init())
	require.Nil(t, plugin.Connect())
}

func TestWrite(t *testing.T) {
	//mock cloudwatch logs endpoint that is used only in plugin.Connect
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w,
			`{
				   "logGroups": [ 
					  { 
						 "arn": "string",
						 "creationTime": 123456789,
						 "kmsKeyId": "string",
						 "logGroupName": "TestLogGroup",
						 "metricFilterCount": 1,
						 "retentionInDays": 1,
						 "storedBytes": 0
					  }
				   ]
				}`)
	}))
	defer ts.Close()

	plugin := &CloudWatchLogs{
		CredentialConfig: internalaws.CredentialConfig{
			Region:      "eu-central-1",
			AccessKey:   "dummy",
			SecretKey:   "dummy",
			EndpointURL: ts.URL,
		},
		LogGroup:     "TestLogGroup",
		LogStream:    "tag:source",
		LDMetricName: "docker_log",
		LDSource:     "field:message",
		Log: testutil.Logger{
			Name: "outputs.cloudwatch_logs",
		},
	}
	require.Nil(t, plugin.Init())
	require.Nil(t, plugin.Connect())

	tests := []struct {
		name                 string
		logStreamName        string
		metrics              []telegraf.Metric
		expectedMetricsOrder map[int]int //map[<index of pushed log event>]<index of corresponding metric>
		expectedMetricsCount int
	}{
		{
			name:                 "Sorted by timestamp log entries",
			logStreamName:        "deadbeef",
			expectedMetricsOrder: map[int]int{0: 0, 1: 1},
			expectedMetricsCount: 2,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "Sorted: message #1",
					},
					time.Now().Add(-time.Minute),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "Sorted: message #2",
					},
					time.Now(),
				),
			},
		},
		{
			name:                 "Unsorted log entries",
			logStreamName:        "deadbeef",
			expectedMetricsOrder: map[int]int{0: 1, 1: 0},
			expectedMetricsCount: 2,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "Unsorted: message #1",
					},
					time.Now(),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "Unsorted: message #2",
					},
					time.Now().Add(-time.Minute),
				),
			},
		},
		{
			name:                 "Too old log entry & log entry in the future",
			logStreamName:        "deadbeef",
			expectedMetricsCount: 0,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "message #1",
					},
					time.Now().Add(-maxPastLogEventTimeOffset).Add(-time.Hour),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "message #2",
					},
					time.Now().Add(maxFutureLogEventTimeOffset).Add(time.Hour),
				),
			},
		},
		{
			name:                 "Oversized log entry",
			logStreamName:        "deadbeef",
			expectedMetricsCount: 0,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						//Here comes very long message
						"message": RandStringBytes(maxLogMessageLength + 1),
					},
					time.Now().Add(-time.Minute),
				),
			},
		},
		{
			name:                 "Batching log entries",
			logStreamName:        "deadbeef",
			expectedMetricsOrder: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4},
			expectedMetricsCount: 5,
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						//Here comes very long message to cause message batching
						"message": "batch1 message1:" + RandStringBytes(maxLogMessageLength-16),
					},
					time.Now().Add(-4*time.Minute),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						//Here comes very long message to cause message batching
						"message": "batch1 message2:" + RandStringBytes(maxLogMessageLength-16),
					},
					time.Now().Add(-3*time.Minute),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						//Here comes very long message to cause message batching
						"message": "batch1 message3:" + RandStringBytes(maxLogMessageLength-16),
					},
					time.Now().Add(-2*time.Minute),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						//Here comes very long message to cause message batching
						"message": "batch1 message4:" + RandStringBytes(maxLogMessageLength-16),
					},
					time.Now().Add(-time.Minute),
				),
				testutil.MustMetric(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "batch2 message1",
					},
					time.Now(),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Overwrite cloud watch log endpoint
			mockCwl := &mockCloudWatchLogs{}
			mockCwl.Init(tt.logStreamName)
			plugin.svc = mockCwl
			require.Nil(t, plugin.Write(tt.metrics))
			require.Equal(t, tt.expectedMetricsCount, len(mockCwl.pushedLogEvents))

			for index, elem := range mockCwl.pushedLogEvents {
				require.Equal(t, *elem.Message, tt.metrics[tt.expectedMetricsOrder[index]].Fields()["message"])
				require.Equal(t, *elem.Timestamp, tt.metrics[tt.expectedMetricsOrder[index]].Time().UnixNano()/1000000)
			}
		})
	}
}
