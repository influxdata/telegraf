# MongoDB Output Plugin

This plugin sends metrics to MongoDB and automatically creates the collections as time series collections when they don't already exist.
**Please note:** Requires MongoDB 5.0+ for Time Series Collections

### Configuration:

```toml
# A plugin that can transmit logs to mongodb
[[outputs.mongodb]]
  connection_string = "mongodb://localhost:27017/admin"
  authentication_type = "NONE"
  # authentication_type = "SCRAM"
  # username = "root"
  # password = "***"
  # authentication_type = "X509"
  # x509clientpem = "client.pem"
  # x509clientpempwd = "changeme"
  # cafile = "ca.pem" #if using X509 authentication
  # allow_tls_insecure = false
  metric_database = "telegraf" #tells telegraf which database to write metrics to. collections are automatically created as time series collections
  metric_granularity = "seconds" # can be seconds, minutes, or hours
  retention_policy = "120s" #set a TTL on the collect. examples: 120m, 24h, or 15d
  data_format = "json" #always set to json for proper serialization
```