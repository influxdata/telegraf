# PostgreSQL Output Plugin

This output plugin writes all metrics to PostgreSQL in batch.

Obs: Currently, you should create your PostgreSQL tables first

### Configuration:

```toml
# Send metrics to PostgreSQL using batch (grouped insert)
[[outputs.postgresql_batch]]
  address = "host=localhost user=postgres sslmode=verify-full"
```