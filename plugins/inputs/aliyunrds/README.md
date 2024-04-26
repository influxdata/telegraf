# Alibaba Cloud RDS (Relational Database Service) Input Plugin

This plugin gathers performance metrics from [Alibaba Cloud RDS][rds] instances.
It uses the RDS Performance API to collect database performance metrics.

⭐ Telegraf v1.38.0
🏷️ cloud
💻 all

[rds]: https://www.alibabacloud.com/product/rds

## Aliyun Authentication

This plugin uses an [AccessKey][1] credential for Authentication with the
Aliyun OpenAPI endpoint. In the following order the plugin will attempt
to authenticate:

1. Ram RoleARN credential if `access_key_id`, `access_key_secret`, `role_arn`,
   `role_session_name` is specified
2. AccessKey STS token credential if `access_key_id`, `access_key_secret`,
   `access_key_sts_token` is specified
3. AccessKey credential if `access_key_id`, `access_key_secret` is specified
4. Ecs Ram Role Credential if `role_name` is specified
5. RSA keypair credential if `private_key`, `public_key_id` is specified
6. Environment variables credential
7. Instance metadata credential

[1]: https://www.alibabacloud.com/help/doc-detail/53045.htm

## Global configuration options

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Pull Metric Statistics from Aliyun RDS (Relational Database Service)
[[inputs.aliyunrds]]
  ## Aliyun Credentials
  ## Credentials are loaded in the following order
  ## 1) Ram RoleArn credential
  ## 2) AccessKey STS token credential
  ## 3) AccessKey credential
  ## 4) Ecs Ram Role credential
  ## 5) RSA keypair credential
  ## 6) Environment variables credential
  ## 7) Instance metadata credential

  # access_key_id = ""
  # access_key_secret = ""
  # access_key_sts_token = ""
  # role_arn = ""
  # role_session_name = ""
  # private_key = ""
  # public_key_id = ""
  # role_name = ""

  ## Specify ali cloud regions to be queried for metric and object discovery
  ## If not set, all supported regions (see below) would be covered, it can
  ## provide a significant load on API, so the recommendation here is to
  ## limit the list as much as possible.
  ## Allowed values: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
  ## Default supported regions are:
  ##   cn-qingdao,cn-beijing,cn-zhangjiakou,cn-huhehaote,cn-hangzhou,
  ##   cn-shanghai, cn-shenzhen, cn-heyuan,cn-chengdu,cn-hongkong,
  ##   ap-southeast-1,ap-southeast-2,ap-southeast-3,ap-southeast-5,
  ##   ap-south-1,ap-northeast-1, us-west-1,us-east-1,eu-central-1,
  ##   eu-west-1,me-east-1
  ##
  ## Discovery is automatically enabled for these regions to find RDS instances.
  regions = ["cn-hongkong"]

  ## Requested aggregation Period (required)
  ## The period must be multiples of 60s and the minimum for RDS metrics
  ## is 1 minute (60s). However not all metrics are made available to the
  ## one minute period. Some are collected at 5 minute or larger intervals.
  ## See: https://help.aliyun.com/document_detail/26316.html
  period = "5m"

  ## Collection Delay (required)
  ## The delay must account for metrics availability via RDS API.
  delay = "1m"

  ## Recommended: use metric 'interval' that is a multiple of 'period'
  ## to avoid gaps or overlap in pulled data
  interval = "5m"

  ## Maximum requests per second, default value is 200
  ratelimit = 200

  ## How often the discovery API call executed (default 1m)
  #discovery_interval = "1m"

  ## NOTE: Due to the way TOML is parsed, tables must be at the END of the
  ## plugin definition, otherwise additional config options are read as part of
  ## the table

  ## Metrics to Pull
  ## At least one metrics definition required
  [[inputs.aliyunrds.metrics]]
    ## Metric names to be requested from RDS Performance API
    ## Full list of available metrics can be found here:
    ## https://help.aliyun.com/document_detail/26316.html
    ## Common RDS metrics include:
    ##   - MySQL_Sessions (Active Sessions)
    ##   - MySQL_QPS (Queries Per Second)
    ##   - MySQL_TPS (Transactions Per Second)
    ##   - MySQL_IOPS (I/O Operations Per Second)
    ##   - MySQL_NetworkInNew (Network Input Traffic)
    ##   - MySQL_NetworkOutNew (Network Output Traffic)
    ##   - CpuUsage (CPU Usage)
    ##   - MemoryUsage (Memory Usage)
    ##   - DiskUsage (Disk Usage)
    ##   - ConnectionUsage (Connection Usage)
    names = ["MySQL_Sessions", "MySQL_QPS", "CpuUsage"]

    ## List of specific RDS instance IDs to monitor (optional)
    ## If not specified, all discovered instances in the configured regions will be monitored
    ## You can specify instance IDs to limit monitoring to specific instances:
    # instances = ["rm-xxxxx", "rm-yyyyy"]

    ## Discovery is enabled by default. All RDS instances in the configured regions
    ## will be discovered and monitored. Discovery data is used to enrich metrics
    ## with tags such as:
    ##   - RegionId
    ##   - DBInstanceType
    ##   - Engine
    ##   - EngineVersion
    ##   - DBInstanceDescription

