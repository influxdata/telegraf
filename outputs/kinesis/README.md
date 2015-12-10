## Amazon Kinesis Output for Telegraf

This is an experimental plugin that is still in the early stages of development. It will batch up all of the Points
in one Put request to Kinesis. This should save the number of API requests by a considerable level.

## About Kinesis

This is not the place to document all of the various Kinesis terms however it
maybe useful for users to review Amazons official docuementation which is available
[here](http://docs.aws.amazon.com/kinesis/latest/dev/key-concepts.html).

## Config

For this output plugin to function correctly the following variables must be configured.

* region
* streamname
* partitionkey

### region

The region is the Amazon region that you wish to connect to. Examples include but are not limited to
* us-west-1
* us-west-2
* us-east-1
* ap-southeast-1
* ap-southeast-2

### streamname

The streamname is used by the plugin to ensure that data is sent to the correct Kinesis stream. It is important to
note that the stream *MUST* be pre-configured for this plugin to function correctly.

### partitionkey

This is used to group data within a stream. Currently this plugin only supports a single partitionkey which means
that data will be entirely sent through a single Shard and Partition from a single host. If you have to scale out the
kinesis throughput using a different partition key on different hosts or host groups might be a workable solution.


## todo

Check if the stream exists so that we have a graceful exit.