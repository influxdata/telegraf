package kinesis

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const zero int64 = 0

func TestPartitionKey(t *testing.T) {

	assert := assert.New(t)
	testPoint := testutil.TestMetric(1)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    "-",
		},
	}
	assert.Equal("-", k.getPartitionKey(testPoint), "PartitionKey should be '-'")

	k = KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "tag",
			Key:    "tag1",
		},
	}
	assert.Equal(testPoint.Tags()["tag1"], k.getPartitionKey(testPoint), "PartitionKey should be value of 'tag1'")

	k = KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method:  "tag",
			Key:     "doesnotexist",
			Default: "somedefault",
		},
	}
	assert.Equal("somedefault", k.getPartitionKey(testPoint), "PartitionKey should use default")

	k = KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "tag",
			Key:    "doesnotexist",
		},
	}
	assert.Equal("telegraf", k.getPartitionKey(testPoint), "PartitionKey should be telegraf")

	k = KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "not supported",
		},
	}
	assert.Equal("", k.getPartitionKey(testPoint), "PartitionKey should be value of ''")

	k = KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "measurement",
		},
	}
	assert.Equal(testPoint.Name(), k.getPartitionKey(testPoint), "PartitionKey should be value of measurement name")

	k = KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "random",
		},
	}
	partitionKey := k.getPartitionKey(testPoint)
	u, err := uuid.FromString(partitionKey)
	assert.Nil(err, "Issue parsing UUID")
	assert.Equal(byte(4), u.Version(), "PartitionKey should be UUIDv4")

	k = KinesisOutput{
		Log:          testutil.Logger{},
		PartitionKey: "-",
	}
	assert.Equal("-", k.getPartitionKey(testPoint), "PartitionKey should be '-'")

	k = KinesisOutput{
		Log:                testutil.Logger{},
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

	k := KinesisOutput{
		Log:        testutil.Logger{},
		StreamName: streamName,
		svc:        svc,
	}

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

	k := KinesisOutput{
		Log:        testutil.Logger{},
		StreamName: streamName,
		svc:        svc,
	}

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

	k := KinesisOutput{
		Log:        testutil.Logger{},
		StreamName: streamName,
		svc:        svc,
	}

	elapsed := k.writeKinesis(records)
	assert.GreaterOrEqual(elapsed.Nanoseconds(), zero)

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records:    records,
		},
	})
}

func TestWrite_NoMetrics(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	svc := &mockKinesisPutRecords{}

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    "partitionKey",
		},
		StreamName: "stream",
		serializer: serializer,
		svc:        svc,
	}

	err := k.Write([]telegraf.Metric{})
	assert.Nil(err, "Should not return error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{})
}

func TestWrite_SingleMetric(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	partitionKey := "partitionKey"
	streamName := "stream"

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(1, 0)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    partitionKey,
		},
		StreamName: streamName,
		serializer: serializer,
		svc:        svc,
	}

	metric, metricData := createTestMetric(t, "metric1", serializer)
	err := k.Write([]telegraf.Metric{metric})
	assert.Nil(err, "Should not return error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				{
					PartitionKey: &partitionKey,
					Data:         metricData,
				},
			},
		},
	})
}

func TestWrite_MultipleMetrics_SinglePartialRequest(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	partitionKey := "partitionKey"
	streamName := "stream"

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(3, 0)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    partitionKey,
		},
		StreamName: streamName,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, 3, serializer)
	err := k.Write(metrics)
	assert.Nil(err, "Should not return error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: createPutRecordsRequestEntries(
				metricsData,
				&partitionKey,
			),
		},
	})
}

func TestWrite_MultipleMetrics_SingleFullRequest(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	partitionKey := "partitionKey"
	streamName := "stream"

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    partitionKey,
		},
		StreamName: streamName,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, maxRecordsPerRequest, serializer)
	err := k.Write(metrics)
	assert.Nil(err, "Should not return error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: createPutRecordsRequestEntries(
				metricsData,
				&partitionKey,
			),
		},
	})
}

func TestWrite_MultipleMetrics_MultipleRequests(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	partitionKey := "partitionKey"
	streamName := "stream"

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)
	svc.SetupGenericResponse(1, 0)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    partitionKey,
		},
		StreamName: streamName,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, maxRecordsPerRequest+1, serializer)
	err := k.Write(metrics)
	assert.Nil(err, "Should not return error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: createPutRecordsRequestEntries(
				metricsData[0:maxRecordsPerRequest],
				&partitionKey,
			),
		},
		{
			StreamName: &streamName,
			Records: createPutRecordsRequestEntries(
				metricsData[maxRecordsPerRequest:],
				&partitionKey,
			),
		},
	})
}

