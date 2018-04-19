# PostgreSQL Output Plugin

This output plugin writes all metrics to PostgreSQL using CopyIn.

Obs: Currently, you should create your PostgreSQL tables first

### Configuration:

```toml
# Send metrics to PostgreSQL using CopyIn
[[outputs.postgresql_copy]]
  address = "postgres://USER:PWD@HOST:PORT/DATABASE?sslmode=disable"
  ignore_insert_errors = false
```
