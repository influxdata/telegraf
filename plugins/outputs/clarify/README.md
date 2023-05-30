# Clarify Output Plugin

This plugin writes to [Clarify][clarify]. To use this plugin you will
need to obtain a set of [credentials][credentials].

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
## Configuration to publish Telegraf metrics to Clarify
[[outputs.clarify]]
  ## Credentials File (Oauth 2.0 from Clarify integration)
  credentials_file = "/path/to/clarify/credentials.json"

  ## Clarify username password (Basic Auth from Clarify integration)
  username = "i-am-bob"
  password = "secret-password"

  ## Timeout for Clarify operations
  # timeout = "20s"

  ## Optional tags to be included when generating the unique ID for a signal in Clarify
  # id_tags = []
  # clarify_id_tag = 'clarify_input_id'
```

You can use either a credentials file or username/password.
If both are present and valid in the configuration the
credentials file will be used.

## How Telegraf Metrics map to Clarify signals

Clarify signal names are formed by joining the Telegraf metric name and the
field key with a `.` character. Telegraf tags are added to signal labels.

If you wish to specify a specific tag to use as the input id, set the config
option `clarify_id_tag` to the tag containing the id to be used.
If this tag is present and there is only one field present in the metric,
this tag will be used as the inputID in Clarify. If there are more fields
available in the metric, the tag will be ignored and normal id generation
will be used.

If information from one or several tags is needed to uniquely identify a metric
field, the id_tags array can be added to the config with the needed tag names.
E.g:

`id_tags = ['sensor']`

Clarify only supports values that can be converted to floating point numbers.
Strings and invalid numbers are ignored.

## Example

The following input would be stored in Clarify with the values shown below:

```text
temperature,host=demo.clarifylocal,sensor=TC0P value=49 1682670910000000000
```

```json
"signal" {
  "id": "temperature.value.TC0P"
  "name": "temperature.value"
  "labels": {
    "host": ["demo.clarifylocal"],
    "sensor": ["TC0P"]
  }
}
"values" {
  "times": ["2023-04-28T08:43:16+00:00"],
  "series": {
    "temperature.value.TC0P": [49]
  }
}
```

[clarify]: https://clarify.io
[credentials]: https://docs.clarify.io/users/admin/integrations/credentials
