package kinesis_consumer

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	consumer "github.com/harlow/kinesis-consumer"
	"github.com/harlow/kinesis-consumer/store/ddb"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type (
	DynamoDB struct {
		AppName   string `toml:"app_name"`
		TableName string `toml:"table_name"`
	}

	KinesisConsumer struct {
		StreamName             string    `toml:"streamname"`
		ShardIteratorType      string    `toml:"shard_iterator_type"`
		DynamoDB               *DynamoDB `toml:"checkpoint_dynamodb"`
		MaxUndeliveredMessages int       `toml:"max_undelivered_messages"`
		ContentEncoding        string    `toml:"content_encoding"`

		Log telegraf.Logger

		cons   *consumer.Consumer
		parser parsers.Parser
		cancel context.CancelFunc
		acc    telegraf.TrackingAccumulator
		sem    chan struct{}

		checkpoint    consumer.Store
		checkpoints   map[string]checkpoint
		records       map[telegraf.TrackingID]string
		checkpointTex sync.Mutex
		recordsTex    sync.Mutex
		wg            sync.WaitGroup

		processContentEncodingFunc processContent

		lastSeqNum *big.Int

		internalaws.CredentialConfig
	}

	checkpoint struct {
		streamName string
		shardID    string
	}
)

const (
	defaultMaxUndeliveredMessages = 1000
)

type processContent func([]byte) ([]byte, error)

// this is the largest sequence number allowed - https://docs.aws.amazon.com/kinesis/latest/APIReference/API_SequenceNumberRange.html
var maxSeq = strToBint(strings.Repeat("9", 129))

func (k *KinesisConsumer) SetParser(parser parsers.Parser) {
	k.parser = parser
}

func (k *KinesisConsumer) connect(ac telegraf.Accumulator) error {
	cfg, err := k.CredentialConfig.Credentials()
	if err != nil {
		return err
	}
	client := kinesis.NewFromConfig(cfg)

	k.checkpoint = &noopStore{}
	if k.DynamoDB != nil {
		var err error
		k.checkpoint, err = ddb.New(
			k.DynamoDB.AppName,
			k.DynamoDB.TableName,
			ddb.WithDynamoClient(dynamodb.NewFromConfig(cfg)),
			ddb.WithMaxInterval(time.Second*10),
		)
		if err != nil {
			return err
		}
	}

	cons, err := consumer.New(
		k.StreamName,
		consumer.WithClient(client),
		consumer.WithShardIteratorType(k.ShardIteratorType),
		consumer.WithStore(k),
	)
	if err != nil {
		return err
	}

	k.cons = cons

	k.acc = ac.WithTracking(k.MaxUndeliveredMessages)
	k.records = make(map[telegraf.TrackingID]string, k.MaxUndeliveredMessages)
	k.checkpoints = make(map[string]checkpoint, k.MaxUndeliveredMessages)
	k.sem = make(chan struct{}, k.MaxUndeliveredMessages)

	ctx := context.Background()
	ctx, k.cancel = context.WithCancel(ctx)

	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		k.onDelivery(ctx)
	}()

	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		err := k.cons.Scan(ctx, func(r *consumer.Record) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case k.sem <- struct{}{}:
				break
			}
			err := k.onMessage(k.acc, r)
			if err != nil {
				<-k.sem
				k.Log.Errorf("Scan parser error: %s", err.Error())
			}

			return nil
		})
		if err != nil {
			k.cancel()
			k.Log.Errorf("Scan encountered an error: %s", err.Error())
			k.cons = nil
		}
	}()

	return nil
}

func (k *KinesisConsumer) Start(ac telegraf.Accumulator) error {
	err := k.connect(ac)
	if err != nil {
		return err
	}

	return nil
}

