# Kafka Consumer Input Plugin

The [Kafka](http://kafka.apache.org/) consumer plugin polls a specified Kafka
topic and adds messages to InfluxDB. The plugin assumes messages follow the
line protocol. [Consumer Group](http://godoc.org/github.com/wvanbergen/kafka/consumergroup)
is used to talk to the Kafka cluster so multiple instances of telegraf can read
from the same topic in parallel.

## Configuration

```toml
# Read metrics from Kafka topic(s)
[[inputs.kafka_consumer]]
  ## topic(s) to consume
  topics = ["telegraf"]
  ## an array of Zookeeper connection strings
  zookeeper_peers = ["localhost:2181"]
  ## the name of the consumer group
  consumer_group = "telegraf_metrics_consumers"
  ## Maximum number of metrics to buffer between collection intervals
  metric_buffer = 100000
  ## Offset (must be either "oldest" or "newest")
  offset = "oldest"

  ## Data format to consume. 

  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Testing

Running integration tests requires running Zookeeper & Kafka. The following
commands assume you're on OS X & using [boot2docker](http://boot2docker.io/) or docker-machine through [Docker Toolbox](https://www.docker.com/docker-toolbox).

To start Kafka & Zookeeper:

```
docker run -d -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST=`boot2docker ip || docker-machine ip <your_machine_name>` --env ADVERTISED_PORT=9092 spotify/kafka
```
