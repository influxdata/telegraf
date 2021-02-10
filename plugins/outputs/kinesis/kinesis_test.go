package kinesis

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

const zero int64 = 0

func TestPartitionKey(t *testing.T) {

	assert := assert.New(t)
	testPoint := testutil.TestMetric(1)

	k := KinesisOutput{
		Partition: &Partition{
			Method: "static",
			Key:    "-",
		},
	}
	assert.Equal("-", k.getPartitionKey(testPoint), "PartitionKey should be '-'")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "tag",
			Key:    "tag1",
		},
	}
	assert.Equal(testPoint.Tags()["tag1"], k.getPartitionKey(testPoint), "PartitionKey should be value of 'tag1'")

	k = KinesisOutput{
		Partition: &Partition{
			Method:  "tag",
			Key:     "doesnotexist",
			Default: "somedefault",
		},
	}
	assert.Equal("somedefault", k.getPartitionKey(testPoint), "PartitionKey should use default")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "tag",
			Key:    "doesnotexist",
		},
	}
	assert.Equal("telegraf", k.getPartitionKey(testPoint), "PartitionKey should be telegraf")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "not supported",
		},
	}
	assert.Equal("", k.getPartitionKey(testPoint), "PartitionKey should be value of ''")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "measurement",
		},
	}
	assert.Equal(testPoint.Name(), k.getPartitionKey(testPoint), "PartitionKey should be value of measurement name")

	k = KinesisOutput{
		Partition: &Partition{
			Method: "random",
		},
	}
	partitionKey := k.getPartitionKey(testPoint)
	u, err := uuid.FromString(partitionKey)
	assert.Nil(err, "Issue parsing UUID")
	assert.Equal(byte(4), u.Version(), "PartitionKey should be UUIDv4")

	k = KinesisOutput{
		PartitionKey: "-",
	}
	assert.Equal("-", k.getPartitionKey(testPoint), "PartitionKey should be '-'")

	k = KinesisOutput{
		RandomPartitionKey: true,
	}
	partitionKey = k.getPartitionKey(testPoint)
	u, err = uuid.FromString(partitionKey)
	assert.Nil(err, "Issue parsing UUID")
	assert.Equal(byte(4), u.Version(), "PartitionKey should be UUIDv4")
}

func TestWriteKinesis_WhenSuccess(t *testing.T) {

	assert := assert.New(t)

	partitionKey := "partitionKey"
	shard := "shard"
	sequenceNumber := "sequenceNumber"
	streamName := "stream"

	records := []*kinesis.PutRecordsRequestEntry{
		{
			PartitionKey: &partitionKey,
			Data:         []byte{0x65},
		},
	}

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(
		0,
		[]*kinesis.PutRecordsResultEntry{
			{
				ErrorCode:      nil,
				ErrorMessage:   nil,
				SequenceNumber: &sequenceNumber,
				ShardId:        &shard,
			},
		},
	)

	k := KinesisOutput{StreamName: streamName}
	k.svc = svc

	elapsed := k.writeKinesis(records)
	assert.GreaterOrEqual(elapsed.Nanoseconds(), zero)

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records:    records,
		},
	})
}

func TestWriteKinesis_WhenRecordErrors(t *testing.T) {

	assert := assert.New(t)

	errorCode := "InternalFailure"
	errorMessage := "Internal Service Failure"
	partitionKey := "partitionKey"
	streamName := "stream"

	records := []*kinesis.PutRecordsRequestEntry{
		{
			PartitionKey: &partitionKey,
			Data:         []byte{0x66},
		},
	}

	svc := &mockKinesisPutRecords{}
	svc.SetupResponse(
		1,
		[]*kinesis.PutRecordsResultEntry{
			{
				ErrorCode:      &errorCode,
				ErrorMessage:   &errorMessage,
				SequenceNumber: nil,
				ShardId:        nil,
			},
		},
	)

	k := KinesisOutput{StreamName: streamName}
	k.svc = svc

	elapsed := k.writeKinesis(records)
	assert.GreaterOrEqual(elapsed.Nanoseconds(), zero)

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records:    records,
		},
	})
}

func TestWriteKinesis_WhenServiceError(t *testing.T) {

	assert := assert.New(t)

	partitionKey := "partitionKey"
	streamName := "stream"

	records := []*kinesis.PutRecordsRequestEntry{
		{
			PartitionKey: &partitionKey,
			Data:         []byte{},
		},
	}

	svc := &mockKinesisPutRecords{}
	svc.SetupErrorResponse(
		awserr.New("InvalidArgumentException", "Invalid record", nil),
	)

	k := KinesisOutput{StreamName: streamName}
	k.svc = svc

	elapsed := k.writeKinesis(records)
	assert.GreaterOrEqual(elapsed.Nanoseconds(), zero)

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records:    records,
		},
	})
}

type mockKinesisPutRecordsResponse struct {
	Output *kinesis.PutRecordsOutput
	Err    error
}

type mockKinesisPutRecords struct {
	kinesisiface.KinesisAPI

	requests  []*kinesis.PutRecordsInput
	responses []*mockKinesisPutRecordsResponse
}

func (m *mockKinesisPutRecords) SetupResponse(
	failedRecordCount int64,
	records []*kinesis.PutRecordsResultEntry,
) {

	m.responses = append(m.responses, &mockKinesisPutRecordsResponse{
		Err: nil,
		Output: &kinesis.PutRecordsOutput{
			FailedRecordCount: &failedRecordCount,
			Records:           records,
		},
	})
}

func (m *mockKinesisPutRecords) SetupErrorResponse(err error) {

	m.responses = append(m.responses, &mockKinesisPutRecordsResponse{
		Err:    err,
		Output: nil,
	})
}

func (m *mockKinesisPutRecords) PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {

	var reqNum = len(m.requests)
	if reqNum > len(m.responses) {
		return nil, fmt.Errorf("Response for request %+v not setup", reqNum)
	}

	m.requests = append(m.requests, input)

	var resp = m.responses[reqNum]
	return resp.Output, resp.Err
}

func (m *mockKinesisPutRecords) AssertRequests(
	assert *assert.Assertions,
	expected []*kinesis.PutRecordsInput,
) {

	assert.Equal(
		len(expected),
		len(m.requests),
		fmt.Sprintf("Expected %v requests", len(expected)),
	)

	for i, expectedInput := range expected {
		actualInput := m.requests[i]

		assert.Equal(
			expectedInput.StreamName,
			actualInput.StreamName,
			fmt.Sprintf("Expected request %v to have correct StreamName", i),
		)

		assert.Equal(
			len(expectedInput.Records),
			len(actualInput.Records),
			fmt.Sprintf("Expected request %v to have %v Records", i, len(expectedInput.Records)),
		)

		for r, expectedRecord := range expectedInput.Records {
			actualRecord := actualInput.Records[r]

			assert.Equal(
				&expectedRecord.PartitionKey,
				&actualRecord.PartitionKey,
				fmt.Sprintf("Expected (request %v, record %v) to have correct PartitionKey", i, r),
			)

			assert.Equal(
				&expectedRecord.ExplicitHashKey,
				&actualRecord.ExplicitHashKey,
				fmt.Sprintf("Expected (request %v, record %v) to have correct ExplicitHashKey", i, r),
			)

			assert.Equal(
				expectedRecord.Data,
				actualRecord.Data,
				fmt.Sprintf("Expected (request %v, record %v) to have correct Data", i, r),
			)
		}
	}
}
