# Kafka Consumer

The [Kafka](http://kafka.apache.org/) consumer plugin polls a specified Kafka
topic and adds messages to InfluxDB. The plugin assumes messages follow the
line protocol. [Consumer Group](http://godoc.org/github.com/wvanbergen/kafka/consumergroup)
is used to talk to the Kafka cluster so multiple instances of telegraf can read
from the same topic in parallel.

## Testing

Running integration tests requires running Zookeeper & Kafka. The following
commands assume you're on OS X & using [boot2docker](http://boot2docker.io/) or docker-machine through [Docker Toolbox](https://www.docker.com/docker-toolbox).

To start Kafka & Zookeeper:

```
docker run -d -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST=`boot2docker ip || docker-machine ip <your_machine_name>` --env ADVERTISED_PORT=9092 spotify/kafka
```

To run tests:

```
go test
```
