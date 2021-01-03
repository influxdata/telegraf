package cloudwatch_logs

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type messageBatch struct {
	logEvents    []*cloudwatchlogs.InputLogEvent
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
	DescribeLogGroups(input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreams(input *cloudwatchlogs.DescribeLogStreamsInput) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
	CreateLogStream(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEvents(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}

type CloudWatchLogs struct {
	Region      string `toml:"region"`
	AccessKey   string `toml:"access_key"`
	SecretKey   string `toml:"secret_key"`
	RoleARN     string `toml:"role_arn"`
	Profile     string `toml:"profile"`
	Filename    string `toml:"shared_credential_file"`
	Token       string `toml:"token"`
	EndpointURL string `toml:"endpoint_url"`

	LogGroup string                   `toml:"log_group"`
	lg       *cloudwatchlogs.LogGroup //log group data

	LogStream string                         `toml:"log_stream"`
	lsKey     string                         //log stream source: tag or field
	lsSource  string                         //log stream source tag or field name
	ls        map[string]*logStreamContainer //log stream info

	LDMetricName string `toml:"log_data_metric_name"`

	LDSource string `toml:"log_data_source"`
	ldKey    string //log data source (tag or field)
	ldSource string //log data source tag or field name

	svc cloudWatchLogs //cloudwatch logs service
}

const (
	logId = "[outputs.cloudwatch_logs]"
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

var sampleConfig = `
## The region is the Amazon region that you wish to connect to.
## Examples include but are not limited to:
## - us-west-1
## - us-west-2
## - us-east-1
## - ap-southeast-1
## - ap-southeast-2
## ...
region = "us-east-1"

## Amazon Credentials
## Credentials are loaded in the following order
## 1) Assumed credentials via STS if role_arn is specified
## 2) explicit credentials from 'access_key' and 'secret_key'
## 3) shared profile from 'profile'
## 4) environment variables
## 5) shared credentials file
## 6) EC2 Instance Profile
#access_key = ""
#secret_key = ""
#token = ""
#role_arn = ""
#profile = ""
#shared_credential_file = ""

## Endpoint to make request against, the correct endpoint is automatically
## determined and this option should only be set if you wish to override the
## default.
##   ex: endpoint_url = "http://localhost:8000"
# endpoint_url = ""

## Cloud watch log group. Must be created in AWS cloudwatch logs upfront!
## For example, you can specify the name of the k8s cluster here to group logs from all cluster in oine place
log_group = "my-group-name" 

## Log stream in log group
## Either log group name or reference to metric attribute, from which it can be parsed:
## tag:<TAG_NAME> or field:<FIELD_NAME>. If log stream is not exist, it will be created.
## Since AWS is not automatically delete logs streams with expired logs entries (i.e. empty log stream) 
## you need to put in place appropriate house-keeping (https://forums.aws.amazon.com/thread.jspa?threadID=178855)
log_stream = "tag:location"

## Source of log data - metric name
## specify the name of the metric, from which the log data should be retrieved.
## I.e., if you  are using docker_log plugin to stream logs from container, then
## specify log_data_metric_name  = "docker_log"
log_data_metric_name  = "docker_log"

## Specify from which metric attribute the log data should be retrieved:
## tag:<TAG_NAME> or field:<FIELD_NAME>.
## I.e., if you  are using docker_log plugin to stream logs from container, then
## specify log_data_source  = "field:message" 
log_data_source  = "field:message"
`

func (c *CloudWatchLogs) SampleConfig() string {
	return sampleConfig
}

func (c *CloudWatchLogs) Description() string {
	return "Configuration for AWS CloudWatchLogs output."
}

func (c *CloudWatchLogs) Connect() error {
	var queryToken *string
	var dummyToken = "dummy"
	var logGroupsOutput = &cloudwatchlogs.DescribeLogGroupsOutput{NextToken: &dummyToken}
	var err error

	if c.LogGroup == "" {
		return fmt.Errorf("Log group is not set!")
	}

	if c.LogStream == "" {
		return fmt.Errorf("Log stream is not set!")
	}

	if c.LDMetricName == "" {
		return fmt.Errorf("Log data metrics name is not set!")
	}

	if c.LDSource == "" {
		return fmt.Errorf("Log data source is not set!")
	} else {
		lsSplitArray := strings.Split(c.LDSource, ":")
		if len(lsSplitArray) > 1 {
			if lsSplitArray[0] == "tag" || lsSplitArray[0] == "field" {
				c.ldKey = lsSplitArray[0]
				c.ldSource = lsSplitArray[1]
				log.Printf("D! %s Log data: key '%s', source '%s'...", logId, c.ldKey, c.ldSource)
			} else {
				return fmt.Errorf("Log data source is not properly formatted.\n" +
					"Should be 'tag:<tag_mame>' or 'field:<field_name>'")
			}
		} else {
			return fmt.Errorf("Log data source is not properly formatted, ':' is missed.\n" +
				"Should be 'tag:<tag_mame>' or 'field:<field_name>'")
		}

		if c.lsSource == "" {
			c.lsSource = c.LogStream
			log.Printf("D! %s Log stream '%s'...", logId, c.lsSource)
		}
	}

	credentialConfig := &internalaws.CredentialConfig{
		Region:      c.Region,
		AccessKey:   c.AccessKey,
		SecretKey:   c.SecretKey,
		RoleARN:     c.RoleARN,
		Profile:     c.Profile,
		Filename:    c.Filename,
		Token:       c.Token,
		EndpointURL: c.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()

	c.svc = cloudwatchlogs.New(configProvider)
	if c.svc == nil {
		return fmt.Errorf("Can't create cloudwatch logs service endpoint...")
	}

	//Find log group with name 'c.LogGroup'
	if c.lg == nil { //In case connection is not retried, first time
		for logGroupsOutput.NextToken != nil {

			logGroupsOutput, err = c.svc.DescribeLogGroups(
				&cloudwatchlogs.DescribeLogGroupsInput{
					LogGroupNamePrefix: &c.LogGroup,
					NextToken:          queryToken})

			if err != nil {
				return err
			}
			queryToken = logGroupsOutput.NextToken

			for _, logGroup := range logGroupsOutput.LogGroups {
				if *(logGroup.LogGroupName) == c.LogGroup {
					log.Printf("D! %s Found log group '%s'", logId, c.LogGroup)
					c.lg = logGroup
				}
			}
		}

		if c.lg == nil {
			return fmt.Errorf("Can't find log group '%s'...", c.LogGroup)
		}

		lsSplitArray := strings.Split(c.LogStream, ":")
		if len(lsSplitArray) > 1 {
			if lsSplitArray[0] == "tag" || lsSplitArray[0] == "field" {
				c.lsKey = lsSplitArray[0]
				c.lsSource = lsSplitArray[1]
				log.Printf("D! %s Log stream: key '%s', source '%s'...", logId, c.lsKey, c.lsSource)
			}
		}

		if c.lsSource == "" {
			c.lsSource = c.LogStream
			log.Printf("D! %s Log stream '%s'...", logId, c.lsSource)
		}

		c.ls = map[string]*logStreamContainer{}

	}

	return nil
}

func (c *CloudWatchLogs) Close() error {
	return nil
}

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
			log.Printf("D! %s Processing metric '%v': Metric is filtered based on TS!", logId, m)
			continue
		}

		tags := m.Tags()
		fields := m.Fields()

		logStream := ""
		logData := ""
		lsContainer := &logStreamContainer{
			currentBatchSizeBytes: 0,
			currentBatchIndex:     0,
			messageBatches:        []messageBatch{{}}}

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
			log.Printf("E! %s Processing metric '%v': log stream: key '%s', source '%s', not found!", logId, m, c.lsKey, c.lsSource)
			continue
		}

		switch c.ldKey {
		case "tag":
			logData = tags[c.ldSource]
		case "field":
			if fields[c.ldSource] != nil {
				logData = fields[c.ldSource].(string)
			}
		}

		if logData == "" {
			log.Printf("E! %s Processing metric '%v': log data: key '%s', source '%s', not found!", logId, m, c.ldKey, c.ldSource)
			continue
		}

		//Check if message size is not fit to batch
		if len(logData) > maxLogMessageLength {
			metricStr := fmt.Sprintf("%v", m)
			log.Printf("E! %s Processing metric '%s...', message is too large to fit to aws max log message size: %d (bytes) !", logId, metricStr[0:maxLogMessageLength/1000], maxLogMessageLength)
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
			lsContainer.messageBatches[0].logEvents = []*cloudwatchlogs.InputLogEvent{}
			c.ls[logStream] = lsContainer
		}

		if lsContainer.currentBatchSizeBytes+messageSizeInBytesForAWS > maxBatchSizeBytes ||
			lsContainer.messageBatches[lsContainer.currentBatchIndex].messageCount >= maxItemsInBatch {
			//Need to start new batch, and reset counters
			lsContainer.currentBatchIndex += 1
			lsContainer.messageBatches = append(lsContainer.messageBatches,
				messageBatch{
					logEvents:    []*cloudwatchlogs.InputLogEvent{},
					messageCount: 0})
			lsContainer.currentBatchSizeBytes = messageSizeInBytesForAWS
		} else {
			lsContainer.currentBatchSizeBytes += messageSizeInBytesForAWS
			lsContainer.messageBatches[lsContainer.currentBatchIndex].messageCount += 1
		}

		//AWS need time in milliseconds. time.UnixNano() returns time in nanoseconds since epoch
		//we store here TS with nanosec precision iun order to have proper ordering, later ts will be reduced to milliseconds
		metricTime := m.Time().UnixNano()
		//Adding metring to batch
		lsContainer.messageBatches[lsContainer.currentBatchIndex].logEvents =
			append(lsContainer.messageBatches[lsContainer.currentBatchIndex].logEvents,
				&cloudwatchlogs.InputLogEvent{
					Message:   &logData,
					Timestamp: &metricTime})
	}

	// Sorting out log events by TS and sending them to cloud watch logs
	for logStream, elem := range c.ls {
		for index, batch := range elem.messageBatches {
			if len(batch.logEvents) == 0 { //can't push empty batch
				//log.Printf("W! Empty batch detected, skipping...")
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
				describeLogStreamOutput, err := c.svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
					LogGroupName:        &c.LogGroup,
					LogStreamNamePrefix: &logStream})
				if err == nil && len(describeLogStreamOutput.LogStreams) == 0 {
					_, err := c.svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
						LogGroupName:  &c.LogGroup,
						LogStreamName: &logStream})
					if err != nil {
						log.Printf("E! %s Can't create log stream '%s' in log group. Reason: %v '%s'.", logId, logStream, c.LogGroup, err)
						continue
					}
					putLogEvents.SequenceToken = nil
				} else if err == nil && len(describeLogStreamOutput.LogStreams) == 1 {
					putLogEvents.SequenceToken = describeLogStreamOutput.LogStreams[0].UploadSequenceToken
				} else if err == nil && len(describeLogStreamOutput.LogStreams) > 1 { //Ambiguity
					log.Printf("E! %s More than 1 log stream found with prefix '%s' in log group '%s'.", logId, logStream, c.LogGroup)
					continue
				} else {
					log.Printf("E! %s Error describing log streams in log group '%s'. Reason: %v", logId, c.LogGroup, err)
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
			putLogEventsOutput, err := c.svc.PutLogEvents(&putLogEvents)
			if err != nil {
				log.Printf("E! %s Can't push logs batch to AWS. Reason: %v", logId, err)
				continue
			}
			//Cleanup batch
			elem.messageBatches[index] = messageBatch{
				logEvents:    []*cloudwatchlogs.InputLogEvent{},
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