```

## Metrics

The plugin collects performance metrics from Aliyun RDS. The specific metrics
available depend on the database engine (MySQL, PostgreSQL, SQL Server, etc.).

### Common MySQL Metrics

- `MySQL_Sessions` - Number of active sessions
- `MySQL_QPS` - Queries per second
- `MySQL_TPS` - Transactions per second
- `MySQL_IOPS` - I/O operations per second (includes read_iops and write_iops)
- `MySQL_NetworkInNew` - Network input traffic (bytes/second)
- `MySQL_NetworkOutNew` - Network output traffic (bytes/second)
- `CpuUsage` - CPU usage percentage
- `MemoryUsage` - Memory usage percentage
- `DiskUsage` - Disk usage percentage
- `ConnectionUsage` - Connection usage percentage

For a complete list of available metrics per database engine, see:
[https://help.aliyun.com/document_detail/26316.html](https://help.aliyun.com/document_detail/26316.html)

### Tags

All metrics include the following tags:

- `instanceId` - RDS instance ID
- `region` - Aliyun region

When discovery is enabled, additional tags are added:

- `RegionId` - Region ID from discovery
- `DBInstanceType` - Instance type (Primary, ReadOnly, Guard, Temp)
- `Engine` - Database engine (MySQL, PostgreSQL, SQLServer, etc.)
- `EngineVersion` - Database engine version
- `DBInstanceDescription` - Instance description/name

### Fields

Field names are formatted as `{metric_name}_{statistic}` in snake_case.
For example:

- `my_sql_sessions_sessions`
- `my_sql_qps_value`
- `cpu_usage_value`

## Example Output

```text
aliyunrds,DBInstanceDescription=Production\ MySQL,DBInstanceType=Primary,Engine=MySQL,EngineVersion=5.7,RegionId=cn-hangzhou,instanceId=rm-xxxxx,region=cn-hangzhou cpu_usage_value=45.2,my_sql_qps_value=100.5,my_sql_sessions_sessions=10 1704067200000000000
```

## Discovery

The plugin automatically discovers RDS instances in the configured regions.
Discovery runs periodically (default: every 1 minute) and updates the list of
monitored instances. Discovery can be disabled by explicitly listing instances
in the `instances` configuration parameter.

## Rate Limiting

The plugin implements rate limiting to avoid exceeding Aliyun API quotas. The
default rate limit is 200 requests per second. Discovery uses 20% of the
configured rate limit.

## Performance Considerations

- Limit the `regions` configuration to only the regions you need
- Use the `instances` parameter to monitor specific instances instead of all
  discovered instances
- Adjust the `period` and `interval` to balance between metric granularity and
  API usage
- The `discovery_interval` can be increased to reduce API calls if your
  infrastructure changes infrequently
