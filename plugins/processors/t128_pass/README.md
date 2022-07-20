# T128 Filter Processor Plugin

The `t128_filter` filters out metrics passing through it. This processor is useful if the sort of filtering you wish to accomplish which is not possible with `tagpass`/`tagdrop`. The most usual case this processor can help is if you have two tags that you want both to be a certain value before passing along the metric.

[Metric selectors](docs/CONFIGURATION.md#selectors) (such as `tagpass` and `tagdrop`) can be used to only apply the filters to specific metrics.

### Configuration:

```toml
[[processors.t128_filter]]
  ## The conditions that must be met to pass a metric through. This is similar
  ## behavior to a tagpass, but the multiple tags are ANDed
  [[processors.t128_filter.condition]]
    ## Mode dictates how to match the condition's tag values
    ## Valid values are:
    ##  * "exact": exact string comparison
    ##  * "glob": go flavored glob comparison (see https://github.com/gobwas/glob)
    ##  * "regex": go flavored regex comparison
    # mode = "exact"

    ## Operation dictates how to combine the condition's tag matching
    ## Valid values are:
    ##  * "and": logical and the results together
    ##  * "or": logical or the results together
    # operation = "and"

    ## Invert dictates whether to invert the final result of the condition
    # invert = false

    ## Whether to ignore if any tag or field keys are missing.
    # ignore_missing_keys = false

  [processors.t128_filter.condition.tags]
    # tag1 = ["value1", "value2"]
    # tag2 = ["value3"]

  ## Fields work the same as tags and can be included in the same condition.
  ## Only string values are accepted and the non-string field values in this section
  ## will be converted to strings before comparison.
  [processors.t128_pass.condition.fields.string]
    # field1 = ["value1", "value2"]
    # field2 = ["value3"]

  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
    # tag1 = ["value3"]
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

### Basic Field Example Filter:

Fields can also be filtered on by themselves or AND'd together with tags.

```toml
[[processors.t128_filter]]
  [[processors.t128_filter.condition]]

  [processors.t128_filter.condition.tags]
    tag1 = ["value1"]

  [processors.t128_filter.condition.fields.string]
    field1 = ["value2"]
```

```diff
measurement tag1=value1 field1=value2 1612214810000000000
- measurement tag1=value2 field1=value2 1612214810000000000
- measurement tag1=value1 field1=value1 1612214810000000000
```
