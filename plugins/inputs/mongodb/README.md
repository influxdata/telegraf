# Telegraf plugin: MongoDB

#### Configuration

```toml
[[inputs.mongodb]]
  servers = ["127.0.0.1:27017"]
```

For authenticated mongodb istances use connection mongdb connection URI

```toml
[[inputs.mongodb]]
  servers = ["mongodb://myuser:mypassword@myinstance.telegraf.com:27601/mydatabasename?authMechanism=PLAIN&authSource=$external"]
```
This connection uri may be different based on your environement and mongodb setup. If the user doesn't have the required priviliges to execute serverStatus command the you will get this error on telegraf

```toml
Error in input [mongodb]: not authorized on admin to execute command { serverStatus: 1, recordStats: 0 }
```

#### Description

The Jolokia plugin collects JVM metrics exposed as MBean's attributes through jolokia REST endpoint. All metrics
are collected for each server configured.

See: https://jolokia.org/

# Measurements:
Jolokia plugin produces one measure for each metric configured, adding Server's `name`, `host` and `port` as tags.
