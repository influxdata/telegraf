# Quix Output Plugin

This plugin writes metrics to a [Quix](https://quix.io/) endpoint.

Please consult Quix's [official documentation][quick] for more
details on the Quix platform architecture and concepts.

⭐ Telegraf v1.33.0 🏷️ cloud, messaging 💻 all

[quix]: https://quix.io/docs/

## Quix Authentication

This plugin uses a SDK token for authentication with Quix. You can generate
one in the settings under the `API and tokens` section by clicking on 
`SDK Token`.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[outputs.quix]]
  workspace = "your_workspace"
  auth_token = "your_auth_token"
  topic = "telegraf_metrics"
  api_url = "https://portal-api.platform.quix.io"
  data_format = "json"
  timestamp_units = "1s"
```

For this output plugin to function correctly the following variables must be
configured.

* workspace
* auth_token
* topic

### workspace

The workspace is the environment of your Quix project and is the `Workspace ID`
or the `Environment ID` used to target your environment. It can be found in the
settings under the `General settings` section.

### auth_token

The auth_token is the `SDK Token` used to authenticate against your Quix
environment and is limited to that environment. It can be found in the settings
under the `API and tokens` section.

### topic

The plugin will send data to this named topic.

### api_url

The Quix platform API URL. Defaults to `https://portal-api.platform.quix.io`.

### data_format

The data format for serializing the messages. Defaults to `json`.

### timestamp_units

The timestamp units for precision. Defaults to `1s` for one second.
