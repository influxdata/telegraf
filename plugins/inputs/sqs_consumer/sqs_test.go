package sqs_consumer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	sqsConsumer "github.com/influxdata/telegraf/plugins/inputs/sqs_consumer"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SqsInputPlugin(t *testing.T) {

	t.Run("with no messages in queue", func(t *testing.T) {
		queueUrl := "http://fake.endpoint/queue"
		plugin := &sqsConsumer.Sqs{
			URL:    queueUrl,
			Region: "eu-west-1",
		}

		var acc testutil.Accumulator
		_ = plugin.Init()
		fSqs := &fakeAwsSqs{
			ReturnsMessages: []*sqs.Message{},
		}
		plugin.SqsSdk = fSqs

		require.NoError(t, acc.GatherError(plugin.Gather))

		t.Run("should have no metrics", func(t *testing.T) {
			require.Len(t, acc.Metrics, 0)
		})

		t.Run("should have called sqs", func(t *testing.T) {
			require.Len(t, fSqs.ReceiveMessageInvocations, 1)
			req := fSqs.ReceiveMessageInvocations[0]
			assert.Equal(t, queueUrl, *req.QueueUrl, "should match path")
		})

	})

}

func Test_SqsInputPlugin_WithMessagesInQueue(t *testing.T) {
	t.Run("with messages in queue", func(t *testing.T) {
		queueUrl := "http://fake.endpoint/queue"
		plugin := &sqsConsumer.Sqs{
			URL:    queueUrl,
			Region: "eu-west-1",
		}

		var acc testutil.Accumulator
		_ = plugin.Init()
		plugin.SetParser(&fakeParser{
			invocations: 0,
		})
		messageId1 := "md1"
		receiptHandle1 := "rh1"
		messageId2 := "md2"
		receiptHandle2 := "rh2"
		msgBody := "body"
		fSqs := &fakeAwsSqs{
			ReturnsMessages: []*sqs.Message{
				{
					Body:          &msgBody,
					MessageId:     &messageId1,
					ReceiptHandle: &receiptHandle1,
				},
				{
					Body:          &msgBody,
					MessageId:     &messageId2,
					ReceiptHandle: &receiptHandle2,
				},
			},
		}
		plugin.SqsSdk = fSqs

		t.Run("should have no errors", func(t *testing.T) {
			require.NoError(t, acc.GatherError(plugin.Gather))
		})

		t.Run("should have two metrics", func(t *testing.T) {
			require.Len(t, acc.Metrics, 2)
			assert.Equal(t, int64(0), acc.Metrics[0].Fields["value"])
			assert.Equal(t, int64(1), acc.Metrics[1].Fields["value"])
		})

		t.Run("should delete messages from queue", func(t *testing.T) {
			require.Len(t, fSqs.DeleteBatchMessageInvocations, 1)
			req := fSqs.DeleteBatchMessageInvocations[0]
			assert.Equal(t, *req.QueueUrl, queueUrl, "Queue url does not match")
			require.Len(t, req.Entries, 2)
			assert.Equal(t, *req.Entries[0].Id, messageId1)
			assert.Equal(t, *req.Entries[1].Id, messageId2)
			assert.Equal(t, *req.Entries[0].ReceiptHandle, receiptHandle1)
			assert.Equal(t, *req.Entries[1].ReceiptHandle, receiptHandle2)
		})

	})
}

func Test_SqsInputPlugin_WithFaultyMessagesInQueue(t *testing.T) {
	t.Run("with messages in queue", func(t *testing.T) {
		queueUrl := "http://fake.endpoint/queue"
		plugin := &sqsConsumer.Sqs{
			URL:    queueUrl,
			Region: "eu-west-1",
		}

		var acc testutil.Accumulator
		_ = plugin.Init()
		failOnLine := "<xml>asdf1</xml>"
		plugin.SetParser(&fakeParser{
			invocations: 0,
			FailOnLine:  failOnLine,
		})

		msgBody := "body"
		messageId1 := "md1"
		receiptHandle1 := "rh1"
		messageId2 := "md2"
		receiptHandle2 := "rh2"
		fSqs := &fakeAwsSqs{
			ReturnsMessages: []*sqs.Message{
				{
					Body:          &failOnLine,
					MessageId:     &messageId1,
					ReceiptHandle: &receiptHandle1,
				},
				{
					Body:          &msgBody,
					MessageId:     &messageId2,
					ReceiptHandle: &receiptHandle2,
				},
			},
		}
		plugin.SqsSdk = fSqs

		t.Run("should have error", func(t *testing.T) {
			require.Error(t, acc.GatherError(plugin.Gather))
		})

		t.Run("should have one metrics", func(t *testing.T) {
			require.Len(t, acc.Metrics, 1)
			assert.Equal(t, int64(1), acc.Metrics[0].Fields["value"])
		})

		t.Run("should delete non faulty messages from queue", func(t *testing.T) {
			require.Len(t, fSqs.DeleteBatchMessageInvocations, 1)
			req := fSqs.DeleteBatchMessageInvocations[0]
			assert.Equal(t, *req.QueueUrl, queueUrl, "Queue url does not match")
			require.Len(t, req.Entries, 1)
			assert.Equal(t, *req.Entries[0].Id, messageId2)
			assert.Equal(t, *req.Entries[0].ReceiptHandle, receiptHandle2)
		})

	})
}

type fakeAwsSqs struct {
	ReceiveMessageInvocations     []*sqs.ReceiveMessageInput
	DeleteBatchMessageInvocations []*sqs.DeleteMessageBatchInput
	ReturnsMessages               []*sqs.Message
}

func (fSqs *fakeAwsSqs) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	fSqs.ReceiveMessageInvocations = append(fSqs.ReceiveMessageInvocations, input)
	response := &sqs.ReceiveMessageOutput{Messages: fSqs.ReturnsMessages}
	fSqs.ReturnsMessages = []*sqs.Message{}
	return response, nil
}
func (fSqs *fakeAwsSqs) DeleteMessageBatch(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	fSqs.DeleteBatchMessageInvocations = append(fSqs.DeleteBatchMessageInvocations, input)
	return nil, nil
}

type fakeParser struct {
	invocations int
	FailOnLine  string
}

// FakeParser satisfies parsers.Parser
var _ parsers.Parser = &fakeParser{
	invocations: 0,
}

func (p *fakeParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	panic("not implemented")
}

func (p *fakeParser) ParseLine(line string) (telegraf.Metric, error) {
	fields := map[string]interface{}{"value": p.invocations}
	p.invocations += 1

	if p.FailOnLine == line {
		return nil, fmt.Errorf("they told me I should fail")
	}

	return metric.New(
		"fake-metric",
		map[string]string{},
		fields,
		time.Now().UTC(),
	)
}

func (p *fakeParser) SetDefaultTags(tags map[string]string) {
	panic("not implemented")
}
