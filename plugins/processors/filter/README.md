# Filter Processor Plugin

This plugin allows specifying a set of rules for metrics with the ability to
_keep_ or _drop_ those metrics. It does _not_ modify the metric. As such a user
might want to apply this processor to remove metrics from the processing/output
stream.

> [!NOTE]
> The filtering is _not_ output specific, but will apply to the metrics
> processed by this processor.

⭐ Telegraf v1.29.0
🏷️ filtering
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Filter metrics by the given criteria
[[processors.filter]]
    ## Default action if no rule applies
    # default = "pass"

    ## Rules to apply on the incoming metrics (multiple rules are possible)
    ## The rules are evaluated in order and the first matching rule is applied.
    ## In case no rule matches the "default" is applied.
    ## All filter criteria in a rule must apply for the rule to match the metric.
    ## The criteria are combined by a logical AND. If a criterion is
    ## omitted, it is NOT applied at all and ignored.
    [[processors.filter.rule]]
        ## List of metric names to match including glob expressions
        # name = []

        ## List of tag key/values pairs to match including glob expressions
        ## ALL given tags keys must exist and at least one value must match
        ## for the metric to match the rule.
        # tags = {}

        ## List of field keys to match including glob expressions
        ## At least one field must exist for the metric to match the rule.
        # fields = []

        ## Action to apply for this rule
        ## "pass" will keep the metric and pass it on, while "drop" will remove
        ## the metric
        # action = "drop"
```

## Examples

Consider a use-case where you collected a bunch of metrics

```text
machine,source="machine1",status="OK" operating_hours=37i,temperature=23.1
machine,source="machine2",status="warning" operating_hours=1433i,temperature=48.9,message="too hot"
machine,source="machine3",status="OK" operating_hours=811i,temperature=29.5
machine,source="machine4",status="failure" operating_hours=1009i,temperature=67.3,message="temperature alert"
```

but only want to keep the ones indicating a `status` of `failure` or `warning`:

```toml
[[processors.filter]]
  namepass = ["machine"]
  default = "drop"

  [[processors.filter.rule]]
    tags = {"status" = ["warning", "failure"]}
    action = "pass"
```

Alternatively, you can "black-list" the `OK` value via

```toml
[[processors.filter]]
  namepass = ["machine"]

  [[processors.filter.rule]]
    tags = {"status" = ["OK"]}
```
