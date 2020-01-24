# AMQP 1.0 Output Plugin

This plugin writes to a AMQP 1-0 queue, for example Microsoft Azure Service Bus or Event Hub

It is an early version, that is functional but has still room for improvements.

Further information can be found here:
- https://github.com/xinchen10/awesome-amqp
- https://docs.microsoft.com/en-us/azure/service-bus-messaging/service-bus-amqp-overview

### Configuration:
```toml
# Publishes metrics to an AMQP 1.0 broker
[[outputs.amqp10]]
  ## Brokers to publish to.  If multiple brokers are specified a random broker
  ## will be selected anytime a connection is established.  This can be
  ## helpful for load balancing when not using a dedicated load balancer.
  ## The SASPolicyKey has to be URL encoded!
  brokers = ["amqps://[SASPolicyName]:[SASPolicyKey]@[namespace].servicebus.windows.net"]

  ## Target address to send the message to.
  topic = "/target"

  ## Authentication credentials.
  # username = ""
  # password = ""

  ## Connection timeout.  If not provided, will default to 5s.  0s means no
  ## timeout (not recommended).
  # timeout = "5s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
```
