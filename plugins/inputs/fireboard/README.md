# Fireboard Input Plugin

The fireboard plugin gathers the real time temperature data from fireboard
thermometers.  In order to use this input plugin, you'll need to sign up to use
the [Fireboard REST API](https://docs.fireboard.io/reference/restapi.html).

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read real time temps from fireboard.io servers
[[inputs.fireboard]]
  ## Specify auth token for your account
  auth_token = "invalidAuthToken"
  ## You can override the fireboard server URL if necessary
  # url = https://fireboard.io/api/v1/devices.json
  ## You can set a different http_timeout if you need to
  ## You should set a string using an number and time indicator
  ## for example "12s" for 12 seconds.
  # http_timeout = "4s"
```

### auth_token

In lieu of requiring a username and password, this plugin requires an
authentication token that you can generate using the [Fireboard REST
API](https://docs.fireboard.io/reference/restapi.html#Authentication).

### url

While there should be no reason to override the URL, the option is available
in case Fireboard changes their site, etc.

### http_timeout

If you need to increase the HTTP timeout, you can do so here. You can set this
value in seconds. The default value is four (4) seconds.

## Metrics

The Fireboard REST API docs have good examples of the data that is available,
currently this input only returns the real time temperatures. Temperature
values are included if they are less than a minute old.

- fireboard
  - tags:
    - channel
    - scale (Celcius; Farenheit)
    - title (name of the Fireboard)
    - uuid (UUID of the Fireboard)
  - fields:
    - temperature (float, unit)

## Example Output

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter <plugin-name> --test` or use the `file` output to get
this information.

```text
fireboard,channel=2,host=patas-mbp,scale=Farenheit,title=telegraf-FireBoard,uuid=b55e766c-b308-49b5-93a4-df89fe31efd0 temperature=78.2 1561690040000000000
```
