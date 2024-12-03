# Quix Output Plugin

This plugin writes metrics to a [Quix][quix] endpoint.

Please consult Quix's [official documentation][docs] for more details on the
Quix platform architecture and concepts.

⭐ Telegraf v1.33.0
🏷️ cloud, messaging
💻 all

[quix]: https://quix.io
[docs]: https://quix.io/docs/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Send metrics to a Quix data processing pipeline
[[outputs.quix]]
  ## Endpoint accepting the data
  # url = "https://portal-api.platform.quix.io"

  ## Workspace and topics to send the metrics to
  workspace = "your_workspace"
  topic = "your_topic"

  ## Authentication token created in Quix
  token = "your_auth_token"
```

The plugin requires a [SDK token][token] for authentication with Quix. You can
generate the `token` in settings under the `API and tokens` section.

Furthermore, the `workspace` parameter must be set to the `Workspace ID` or the
`Environment ID` of your Quix project. Those values can be found in settings
under the `General settings` section.

[token]: https://quix.io/docs/develop/authentication/personal-access-token.html
