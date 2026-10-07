# Classify Processor Plugin

The `classify` plugin classifies metrics by matching a designated tag or field
value against groups of regular expressions. Each input metric either passes
through with a new result tag/field set to the matched category name, or is
dropped. Apart from the result tag or field, which replaces any existing tag
or field of the same name, the metric is not modified.

The plugin optionally supports a selector mapping step: a tag or field value
is mapped to the name of a regex group, allowing distinct sets of regexes to
be applied to different classes of input without repeating configuration.

In addition to classification, the plugin can accumulate per-period statistics
and emit them as separate metrics, acting as a lightweight aggregator.

⭐ Telegraf v1.41.0
🏷️ filtering, transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## When to use classify

The plugin turns an open-ended value, such as a log message, a URL path or an
SNMP trap description, into one of a few named categories, and can count how
often each category was seen. Typical uses, shown in the [examples](#examples):

- bounding tag cardinality, by replacing unbounded values with a few classes;
- normalizing severities or log levels that each source words differently;
- dropping known noise while measuring how much of it there is, and where it
  comes from;
- deriving the states that an alerting or monitoring system expects.

The stock plugins cover parts of this, but not the combination:

| Need | `classify` | Closest stock alternative |
| --- | --- | --- |
| Map a value to one of a few names by trying an ordered list of regexes, first match wins | Categories are tried in order; each may hold any number of regexes; the category name is written as is | `processors.regex` runs every conversion in turn, so the last match wins and a catch-all must come first; each regex is its own block with its own replacement. `enum` and `lookup` only match exact values. |
| Fall back to a default state when nothing matches | `default_category` | `processors.defaults` only fills a missing tag or field, so it needs the regex result written to a separate key first. |
| Apply different rules to different kinds of sources | One `selector_mapping` picks a regex group per metric, with a fallback group | Several processor instances, each limited with `tagpass` globs. Telegraf metric filters cannot express "every source the other instances did not take", and every instance repeats the shared rules. |
| Drop uninteresting messages as part of the same decision | `drop_categories`, or no match without `default_category` | `processors.filter` matches tag values with globs only; `metricpass` can test a regex, but as a separate rule set that duplicates the classification. |
| Count states per period, including dropped metrics | `aggregation_*` options count every metric seen, bucketed by summary, regex group or selector value | `aggregators.valuecounter` runs after the processors, so it never sees dropped metrics; it counts per series (all tags) and names fields `<field>_<value>`. |

Prefer the stock plugins when they are enough: `enum` or `lookup` for
exact-value lookups, `template` for building values from other tags, `regex`
for rewriting parts of a value, and `starlark` for logic that does not fit any
declarative plugin.

## Processing Model

```text
input metric
    │
    ├─ resolve regex group (selector → group mapping, or default_regex_group)
    │
    ├─ match against category regexes in that group (first match wins)
    │
    ├─ apply result tag/field to metric
    │
    └─ pass downstream  ─OR─  drop (category in drop_categories, or no match
                                    and no default_category)
```

The selector resolves to one of several regex groups; only the selected group's
category regexes are tested against the match item:

```text
                    ┌─────────────────────────────────────┐
        selector ──►│   selector to regex-group mapping   │
                    └─────────────────────────────────────┘
                                        │
                                  regex group name
                                        │
                                        ▼
         ┌─ regex group 1 ────────────────────────────────┐
         │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
         │  │categoryA │  │categoryB │  │categoryC │     │
         │  │ regexes  │  │ regexes  │  │ regexes  │     │
         │  └──────────┘  └──────────┘  └──────────┘     │
         └────────────────────────────────────────────────┘

match ──►┌─ regex group 2 (selected) ─────────────────────┐──► result
item     │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
         │  │categoryA │  │categoryB │  │categoryC │     │
         │  │ regexes  │  │ regexes  │  │ regexes  │     │
         │  └──────────┘  └──────────┘  └──────────┘     │
         └────────────────────────────────────────────────┘

         ┌─ regex group 3 ────────────────────────────────┐
         │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
         │  │categoryA │  │categoryB │  │categoryC │     │
         │  │ regexes  │  │ regexes  │  │ regexes  │     │
         │  └──────────┘  └──────────┘  └──────────┘     │
         └────────────────────────────────────────────────┘

all data point tags and fields ──────────────────────────────────────────────►
```

### Choosing the regex group

When `selector_tag` or `selector_field` is set, the regex group is resolved
as follows:

1. A metric without the selector item is dropped, as is one whose selector
   field is not a string.
2. The `selector_mapping` entries are tried in order and the first one whose
   regex matches the selector value gives the group name. `"*"` uses the
   selector value itself as the group name, and `""` drops the metric.
3. When no entry matches, including when `selector_mapping` is not set,
   `default_regex_group` is used. When that is not set either, the metric is
   dropped.
4. When the resolved name is not a group with category regexes in
   `mapped_selector_regexes`, for example a `"*"` value that has no group of
   its own, `default_regex_group` is used, or the metric is dropped when it is
   not set.

Regexes are unanchored, so `pg\\d{3}` also matches `xpg1234`; anchor them
with `^` and `$` to match the whole value. To drop metrics whose selector is
empty, or all selectors not listed, end the mapping with a catch-all:

```toml
selector_mapping = [
  { "^fire\\d{3}$" = "firewall" },
  { "^pg\\d{3}$"   = "database" },
  { "^$"           = ""         },  # empty selector: drop
  { ".*"           = ""         },  # anything else: drop
]
```

Without a selector, every metric uses `default_regex_group`.

### Choosing the category

The match item is read from `match_tag`, or from `match_field`, which must
hold a string; a metric without it is dropped. The categories of the selected
group are then tried in the order they are listed, and within a category its
regexes are tried in order; the first regex that matches decides the category.

- When nothing matches, `default_category` is used, or the metric is dropped
  when it is not set. A last category with the regex `.*` has the same effect
  as `default_category`, but is counted under its own name.
- A metric whose category is in `drop_categories` is dropped.
- Otherwise the category name is written to `result_tag` or `result_field`,
  replacing any value already there, and the metric passes downstream.

A category may be listed with no regexes, such as `{ warning = [] }`, as a
placeholder for a state that a group has no rules for yet. It never matches,
but its name may still be used in `drop_categories` and the aggregation field
lists.

## Aggregation

When `aggregation_period` is set, the plugin emits classification counters as
separate metrics, named by `aggregation_measurement`, at the end of each period.
Omitting `aggregation_period` disables aggregation even if the other aggregation
options are set.

Three independent aggregation types can be enabled simultaneously:

- **Summary**: one metric per period with a fixed tag value, counting across all
  regex groups and selectors.
- **By group**: one metric per active regex group per period.
- **By selector**: one metric per observed selector value per period.

Each aggregation type emits only the fields listed in its `*_fields` option.
Fields with a zero count are omitted unless `aggregation_includes_zeroes` is
enabled; a data point whose listed fields are all zero is never emitted.

Every metric that reaches the plugin is counted, whether it passes or is
dropped:

- `aggregation_total_field` counts every metric.
- `aggregation_dropped_field` counts every dropped metric, for any reason.
- A category field counts the metrics classified as that category, including
  `default_category` and the categories in `drop_categories`. A metric dropped
  in `drop_categories` is therefore counted under both its category and the
  dropped field.

The summary counts all metrics. The per-group counts only include metrics for
which a regex group was found, under that group's name, which is
`default_regex_group` for metrics that fell back to it. The per-selector
counts only include metrics with a non-empty selector value, under the value
as read from the metric, before `selector_mapping` is applied. Mind the number
of distinct selector values, since each one becomes a separate series.

The counts are emitted at the end of each period, with that time as the
timestamp. Periods are aligned to the UTC clock, so a `10m` period emits at
`:00`, `:10` and so on, and a `1h` period on the hour. The counts are reset
after each emission, and the remaining counts are emitted when Telegraf stops.
Unlike an aggregator plugin, this plugin sees the metrics it drops, which is
what makes the dropped and total counts possible.

The aggregation metrics are passed downstream like the classified metrics:
through the processors with a higher `order` and to every output. Use
`namepass` or `namedrop` with the `aggregation_measurement` name to send them
only to the outputs that should receive them.

## Example output

```text
# Passthrough metric with result field added:
syslog,host=pg123 message="Tablespace users free space is low",status="warning" 1700000000

# Summary aggregation after one period:
aggregated_status,summary=full okay=8i,warning=3i,critical=1i,dropped=6i,total=18i 1700000400

# Per-group aggregation:
aggregated_status,host_type=database okay=8i,warning=2i,total=10i 1700000400
aggregated_status,host_type=firewall warning=1i,critical=1i,dropped=6i,total=8i 1700000400
```

## Configuration

```toml @sample.conf
# Classify Telegraf data points according to user-specified rules.
[[processors.classify]]
  ## Tag or field used to select which regex group to apply.
  ## These are mutually exclusive. Omit both to use default_regex_group directly,
  ## which is then required.
  # selector_tag = "host"
  # selector_field = ""

  ## Ordered list of regex-to-group-name mappings for the selector value.
  ## Each element must have exactly one key (the regex) and one value (the group name).
  ## Use "*" as the group name to pass the selector value through unchanged.
  ## selector_mapping = [
  ##   { "fire\\d{3}" = "firewall" },
  ##   { "pg\\d{3}"   = "database" },
  ## ]

  ## Regex group to use when no selector_mapping entry matches, or when the
  ## matching entry names a group that is not defined. It must name a group
  ## defined in mapped_selector_regexes. An entry that maps to "" drops the
  ## metric without this fallback. If empty and no defined group is found, the
  ## metric is dropped.
  # default_regex_group = ""

  ## Tag or field whose value is matched against the category regexes.
  ## Exactly one must be defined.
  # match_tag = ""
  # match_field = "message"

  ## Category applied when no regex matches. Metrics are dropped if unset.
  # default_category = ""

  ## Categories whose metrics are dropped after classification.
  ## Accepts a single string or an array of strings.
  ## drop_categories = ["ignore", "unknown"]

  ## Tag or field where the classification result is written.
  ## Exactly one must be defined.
  # result_tag = ""
  # result_field = "status"

  ## Aggregation options. Aggregation is enabled by setting aggregation_period
  ## along with at least one of the summary/group/selector sets below; it then
  ## requires aggregation_measurement. Omit aggregation_period to disable it.
  # aggregation_period = "10m"
  # aggregation_measurement = "aggregated_status"

  ## Field name for dropped-metric counts in aggregation output.
  ## Each *_fields list below names the fields emitted for that aggregation:
  ## regex categories, default_category, and these dropped/total fields.
  # aggregation_dropped_field = "dropped"

  ## Field name for total-metric counts in aggregation output.
  # aggregation_total_field = "total"

  ## Summary aggregation: one metric per period with a fixed tag value.
  ## All three options must be set together, or none of them.
  # aggregation_summary_tag = "summary"
  # aggregation_summary_value = "full"
  # aggregation_summary_fields = ["okay", "warning", "critical", "unknown", "dropped", "total"]

  ## Per-regex-group aggregation: one metric per active group per period.
  ## Both options must be set together, or neither.
  # aggregation_group_tag = "host_type"
  # aggregation_group_fields = ["okay", "warning", "critical", "unknown", "dropped", "total"]

  ## Per-selector-value aggregation: one metric per observed selector per period.
  ## Both options must be set together, or neither; requires selector_tag or
  ## selector_field.
  # aggregation_selector_tag = "host"
  # aggregation_selector_fields = ["okay", "warning", "critical", "unknown", "dropped", "total"]

  ## Include zero-value fields in aggregation output metrics. Metrics whose
  ## fields would all be zero are never emitted.
  # aggregation_includes_zeroes = false

  ## Per-group ordered category definitions.
  ## Each category value may be a single regex string, a multi-line string
  ## (one regex per non-blank, trimmed line), or an array of regex strings.
  ## A category with no regexes never matches, but its name may still be used
  ## in drop_categories and the aggregation *_fields lists.
  ## This table must come last: every key after its header belongs to it.
  ## [processors.classify.mapped_selector_regexes]
  ##   database = [
  ##     { ignore   = "DB client connected" },
  ##     { okay     = "Database is starting up" },
  ##     { warning  = "Tablespace \\w+ free space is low" },
  ##     { critical = "Database is shutting down" },
  ##     { unknown  = ".*" },
  ##   ]
  ##   firewall = [
  ##     { ignore   = "low-priority traffic" },
  ##     { warning  = "login attempt" },
  ##     { critical = "intrusion detected" },
  ##     { unknown  = ".*" },
  ##   ]
```

## Options

### Selector options

| Option | Description |
| --- | --- |
| `selector_tag` | Tag whose value selects the regex group. Mutually exclusive with `selector_field`. |
| `selector_field` | Field whose value selects the regex group. |
| `selector_mapping` | Ordered list of `{regex: group_name}` elements; the first match wins. Use `"*"` as the group name to pass the selector value through unchanged. A group name that is not in `mapped_selector_regexes` is logged as a warning at startup. |
| `default_regex_group` | Regex group to use when no selector mapping entry matches or the mapped group is not defined; an entry that maps to `""` drops the metric instead. Must name a group in `mapped_selector_regexes`; required when neither `selector_tag` nor `selector_field` is set. |

### Classification options

| Option | Description |
| --- | --- |
| `match_tag` | Tag whose value is matched against category regexes. Exactly one of `match_tag`/`match_field` is required. |
| `match_field` | Field whose value is matched against category regexes. |
| `mapped_selector_regexes` | TOML table mapping each group name to an ordered list of `{category: regex}` entries, each with exactly one category. Category values may be a single regex string, a multi-line string (one trimmed regex per non-blank line), or an array of strings. Category names must be non-empty; a category with no regexes never matches, but its name may still be used in `drop_categories` and the aggregation `*_fields` lists. |
| `default_category` | Category to apply when no regex matches. Metrics are dropped if unset and no match is found. |
| `drop_categories` | Category name or list of names whose matched metrics are dropped after classification. |
| `result_tag` | Tag to set to the matched category name, replacing any existing tag of that name. Exactly one of `result_tag`/`result_field` is required. |
| `result_field` | Field to set to the matched category name, replacing any existing field of that name. |

### Aggregation options

| Option | Description |
| --- | --- |
| `aggregation_period` | How often to emit aggregation metrics (e.g. `"5m"`). Must be ≥ 1s. |
| `aggregation_measurement` | Measurement name for aggregation metrics. Required when `aggregation_period` is set. |
| `aggregation_dropped_field` | Field name for the count of dropped metrics. |
| `aggregation_total_field` | Field name for the total count of all metrics processed. |
| `aggregation_summary_tag` | Tag name for summary aggregation. |
| `aggregation_summary_value` | Tag value for summary aggregation. |
| `aggregation_summary_fields` | Fields to emit in summary aggregation metrics: regex categories, `default_category`, `aggregation_dropped_field`, or `aggregation_total_field`. Set together with the summary tag and value. |
| `aggregation_group_tag` | Tag name for per-group aggregation. |
| `aggregation_group_fields` | Fields to emit in per-group aggregation metrics. Set together with `aggregation_group_tag`. |
| `aggregation_selector_tag` | Tag name for per-selector aggregation. Requires `selector_tag` or `selector_field`. |
| `aggregation_selector_fields` | Fields to emit in per-selector aggregation metrics. Set together with `aggregation_selector_tag`. |
| `aggregation_includes_zeroes` | Include listed fields with zero counts in aggregation output; all-zero points are still suppressed. Default: `false`. |

## Examples

### Bounding the cardinality of request paths

Storing raw URL paths as tags creates a new series for every distinct path.
Classifying them into a few route classes keeps the number of series fixed,
and drops health checks on the way. The access log is parsed with the `grok`
data format, which puts the path into the `request` field:

```toml
[[inputs.tail]]
  files = ["/var/log/nginx/access.log"]
  data_format = "grok"
  grok_patterns = ["%{COMBINED_LOG_FORMAT}"]

[[processors.classify]]
  default_regex_group = "paths"
  match_field = "request"
  default_category = "other"
  drop_categories = "health"
  result_tag = "route"

  [processors.classify.mapped_selector_regexes]
    paths = [
      { health = ["^/healthz", "^/ready"] },
      { api    = "^/api/" },
      { static = ['\.(css|js|png|svg|woff2?)$'] },
      { login  = "^/(login|logout|oauth)" },
    ]
```

Add `fieldexclude = ["request"]` to the outputs to drop the raw path once it
has been classified.

### Normalizing log levels across applications

Each application words its severities differently. A selector on the syslog
`appname` tag picks a rule set per application; with `"*"` the application
name is the group name, so adding an application only takes a new group.
Applications without a group of their own use the `generic` rules:

```toml
[[processors.classify]]
  selector_tag = "appname"
  selector_mapping = [
    { "^postgres" = "postgres" },
    { ".*"        = "*"        },
  ]
  default_regex_group = "generic"
  match_field = "message"
  default_category = "info"
  result_tag = "level"

  [processors.classify.mapped_selector_regexes]
    postgres = [
      { error = ["^(ERROR|FATAL|PANIC):"] },
      { warn  = ["^WARNING:"] },
    ]
    nginx = [
      { error = ['\[(emerg|alert|crit|error)\]'] },
      { warn  = ['\[warn\]'] },
    ]
    generic = [
      { error = ["(?i)\\b(error|fatal|exception)\\b"] },
      { warn  = ["(?i)\\bwarn(ing)?\\b"] },
    ]
```

### Measuring noise before dropping it

Dropping chatty messages saves storage, but it is worth knowing how much is
dropped and where it comes from. Here known noise is dropped, and the counts
per host show which sources produce it. The counts go to InfluxDB, while the
messages themselves go to a log store:

```toml
[[processors.classify]]
  selector_tag = "hostname"
  default_regex_group = "all"
  match_field = "message"
  default_category = "kept"
  drop_categories = ["noise"]
  result_tag = "classified"

  aggregation_period = "1m"
  aggregation_measurement = "log_noise"
  aggregation_total_field = "total"
  aggregation_selector_tag = "host"
  aggregation_selector_fields = ["noise", "kept", "total"]

  [processors.classify.mapped_selector_regexes]
    all = [
      { noise = """
          session opened for user
          session closed for user
          Started Session \\d+ of user
          CRON\\[\\d+\\]
      """ },
    ]

[[outputs.influxdb_v2]]
  namepass = ["log_noise"]
  # ...

[[outputs.loki]]
  namedrop = ["log_noise"]
  # ...
```

Every minute, one metric per host is emitted:

```text
log_noise,host=web1 noise=412i,kept=38i,total=450i 1700000040000000000
```

### States for a monitoring system

The category names can be the states another system expects. The
[GroundWork output][groundwork] reads the service status from the `status`
tag, so classifying syslog messages into GroundWork statuses turns them into
service states:

```toml
[[processors.classify]]
  selector_tag = "hostname"
  selector_mapping = [
    { "^fire\\d{3}$" = "firewall" },
  ]
  default_regex_group = "database"
  match_field = "message"
  drop_categories = "ignore"
  result_tag = "status"

  [processors.classify.mapped_selector_regexes]
    database = [
      { ignore                       = "DB client connected" },
      { SERVICE_OK                   = "Database is starting up" },
      { SERVICE_WARNING              = 'Tablespace \w+ free space is low' },
      { SERVICE_UNSCHEDULED_CRITICAL = "Database is shutting down" },
      { SERVICE_UNKNOWN              = ".*" },
    ]
    firewall = [
      { ignore                       = "snort.+Priority: 3" },
      { SERVICE_WARNING              = "snort.+Priority: 2" },
      { SERVICE_UNSCHEDULED_CRITICAL = "snort.+Priority: 1" },
      { SERVICE_UNKNOWN              = ".*" },
    ]
```

The catch-all last category makes sure every message that is not dropped gets
a valid state. A category name the receiving system does not accept is
ignored there, so keep the names in line with what it expects.

[groundwork]: ../../outputs/groundwork/README.md
