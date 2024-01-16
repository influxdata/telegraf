# Sample Configuration

The sample config file is generated from a results of the `SampleConfig()` functions of the plugin.

You can generate a full sample
config:

```shell
telegraf config
```

You can also generate the config for a particular plugin using the `-usage`
option:

```shell
telegraf --usage influxdb
```

## Style

In the config file we use 2-space indention.  Since the config is
[TOML](https://github.com/toml-lang/toml) the indention has no meaning.

Documentation is double commented, full sentences, and ends with a period.

```toml
  ## This text describes what an the exchange_type option does.
  # exchange_type = "topic"
```

Try to give every parameter a default value whenever possible.  If a
parameter does not have a default or must frequently be changed then have it
uncommented.

```toml
  ## Brokers are the AMQP brokers to connect to.
  brokers = ["amqp://localhost:5672"]
```

Options where the default value is usually sufficient are normally commented
out.  The commented out value is the default.

```toml
  ## What an exchange type is.
  # exchange_type = "topic"
```

If you want to show an example of a possible setting filled out that is
different from the default, show both:

```toml
  ## Static routing key.  Used when no routing_tag is set or as a fallback
  ## when the tag specified in routing tag is not found.
  ##   example: routing_key = "telegraf"
  # routing_key = ""
```

Unless parameters are closely related, add a space between them.  Usually
parameters is closely related have a single description.

```toml
  ## If true, queue will be declared as an exclusive queue.
  # queue_exclusive = false

  ## If true, queue will be declared as an auto deleted queue.
  # queue_auto_delete = false

  ## Authentication credentials for the PLAIN auth_method.
  # username = ""
  # password = ""
```

Parameters should usually be describable in a few sentences.  If it takes
much more than this, try to provide a shorter explanation and provide a more
complex description in the Configuration section of the plugins
[README](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/example)

Boolean parameters should be used judiciously.  You should try to think of
something better since they don't scale well, things are often not truly
boolean, and frequently end up with implicit dependencies: this option does
something if this and this are also set.
