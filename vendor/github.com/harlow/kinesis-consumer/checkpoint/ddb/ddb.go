package ddb

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// Option is used to override defaults when creating a new Checkpoint
type Option func(*Checkpoint)

// WithMaxInterval sets the flush interval
func WithMaxInterval(maxInterval time.Duration) Option {
	return func(c *Checkpoint) {
		c.maxInterval = maxInterval
	}
}

// WithDynamoClient sets the dynamoDb client
func WithDynamoClient(svc dynamodbiface.DynamoDBAPI) Option {
	return func(c *Checkpoint) {
		c.client = svc
	}
}

// WithRetryer sets the retryer
func WithRetryer(r Retryer) Option {
	return func(c *Checkpoint) {
		c.retryer = r
	}
}

// New returns a checkpoint that uses DynamoDB for underlying storage
func New(appName, tableName string, opts ...Option) (*Checkpoint, error) {
	client := dynamodb.New(session.New(aws.NewConfig()))

	ck := &Checkpoint{
		tableName:   tableName,
		appName:     appName,
		client:      client,
		maxInterval: time.Duration(1 * time.Minute),
		done:        make(chan struct{}),
		mu:          &sync.Mutex{},
		checkpoints: map[key]string{},
		retryer:     &DefaultRetryer{},
	}

	for _, opt := range opts {
		opt(ck)
	}

	go ck.loop()

	return ck, nil
}

// Checkpoint stores and retreives the last evaluated key from a DDB scan
type Checkpoint struct {
	tableName   string
	appName     string
	client      dynamodbiface.DynamoDBAPI
	maxInterval time.Duration
	mu          *sync.Mutex // protects the checkpoints
	checkpoints map[key]string
	done        chan struct{}
	retryer     Retryer
}

type key struct {
	streamName string
	shardID    string
}

type item struct {
	Namespace      string `json:"namespace"`
	ShardID        string `json:"shard_id"`
	SequenceNumber string `json:"sequence_number"`
}

// Get determines if a checkpoint for a particular Shard exists.
// Typically used to determine whether we should start processing the shard with
// TRIM_HORIZON or AFTER_SEQUENCE_NUMBER (if checkpoint exists).
func (c *Checkpoint) Get(streamName, shardID string) (string, error) {
	namespace := fmt.Sprintf("%s-%s", c.appName, streamName)

	params := &dynamodb.GetItemInput{
		TableName:      aws.String(c.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"namespace": &dynamodb.AttributeValue{
				S: aws.String(namespace),
			},
			"shard_id": &dynamodb.AttributeValue{
				S: aws.String(shardID),
			},
		},
	}

	resp, err := c.client.GetItem(params)
	if err != nil {
		if c.retryer.ShouldRetry(err) {
			return c.Get(streamName, shardID)
		}
		return "", err
	}

	var i item
	dynamodbattribute.UnmarshalMap(resp.Item, &i)
	return i.SequenceNumber, nil
}

// Set stores a checkpoint for a shard (e.g. sequence number of last record processed by application).
// Upon failover, record processing is resumed from this point.
func (c *Checkpoint) Set(streamName, shardID, sequenceNumber string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if sequenceNumber == "" {
		return fmt.Errorf("sequence number should not be empty")
	}

	key := key{
		streamName: streamName,
		shardID:    shardID,
	}
	c.checkpoints[key] = sequenceNumber

	return nil
}

// Shutdown the checkpoint. Save any in-flight data.
func (c *Checkpoint) Shutdown() error {
	c.done <- struct{}{}
	return c.save()
}

func (c *Checkpoint) loop() {
	tick := time.NewTicker(c.maxInterval)
	defer tick.Stop()
	defer close(c.done)

	for {
		select {
		case <-tick.C:
			c.save()
		case <-c.done:
			return
		}
	}
}

func (c *Checkpoint) save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, sequenceNumber := range c.checkpoints {
		item, err := dynamodbattribute.MarshalMap(item{
			Namespace:      fmt.Sprintf("%s-%s", c.appName, key.streamName),
			ShardID:        key.shardID,
			SequenceNumber: sequenceNumber,
		})
		if err != nil {
			log.Printf("marshal map error: %v", err)
			return nil
		}

		_, err = c.client.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(c.tableName),
			Item:      item,
		})
		if err != nil {
			if !c.retryer.ShouldRetry(err) {
				return err
			}
			return c.save()
		}
	}

	return nil
}
