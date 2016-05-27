# Webhooks

This is a Telegraf service plugin that start an http server and start multiple webhook listeners.

```sh
$ telegraf -sample-config -input-filter webhooks -output-filter influxdb > config.conf.new
```

Change the config file to point to the InfluxDB server you are using and adjust the settings to match your environment. Once that is complete:

```sh
$ cp config.conf.new /etc/telegraf/telegraf.conf
$ sudo service telegraf start
```

## Available webhooks

- [Github](github/)
- [Rollbar](rollbar/)

## Adding new webhook services

TODO