func TestWrite_MultipleMetrics_MultipleFullRequests(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	partitionKey := "partitionKey"
	streamName := "stream"

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)
	svc.SetupGenericResponse(maxRecordsPerRequest, 0)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    partitionKey,
		},
		StreamName: streamName,
		serializer: serializer,
		svc:        svc,
	}

	metrics, metricsData := createTestMetrics(t, maxRecordsPerRequest*2, serializer)
	err := k.Write(metrics)
	assert.Nil(err, "Should not return error")

	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: createPutRecordsRequestEntries(
				metricsData[0:maxRecordsPerRequest],
				&partitionKey,
			),
		},
		{
			StreamName: &streamName,
			Records: createPutRecordsRequestEntries(
				metricsData[maxRecordsPerRequest:],
				&partitionKey,
			),
		},
	})
}

func TestWrite_SerializerError(t *testing.T) {
	assert := assert.New(t)
	serializer := influx.NewSerializer()
	partitionKey := "partitionKey"
	streamName := "stream"

	svc := &mockKinesisPutRecords{}
	svc.SetupGenericResponse(2, 0)

	k := KinesisOutput{
		Log: testutil.Logger{},
		Partition: &Partition{
			Method: "static",
			Key:    partitionKey,
		},
		StreamName: streamName,
		serializer: serializer,
		svc:        svc,
	}

	metric1, metric1Data := createTestMetric(t, "metric1", serializer)
	metric2, metric2Data := createTestMetric(t, "metric2", serializer)

	// metric is invalid because of empty name
	invalidMetric := testutil.TestMetric(3, "")

	err := k.Write([]telegraf.Metric{
		metric1,
		invalidMetric,
		metric2,
	})
	assert.Nil(err, "Should not return error")

	// remaining valid metrics should still get written
	svc.AssertRequests(assert, []*kinesis.PutRecordsInput{
		{
			StreamName: &streamName,
			Records: []*kinesis.PutRecordsRequestEntry{
				{
					PartitionKey: &partitionKey,
					Data:         metric1Data,
				},
				{
					PartitionKey: &partitionKey,
					Data:         metric2Data,
				},
			},
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

func (m *mockKinesisPutRecords) SetupGenericResponse(
	successfulRecordCount uint32,
	failedRecordCount uint32,
) {

	errorCode := "InternalFailure"
	errorMessage := "Internal Service Failure"
	shard := "shardId-000000000003"

	records := []*kinesis.PutRecordsResultEntry{}

	for i := uint32(0); i < successfulRecordCount; i++ {
		sequenceNumber := fmt.Sprintf("%d", i)
		records = append(records, &kinesis.PutRecordsResultEntry{
			SequenceNumber: &sequenceNumber,
			ShardId:        &shard,
		})
	}

	for i := uint32(0); i < failedRecordCount; i++ {
		records = append(records, &kinesis.PutRecordsResultEntry{
			ErrorCode:    &errorCode,
			ErrorMessage: &errorMessage,
		})
	}

	m.SetupResponse(int64(failedRecordCount), records)
}

func (m *mockKinesisPutRecords) SetupErrorResponse(err error) {

	m.responses = append(m.responses, &mockKinesisPutRecordsResponse{
		Err:    err,
		Output: nil,
	})
}

func (m *mockKinesisPutRecords) PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {

	reqNum := len(m.requests)
	if reqNum > len(m.responses) {
		return nil, fmt.Errorf("Response for request %+v not setup", reqNum)
	}

	m.requests = append(m.requests, input)

	resp := m.responses[reqNum]
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

func createTestMetric(
	t *testing.T,
	name string,
	serializer serializers.Serializer,
) (telegraf.Metric, []byte) {

	metric := testutil.TestMetric(1, name)

	data, err := serializer.Serialize(metric)
	require.NoError(t, err)

	return metric, data
}

func createTestMetrics(
	t *testing.T,
	count uint32,
	serializer serializers.Serializer,
) ([]telegraf.Metric, [][]byte) {

	metrics := make([]telegraf.Metric, count)
	metricsData := make([][]byte, count)

	for i := uint32(0); i < count; i++ {
		name := fmt.Sprintf("metric%d", i)
		metric, data := createTestMetric(t, name, serializer)
		metrics[i] = metric
		metricsData[i] = data
	}

	return metrics, metricsData
}

func createPutRecordsRequestEntries(
	metricsData [][]byte,
	partitionKey *string,
) []*kinesis.PutRecordsRequestEntry {

	count := len(metricsData)
	records := make([]*kinesis.PutRecordsRequestEntry, count)

	for i := 0; i < count; i++ {
		records[i] = &kinesis.PutRecordsRequestEntry{
			PartitionKey: partitionKey,
			Data:         metricsData[i],
		}
	}

	return records
}
