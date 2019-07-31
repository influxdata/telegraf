## Amazon Kinesis Output for Telegraf

This is an experimental plugin that is still in the early stages of development. It will batch up all of the Points
in one Put request to Kinesis. This should save the number of API requests by a considerable level.

## About Kinesis

This is not the place to document all of the various Kinesis terms however it
maybe useful for users to review Amazons official documentation which is available
[here](http://docs.aws.amazon.com/kinesis/latest/dev/key-concepts.html).

## Amazon Authentication

This plugin uses a credential chain for Authentication with the Kinesis API endpoint. In the following order the plugin
will attempt to authenticate.
1. Assumed credentials via STS if `role_arn` attribute is specified (source credentials are evaluated from subsequent rules)
2. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
3. Shared profile from `profile` attribute
4. [Environment Variables](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables)
5. [Shared Credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)
6. [EC2 Instance Profile](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)


## Config

For this output plugin to function correctly the following variables must be configured.

* region
* streamname

### region

The region is the Amazon region that you wish to connect to. Examples include but are not limited to
* us-west-1
* us-west-2
* us-east-1
* ap-southeast-1
* ap-southeast-2

### streamname

The streamname is used by the plugin to ensure that data is sent to the correct Kinesis stream. It is important to
note that the stream *MUST* be pre-configured for this plugin to function correctly. If the stream does not exist the
plugin will result in telegraf exiting with an exit code of 1.

### partitionkey [DEPRECATED]

This is used to group data within a stream. Currently this plugin only supports a single partitionkey.
Manually configuring different hosts, or groups of hosts with manually selected partitionkeys might be a workable
solution to scale out.

### use_random_partitionkey [DEPRECATED]

When true a random UUID will be generated and used as the partitionkey when sending data to Kinesis. This allows data to evenly spread across multiple shards in the stream. Due to using a random paritionKey there can be no guarantee of ordering when consuming the data off the shards.
If true then the partitionkey option will be ignored.

### partition

This is used to group data within a stream. Currently four methods are supported: random, static, tag or measurement

#### random

This will generate a UUIDv4 for each metric to spread them across shards.
Any guarantee of ordering is lost with this method

#### static

This uses a static string as a partitionkey.
All metrics will be mapped to the same shard which may limit throughput.

#### tag

This will take the value of the specified tag from each metric as the paritionKey.
If the tag is not found the `default` value will be used or `telegraf` if unspecified

#### measurement

This will use the measurement's name as the partitionKey.

### format

The format configuration value has been designated to allow people to change the format of the Point as written to
Kinesis. Right now there are two supported formats string and custom.

#### string

String is defined using the default Point.String() value and translated to []byte for the Kinesis stream.

#### custom

Custom is a string defined by a number of values in the FormatMetric() function.
