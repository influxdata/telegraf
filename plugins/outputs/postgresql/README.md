# PostgreSQL Output Plugin

This output plugin writes all metrics to PostgreSQL.

### Configuration:

```toml
# Send metrics to postgres
[[outputs.postgresql]]
  address = "host=localhost user=postgres sslmode=verify-full"

  ## A list of tags to exclude from storing. If not specified, all tags are stored.
  # ignored_tags = ["foo", "bar"]

  ## Store tags as foreign keys in the metrics table. Default is false.
  # tags_as_foreignkeys = false
```
