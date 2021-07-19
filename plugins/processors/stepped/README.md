# Stepped Processor Plugin

Emit a metric of the previous unique field value and tag with the timestamp set just before the current one to display the field as stepped. See [Step Function](https://en.wikipedia.org/wiki/Step_function). 

## How?

The stepped processor plugin caches the last field value for a set of tags. When processing, it compares the processing metrics against the cache, if a field in `unique_fields` has changed between the current value and cache, then this processor emits an additional metric that is the last value at the timestamp of the current record minus the `step_offset`. Add a metric to the cache if it has a `unique_field` and isn't already in the cache. Remove a metric from the cache if it hasn't been updated in `cache_interval`. Identical `unique_field` values will update the timestamp in the cache.

Note that different fields don't overwrite the cache metric and get merged. Consider the following stream:

```
m1,tag1=1 value=0  1600000000000000000
m1,tag1=1 temp=295 1600000001000000000
```

The cache would update on the second metric to include the field from the first.

```diff
- m1,tag1=1 value=0          1600000000000000000
+ m1,tag1=1 value=0,temp=295 1600000001000000000
```

## Properties

| Property       | Description                                                                                                                                 |
|----------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| unique_fields  | The fields to compare when stepping the signal, the stepped plugin will only emit a stepped metric if a value in this field(s) has changed. |
| step_offset    | The duration from the current record to step back from                                                                                      |
| cache_interval | How long to cache the last value for, before purging. Used to drop stale metrics.                                                           |


## Typical Use Cases

The use case for a stepped function is for fields that cannot be interpolated by default. Examples include strings, booleans, and storing states/levels as integers.

### Configuration

```toml
[[processors.stepped]]
	## Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

	## Unique Fields
	unique_fields = ["value"]

	## Step value offset
	step_offset = "1ns"

	## Maximum time to cache last value
	cache_interval = "720h"
```


### Example

```diff
m1,tag1=1 value=0 1600000000000000000
+ m1,tag1=1 value=0 1609999999999999999
m1,tag1=1 value=1 1610000000000000000
```

As Graph

```
    __             __
   ╱              |
  ╱       ->      |
_╱             ___|
```

## Contact

- Author: Tom Hollingworth
- Email: tom.hollingworth@spruiktec.com
- Github: [@tomhollingworth](https://github.com/tomhollingworth)
- Influx Community Slack: [@tomhollingworth](https://influxcommunity.slack.com)