func (k *KinesisConsumer) onMessage(acc telegraf.TrackingAccumulator, r *consumer.Record) error {
	data, err := k.processContentEncodingFunc(r.Data)
	if err != nil {
		return err
	}
	metrics, err := k.parser.Parse(data)
	if err != nil {
		return err
	}

	k.recordsTex.Lock()
	id := acc.AddTrackingMetricGroup(metrics)
	k.records[id] = *r.SequenceNumber
	k.recordsTex.Unlock()

	return nil
}

func (k *KinesisConsumer) onDelivery(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case info := <-k.acc.Delivered():
			k.recordsTex.Lock()
			sequenceNum, ok := k.records[info.ID()]
			if !ok {
				k.recordsTex.Unlock()
				continue
			}
			<-k.sem
			delete(k.records, info.ID())
			k.recordsTex.Unlock()

			if info.Delivered() {
				k.checkpointTex.Lock()
				chk, ok := k.checkpoints[sequenceNum]
				if !ok {
					k.checkpointTex.Unlock()
					continue
				}
				delete(k.checkpoints, sequenceNum)
				k.checkpointTex.Unlock()

				// at least once
				if strToBint(sequenceNum).Cmp(k.lastSeqNum) > 0 {
					continue
				}

				k.lastSeqNum = strToBint(sequenceNum)
				if err := k.checkpoint.SetCheckpoint(chk.streamName, chk.shardID, sequenceNum); err != nil {
					k.Log.Debug("Setting checkpoint failed: %v", err)
				}
			} else {
				k.Log.Debug("Metric group failed to process")
			}
		}
	}
}

var negOne *big.Int

func strToBint(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return negOne
	}
	return n
}

func (k *KinesisConsumer) Stop() {
	k.cancel()
	k.wg.Wait()
}

func (k *KinesisConsumer) Gather(acc telegraf.Accumulator) error {
	if k.cons == nil {
		return k.connect(acc)
	}
	k.lastSeqNum = maxSeq

	return nil
}

// Get wraps the checkpoint's GetCheckpoint function (called by consumer library)
func (k *KinesisConsumer) GetCheckpoint(streamName, shardID string) (string, error) {
	return k.checkpoint.GetCheckpoint(streamName, shardID)
}

// Set wraps the checkpoint's SetCheckpoint function (called by consumer library)
func (k *KinesisConsumer) SetCheckpoint(streamName, shardID, sequenceNumber string) error {
	if sequenceNumber == "" {
		return fmt.Errorf("sequence number should not be empty")
	}

	k.checkpointTex.Lock()
	k.checkpoints[sequenceNumber] = checkpoint{streamName: streamName, shardID: shardID}
	k.checkpointTex.Unlock()

	return nil
}

func processGzip(data []byte) ([]byte, error) {
	zipData, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zipData.Close()
	return io.ReadAll(zipData)
}

func processZlib(data []byte) ([]byte, error) {
	zlibData, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zlibData.Close()
	return io.ReadAll(zlibData)
}

func processNoOp(data []byte) ([]byte, error) {
	return data, nil
}

func (k *KinesisConsumer) configureProcessContentEncodingFunc() error {
	switch k.ContentEncoding {
	case "gzip":
		k.processContentEncodingFunc = processGzip
	case "zlib":
		k.processContentEncodingFunc = processZlib
	case "none", "identity", "":
		k.processContentEncodingFunc = processNoOp
	default:
		return fmt.Errorf("unknown content encoding %q", k.ContentEncoding)
	}
	return nil
}

func (k *KinesisConsumer) Init() error {
	return k.configureProcessContentEncodingFunc()
}

type noopStore struct{}

func (n noopStore) SetCheckpoint(string, string, string) error   { return nil }
func (n noopStore) GetCheckpoint(string, string) (string, error) { return "", nil }

func init() {
	negOne, _ = new(big.Int).SetString("-1", 10)

	inputs.Add("kinesis_consumer", func() telegraf.Input {
		return &KinesisConsumer{
			ShardIteratorType:      "TRIM_HORIZON",
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
			lastSeqNum:             maxSeq,
			ContentEncoding:        "identity",
		}
	})
}
