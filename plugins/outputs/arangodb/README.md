# ArangoDb Output Plugin

This plugin writes telegraf metrics to ArangoDb

### Configuration

```toml
[[outputs.file]]
  ## ArangoDb URL to connect to
  url = "http://192.168.1.100:8529"

  ## Username to connect to ArangoDb
  username = "user"
  ## Password to connect to ArangoDb
  password = "password"

  ## Database to write logs to
  database = "logdb"

  ## Collection to write logs to
  collection = "log_data"
  
  ## Uses batch transactions to insert log records into ArangoDb
  use_batch_format = true

```
