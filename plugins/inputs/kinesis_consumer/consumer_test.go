package kinesis_consumer

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf/testutil"
)

// fakeKinesisClient drives the shard consumer through the
// expired-iterator recovery path and records every iterator request.
type fakeKinesisClient struct {
	iterRequests  []kinesis.GetShardIteratorInput
	getRecordsSeq int
	recordsPerGet [][]types.Record
	expireAt      map[int]bool
}

func (f *fakeKinesisClient) GetShardIterator(
	_ context.Context,
	params *kinesis.GetShardIteratorInput,
	_ ...func(*kinesis.Options),
) (*kinesis.GetShardIteratorOutput, error) {
	f.iterRequests = append(f.iterRequests, *params)
	iter := aws.String("iter-after-" + *params.ShardId)
	return &kinesis.GetShardIteratorOutput{ShardIterator: iter}, nil
}

func (f *fakeKinesisClient) GetRecords(
	_ context.Context,
	_ *kinesis.GetRecordsInput,
	_ ...func(*kinesis.Options),
) (*kinesis.GetRecordsOutput, error) {
	i := f.getRecordsSeq
	f.getRecordsSeq++

	if f.expireAt[i] {
		return nil, &types.ExpiredIteratorException{Message: aws.String("Iterator expired")}
	}

	next := aws.String("iter-next")
	if i == len(f.recordsPerGet)-1 {
		// no next iterator ends the consume loop
		next = nil
	}
	return &kinesis.GetRecordsOutput{
		Records:          f.recordsPerGet[i],
		NextShardIterator: next,
	}, nil
}

func TestConsumeRecreatesExpiredIteratorFromLatestSequenceNumber(t *testing.T) {
	shardID := "shard-1"
	startupSeqnr := "100"
	latestSeqnr := "999"

	fake := &fakeKinesisClient{
		// first GetRecords succeeds with a record far newer than the
		// startup checkpoint, the second one hits an expired iterator,
		// the third returns no records and no next iterator to end the loop
		recordsPerGet: [][]types.Record{
			{{SequenceNumber: aws.String(latestSeqnr), Data: []byte(`{}`)}},
			{},
			{},
		},
		expireAt: map[int]bool{1: true},
	}

	var delivered int
	sc := &shardConsumer{
		seqnr:    startupSeqnr,
		interval: time.Millisecond,
		log:      testutil.Logger{},
		client:   fake,
		params: &kinesis.GetShardIteratorInput{
			ShardId:                &shardID,
			ShardIteratorType:      types.ShardIteratorTypeAfterSequenceNumber,
			StartingSequenceNumber: &startupSeqnr,
			StreamName:             aws.String("stream"),
		},
		onMessage: func(_ context.Context, _ string, _ *types.Record) {
			delivered++
		},
	}

	_, err := sc.consume(context.Background(), shardID)
	require.NoError(t, err)

	// the single record was processed exactly once
	require.Equal(t, 1, delivered)

	// two iterator requests: the initial one and the recreation after expiry
	require.Len(t, fake.iterRequests, 2)

	initial, recreated := fake.iterRequests[0], fake.iterRequests[1]
	assert.Equal(t, types.ShardIteratorTypeAfterSequenceNumber, initial.ShardIteratorType)
	require.NotNil(t, recreated.StartingSequenceNumber)

	// the recreated iterator must resume from the latest consumed sequence
	// number, not the (potentially far older) startup one (#19505)
	assert.Equal(t, types.ShardIteratorTypeAfterSequenceNumber, recreated.ShardIteratorType)
	assert.Equal(t, latestSeqnr, *recreated.StartingSequenceNumber)
}
