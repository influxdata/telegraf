# MongoDB Output Plugin

This plugin writes metrics to [MongoDB][mongodb] automatically creating
collections as time series collections if they don't exist.

> [!NOTE]
> This plugin requires MongoDB v5 or later for time series collections.

⭐ Telegraf v1.21.0
🏷️ datastore
💻 all

[mongodb]: https://www.mongodb.com

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret store support

This plugin supports secrets from secret stores for the `username` and
`password` option.
See the [secret store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# A plugin that can transmit logs to mongodb
[[outputs.mongodb]]
  ## Connection string
  ## example: dsn = "mongodb://mongod1:27017,mongod2:27017,mongod3:27017/admin&replicaSet=myReplSet&w=1"
  dsn = "mongodb://localhost:27017"

  ## Overrides serverSelectionTimeoutMS in DSN if set
  # timeout = "30s"

  ## Authentication method, available options are NONE, PLAIN, SCRAM, X509
  # authentication = "NONE"

  # ## for SCRAM-SHA-256 authentication
  # # authentication = "SCRAM"
  # # username = "root"
  # # password = "***"

  # ## for PLAIN authentication (e.g., LDAP)
  # ## IMPORTANT: PLAIN authentication sends credentials in plaintext during the
  # ## authentication handshake. Always use TLS to encrypt credentials in transit.
  # # authentication = "PLAIN"
  # # username = "myuser"
  # # password = "***"

  # ## X509 based certificate authentication
  # # authentication = "X509"
  # # tls_ca = "ca.pem"
  # # tls_key = "client.pem"
  # # # tls_key_pwd = "changeme" # required for encrypted tls_key
  # # insecure_skip_verify = false

  ## Database to store measurements and time series collections
  # database = "telegraf"

  ## If true, write multiple metrics for the same collection in a batched
  ## fashion. Otherwise, write each metric individually.
  # write_batch = false

  ## List of tags to use as metadata (wildcards accepted)
  ## If empty the "tags" element will be used as metadata field, otherwise
  ## a "metadata" element will be created containing the matching, existing
  ## tag key-values.
  # metadata_keys = []

  ## Behavior when using explicit metadata keys; ignored for empty metadata_keys
  ##  keep  -- keep all tags as tags
  ##  move  -- move tags matching metadata_keys and delete them in the tags field
  ##  clear -- clear ALL tags and remove the tags field entirely from the document
  # metadata_tag_strategy = "keep"

  ## Granularity can be seconds, minutes, or hours.
  ## Configuring this value will be based on your input collection frequency
  ## see https://docs.mongodb.com/manual/core/timeseries-collections/#create-a-time-series-collection
  # granularity = "seconds"

  ## TTL to automatically expire documents from the measurement collections.
  # ttl = "360h"

```
