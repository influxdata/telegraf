package cloudwatch_logs

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type messageBatch struct {
	logEvents    []types.InputLogEvent
	messageCount int
}
type logStreamContainer struct {
	currentBatchSizeBytes int
	currentBatchIndex     int
	messageBatches        []messageBatch
	sequenceToken         string
}

//Cloudwatch Logs service interface
type cloudWatchLogs interface {
	DescribeLogGroups(context.Context, *cloudwatchlogs.DescribeLogGroupsInput, ...func(options *cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreams(context.Context, *cloudwatchlogs.DescribeLogStreamsInput, ...func(options *cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
	CreateLogStream(context.Context, *cloudwatchlogs.CreateLogStreamInput, ...func(options *cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEvents(context.Context, *cloudwatchlogs.PutLogEventsInput, ...func(options *cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error)
}

// CloudWatchLogs plugin object definition
type CloudWatchLogs struct {
	LogGroup string          `toml:"log_group"`
	lg       *types.LogGroup //log group data

	LogStream string                         `toml:"log_stream"`
	lsKey     string                         //log stream source: tag or field
	lsSource  string                         //log stream source tag or field name
	ls        map[string]*logStreamContainer //log stream info

	LDMetricName string `toml:"log_data_metric_name"`

	LDSource      string `toml:"log_data_source"`
	logDatKey     string //log data source (tag or field)
	logDataSource string //log data source tag or field name

	svc cloudWatchLogs //cloudwatch logs service

	Log telegraf.Logger `toml:"-"`

	internalaws.CredentialConfig
}

const (
	// Log events must comply with the following
	// (https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatchlogs/#CloudWatchLogs.PutLogEvents):
	maxLogMessageLength           = 262144 - awsOverheadPerLogMessageBytes //In bytes
	maxBatchSizeBytes             = 1048576                                // The sum of all event messages in UTF-8, plus 26 bytes for each log event
	awsOverheadPerLogMessageBytes = 26
	maxFutureLogEventTimeOffset   = time.Hour * 2 // None of the log events in the batch can be more than 2 hours in the future.

	maxPastLogEventTimeOffset = time.Hour * 24 * 14 // None of the log events in the batch can be older than 14 days or older
	// than the retention period of the log group.

	maxItemsInBatch = 10000 // The maximum number of log events in a batch is 10,000.

	//maxTimeSpanInBatch = time.Hour * 24 // A batch of log events in a single request cannot span more than 24 hours.
	// Otherwise, the operation fails.
)

// Init initialize plugin with checking configuration parameters
func (c *CloudWatchLogs) Init() error {
	if c.LogGroup == "" {
		return fmt.Errorf("log group is not set")
	}

	if c.LogStream == "" {
		return fmt.Errorf("log stream is not set")
	}

	if c.LDMetricName == "" {
		return fmt.Errorf("log data metrics name is not set")
	}

	if c.LDSource == "" {
		return fmt.Errorf("log data source is not set")
	}
	lsSplitArray := strings.Split(c.LDSource, ":")
	if len(lsSplitArray) != 2 {
		return fmt.Errorf("log data source is not properly formatted, ':' is missed.\n" +
			"Should be 'tag:<tag_mame>' or 'field:<field_name>'")
	}

	if lsSplitArray[0] != "tag" && lsSplitArray[0] != "field" {
		return fmt.Errorf("log data source is not properly formatted.\n" +
			"Should be 'tag:<tag_mame>' or 'field:<field_name>'")
	}

	c.logDatKey = lsSplitArray[0]
	c.logDataSource = lsSplitArray[1]
	c.Log.Debugf("Log data: key '%s', source '%s'...", c.logDatKey, c.logDataSource)

	if c.lsSource == "" {
		c.lsSource = c.LogStream
		c.Log.Debugf("Log stream '%s'...", c.lsSource)
	}

	return nil
}

// Connect connects plugin with to receiver of metrics
func (c *CloudWatchLogs) Connect() error {
	var queryToken *string
	var dummyToken = "dummy"
	var logGroupsOutput = &cloudwatchlogs.DescribeLogGroupsOutput{NextToken: &dummyToken}
	var err error

	cfg, err := c.CredentialConfig.Credentials()
	if err != nil {
		return err
	}
	c.svc = cloudwatchlogs.NewFromConfig(cfg)

	//Find log group with name 'c.LogGroup'
	if c.lg == nil { //In case connection is not retried, first time
		for logGroupsOutput.NextToken != nil {
			logGroupsOutput, err = c.svc.DescribeLogGroups(
				context.Background(),
				&cloudwatchlogs.DescribeLogGroupsInput{
					LogGroupNamePrefix: &c.LogGroup,
					NextToken:          queryToken})

			if err != nil {
				return err
			}
			queryToken = logGroupsOutput.NextToken

			for _, logGroup := range logGroupsOutput.LogGroups {
				if *(logGroup.LogGroupName) == c.LogGroup {
					c.Log.Debugf("Found log group %q", c.LogGroup)
					c.lg = &logGroup //nolint:revive
				}
			}
		}

		if c.lg == nil {
			return fmt.Errorf("can't find log group %q", c.LogGroup)
		}

		lsSplitArray := strings.Split(c.LogStream, ":")
		if len(lsSplitArray) > 1 {
			if lsSplitArray[0] == "tag" || lsSplitArray[0] == "field" {
				c.lsKey = lsSplitArray[0]
				c.lsSource = lsSplitArray[1]
				c.Log.Debugf("Log stream: key %q, source %q...", c.lsKey, c.lsSource)
			}
		}

		if c.lsSource == "" {
			c.lsSource = c.LogStream
			c.Log.Debugf("Log stream %q...", c.lsSource)
		}

		c.ls = map[string]*logStreamContainer{}
	}

	return nil
}

// Close closes plugin connection with remote receiver
func (c *CloudWatchLogs) Close() error {
	return nil
}

// Write perform metrics write to receiver of metrics
func (c *CloudWatchLogs) Write(metrics []telegraf.Metric) error {
	minTime := time.Now()
	if c.lg.RetentionInDays != nil {
		minTime = minTime.Add(-time.Hour * 24 * time.Duration(*c.lg.RetentionInDays))
	} else {
		minTime = minTime.Add(-maxPastLogEventTimeOffset)
	}

	maxTime := time.Now().Add(maxFutureLogEventTimeOffset)

	for _, m := range metrics {
		//Filtering metrics
		if m.Name() != c.LDMetricName {
			continue
		}

		if m.Time().After(maxTime) || m.Time().Before(minTime) {
			c.Log.Debugf("Processing metric '%v': Metric is filtered based on TS!", m)
			continue
		}

		tags := m.Tags()
		fields := m.Fields()

		logStream := ""
		logData := ""
		lsContainer := &logStreamContainer{
			currentBatchSizeBytes: 0,
			currentBatchIndex:     0,
			messageBatches: []messageBatch{
				{},
			},
		}

		switch c.lsKey {
		case "tag":
			logStream = tags[c.lsSource]
		case "field":
			if fields[c.lsSource] != nil {
				logStream = fields[c.lsSource].(string)
			}
		default:
			logStream = c.lsSource
		}

		if logStream == "" {
			c.Log.Errorf("Processing metric '%v': log stream: key %q, source %q, not found!", m, c.lsKey, c.lsSource)
			continue
		}

		switch c.logDatKey {
		case "tag":
			logData = tags[c.logDataSource]
		case "field":
			if fields[c.logDataSource] != nil {
				logData = fields[c.logDataSource].(string)
			}
		}

		if logData == "" {
			c.Log.Errorf("Processing metric '%v': log data: key %q, source %q, not found!", m, c.logDatKey, c.logDataSource)
			continue
		}

		//Check if message size is not fit to batch
		if len(logData) > maxLogMessageLength {
			metricStr := fmt.Sprintf("%v", m)
			c.Log.Errorf("Processing metric '%s...', message is too large to fit to aws max log message size: %d (bytes) !", metricStr[0:maxLogMessageLength/1000], maxLogMessageLength)
			continue
		}
		//Batching log messages
		//awsOverheadPerLogMessageBytes - is mandatory aws overhead per each log message
		messageSizeInBytesForAWS := len(logData) + awsOverheadPerLogMessageBytes

		//Pick up existing or prepare new log stream container.
		//Log stream container stores logs per log stream in
		//the AWS Cloudwatch logs API friendly structure
		if val, ok := c.ls[logStream]; ok {
			lsContainer = val
		} else {
			lsContainer.messageBatches[0].messageCount = 0
			lsContainer.messageBatches[0].logEvents = []types.InputLogEvent{}
			c.ls[logStream] = lsContainer
		}

		if lsContainer.currentBatchSizeBytes+messageSizeInBytesForAWS > maxBatchSizeBytes ||
			lsContainer.messageBatches[lsContainer.currentBatchIndex].messageCount >= maxItemsInBatch {
			//Need to start new batch, and reset counters
			lsContainer.currentBatchIndex++
			lsContainer.messageBatches = append(lsContainer.messageBatches,
				messageBatch{
					logEvents:    []types.InputLogEvent{},
					messageCount: 0})
			lsContainer.currentBatchSizeBytes = messageSizeInBytesForAWS
		} else {
			lsContainer.currentBatchSizeBytes += messageSizeInBytesForAWS
			lsContainer.messageBatches[lsContainer.currentBatchIndex].messageCount++
		}

		//AWS need time in milliseconds. time.UnixNano() returns time in nanoseconds since epoch
		//we store here TS with nanosec precision iun order to have proper ordering, later ts will be reduced to milliseconds
		metricTime := m.Time().UnixNano()
		//Adding metring to batch
		lsContainer.messageBatches[lsContainer.currentBatchIndex].logEvents =
			append(lsContainer.messageBatches[lsContainer.currentBatchIndex].logEvents,
				types.InputLogEvent{
					Message:   &logData,
					Timestamp: &metricTime})
	}

	// Sorting out log events by TS and sending them to cloud watch logs
	for logStream, elem := range c.ls {
		for index, batch := range elem.messageBatches {
			if len(batch.logEvents) == 0 { //can't push empty batch
				//c.Log.Warnf("Empty batch detected, skipping...")
				continue
			}
			//Sorting
			sort.Slice(batch.logEvents[:], func(i, j int) bool {
				return *batch.logEvents[i].Timestamp < *batch.logEvents[j].Timestamp
			})

			putLogEvents := cloudwatchlogs.PutLogEventsInput{LogGroupName: &c.LogGroup, LogStreamName: &logStream}
			if elem.sequenceToken == "" {
				//This is the first attempt to write to log stream,
				//need to check log stream existence and create it if necessary
				describeLogStreamOutput, err := c.svc.DescribeLogStreams(context.Background(), &cloudwatchlogs.DescribeLogStreamsInput{
					LogGroupName:        &c.LogGroup,
					LogStreamNamePrefix: &logStream})
				if err == nil && len(describeLogStreamOutput.LogStreams) == 0 {
					_, err := c.svc.CreateLogStream(context.Background(), &cloudwatchlogs.CreateLogStreamInput{
						LogGroupName:  &c.LogGroup,
						LogStreamName: &logStream})
					if err != nil {
						c.Log.Errorf("Can't create log stream %q in log group. Reason: %v %q.", logStream, c.LogGroup, err)
						continue
					}
					putLogEvents.SequenceToken = nil
				} else if err == nil && len(describeLogStreamOutput.LogStreams) == 1 {
					putLogEvents.SequenceToken = describeLogStreamOutput.LogStreams[0].UploadSequenceToken
				} else if err == nil && len(describeLogStreamOutput.LogStreams) > 1 { //Ambiguity
					c.Log.Errorf("More than 1 log stream found with prefix %q in log group %q.", logStream, c.LogGroup)
					continue
				} else {
					c.Log.Errorf("Error describing log streams in log group %q. Reason: %v", c.LogGroup, err)
					continue
				}
			} else {
				putLogEvents.SequenceToken = &elem.sequenceToken
			}

			//Upload log events
			//Adjusting TS to be in align with cloudwatch logs requirements
			for _, event := range batch.logEvents {
				*event.Timestamp = *event.Timestamp / 1000000
			}
			putLogEvents.LogEvents = batch.logEvents

			//There is a quota of 5 requests per second per log stream. Additional
			//requests are throttled. This quota can't be changed.
			putLogEventsOutput, err := c.svc.PutLogEvents(context.Background(), &putLogEvents)
			if err != nil {
				c.Log.Errorf("Can't push logs batch to AWS. Reason: %v", err)
				continue
			}
			//Cleanup batch
			elem.messageBatches[index] = messageBatch{
				logEvents:    []types.InputLogEvent{},
				messageCount: 0}

			elem.sequenceToken = *putLogEventsOutput.NextSequenceToken
		}
	}

	return nil
}

func init() {
	outputs.Add("cloudwatch_logs", func() telegraf.Output {
		return &CloudWatchLogs{}
	})
}
