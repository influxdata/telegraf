# Snowpipe Streaming Output Plugin

This plugin writes metrics to [Snowflake][snowflake] using efficient batch
inserts via the [gosnowflake][gosnowflake] driver with array binding, which
leverages Snowpipe Streaming internally for low-latency, high-throughput
ingest without staging files.

[snowflake]: https://www.snowflake.com/
[gosnowflake]: https://github.com/snowflakedb/gosnowflake

⭐ Telegraf v1.35.0
🏷️ cloud, datastore
💻 all

## Prerequisites

1. A Snowflake account with a database and schema already created.
2. Key-pair authentication configured for the Snowflake user:
   - Generate an RSA key pair:

     ```bash
     openssl genrsa 2048 | openssl pkcs8 -topk8 -inform PEM -out rsa_key.p8 -nocrypt
     openssl rsa -in rsa_key.p8 -pubout -out rsa_key.pub
     ```

   - Assign the public key to the user:

     ```sql
     ALTER USER my_user SET RSA_PUBLIC_KEY='<public key contents>';
     ```

3. The user must have INSERT privileges on the target table(s).
4. If `create_table = true`, the user must also have CREATE TABLE privileges.

## Global configuration options

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Stream metrics to Snowflake via Snowpipe Streaming
[[outputs.snowpipe_streaming]]
  ## Snowflake account identifier (e.g. "xy12345.us-east-1")
  account = ""

  ## Snowflake username for key-pair authentication
  user = ""

  ## Path to RSA private key file (PEM format) for key-pair auth
  private_key_path = ""

  ## Optional passphrase for the RSA private key
  # private_key_passphrase = ""

  ## Snowflake role to use
  # role = ""

  ## Target database name
  database = ""

  ## Target schema name
  schema = ""

  ## Target table name
  ## Supports Go templates with access to metric properties:
  ##   {{.Name}} - metric name
  ##   {{.Tag "key"}} - tag value
  ## Example: "metrics_{{.Name}}" routes each metric name to a separate table
  table = ""

  ## Number of rows per insert batch
  # batch_size = 1000

  ## Maximum number of retries on transient errors
  # retry_max = 3

  ## Delay between retries (exponential backoff base)
  # retry_delay = "1s"

  ## Column name to store the metric timestamp
  # timestamp_column = "timestamp"

  ## Restrict which tags to include as columns (empty = all tags)
  # tag_columns = []

  ## Restrict which fields to include as columns (empty = all fields)
  # field_columns = []

  ## Automatically create the target table if it does not exist
  # create_table = false

  ## How long to cache table schema information
  # table_schema_cache_ttl = "5m"
```

## Table Schema

Each metric is stored as a row with the following column mapping:

| Column             | Type          | Source                |
|--------------------|---------------|-----------------------|
| `timestamp`        | TIMESTAMP_NTZ | Metric timestamp      |
| `name`             | VARCHAR       | Metric name           |
| *(each tag key)*   | VARCHAR       | Tag value             |
| *(each field key)* | varies        | Field value           |

Field type mapping:

| Go Type         | Snowflake Type |
|-----------------|----------------|
| int64, uint64   | NUMBER         |
| float64         | DOUBLE         |
| bool            | BOOLEAN        |
| string          | VARCHAR        |

When `create_table = true`, the plugin will create the table with appropriate
types. When new tags or fields appear, columns are automatically added via
`ALTER TABLE ADD COLUMN`.

## Example Configurations

### Basic — single table

```toml
[[outputs.snowpipe_streaming]]
  account = "xy12345.us-east-1"
  user = "TELEGRAF_USER"
  private_key_path = "/etc/telegraf/snowflake_key.p8"
  database = "TELEMETRY"
  schema = "PUBLIC"
  table = "METRICS"
  create_table = true
```

### Template-based table routing

```toml
[[outputs.snowpipe_streaming]]
  account = "xy12345.us-east-1"
  user = "TELEGRAF_USER"
  private_key_path = "/etc/telegraf/snowflake_key.p8"
  database = "TELEMETRY"
  schema = "RAW"
  table = "metrics_{{.Name}}"
  create_table = true
```

### Specific columns only

```toml
[[outputs.snowpipe_streaming]]
  account = "xy12345.us-east-1"
  user = "TELEGRAF_USER"
  private_key_path = "/etc/telegraf/snowflake_key.p8"
  database = "TELEMETRY"
  schema = "PUBLIC"
  table = "CPU_METRICS"
  tag_columns = ["host", "cpu"]
  field_columns = ["usage_idle", "usage_user", "usage_system"]
  batch_size = 5000
```

## Troubleshooting

### Authentication errors

Ensure your RSA key pair is correctly configured:

```sql
DESC USER my_user;
```

Check that `RSA_PUBLIC_KEY_FP` is set and matches your key.

### Permission errors

The user/role must have the required grants:

```sql
GRANT USAGE ON DATABASE telemetry TO ROLE my_role;
GRANT USAGE ON SCHEMA telemetry.public TO ROLE my_role;
GRANT INSERT ON TABLE telemetry.public.metrics TO ROLE my_role;
-- If using create_table = true:
GRANT CREATE TABLE ON SCHEMA telemetry.public TO ROLE my_role;
```

### Transient errors and retries

The plugin automatically retries on transient errors (connection resets,
timeouts, service unavailable) with exponential backoff. Increase `retry_max`
and `retry_delay` for unreliable networks.

### NaN/Inf field values

Fields containing NaN or Inf float values are inserted as NULL to avoid
Snowflake errors.
