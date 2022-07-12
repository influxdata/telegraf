# T128 Filter Processor Plugin

The `t128_filter` filters out metrics passing through it. This processor is useful if the sort of filtering you wish to accomplish which is not possible with `tagpass`/`tagdrop`. The most usual case this processor can help is if you have two tags that you want both to be a certain value before passing along the metric.

[Metric selectors](docs/CONFIGURATION.md#selectors) (such as `tagpass` and `tagdrop`) can be used to only apply the filters to specific metrics.

### Configuration:

```toml
[[processors.t128_filter]]
  ## The conditions that must be met to pass a metric through. This is similar
  ## behavior to a tagpass, but multiple tags are ANDed
  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
     #tag1 = ["value1", "value2"]
     #tag2 = ["value3"]
```

### Basic Example Filter:

Multiple values for a specific tag are OR'd together. A metric without a tag is also dropped.

```toml
[[processors.t128_filter]]
  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
    tag1 = ["value1", "value2"]
```

```diff
measurement tag1=value1 1612214810000000000
measurement tag1=value2 1612214810000000000
- measurement tag1=value3 1612214805000000000
- measurement 1612214805000000000
```

### Multiple Keys Example Filter:

Multiple keys are AND'd together.

```toml
[[processors.t128_filter]]
  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
    tag1 = ["value1"]
    tag2 - ["value2"]
```

```diff
measurement tag1=value1,tag2=value2 1612214810000000000
- measurement tag2=value2 1612214810000000000
- measurement tag1=value1 1612214805000000000
- measurement tag1=value1,tag2=value3 1612214805000000000
```

### Multiple Conditions Example Filter:

Multiple conditions are OR'd together.

```toml
[[processors.t128_filter]]
  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
    tag1 = ["value1"]

  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
    tag1 = ["value2"]
```

```diff
measurement tag1=value1 1612214810000000000
measurement tag1=value2 1612214810000000000
- measurement tag1=value3 1612214810000000000
```
