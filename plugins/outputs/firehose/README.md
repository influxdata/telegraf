## Amazon Kinesis Firehose Output for Telegraf

This is an experimental plugin that is still in the early stages of development. It will submit Points in batches up to 500
in one Put request to Kinesis Firehose. This should reduce the number of API requests considerably.

## About Kinesis Firehose

This is not the place to document all of the various Kinesis Firehose terms however it
maybe useful for users to review Amazons official documentation which is available
[here](http://docs.aws.amazon.com/firehose/latest/dev/what-is-this-service.html).

## Amazon Authentication

This plugin uses a credential chain for Authentication with the Firehose API endpoint. In the following order the plugin
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
* delivery_stream_name

### region

The region is the Amazon region that you wish to connect to. Examples include but are not limited to
* us-west-1
* us-west-2
* us-east-1
* ap-southeast-1
* ap-southeast-2

### delivery_stream_name

The delivery_stream_name config variable is used by the plugin to ensure that data is sent to the correct Kinesis Firehose delivery stream. 
It is important to note that the stream *MUST* be pre-configured for this plugin to function correctly.

### max_submit_attempts
The maximum number of times to attempt resubmitting a single metric, the sample config defautls to 10.

### data_format
Each data format has its own unique set of configuration options, read
more about them here:
https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
