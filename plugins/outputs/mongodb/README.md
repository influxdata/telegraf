# MongoDB Output Plugin

This plugin sends metrics to MongoDB and automatically creates the collections
as time series collections when they don't already exist.  **Please note:**
Requires MongoDB 5.0+ for Time Series Collections

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# A plugin that can transmit logs to mongodb
[[outputs.mongodb]]
  # connection string examples for mongodb
  dsn = "mongodb://localhost:27017"
  # dsn = "mongodb://mongod1:27017,mongod2:27017,mongod3:27017/admin&replicaSet=myReplSet&w=1"

  # overrides serverSelectionTimeoutMS in dsn if set
  # timeout = "30s"

  # default authentication, optional
  # authentication = "NONE"

  # for SCRAM-SHA-256 authentication
  # authentication = "SCRAM"
  # username = "root"
  # password = "***"

  # for x509 certificate authentication
  # authentication = "X509"
  # tls_ca = "ca.pem"
  # tls_key = "client.pem"
  # # tls_key_pwd = "changeme" # required for encrypted tls_key
  # insecure_skip_verify = false

  # database to store measurements and time series collections
  # database = "telegraf"

  # granularity can be seconds, minutes, or hours.
  # configuring this value will be based on your input collection frequency.
  # see https://docs.mongodb.com/manual/core/timeseries-collections/#create-a-time-series-collection
  # granularity = "seconds"

  # optionally set a TTL to automatically expire documents from the measurement collections.
  # ttl = "360h"
```
