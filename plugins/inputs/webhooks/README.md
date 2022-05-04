# Webhooks Input Plugin

This is a Telegraf service plugin that start an http server and register multiple webhook listeners.

```sh
telegraf config -input-filter webhooks -output-filter influxdb > config.conf.new
```

Change the config file to point to the InfluxDB server you are using and adjust the settings to match your environment. Once that is complete:

```sh
cp config.conf.new /etc/telegraf/telegraf.conf
sudo service telegraf start
```

## Configuration

```toml
# A Webhooks Event collector
[[inputs.webhooks]]
  ## Address and port to host Webhook listener on
  service_address = ":1619"

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
```

## Available webhooks

- [Filestack](filestack/)
- [Github](github/)
- [Mandrill](mandrill/)
- [Rollbar](rollbar/)
- [Papertrail](papertrail/)
- [Particle](particle/)

## Adding new webhooks plugin

1. Add your webhook plugin inside the `webhooks` folder
1. Your plugin must implement the `Webhook` interface
1. Import your plugin in the `webhooks.go` file and add it to the `Webhooks` struct

Both [Github](github/) and [Rollbar](rollbar/) are good example to follow.
