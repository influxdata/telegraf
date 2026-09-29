package kinesis_consumer

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

// kinesisStub serves the two operations a shard consumer performs. It returns
// one record, expires the iterator, then closes the shard with a second record
// so the consumer has to resume after the record it already consumed.
type kinesisStub struct {
	t  *testing.T
	mu sync.Mutex

	// requested starting points of every GetShardIterator call
	iteratorTypes []string
	startingSeqnr []string

	getRecords int
}

func (s *kinesisStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var request map[string]any
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	target := r.Header.Get("X-Amz-Target")
	switch {
	case strings.HasSuffix(target, "GetShardIterator"):
		iteratorType, _ := request["ShardIteratorType"].(string)
		seqnr, _ := request["StartingSequenceNumber"].(string)
		s.iteratorTypes = append(s.iteratorTypes, iteratorType)
		s.startingSeqnr = append(s.startingSeqnr, seqnr)
		writeJSON(s.t, w, map[string]any{"ShardIterator": "iterator"})
	case strings.HasSuffix(target, "GetRecords"):
		s.getRecords++
		switch s.getRecords {
		case 1:
			writeJSON(s.t, w, map[string]any{
				"Records":           []map[string]any{record("43")},
				"NextShardIterator": "iterator",
			})
		case 2:
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(s.t, w, map[string]any{
				"__type":  "ExpiredIteratorException",
				"message": "iterator expired",
			})
		default:
			// A nil iterator closes the shard and ends the consumer
			writeJSON(s.t, w, map[string]any{"Records": []map[string]any{record("44")}})
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func record(seqnr string) map[string]any {
	return map[string]any{
		"SequenceNumber": seqnr,
		"Data":           base64.StdEncoding.EncodeToString([]byte(seqnr)),
		"PartitionKey":   "key",
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, body map[string]any) {
	if err := json.NewEncoder(w).Encode(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		t.Error(err)
	}
}

func TestConsumeResumesAfterIteratorExpired(t *testing.T) {
	stub := &kinesisStub{t: t}
	server := httptest.NewServer(stub)
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
	require.Equal(t, []string{"AFTER_SEQUENCE_NUMBER", "AFTER_SEQUENCE_NUMBER"}, stub.iteratorTypes)
	require.Equal(t, []string{"42", "43"}, stub.startingSeqnr)
}

// closedShardStub serves a single GetRecords call that returns records
// together with a nil iterator, which is how Kinesis delivers the tail of a
// shard that was split or merged.
type closedShardStub struct {
	t *testing.T
}

func (s *closedShardStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	if strings.HasSuffix(r.Header.Get("X-Amz-Target"), "GetShardIterator") {
		writeJSON(s.t, w, map[string]any{"ShardIterator": "iterator"})
		return
	}
	writeJSON(s.t, w, map[string]any{
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
	})
}

func TestConsumeEmitsRecordsOfClosedShard(t *testing.T) {
	server := httptest.NewServer(&closedShardStub{t: t})
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
