# Classify Processor Plugin

The `classify` plugin classifies metrics by matching a designated tag or field
value against groups of regular expressions. Each input metric either passes
through with a new result tag/field set to the matched category name, or is
dropped. Apart from the result tag or field, which replaces any existing tag
or field of the same name, the metric is not modified.

The plugin optionally supports a selector mapping step: a tag or field value
is mapped to the name of a regex group, allowing distinct sets of regexes to
be applied to different classes of input without repeating configuration.

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
SNMP trap description, into one of a few named categories. Typical uses, shown
in the [examples](#examples):

- bounding tag cardinality, by replacing unbounded values with a few classes;
- normalizing severities or log levels that each source words differently;
- telling known noise apart, to drop it or to count it;
- deriving the states that an alerting or monitoring system expects.

The stock plugins cover parts of this, but not the combination:

| Need | `classify` | Closest stock alternative |
| --- | --- | --- |
| Map a value to one of a few names by trying an ordered list of regexes, first match wins | Categories are tried in order; each may hold any number of regexes; the category name is written as is | `processors.regex` runs every conversion in turn, so the last match wins and a catch-all must come first; each regex is its own block with its own replacement. `enum` and `lookup` only match exact values. |
| Fall back to a default state when nothing matches | `default_category` | `processors.defaults` only fills a missing tag or field, so it needs the regex result written to a separate key first. |
| Apply different rules to different kinds of sources | One `selector_mapping` picks a regex group per metric, with a fallback group | Several processor instances, each limited with `tagpass` globs. Telegraf metric filters cannot express "every source the other instances did not take", and every instance repeats the shared rules. |
| Drop uninteresting messages as part of the same decision | `drop_categories`, or no match without `default_category` | `processors.filter` matches tag values with globs only; `metricpass` can test a regex, but as a separate rule set that duplicates the classification. |

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
4. When a `"*"` value is not a group in `mapped_selector_regexes`,
   `default_regex_group` is used, or the metric is dropped when it is not set.

Patterns are unanchored, so `pg\d{3}` also matches `xpg1234`; anchor them
with `^` and `$` to match the whole value. To drop metrics whose selector is
empty, or all selectors not listed, end the mapping with a catch-all:

```toml
selector_mapping = [
  { pattern = '^fire\d{3}$', group = "firewall" },
  { pattern = '^pg\d{3}$', group = "database" },
  { pattern = '^$', group = "" },  # empty selector: drop
  { pattern = '.*', group = "" },  # anything else: drop
]
```

A group name in `selector_mapping` other than `"*"` and `""` must be defined
in `mapped_selector_regexes`, otherwise the plugin fails to start.

Without a selector, every metric uses `default_regex_group`.

### Choosing the category

The match item is read from `match_tag`, or from `match_field`, which must
hold a string; a metric without it is dropped. The categories of the selected
group are then tried in the order they are listed, and within a category its
patterns are tried in order; the first pattern that matches decides the
category.

- When nothing matches, `default_category` is used, or the metric is dropped
  when it is not set. A last category with the pattern `.*` has the same
  effect.
- A metric whose category is in `drop_categories` is dropped.
- Otherwise the category name is written to `result_tag` or `result_field`,
  replacing any value already there, and the metric passes downstream.

## Example output

```text
syslog,host=pg123,status=warning message="Tablespace users free space is low" 1700000000000000000
```

## Configuration

```toml @sample.conf
# Classify metrics by matching a tag or field against ordered regexes
[[processors.classify]]
  ## Tag or field selecting the regex group; set at most one of them.
  # selector_tag = ""
  # selector_field = ""

  ## Ordered selector patterns, the first match picks the group.
  ## Group "*" uses the selector value as the group, "" drops the metric.
  # selector_mapping = [
  #   { pattern = 'fire\d{3}', group = "firewall" },
  #   { pattern = 'pg\d{3}', group = "database" },
  # ]

  ## Group used when no selector pattern matches; required without a selector.
  default_regex_group = "database"

  ## Tag or field matched against the category patterns; set exactly one.
  match_field = "message"
  # match_tag = ""

  ## Category used when no pattern matches; metrics are dropped if empty.
  # default_category = ""

  ## Categories whose metrics are dropped.
  # drop_categories = []

  ## Tag or field the category is written to; set exactly one.
  result_tag = "status"
  # result_field = ""

  ## Ordered categories per group, the first matching category wins.
  [processors.classify.mapped_selector_regexes]
    database = [
      { category = "ignore", patterns = ["DB client connected"] },
      { category = "critical", patterns = ["Database is shutting down"] },
      { category = "unknown", patterns = [".*"] },
    ]
    firewall = [
      { category = "warning", patterns = ["login attempt"] },
      { category = "critical", patterns = ["intrusion detected"] },
    ]
```

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
  drop_categories = ["health"]
  result_tag = "route"

  [processors.classify.mapped_selector_regexes]
    paths = [
      { category = "health", patterns = ["^/healthz", "^/ready"] },
      { category = "api", patterns = ["^/api/"] },
      { category = "static", patterns = ['\.(css|js|png|svg|woff2?)$'] },
      { category = "login", patterns = ["^/(login|logout|oauth)"] },
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
    { pattern = "^postgres", group = "postgres" },
    { pattern = ".*", group = "*" },
  ]
  default_regex_group = "generic"
  match_field = "message"
  default_category = "info"
  result_tag = "level"

  [processors.classify.mapped_selector_regexes]
    postgres = [
      { category = "error", patterns = ["^(ERROR|FATAL|PANIC):"] },
      { category = "warn", patterns = ["^WARNING:"] },
    ]
    nginx = [
      { category = "error", patterns = ['\[(emerg|alert|crit|error)\]'] },
      { category = "warn", patterns = ['\[warn\]'] },
    ]
    generic = [
      { category = "error", patterns = ['(?i)\b(error|fatal|exception)\b'] },
      { category = "warn", patterns = ['(?i)\bwarn(ing)?\b'] },
    ]
```

### Measuring noise before dropping it

Dropping chatty messages saves storage, but it is worth knowing how much is
dropped and where it comes from. Here the plugin only marks known noise in the
`class` field. The `valuecounter` aggregator counts the classes per host,
since aggregators see every metric before the outputs filter them, and the log
store output drops the noise:

```toml
[[processors.classify]]
  default_regex_group = "all"
  match_field = "message"
  default_category = "kept"
  result_field = "class"

  [processors.classify.mapped_selector_regexes]
    all = [
      { category = "noise", patterns = [
          "session opened for user",
          "session closed for user",
          'Started Session \d+ of user',
          'CRON\[\d+\]',
      ] },
    ]

[[aggregators.valuecounter]]
  period = "1m"
  fields = ["class"]
  taginclude = ["host"]
  name_suffix = "_classes"

[[outputs.influxdb_v2]]
  namepass = ["syslog_classes"]
  # ...

[[outputs.loki]]
  namedrop = ["syslog_classes"]
  metricpass = 'fields["class"] != "noise"'
  # ...
```

Every minute, one metric per host is emitted:

```text
syslog_classes,host=web1 class_noise=412i,class_kept=38i 1700000040000000000
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
    { pattern = '^fire\d{3}$', group = "firewall" },
  ]
  default_regex_group = "database"
  match_field = "message"
  drop_categories = ["ignore"]
  result_tag = "status"

  [processors.classify.mapped_selector_regexes]
    database = [
      { category = "ignore", patterns = ["DB client connected"] },
      { category = "SERVICE_OK", patterns = ["Database is starting up"] },
      { category = "SERVICE_WARNING", patterns = ['Tablespace \w+ free space is low'] },
      { category = "SERVICE_UNSCHEDULED_CRITICAL", patterns = ["Database is shutting down"] },
      { category = "SERVICE_UNKNOWN", patterns = [".*"] },
    ]
    firewall = [
      { category = "ignore", patterns = ["snort.+Priority: 3"] },
      { category = "SERVICE_WARNING", patterns = ["snort.+Priority: 2"] },
      { category = "SERVICE_UNSCHEDULED_CRITICAL", patterns = ["snort.+Priority: 1"] },
      { category = "SERVICE_UNKNOWN", patterns = [".*"] },
    ]
```

The catch-all last category makes sure every message that is not dropped gets
a valid state. A category name the receiving system does not accept is
ignored there, so keep the names in line with what it expects.

[groundwork]: ../../outputs/groundwork/README.md
