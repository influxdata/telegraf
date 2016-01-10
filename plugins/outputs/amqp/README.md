# AMQP Output Plugin

This plugin writes to a AMQP exchange using tag, defined in configuration file
as RoutingTag, as a routing key.

If RoutingTag is empty, then empty routing key will be used.
Metrics are grouped in batches by RoutingTag.

This plugin doesn't bind exchange to a queue, so it should be done by consumer.
