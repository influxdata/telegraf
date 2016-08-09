# Telegraf plugin: Spark

#### Plugin arguments:
- **SparkServer** []string: List of spark nodes with the format ["host:port"] (optional)
- **YarnServer** string: Address of Yarn resource manager (optional)

#### Description

The Spark plugin collects metrics in 2 ways,both being optional: <br />
**1.** Spark-JVM metrics exposed as MBean's attributes through jolokia REST endpoint. Metrics are collected for each server configured. See: https://jolokia.org/. <br />
**2.** Spark application metrics if managed by Yarn Resource manager. The plugin collects through the Yarn API. If some spark job has been submitted then only it will fetch else it will not produce any spark application result.

# Measurements:
Spark plugin produces one or more measurements according to the SparkServer or YarnServer provided.

Given a configuration like:

```toml
[[inputs.spark]]
  SparkServer = ["127.0.0.1:8778"]
  YarnServer = "127.0.0.1:8088"
```

The maximum collected measurements will be:

```
spark_HeapMemoryUsage , spark_Threading , spark_apps , spark_clusterInfo , spark_clusterMetrics , spark_jolokiaMetrics , spark_jvmMetrics , spark_nodes
```

# Useful Metrics:

Here is a list of metrics that are fetched and might be useful to monitor your spark cluster.

####measurement domains collected through Jolokia

- "/metrics:*"
- "/java.lang:type=Memory/HeapMemoryUsage"
- "/java.lang:type=Threading

####measurement domains collected through YarnServer
- "/cluster"
- "/cluster/metrics"
- "/cluster/apps"
- "/cluster/nodes"

 

