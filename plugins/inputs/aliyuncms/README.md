# Aliyun CloudMonitor Service Statistics Input

This plugin will pull Metric Statistics from Aliyun CMS.

### Aliyun Authentication

This plugin uses an [AccessKey](https://www.alibabacloud.com/help/doc-detail/53045.htm?spm=a2c63.p38356.b99.127.5cba21fdt5MJKr&parentId=28572) credential for Authentication with the Aliyun OpenAPI endpoint.
In the following order the plugin will attempt to authenticate.
1. Ram RoleARN credential if `access_key_id`, `access_key_secret`, `role_arn`, `role_session_name` is specified
2. AccessKey STS token credential if `access_key_id`, `access_key_secret`, `access_key_sts_token` is specified
3. AccessKey credentail if `access_key_id`, `access_key_secret` is specified
4. Ecs Ram Role Credential if `role_name` is specified
5. RSA keypair credential if `private_key`, `public_key_id` is specified
6. Environment variables credential
7. Instance metadata credential

### Configuration:

```toml
	## Aliyun Region
	region_id = "cn-hangzhou"
  
	## Aliyun Credentials
	## Credentials are loaded in the following order
	## 1) Ram RoleArn credential
	## 2) AccessKey STS token credential
	## 3) AccessKey credential
	## 4) Ecs Ram Role credential
	## 5) RSA keypair cendential
	## 6) Environment varialbes credential
	## 7) Instance metadata credential
	# access_key_id = ""
	# access_key_secret = ""
	# access_key_sts_token = ""
	# role_arn = ""
	# role_session_name = ""
	# private_key = ""
	# public_key_id = ""
	# role_name = ""

	# The minimum period for AliyunCMS metrics is 1 minute (60s). However not all
	# metrics are made available to the 1 minute period. Some are collected at
	# 3 minute, 5 minute, or larger intervals.
	# See: https://help.aliyun.com/document_detail/51936.html?spm=a2c4g.11186623.2.18.2bc1750eeOw1Pv
	# Note that if a period is configured that is smaller than the minimum for a
	# particular metric, that metric will not be returned by the Aliyun OpenAPI
	# and will not be collected by Telegraf.
	#
	## Requested AliyunCMS aggregation Period (required - must be a multiple of 60s)
	period = "5m"
  
	## Collection Delay (required - must account for metrics availability via AliyunCMS API)
	delay = "1m"
  
	## Recommended: use metric 'interval' that is a multiple of 'period' to avoid
	## gaps or overlap in pulled data
	interval = "5m"
  
	## Metric Statistic Project (required)
	project = "acs_slb_dashboard"

	## Maximum requests per second, default value is 200
	ratelimit = 200
  
	## Metrics to Pull (Required)
	## Defaults to all Metrics in Namespace if nothing is provided
	## Refreshes Namespace available metrics every 1h
	[[inputs.aliyuncms.metrics]]
	  names = ["InstanceActiveConnection", "InstanceNewConnection"]
	
	  ## Dimension filters for Metric.  These are optional however all dimensions
	  ## defined for the metric names must be specified in order to retrieve
	  ## the metric statistics.
	  [[inputs.aliyuncms.metrics.dimensions]]
		value = '{"instanceId": "p-example"}'

	  [[inputs.aliyuncms.metrics.dimensions]]
		value = '{"instanceId": "q-example"}'
```

#### Requirements and Terminology

Plugin Configuration utilizes [preset metric items references](https://www.alibabacloud.com/help/doc-detail/28619.htm?spm=a2c63.p38356.a3.2.389f233d0kPJn0)

- `region` must be a valid Aliyun [Region](https://www.alibabacloud.com/help/doc-detail/40654.htm) value
- `period` must be a valid duration value
- `project` must be a preset project value
- `names` must be preset metric names
- `dimensions` must be preset dimension values

Each of dimension values is a JSON string, for example, {"instanceId":"i-23gyb3kkd"}. If specifying >1 dimension, then the metric matches *any* of the configured dimensions.

Example:
```
[[inputs.aliyuncms.metrics]]
  names = ["Latency"]

  ## Dimension filters for Metric (required)
  [[inputs.aliyuncms.metrics.dimensions]]
	value = '{"instanceId": "p-example"}'

  [[inputs.aliyuncms.metrics.dimensions]]
	value = '{"instanceId": "q-example"}'
```

Then 2 metrics will be output:
- instanceId: `p-example`
- instanceId: `q-example`

### Measurements & Fields:

Each Aliyun CMS Project monitored records a measurement with fields for each available Metric Statistic
Project and Metrics are represented in [snake case](https://en.wikipedia.org/wiki/Snake_case)

- aliyuncms_{project}
  - {metric}_average     (metric Average value)
  - {metric}_minimum     (metric Minimum value)
  - {metric}_maximum     (metric Maximum value)
  - {metric}_value       (metric Value value)

### Tags:
Each measurement is tagged with the following identifiers to uniquely identify the associated metric
Tag Dimension names are represented in [snake case](https://en.wikipedia.org/wiki/Snake_case)

- All measurements have the following tags:
  - region           (Aliyun Region ID)
  - userId           (Aliyun Account ID)
  - {dimension-name} (Dimension value - one for each metric dimension)

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter aliyuncms --test
> aliyuncms_acs_slb_dashboard,instanceId=p-example,region=cn-hangzhou,userId=1234567890 latency_average=0.004810798017284538,latency_maximum=0.1100282669067383,latency_minimum=0.0006084442138671875
```
