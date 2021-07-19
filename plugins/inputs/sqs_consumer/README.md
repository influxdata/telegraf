# SQS Consumer input plugin

The `sqs_consumer` input plugin reads messages from AWS SQS queue and creates metrics
using one of the supported [input data formats](/docs/DATA_FORMATS_INPUT.md).

### Configuration

```toml
# Read metrics from AWS SQS
[[inputs.sqs_consumer]]
  ## Required. Amazon REGION of SQS endpoint.
  region = "ap-southeast-2"

  ## Optional. Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # profile = ""
  # shared_credential_file = ""

  ## Optional. Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Optional. Use together with queue_owner_account_id to discover queue url
  # queue_name = ""

  ## Optional. Use together with queue_name to discover queue url
  # queue_owner_account_id = ""

  ## Optional. Required if queue_name and queue_owner_account_id were not provided
  queue_url = ""

  ## Optional. The maximum number of messages to return. Defaults to 10
  # max_number_of_messages = 10

  ## Optional. The duration (in seconds) that the received messages are hidden
  ## from subsequent retrieve requests. If not set defaults to queue settings.
  # visibility_timeout = 30

  ## Optional. The duration (in seconds) for which the call waits for a message
  ## to arrive in the queue before returning. If set to higher then 0, enables long polling
  ## 0 enables short polling. Defaults to 20
  # wait_time_seconds = 20

  ## Optional. When > 1 messages will be deleted from a queue in batches of provided size.
  ## Defaults to 10
  # delete_batch_size = 10

  ## Optional. If batch delete is enabled - flush messages when no new messages were received for
  ## this period of time to avoid redeliveries. Defaults to 20
  # delete_batch_flush_seconds = 20

  ## Optional. Number of seconds to wait before attempting receiving messages
  ## after failed attempt. Defaults to 5
  # retry_receive_delay_seconds = 5

  ## Optional. Maximum byte length of a message to consume.
  ## Larger messages are dropped with an error. If less than 0 or unspecified,
  ## treated as no limit.
  # max_message_len = 1000000

  ## Optional. Maximum messages to read from the queue that have not been written by an
  ## output. Defaults to 1000.
  ## For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

#### Required AWS IAM permissions

Minimum required permissions for queue:
 - `sqs:ReceiveMessage`
 - `sqs:DeleteMessage`

If you are using `queue_name` and `queue_owner_account_id` fields to discover queue URL you will also need to add this permission:
 - `sqs:GetQueueUrl`

Example policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "sqs:DeleteMessage",
                "sqs:ReceiveMessage"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:sqs:us-east-1:12345678:queue_name"
        }
    ]
}
```

### Developing

Follow [input development doc](/docs/INPUTS.md#Development) to bootstrap local environment.

To send messages to local SQS queue use aws cli:

```
AWS_ACCESS_KEY_ID=x AWS_SECRET_ACCESS_KEY=x aws --region elasticmq --endpoint-url http://localhost:9324 sqs send-message --queue-url http://localhost:9324/queue/telegraf --message-body "system,host=tyrion uptime=1249632i 1483964144000000000"
```

To check queue status:

```
AWS_ACCESS_KEY_ID=x AWS_SECRET_ACCESS_KEY=x aws --region elasticmq --endpoint-url http://localhost:9324 sqs get-queue-attributes --queue-url http://localhost:9324/queue/telegraf --attribute-names "All"
```
