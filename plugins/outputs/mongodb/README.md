# MongoDB Output Plugin

This plugin sends metrics to MongoDB and automatically creates the collections as time series collections when they don't already exist.
**Please note:** Requires MongoDB 5.0+ for Time Series Collections

### Configuration:

```toml
# A plugin that can transmit logs to mongodb
[[outputs.mongodb]]
  dsn = "mongodb://localhost:27017"
  # dsn = "mongodb://mongod1:27017,mongod2:27017,mongod3:27017/admin&replicaSet=myReplSet&w=1"
  authentication = "NONE"
  # authentication = "SCRAM"
  # username = "root"
  # password = "***"
  # authentication = "X509"
  # x509clientpem = "clientpwd.pem"
  # x509clientpempwd = "changeme"
  # cafile = "ca.pem"
  # allow_tls_insecure = false
  database = "telegraf"
  granularity = "seconds"
  ttl = "120s"
```