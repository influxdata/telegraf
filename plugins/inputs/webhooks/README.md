# Webhooks Input Plugin

This service plugin provides an HTTP server and register for multiple webhook
listeners.

⭐ Telegraf v1.0.0
🏷️ applications, web
💻 all

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listen and wait for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

### Available webhooks

- Artifactory
- Filestack
- Github
- Mandrill
- Papertrail
- Particle
- Rollbar

## Metrics

The produced metrics depend on the configured webhook.

## Example Output

The produced metrics depend on the configured webhook.
