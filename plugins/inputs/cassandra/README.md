# Telegraf plugin: Cassandra

#### Plugin arguments:
- **context** string: Context root used for jolokia url
- **servers** []Server: List of servers
  + **host** string: Server's ip address or hostname
  + **port** string: Server's listening port
  + **username** string: Server's username for authentication (optional)
  + **password** string: Server's password for authentication (optional)
- **metrics** []Metric
  + **jmx** string: Jmx path that identifies mbeans attributes
  + **pass** []string: Attributes to retain when collecting values (TODO)
  + **drop** []string: Attributes to drop when collecting values (TODO)

#### Description

The Cassandra plugin collects Cassandra/JVM metrics exposed as MBean's attributes through jolokia REST endpoint. All metrics are collected for each server configured.

See: https://jolokia.org/

# Measurements:
Cassandra plugin produces one or more measurements for each metric configured, adding Server's name  as `host` tag. More than one measurement is generated when querying table metrics with a wildcard for the keyspace or table name.

Given a configuration like:

```ini
[cassandra]

[[cassandra.servers]]
  host = "127.0.0.1"
  port = "878"

[[cassandra.metrics]]
  jmx  = "/java.lang:type=Memory/HeapMemoryUsage"
```

The collected metrics will be:

```
javaMemory,host=myHost,mname=HeapMemoryUsage HeapMemoryUsage_committed=1040187392,HeapMemoryUsage_init=1050673152,HeapMemoryUsage_max=1040187392,HeapMemoryUsage_used=368155000 1459551767230567084
```
