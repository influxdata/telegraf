# Amazon Kinesis Output Plugin

This is an experimental plugin that is still in the early stages of
development. It will batch up all of the Points in one Put request to
Kinesis. This should save the number of API requests by a considerable level.

## About Kinesis

This is not the place to document all of the various Kinesis terms however it
maybe useful for users to review Amazons official documentation which is
available
[here](http://docs.aws.amazon.com/kinesis/latest/dev/key-concepts.html).

## Amazon Authentication

This plugin uses a credential chain for Authentication with the Kinesis API
endpoint. In the following order the plugin will attempt to authenticate.

1. Web identity provider credentials via STS if `role_arn` and
   `web_identity_token_file` are specified
1. Assumed credentials via STS if `role_arn` attribute is specified (source
   credentials are evaluated from subsequent rules)
1. Explicit credentials from `access_key`, `secret_key`, and `token` attributes
1. Shared profile from `profile` attribute
1. [Environment Variables][1]
1. [Shared Credentials][2]
1. [EC2 Instance Profile][3]

If you are using credentials from a web identity provider, you can specify the
session name using `role_session_name`. If left empty, the current timestamp
will be used.

[1]: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables
[2]: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file
[3]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html

## Configuration

```toml
# Configuration for the AWS Kinesis output.
[[outputs.kinesis]]
  ## Amazon REGION of kinesis endpoint.
  region = "ap-southeast-2"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #web_identity_token_file = ""
  #role_session_name = ""
  #profile = ""
  #shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Kinesis StreamName must exist prior to starting telegraf.
  streamname = "StreamName"

  ## The partition key can be calculated using one of several methods:
  ##
  ## Use a static value for all writes:
  #  [outputs.kinesis.partition]
  #    method = "static"
  #    key = "howdy"
  #
  ## Use a random partition key on each write:
  #  [outputs.kinesis.partition]
  #    method = "random"
  #
  ## Use the measurement name as the partition key:
  #  [outputs.kinesis.partition]
  #    method = "measurement"
  #
  ## Use the value of a tag for all writes, if the tag is not set the empty
  ## default option will be used. When no default, defaults to "telegraf"
  #  [outputs.kinesis.partition]
  #    method = "tag"
  #    key = "host"
  #    default = "mykey"


  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## debug will show upstream aws messages.
  debug = false
```

For this output plugin to function correctly the following variables must be
configured.

* region
* streamname

### region

The region is the Amazon region that you wish to connect to. Examples include
but are not limited to

* us-west-1
* us-west-2
* us-east-1
* ap-southeast-1
* ap-southeast-2

### streamname

The streamname is used by the plugin to ensure that data is sent to the correct
Kinesis stream. It is important to note that the stream *MUST* be pre-configured
for this plugin to function correctly. If the stream does not exist the plugin
will result in telegraf exiting with an exit code of 1.

### partitionkey [DEPRECATED]

This is used to group data within a stream. Currently this plugin only supports
a single partitionkey.  Manually configuring different hosts, or groups of hosts
with manually selected partitionkeys might be a workable solution to scale out.

### use_random_partitionkey [DEPRECATED]

When true a random UUID will be generated and used as the partitionkey when
sending data to Kinesis. This allows data to evenly spread across multiple
shards in the stream. Due to using a random partitionKey there can be no
guarantee of ordering when consuming the data off the shards.  If true then the
partitionkey option will be ignored.

### partition

This is used to group data within a stream. Currently four methods are
supported: random, static, tag or measurement

#### random

This will generate a UUIDv4 for each metric to spread them across shards.  Any
guarantee of ordering is lost with this method

#### static

This uses a static string as a partitionkey.  All metrics will be mapped to the
same shard which may limit throughput.

#### tag

This will take the value of the specified tag from each metric as the
partitionKey.  If the tag is not found the `default` value will be used or
`telegraf` if unspecified

#### measurement

This will use the measurement's name as the partitionKey.

### format

The format configuration value has been designated to allow people to change the
format of the Point as written to Kinesis. Right now there are two supported
formats string and custom.

#### string

String is defined using the default Point.String() value and translated to
[]byte for the Kinesis stream.

#### custom

Custom is a string defined by a number of values in the FormatMetric() function.
