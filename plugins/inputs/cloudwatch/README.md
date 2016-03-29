# Telegraf Plugin: Cloudwatch

## Configuration:

```
# Read metrics from Cloudwatch
[[inputs.cloudwatch]]
  ## AWS region
  region = "us-east-1"
  ## specify namespaces as strings
  namespaces = ["AWS/EC2", "AWS/DynamoDB"]
```

In each iteration, the cloudwatch input plugin queries Cloudwatch for
datapoints that are timestamped since the last iteration. (The first
iteration does not retrieve any data.)

Some AWS services post timepoints late, in the sense that new
timepoints are timestamped 2–3 minutes in the past. These points will
be missed by the cloudwatch input plugin if you run telegraf with a
short `interval`. Running telegraf with a longer interval, such as
`5m`, is recommended.


## Measurements:

Each namespace becomes a measurement, converted to snake_case and
prefixed with `cloudwatch`. For instance, the namespace `AWS/EC2`
becomes the measurement `cloudwatch_aws_ec2`, and the namespace
`AWS/DynamoDB` becomes the measurement `cloudwatch_aws_dynamo_db`.

### Fields

The cloudwatch input plugin retrieves all statistics from all
measurements in the specified namespaces. Each combination of metric
and statistic becomes a field. Metric names are converted to
snake_case. As an example, the metric `DiskWriteBytes` leads to the fields

* `disk_write_bytes_average`,
* `disk_write_bytes_maximum`,
* `disk_write_bytes_minimum`,
* `disk_write_bytes_sample_count`, and
* `disk_write_bytes_sum`.

To learn about the metrics for a particular Cloudwatch namespace, and
to see units and example data, see the AWS documentation for the
service in question, or explore using the AWS Console or AWS CLI.

### Tags

The cloudwatch input plugin tags the metrics it collects with the AWS
region (key: `region`) and the cloudwatch dimensions of the metric in
question. The dimensions names are converted to snake_keys, but the
values are left as is. As an example, the `AWS/DynamoDB` namespace has
a metric called `ThrottledRequest` with the dimensions `Operation` and
`TableName`. An example set of tags for this metric could be:

* `region`: `us-east-1`
* `operation`: `GetItem`
* `table_name`: `foo`

## Example output

```
$ ~/src/go/bin/telegraf -config ~/tmp/kapacitor/telegraf.conf -input-filter cloudwatch -debug
2016/03/29 10:47:15 Attempting connection to output: influxdb
2016/03/29 10:47:16 Successfully connected to output: influxdb
2016/03/29 10:47:16 Starting Telegraf (version 0.11.1-54-ge07c792)
2016/03/29 10:47:16 Loaded outputs: influxdb
2016/03/29 10:47:16 Loaded inputs: cloudwatch
2016/03/29 10:47:16 Tags enabled: host=vljosa-yieldbot.local
2016/03/29 10:47:16 Agent Config: Interval:5m0s, Debug:true, Quiet:false, Hostname:"vljosa-yieldbot.local", Flush Interval:11.228634593s 
2016/03/29 10:50:00 Found 26 cloudwatch metrics
2016/03/29 10:50:00 Gathered metrics, (5m0s interval), from 1 inputs in 709.03259ms
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,table_name=prd-platform-datomic-config consumed_write_capacity_units_average=1,consumed_write_capacity_units_maximum=1,consumed_write_capacity_units_minimum=1,consumed_write_capacity_units_sample_count=65,consumed_write_capacity_units_sum=65 1459263300548895733
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,table_name=dev-platform-datomic-config consumed_read_capacity_units_average=1,consumed_read_capacity_units_maximum=1,consumed_read_capacity_units_minimum=1,consumed_read_capacity_units_sample_count=48,consumed_read_capacity_units_sum=48 1459263300642785107
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,operation=PutItem,table_name=dev-platform-datomic-config successful_request_latency_average=8.910156727272726,successful_request_latency_maximum=26.477,successful_request_latency_minimum=4.243,successful_request_latency_sample_count=44,successful_request_latency_sum=392.04689599999995 1459263300688946255
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,operation=GetItem,table_name=dev-platform-datomic-config successful_request_latency_average=6.982394130434781,successful_request_latency_maximum=38.125,successful_request_latency_minimum=2.156,successful_request_latency_sample_count=46,successful_request_latency_sum=321.19012999999995 1459263300727174463
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,table_name=dev-platform-datomic-config consumed_write_capacity_units_average=1,consumed_write_capacity_units_maximum=1,consumed_write_capacity_units_minimum=1,consumed_write_capacity_units_sample_count=48,consumed_write_capacity_units_sum=48 1459263300760742003
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,table_name=prd-platform-datomic-config consumed_read_capacity_units_average=1.8539288112827401,consumed_read_capacity_units_maximum=8,consumed_read_capacity_units_minimum=0.5,consumed_read_capacity_units_sample_count=1489,consumed_read_capacity_units_sum=2760.5 1459263300832984286
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,operation=GetItem,table_name=prd-platform-datomic-config successful_request_latency_average=5.014870193569996,successful_request_latency_maximum=56.839959,successful_request_latency_minimum=2.030115,successful_request_latency_sample_count=1493,successful_request_latency_sum=7487.201199000004 1459263300868485699
> cloudwatch_aws_dynamo_db,host=vljosa-yieldbot.local,operation=PutItem,table_name=prd-platform-datomic-config successful_request_latency_average=9.664095158730158,successful_request_latency_maximum=25.262,successful_request_latency_minimum=3.533,successful_request_latency_sample_count=63,successful_request_latency_sum=608.837995 1459263300990375223
2016/03/29 10:55:01 Gathered metrics, (5m0s interval), from 1 inputs in 1.048629086s
2016/03/29 10:55:03 Wrote 8 metrics to output influxdb in 90.0269ms
```
