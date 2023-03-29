# CloudWatch Metric Streams Input Plugin

The CloudWatch Metric Streams plugin is a service input plugin that listens
for metrics sent via HTTP and performs the required processing for
[Metric Streams from AWS](#troubleshooting-documentation).

For cost, see the Metric Streams example in
[CloudWatch pricing](#troubleshooting-documentation).

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# AWS Metric Streams listener
[[inputs.cloudwatch_metric_streams]]
  ## Address and port to host HTTP listener on
  service_address = ":443"

  ## Paths to listen to.
  # paths = ["/telegraf"]

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"

  ## maximum duration before timing out write of the response
  # write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 524,288,000 bytes (500 mebibytes)
  # max_body_size = "500MB"

  ## Optional access key for Firehose security.
  # access_key = "test-key"

  ## An optional flag to keep Metric Streams metrics compatible with
  ## CloudWatch's API naming
  # api_compatability = false

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
```

## Metrics

Metrics sent by AWS are Base64 encoded blocks of JSON data.
The JSON block below is the Base64 decoded data in the `data`
field of a `record`.
There can be multiple blocks of JSON for each `data` field
in each `record` and there can be multiple `record` fields in
a `record`.

The metric when decoded may look like this:

```json
{
    "metric_stream_name": "sandbox-dev-cloudwatch-metric-stream",
    "account_id": "541737779709",
    "region": "us-west-2",
    "namespace": "AWS/EC2",
    "metric_name": "CPUUtilization",
    "dimensions": {
        "InstanceId": "i-0efc7ghy09c123428"
    },
    "timestamp": 1651679580000,
    "value": {
        "max": 10.011666666666667,
        "min": 10.011666666666667,
        "sum": 10.011666666666667,
        "count": 1
    },
    "unit": "Percent"
}
```

### Tags

All tags in the `dimensions` list are added as tags to the metric.

The `account_id` and `region` tag are added to each metric as well.

### Measurements and Fields

The metric name is a combination of `namespace` and `metric_name`,
separated by `_` and lowercased.

The fields are each aggregate in the `value` list.

These fields are optionally renamed to match the CloudWatch API for
easier transition from the API to Metric Streams. This relies on
setting the `api_compatability` flag in the configuration.

The timestamp applied is the timestamp from the metric,
typically 3-5 minutes older than the time processed due
to CloudWatch delays.

## Example Output

Example output based on the above JSON & compatability flag is:

**Standard Metric Streams format:**

```text
aws_ec2_cpuutilization,accountId=541737779709,region=us-west-2,InstanceId=i-0efc7ghy09c123428 max=10.011666666666667,min=10.011666666666667,sum=10.011666666666667,count=1 1651679580000
```

**API Compatability format:**

```text
aws_ec2_cpuutilization,accountId=541737779709,region=us-west-2,InstanceId=i-0efc7ghy09c123428 maximum=10.011666666666667,minimum=10.011666666666667,sum=10.011666666666667,samplecount=1 1651679580000
```

## Troubleshooting

The plugin has its own internal metrics for troubleshooting:

* Requests Received
  * The number of requests received by the listener.
* Writes Served
  * The number of writes served by the listener.
* Bad Requests
  * The number of bad requests, separated by the error code as a tag.
* Request Time
  * The duration of the request measured in ns.
* Age Max
  * The maximum age of a metric in this interval. This is useful for offsetting
    any lag or latency measurements in a metrics pipeline that measures based
    on the timestamp.
* Age Min
  * The minimum age of a metric in this interval.

Specific errors will be logged and an error will be returned to AWS.

### Troubleshooting Documentation

Additional troubleshooting for a Metric Stream can be found
in AWS's documentation:

* [CloudWatch Metric Streams](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-Metric-Streams.html)
* [AWS HTTP Specifications](https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html)
* [Firehose Troubleshooting](https://docs.aws.amazon.com/firehose/latest/dev/http_troubleshooting.html)
* [CloudWatch Pricing](https://aws.amazon.com/cloudwatch/pricing/)
