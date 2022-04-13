# Tag Limit Processor Plugin

Use the `tag_limit` processor to ensure that only a certain number of tags are
preserved for any given metric, and to choose the tags to preserve when the
number of tags appended by the data source is over the limit.

This can be useful when dealing with output systems (e.g. Stackdriver) that
impose hard limits on the number of tags/labels per metric or where high
levels of cardinality are computationally and/or financially expensive.

## Configuration

```toml
# Restricts the number of tags that can pass through this filter and chooses which tags to preserve when over the limit.
[[processors.tag_limit]]
  ## Maximum number of tags to preserve
  limit = 3

  ## List of tags to preferentially preserve
  keep = ["environment", "region"]
```

## Example

```diff
+ throughput month=Jun,environment=qa,region=us-east1,lower=10i,upper=1000i,mean=500i 1560540094000000000
+ throughput environment=qa,region=us-east1,lower=10i 1560540094000000000
```
