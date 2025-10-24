# Arc Output Plugin

This plugin writes metrics to [Arc](https://github.com/basekick-labs/arc),
a high-performance time-series database, using the MessagePack binary protocol.

Arc's MessagePack protocol provides **3-5x better performance** than traditional
line protocol formats through binary serialization and direct Arrow/Parquet
writes.

⭐ Telegraf v1.32.0
🏷️ datastore
💻 all

## Global configuration options

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Arc Time-Series Database Output Plugin
# High-performance MessagePack binary protocol (3-5x faster than line protocol)
[[outputs.arc]]
  ## Arc MessagePack API URL
  ## The endpoint where Arc is listening for MessagePack writes
  url = "http://localhost:8000/api/v1/write/msgpack"

  ## Timeout for HTTP writes
  # timeout = "5s"

  ## API Key for authentication
  ## Generate an API key in Arc with write permissions
  ## Example: python3 cli.py auth create-token --name telegraf --permissions write
  # api_key = "your-arc-api-key-here"

  ## Database name for multi-database architecture (optional)
  ## Routes metrics to a specific database namespace
  ## Examples: "production", "staging", "development", "default"
  ## If not specified, Arc uses the default database configured in arc.conf
  # database = "production"

  ## HTTP User-Agent
  # user_agent = "Telegraf-Arc-Output-Plugin"

  ## Content encoding for request body
  ## Options: "gzip" (default, recommended), "identity" (no compression)
  ## Gzip compression reduces network bandwidth by ~10x
  # content_encoding = "gzip"

  ## Batch size for MessagePack writes
  ## Higher values improve throughput but use more memory
  # batch_size = 1000

  ## Additional HTTP headers
  # [outputs.arc.headers]
  #   X-Custom-Header = "custom-value"

  ## Arc uses MessagePack binary protocol with columnar format for maximum performance
  ## Columnar format (2.66x faster than row format):
  ##   {
  ##     "m": "cpu",              # measurement name
  ##     "columns": {             # all data organized as columns (arrays)
  ##       "time": [1633024800000, 1633024801000, 1633024802000],
  ##       "host": ["server01", "server02", "server03"],
  ##       "region": ["us-east", "us-west", "eu-central"],
  ##       "usage_idle": [95.0, 85.0, 92.0],
  ##       "usage_user": [3.2, 10.5, 5.8]
  ##     }
  ##   }
```

## Performance

Arc's MessagePack binary protocol with columnar format provides exceptional
write performance:

- **Columnar Format:** 2.66x faster than row format (2.42M records/sec throughput)
- **Compression:** Gzip reduces bandwidth by ~10x with minimal CPU
  overhead
- **Batching:** Default batch size of 1000 provides optimal throughput
- **Efficiency:** Columnar format enables direct Arrow/Parquet writes for
  maximum performance

### Performance Tuning

For maximum throughput:

1. Use `content_encoding = "gzip"` (default) to reduce network
   bandwidth
2. Increase `batch_size` to 5000-10000 for high-volume scenarios
3. Adjust Telegraf's `flush_interval` to match your latency
   requirements
4. Use multiple Telegraf instances for >1M RPS workloads

## Multi-Database Support

Arc supports multiple databases (namespaces) within a single instance, allowing
you to organize and isolate metrics by environment, tenant, or application.

### Use Cases

1. **Environment Separation**: Route production, staging, and development
   metrics to separate databases
2. **Multi-Tenancy**: Isolate metrics for different customers or teams
3. **Data Lifecycle**: Separate hot, warm, and cold data storage

### Database Configuration

```toml
# Route metrics to a specific database using the database
# parameter
[[outputs.arc]]
  url = "http://arc:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
  database = "production"
```

If no database is specified, metrics are written to the default database
configured in Arc's `arc.conf`.

### Cross-Database Queries

Arc allows querying across databases using SQL:

```sql
-- Query specific database
SELECT * FROM production.cpu WHERE time > NOW() - INTERVAL 1 HOUR

-- Compare production vs staging
SELECT p.time, p.usage as prod_usage, s.usage as staging_usage
FROM production.cpu p
JOIN staging.cpu s ON p.time = s.time AND p.host = s.host
```

## Authentication

Arc uses API key authentication via the `x-api-key` header. Generate a token
with write permissions:

```bash
# Using Arc CLI
python3 cli.py auth create-token --name telegraf --permissions write
```

Add the generated token to your Telegraf configuration:

```toml
[[outputs.arc]]
  url = "http://localhost:8000/api/v1/write/msgpack"
  api_key = "your-generated-token-here"
```

## MessagePack Format

Arc uses a columnar MessagePack binary format optimized for time-series data.
The columnar format provides 2.66x better performance than traditional
row-based formats by organizing data as arrays instead of individual records.

### Columnar Format (Recommended)

All data is organized as columns (arrays), not rows:

```json
{
  "m": "cpu",
  "columns": {
    "time": [1633024800000, 1633024801000, 1633024802000],
    "host": ["server01", "server02", "server03"],
    "region": ["us-east", "us-west", "eu-central"],
    "datacenter": ["aws", "gcp", "azure"],
    "usage_idle": [95.0, 85.0, 92.0],
    "usage_user": [3.2, 10.5, 5.8],
    "usage_system": [1.8, 4.5, 2.2]
  }
}
```

### How It Works

1. **Grouping:** Metrics are automatically grouped by measurement name
2. **Column Creation:** Each field and tag becomes a column (array)
3. **Alignment:** All columns have the same length, with values aligned by
   index
4. **Performance:** Enables direct Arrow/Parquet writes for 2.66x faster
   throughput

### Multiple Measurements

When sending metrics from multiple measurements, the plugin sends an array of
columnar data structures:

```json
[
  {
    "m": "cpu",
    "columns": {"time": [...], "usage_idle": [...], ...}
  },
  {
    "m": "mem",
    "columns": {"time": [...], "usage_percent": [...], ...}
  }
]
```

## Example Configuration

### Basic Configuration

```toml
[[outputs.arc]]
  url = "http://localhost:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
```

### High-Performance Configuration

```toml
[[outputs.arc]]
  url = "http://arc-production:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
  timeout = "10s"
  content_encoding = "gzip"
  batch_size = 5000

  [outputs.arc.headers]
    X-Environment = "production"
```

### Multi-Database Configuration

```toml
# Route metrics to different databases based on environment
[[outputs.arc]]
  url = "http://arc:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
  database = "production"  # Production metrics

[[outputs.arc]]
  url = "http://arc:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
  database = "staging"     # Staging metrics

  # Optional: Use filters to route specific metrics
  namepass = ["cpu", "mem", "disk"]
```

### Load-Balanced Configuration

```toml
# Use multiple Arc instances for >1M RPS
[[outputs.arc]]
  url = "http://arc-01:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
  database = "production"

[[outputs.arc]]
  url = "http://arc-02:8000/api/v1/write/msgpack"
  api_key = "$ARC_API_KEY"
  database = "production"
```

## Metrics

The Arc output plugin does not produce any metrics.

## Troubleshooting

### Connection Issues

1. Verify Arc is running: `curl http://localhost:8000/health`
2. Test authentication:
   `curl -H "x-api-key: YOUR_KEY" http://localhost:8000/health`
3. Check Telegraf logs: `telegraf --config telegraf.conf --debug`

### Performance Issues

1. Enable gzip compression if not already enabled
2. Increase batch_size for higher throughput
3. Check network latency between Telegraf and Arc
4. Monitor Arc's `/metrics` endpoint for bottlenecks

### Authentication Errors

```text
Error: arc returned status 401: Unauthorized
```

Solution: Generate a valid API key and add it to your configuration:

```bash
python3 cli.py auth create-token --name telegraf --permissions write
```

## See Also

- [Arc GitHub Repository](https://github.com/basekick-labs/arc)
- [Arc Documentation](https://docs.basekick.net/arc)
- [ClickBench Results](https://benchmark.clickhouse.com) - Arc ranks #3
  on analytical queries
