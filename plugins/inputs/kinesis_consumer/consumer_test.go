package kinesis_consumer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestConsumeResumesAfterIteratorExpired(t *testing.T) {
	// Return one record, expire the iterator, then close the shard with a
	// second record so the consumer has to resume after the consumed record.
	mock := &kinesisMock{
		responses: []map[string]any{
			{
				"Records": []map[string]any{{
					"SequenceNumber": "43",
					"Data":           base64.StdEncoding.EncodeToString([]byte("43")),
					"PartitionKey":   "key",
				}},
				"NextShardIterator": "iterator",
			},
			{
				"__type":  "ExpiredIteratorException",
				"message": "iterator expired",
			},
			{
				"Records": []map[string]any{{
					"SequenceNumber": "44",
					"Data":           base64.StdEncoding.EncodeToString([]byte("44")),
					"PartitionKey":   "key",
				}},
			},
		},
	}
	server := httptest.NewServer(mock)
	defer server.Close()

	client := kinesis.New(kinesis.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("id", "secret", ""),
		BaseEndpoint: aws.String(server.URL),
	})

	var records []string
	consumer := &shardConsumer{
		// Pretend we already consumed record 42 before the iterator expired.
		seqnr:    "42",
		interval: time.Millisecond,
		log:      testutil.Logger{},
		client:   client,
		params: &kinesis.GetShardIteratorInput{
			StreamName:        aws.String("stream"),
			ShardId:           aws.String("shard-0"),
			ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
		},
		onMessage: func(_ context.Context, _ string, r *types.Record) {
			records = append(records, string(r.Data))
		},
	}

	_, err := consumer.consume(t.Context(), "shard-0")
	require.NoError(t, err)

	// The expired iterator must be replaced rather than aborting the shard.
	require.Equal(t, []string{"43", "44"}, records)
	require.Equal(t, "44", consumer.seqnr)

	// The replacement iterator must resume after the last consumed record,
	// otherwise the shard is replayed after every expiry.
	require.Equal(t, []string{"AFTER_SEQUENCE_NUMBER", "AFTER_SEQUENCE_NUMBER"}, mock.iteratorTypes)
	require.Equal(t, []string{"42", "43"}, mock.startingSeqnr)
}

func TestConsumeEmitsRecordsOfClosedShard(t *testing.T) {
	// A nil iterator together with records is how Kinesis delivers the tail
	// of a shard that was split or merged.
	mock := &kinesisMock{
		responses: []map[string]any{
			{
				"Records": []map[string]any{{
					"SequenceNumber": "7",
					"Data":           base64.StdEncoding.EncodeToString([]byte("last")),
					"PartitionKey":   "key",
				}},
				"ChildShards": []map[string]any{{
					"ShardId":      "shard-1",
					"ParentShards": []string{"shard-0"},
					"HashKeyRange": map[string]any{"StartingHashKey": "0", "EndingHashKey": "1"},
				}},
			},
		},
	}
	server := httptest.NewServer(mock)
	defer server.Close()

	client := kinesis.New(kinesis.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("id", "secret", ""),
		BaseEndpoint: aws.String(server.URL),
	})

	var records []string
	consumer := &shardConsumer{
		interval: time.Millisecond,
		log:      testutil.Logger{},
		client:   client,
		params: &kinesis.GetShardIteratorInput{
			StreamName:        aws.String("stream"),
			ShardId:           aws.String("shard-0"),
			ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
		},
		onMessage: func(_ context.Context, _ string, r *types.Record) {
			records = append(records, string(r.Data))
		},
	}

	children, err := consumer.consume(t.Context(), "shard-0")
	require.NoError(t, err)

	// The final batch of a closed shard must not be dropped along with the
	// iterator, otherwise every split or merge loses records.
	require.Equal(t, []string{"last"}, records)
	require.Equal(t, "7", consumer.seqnr)
	require.Len(t, children, 1)
}

// kinesisMock serves the two operations a shard consumer performs. Every
// GetShardIterator call is recorded, GetRecords calls get the configured
// responses in order where a response with "__type" is an AWS error.
type kinesisMock struct {
	responses []map[string]any

	iteratorTypes []string
	startingSeqnr []string

	served int
	sync.Mutex
}

func (m *kinesisMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	var request map[string]any
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response map[string]any
	target := r.Header.Get("X-Amz-Target")
	switch {
	case strings.HasSuffix(target, "GetShardIterator"):
		iteratorType, _ := request["ShardIteratorType"].(string)
		seqnr, _ := request["StartingSequenceNumber"].(string)
		m.iteratorTypes = append(m.iteratorTypes, iteratorType)
		m.startingSeqnr = append(m.startingSeqnr, seqnr)
		response = map[string]any{"ShardIterator": "iterator"}
	case strings.HasSuffix(target, "GetRecords"):
		if m.served >= len(m.responses) {
			http.Error(w, fmt.Sprintf("unexpected GetRecords call %d", m.served+1), http.StatusInternalServerError)
			return
		}
		response = m.responses[m.served]
		m.served++
	default:
		http.Error(w, "unexpected target "+target, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	if _, isError := response["__type"]; isError {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
