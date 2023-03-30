# Webhooks Input Plugin

This is a Telegraf service plugin that start a http server and register
multiple webhook listeners.

```sh
telegraf config -input-filter webhooks -output-filter influxdb > config.conf.new
```

Change the config file to point to the InfluxDB server you are using and adjust
the settings to match your environment. Once that is complete:

```sh
cp config.conf.new /etc/telegraf/telegraf.conf
sudo service telegraf start
```

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# A Webhooks Event collector
[[inputs.webhooks]]
  ## Address and port to host Webhook listener on
  service_address = ":1619"

  ## Maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## Maximum duration before timing out write of the response
  # write_timeout = "10s"

  [inputs.webhooks.filestack]
    path = "/filestack"

    ## HTTP basic auth
    #username = ""
    #password = ""

  [inputs.webhooks.github]
    path = "/github"
    # secret = ""

    ## HTTP basic auth
    #username = ""
    #password = ""

  [inputs.webhooks.mandrill]
    path = "/mandrill"

    ## HTTP basic auth
    #username = ""
    #password = ""

  [inputs.webhooks.rollbar]
    path = "/rollbar"

    ## HTTP basic auth
    #username = ""
    #password = ""

  [inputs.webhooks.papertrail]
    path = "/papertrail"

    ## HTTP basic auth
    #username = ""
    #password = ""

  [inputs.webhooks.particle]
    path = "/particle"

    ## HTTP basic auth
    #username = ""
    #password = ""

  [inputs.webhooks.artifactory]
    path = "/artifactory"
```

## Available webhooks

- [Filestack](filestack/)
- [Github](github/)
- [Mandrill](mandrill/)
- [Rollbar](rollbar/)
- [Papertrail](papertrail/)
- [Particle](particle/)
- [Artifactory](artifactory/)

## Adding new webhooks plugin

1. Add your webhook plugin inside the `webhooks` folder
1. Your plugin must implement the `Webhook` interface
1. Import your plugin in the `webhooks.go` file and add it to the `Webhooks` struct

Both [Github](github/) and [Rollbar](rollbar/) are good example to follow.

## Metrics

## Example Output
