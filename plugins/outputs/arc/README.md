# Arc Output Plugin

This plugin writes metrics to [Arc][arc], a high-performance time-series
database, via MessagePack binary protocol messages providing a **3-5x better
performance** than the line-protocol format.

⭐ Telegraf v1.37.0
🏷️ datastore
💻 all

[arc]: https://github.com/basekick-labs/arc

## Global configuration options  <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Arc Time-Series Database Output Plugin
[[outputs.arc]]
  ## Arc MessagePack API URL
  url = "http://localhost:8000/api/v1/write/msgpack"

  ## API Key for authentication (required, auth is enabled by default)
  api_key = ""

  ## Database name for multi-database architecture
  ## Defaults to the server configured DB if not specified or empty
  # database = ""

  ## Content encoding for request body
  ## Options: "gzip" (default), "identity"
  # content_encoding = "gzip"

  ## Timeout for HTTP writes
  # timeout = "5s"

  ## Additional HTTP headers
  # [outputs.arc.headers]
  #   X-Custom-Header = "custom-value"
```

## Troubleshooting

For authentication issues, ensure you have generated a valid API key with write
permissions. See the [Arc documentation](https://docs.basekick.net/arc) for
details on authentication and configuration.

For connection or performance issues, check that Arc is running and accessible,
and review the Telegraf debug logs.
