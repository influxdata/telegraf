# Telegraf plugin: Cassandra

#### Plugin arguments:
- **context** string: Context root used for jolokia url
- **servers** []string: List of servers with the format "<user:passwd@><host>:port"
- **metrics** []string: List of Jmx paths that identify mbeans attributes

#### Description

The Cassandra plugin collects Cassandra/JVM metrics exposed as MBean's attributes through jolokia REST endpoint. All metrics are collected for each server configured.

See: https://jolokia.org/

# Measurements:
Cassandra plugin produces one or more measurements for each metric configured, adding Server's name  as `host` tag. More than one measurement is generated when querying table metrics with a wildcard for the keyspace or table name.

Given a configuration like:

```ini
[inputs.cassandra]
  context = "/jolokia/read"
  servers = [":8778"]
  metrics = ["/java.lang:type=Memory/HeapMemoryUsage"]
```

The collected metrics will be:

```
javaMemory,host=myHost,mname=HeapMemoryUsage HeapMemoryUsage_committed=1040187392,HeapMemoryUsage_init=1050673152,HeapMemoryUsage_max=1040187392,HeapMemoryUsage_used=368155000 1459551767230567084
```
