# Configuration

## Telegraf-level plugin configuration

The `classify` plugin needs the full expressiveness of TOML v1.0.0 for
its configuration data.  To make that available, the detailed setup
for this plugin is moved to a separate file, and the standard Telegraf
configuration for this plugin is just a single option.

### `classify_config_file`

The path to a file containing the full plugin configuration data as
described below.  By convention, we use a `.toml` filename extension
on that file, so we can park it alongside the other Telegraf config
files but have Telegraf itself ignore the file.

With that in mind, here is the full standard plugin setup as will be
seen in the `telegraf.conf` file:

```toml
# Classify Telegraf data points according to user-specified rules.
[[processors.classify]]
  ## The detailed configuration data for the classify plugin lives
  ## in a separate file, whose path is given here.
  classify_config_file = '/etc/telegraf/telegraf.d/classify.toml'
```

## Detailed configuration options

This section describes what is in the secondary config file pointed
to by the main-plugin-configuration `classify_config_file` option.
Additional information about how to set these options is given in the
comments in the sample configuration later in this document.

### `selector_tag`, `selector_field`

The name of the tag or field ("selector item") to be used to determine
the `mapped_selector_regexes` group to use for regex matching.

These two options are mutually exclusive.  If neither one is defined,
the `selector_mapping` is irrelevant, and the separately configured
`default_regex_group` value will be used unconditionally as the
`mapped_selector_regexes` group to use for regex matching.

### `selector_mapping`

The mapping of selector values to regex-group names.  It is applied
according to the following rules:

* If the selector item is missing from the input data point, this data
point will be dropped silently (except for being counted as such in
aggregation statistics).

* If this mapping is not defined at all, or if the selector item's
value matches a key in the mapping and the mapped value is the
special string `*`, the selector item's value will be used unmodified
as the mapped value.  That means it will generally be the chosen
`mapped_selector_regexes` group name, unless there is no such group name;
see the special handling described below for that corner case.

  ```toml
  selector_mapping = [
    { 'unmodified_value' = '*' },
  ]
  ```

* If this mapping is defined, the selector item's value matches a key
in the mapping, and the mapped value is an empty string, this data
point will be dropped silently (except for being counted as such in
aggregation statistics).

  ```toml
  selector_mapping = [
    { 'dropped_value' = '' },
  ]
  ```

  Speaking of dropping input data points, if there is any possibility
  of having the selector item be present in an input data point but with
  an empty string as its value, you should consider what you want to do
  in that situation.

  * You can explicitly declare that input data points with an empty
  string as the selector item value should be dropped silently (except
  for being counted as such in aggregation statistics).  This would
  probably be the most common choice.

    ```toml
    selector_mapping = [
      { '^$' = '' },
    ]
    ```

  * You can explicitly declare that empty selector item values should
  map to some particular non-empty regex-group name:

    ```toml
    selector_mapping = [
      { '^$' = 'my_regex_group' },
    ]
    ```

  * You can allow such an input data point to be defaulted as not
  matching any key in the `selector_mapping`, and have the later rule
  kick in that will use the value of the `default_regex_group` option
  as the selector mapping value.

* If this mapping is defined, the selector item's value matches a key
in the mapping, and the mapped value is present in the configuration as
the name of a `mapped_selector_regexes` group, that group will be used
for regex matching.

* Under any of the following conditions, the value of the
`default_regex_group` option will be used as the selector mapping value.

  * Some `selector_mapping` mappings are defined, but the selector item's
  value does not match any key in the mapping.

  * The mapped selector item value, however that was calculated, does
  not match the name of any `mapped_selector_regexes` group.

  This design decision (using the value of the `default_regex_group`
  option) is intended mostly to avoid situations where an unintended
  misconfiguration (e.g., a typo) might otherwise go unnoticed because
  the data point would be dropped.  If you don't want that behavior,
  you have several choices:

  * Extend the set of mapping keys to cover the cases you are concerned
  with.

  * Define the `default_regex_group` option to be an empty string. or
  equivalently just leave it undefined, so those data points will be
  intentionally dropped.

  * Put in place a general wildcard rule at the end of the
  selector_mapping`` that maps all other selector values to an empty
  string, so those data points will be intentionally dropped.

    ```toml
    selector_mapping = [
      { 'fire\d{3}' = 'firewall' },
      { 'pg789'     = 'database' },
      { 'rout846'   = 'router'   },
      { '.*'        = ''         },
    ]
    ```

* If the value of the `default_regex_group` option is needed under
the previous rule but that option is not defined, it will be treated
as though it were defined as an empty string.  Which means, this data
point will be dropped silently (except for being counted as such in
aggregation statistics).

The `selector_mapping` keys are regular expressions, not just literal
strings.  Thus for example if you are using hostnames as selectors,
and you have cleverly chosen the names of all your firewall machines
from the set `fire000` through `fire999`, you could use the following
as one of your mapping elements and not have to deal with the individual
machine names.

```toml
selector_mapping = [
  { 'fire\d{3}' = 'firewall' },
]
```

Because regexes are allowed as mapping keys, not just fixed literal
strings, there is the possibility that a given selector item value might
match more than one key in the mapping.  To eliminate that ambiguity,
the format of the `selector_mapping` option has been chosen to specify a
particular order in which the entries will be used at runtime to attempt
to match a given input data point.  The `classify` plugin will always
attempt selector item value matching using patterns in order as you list
them.  This gives you complete control over which pattern will match first
and thereby provide the regex group name for the next phase of processing.

```toml
## To prevent ambiguity in the order of matching attempts, only
## one key=value setting is allowed in each inline table which is
## enclosed in braces.  The TOML config-file syntax demands that
## everything in each inline table must appear on a single line.
selector_mapping = [
  { 'fire\d{3}' = 'firewall' },
  { 'ora456'    = 'database' },
  { 'pg789'     = 'database' },
  { 'rout237'   = 'router'   },
  { 'rout846'   = 'router'   },
]
```

This ordering allows you to provide a final wildcard entry.  So for
instance if you want all otherwise unmapped selector item values to
have the input data point be matched against the `compute` group of
regexes, without depending on the rule noted above that the value of
the `default_regex_group` option will be used in this case, specify a
corresponding wildcard match as the last `selector_mapping` entry in
the configuration:

```toml
selector_mapping = [
  { 'desk\d{3}' = 'desktop' },
  { 'file\d{3}' = 'storage' },
  { 'tape\d{3}' = 'storage' },
  { 'back\d{3}' = 'backups' },
  { 'wire\d{3}' = 'network' },
  { '.*'        = 'compute' },
]
```

### `default_regex_group`

A regex group name to be used as the value of applying the
`selector_mapping` under any of the following conditions:

* Neither the `selector_tag` option nor the `selector_field` option
is defined.

* No `selector_mapping` match to an input data point's selector item
value is found.  This includes the case where the selector item value
is an empty string and you do not have any `selector_mapping` entry that
matches an empty string.

* The mapped selector item value does not match the name of any
`mapped_selector_regexes` group.

If the `default_regex_group` option is needed for a given input data
point, that data point will be dropped (and counted as such in aggregation
totals) under any of the following conditions:

* The `default_regex_group` option is not defined.

* The `default_regex_group` option is defined as an empty string.

* The `default_regex_group` option is defined as a non-empty string, but
its value does not match the name of any `mapped_selector_regexes` group.

### `match_tag`, `match_field`

The name of the tag or field ("match item") within the input data point
to be used to match against regular expressions.

These two options are mutually exclusive.  One of them must be specified,
as otherwise there will be no way to calculate the classification of
input data points.

### `mapped_selector_regexes`

An ordered group of category-to-regexes mappings, each of which defines an
ordered array of regular expressions.  You will define one or more such
groups in the plugin configuration.  Names of the groups are chosen by
the user, and are not predefined by this plugin.  Those names correspond
to the values specified in the `selector_mapping` mapping.

All the mappings for a given group, taken together, represent successive
patterns to be used for matching if the mapped selector value is the name
of that group.  Each group can contain zero or more category-to-regexes
mappings, from which the regexes are pattern-matched against the match
item (`match_tag` or `match_field`) in order as seen in the configuration
of the group.

Each category-to-regexes mapping can contain zero or more regular
expressions, expressed in any of several possible forms as shown in the
sample setup toward the end of this document.  A match on any one of
the regular expressions in this set will end the classification logic
for this data point, with the result being the name of this category.
Names of the categories are chosen by the user, and are not predefined
by this plugin.  Typically, you will use the same set of categories for
every `mapped_selector_regexes` group.  Failure to match any of the
regular expressions in a given category-to-regexes mapping means the
logic will proceed to attempt matching of the next such mapping defined
in the same `mapped_selector_regexes` group.  If matching attempts
reach the end of all the category-to-regexes mappings in the selected
`mapped_selector_regexes` group and no match has been achieved, this
data point will be dropped silently (except for being counted as such
in aggregation statistics) unless the `default_category` is defined as
a non-empty string and the `default_category` value is not listed in the
`drop_categories`.

The config-file syntax for specifying each category-to-regexes mapping
is a bit touchy, due to limitations of the TOML config-file language.
In particular, no newlines are allowed anywhere between the opening and
closing braces that enclose each individual mapping, except for those
that you can include within a set of regexes in the multi-line format,
which is shown just below.  In this regard, follow the examples carefully.

In the multi-line format for specifying a set of category regexes, all
leading and trailing whitespace in each regex line will be automatically
trimmed and not be part of the regex, and blank lines within a given
set of regexes will be ignored.  That allows you to visually separate
distinct clusters of related regexes, making for easier understanding
and maintenance.  Here is an example:

```toml
mapped_selector_regexes.firewall = [
  { ignore = '''
      snort.+Priority: 3

      snort.+portscan.+192.168.3
      snort.+portscan.+ff02

      snort.+128:4:1.+spp_ssh.+192.168.3
      snort.+136:1:1.+spp_reputation.+Priority: 2.+192.168.[37]
      snort.+140:3:2.+spp_sip.+URI is too long
  ''' },
  { okay = '''
  ''' },
  { warning = '''
      snort.+Priority: 2
  ''' },
]
```

### `default_category`

A category name to be used as the value of the result item if
regular-expression matching was in fact attempted but no pattern match
succeeded.  If the `default_category` is not defined, or is defined
as an empty string, data points having no regex match will be dropped
(and counted as such in aggregation totals).  The category named by this
option is subject to filtering by being mentioned in `drop_categories`,
just as any successful-regex-match category would be.

Note that the `default_category` is not used as the value of the result
item if the classification logic decided that no regular-expression
matching could be attempted and the input data point was dropped.  So for
example if the `selector_mapping` does not end up generating a regex
group name in the `mapped_selector_regexes` and the `default_regex_group`
also has the same fate, the input data point is doomed.

One use for this option would be to name a "normal", "don't care", or
other anodyne category, to be applied if all unmatched data points are
to be treated as unexceptional.  The contrary use is to have this option
name a serious-state category, so you can be notified when something
happens that you have never seen before.

### `drop_categories`

Zero or more regex-match category names wherein if the regex matching
result is one of those categories, however that matching result was
calculated, that data point will be dropped from the plugin output.  Such
a data point will be counted in the aggregation statistics both under the
respective match category, and under the named `aggregation_dropped_field`
if that is defined.

```toml
drop_categories = 'ignore'
```

### `result_tag`, `result_field`

The name of the tag or field where the classification result ("result
item") is stored, if the data point is not dropped.  The result item
value will either be the name of the first category one of whose regexes
pattern-matched the match item value, or the value of `default_category`
if no such match was found.  This item will be added to the data point
if not present, or overwritten if already present.

These two options are mutually exclusive.  One of them must be specified,
as otherwise there will be no way to report out the classification of
input data points.

### `aggregation_period`

How often to spill out the aggregation counters as a measurement separate
from the data points which are streaming through this plugin.  If left
undefined or set to a non-positive value, aggregation will be disabled.
Specified as a number with a trailing letter for time units (`s`,
`m`, `h`).  This field is required if you wish to output any kind of
aggregated statistics.

```toml
aggregation_period = '10m'
```

### `aggregation_measurement`

The name of the measurement that will be used to record aggregated
classification statistics that are sent downstream.  This one
measurement name will be used for all generated aggregation-summary,
aggregation-by-group, and aggregation-by-selector data points.  This field
is required if you wish to output any kind of aggregated statistics.

```toml
aggregation_measurement = 'message_counts'
```

### `aggregation_dropped_field`

The name of a field that may be mentioned in `aggregation_summary_fields`
or `aggregation_group_fields` or `aggregation_selector_fields` to report
the number of data points that were dropped for any reason whatsoever
during their transit through this plugin.  There is only one such field
name supported for all of those types of aggregation statistics, though
like category names, the count it represents varies slightly depending
on which type of aggregation is being reported.

```toml
aggregation_dropped_field = 'dropped'
```

### `aggregation_total_field`

The name of a field that may be mentioned in `aggregation_summary_fields`
or `aggregation_group_fields` or `aggregation_selector_fields` to report
the total number of data points that were processed through this plugin,
whether or not they were ultimately dropped or output.  There is only
one such field name supported for all of those types of aggregation
statistics, though like category names, the count it represents varies
slightly depending on which type of aggregation is being reported.

```toml
aggregation_total_field = 'total'
```

### `aggregation_summary_tag`

The name of a tag to be used when reporting aggregation-summary level
statistics, representing the full volume of input data points during
each `aggregation_period` without any breakdown into finer granularity.
This field is required if you wish to output that level of statistics.
Its purpose is to provide a handle for querying data and retrieving only
the summary-level statistics.

```toml
aggregation_summary_tag = 'severity'
```

### `aggregation_summary_value`

The single fixed value of the `aggregation_summary_tag` to be used when
reporting aggregation-summary level statistics.  This field is required
if you wish to output that level of statistics.

```toml
aggregation_summary_value = 'all'
```

### `aggregation_summary_fields`

A list of the fields that should appear in each aggregation-summary
output data point.  This field is required if you wish to output
that level of statistics.  These fields may be the names of any of
the regex categories defined in `mapped_selector_regexes`, or the
field named by `aggregation_dropped_field`, or the field named by
`aggregation_total_field`.

```toml
aggregation_summary_fields = [
    'ignore', 'okay', 'warning', 'critical', 'unknown',
    'dropped', 'total'
]
```

### `aggregation_group_tag`

The name of the tag to attach to each aggregated-data output data point
that bins input data points firstly by which regex group the selector
mapped to.  This field is required if you wish to output that level of
statistics.  At the end of each `aggregation_period`, one output data
point is produced for each such group that had at least one input data
point mapped to that group during that period.  The value of this tag
will be the name of the regex group whose counts are summarized in that
output data point.

```toml
aggregation_group_tag = 'host_type'
```

### `aggregation_group_fields`

A list of the fields that should appear in each aggregation-by-group
output data point.  This field is required if you wish to output
that level of statistics.  These fields may be the names of any of
the regex categories defined in `mapped_selector_regexes`, or the
field named by `aggregation_dropped_field`, or the field named by
`aggregation_total_field`.

```toml
aggregation_group_fields = [
    'okay', 'warning', 'critical', 'unknown', 'dropped', 'total'
]
```

### `aggregation_selector_tag`

The name of the tag to attach to each aggregated-data output data point
that bins input data points firstly by which selector value was present
in the input data point.  This field is required if you wish to output
that level of statistics.  At the end of each `aggregation_period`, one
output data point is produced for each selector value that had at least
one input data point with that value during that period.  The value of
this tag will be that selector value, corresponding to the counts which
are summarized in that output data point.

```toml
aggregation_selector_tag = 'host'
```

### `aggregation_selector_fields`

A list of the fields that should appear in each aggregation-by-selector
output data point.  This field is required if you wish to output
that level of statistics.  These fields may be the names of any of
the regex categories defined in `mapped_selector_regexes`, or the
field named by `aggregation_dropped_field`, or the field named by
`aggregation_total_field`.

```toml
aggregation_selector_fields = [
    'okay', 'warning', 'critical', 'unknown', 'dropped', 'total'
]
```

### `aggregation_includes_zeroes`

Whether or not to include fields for categories that have a zero count in
an aggregation data point.  By default (if this option is not defined),
such fields are not included in aggregation data points, to reduce
downstream load and storage requirements.  If you want explicit zeroes
to show up for such categories (though only when there is at least one
non-zero counter to force out the aggregation data point), define the
`aggregation_includes_zeroes` option to be `true`.

```toml
aggregation_includes_zeroes = false
```

## Sample detailed configuration

The following example setup shows how you might classify syslog
messages, as described in the [README.md](README.md) file.  This is a
sample `classify.toml` file which would be referred to by the plugin
`classify_config_file` option.

```toml
## Classify telegraf data obtained from the syslog plugin according to
## sets of regular expressions defined here.
##
## Inasmuch as almost every option for the classify plugin has no default
## value, we don't show any option defaults in this sample configuration.
## All settings here simply reflect an extended example showing how this
## plugin might be configured for processing syslog messages.

## The input data point field which discriminates which set of
## regexes to attempt to match.
selector_tag = 'host'

## There are too many individual hosts sending messages our way
## for us to specify corresponding sets of regexes at that level
## of granularity.  That would be excessive administrative burden.
## Instead, we map the hostnames to groups of regexes that reflect
## the functionality of each kind of host.  That way, the desired
## commonality of processing is reflected in the reduced setup.
selector_mapping = [
  { 'fire\d{3}' = 'firewall' },
  { 'ora456'    = 'database' },
  { 'pg789'     = 'database' },
  { 'rout237'   = 'router'   },
  { 'rout846'   = 'router'   },
]

## The regex group to use if selector item value matching does not
## yield a match.  In this example setup, we just want to drop such
## points, so we can either leave this option undefined or define it
## as an empty string.
# default_regex_group = ''

## The Telegraf syslog input plugin documentation is remiss in
## not clearly documenting that the syslog text message will
## appear in the "message" field.  That field should be listed
## in the "Metrics" section of that documentation.
match_field = 'message'

## This simple form of specifying a regex as a TOML literal string
## can be used if you have only one regex to define in a particular
## category.  An empty string used as a regex here would specify
## that no matching is to be attempted for that category, while
## documenting that fact explicitly.
mapped_selector_regexes.database = [
  { ignore   = 'DB client connected' },
  { okay     = 'Database is starting up' },
  { warning  = 'Tablespace \w+ free space is low' },
  { critical = 'Database is shutting down' },
  { unknown  = '.*' },
]

## If you specify regular expressions in the following form, each
## regex occupies one line within a TOML multi-line literal string,
## and all leading and trailing whitespace in each regex line will be
## automatically trimmed and not be part of the regex.  If you need to
## match some whitespace at the start or end of your regex, consider
## using \040 or \o{040} or \x{20} or [ ] (all of which represent
## a single space character), or \t (a single tab character) or \s
## (generalized whitespace) or \h (horizontal whitespace), whatever
## fits your needs.  Or switch to using the other form for specifying
## multiple regular expressions in a single category, shown below.
## You don't have to use the same form for all categories in a group.
mapped_selector_regexes.firewall = [
  { ignore = '''
      snort.+Priority: 3
      snort.+portscan.+192.168.3
      snort.+portscan.+ff02
      snort.+124:3:2.+smtp.+Attempted response buffer overflow:.+192.168.3.11
      snort.+128:4:1.+spp_ssh.+192.168.3
      snort.+136:1:1.+spp_reputation.+Priority: 2.+192.168.[37]
      snort.+140:3:2.+spp_sip.+URI is too long
      snort.+140:8:2.+spp_sip.+Classification: Potentially Bad Traffic.+192.168.7
      snort.+1:49666:2.+SQL HTTP URI blind injection attempt.+192.168.3
      snort.+1:21516.+SERVER-WEBAPP JBoss JMX console access attempt.+192.168.3
  ''' },
  { okay = '''
  ''' },
  { warning = '''
      snort.+Priority: 2
  ''' },
  { critical = '''
      snort.+Priority: 1
  ''' },
  { unknown = '''
      .*
  ''' },
]

## The following format must be used if a regex contains three
## consecutive single-quote characters, since that construction
## is not allowed in the TOML multi-line literal strings used in
## the previous format.  We use single-quoted TOML literal strings
## here because that way, all content between the single-quote
## delimiters is interpreted as-is without modification.  That
## eliminates the need to escape some characters in the regex
## simply because the regex is being stored in a TOML-format file.
## Also, the TOML processing won't try to itself interpret any
## escape sequences that are part of your regex.  That said, there
## is no way to write a single-quote character as part of a regex
## expressed as a single-quoted string.  If you need that, use
## the multi-line form shown above, or use a double-quoted
## string instead, keeping in mind that you will then need to
## backslash-escape all backslash and double-quote characters
## in your regex, including backslashes used to specify a special
## character such as \t (a tab character).
mapped_selector_regexes.router = [
  { ignore = [ 'SYS-5-CONFIG_I: Configured from console', ] },
  { okay = [ 'TRACK-6-STATE: 100 ip sla 1 reachability Down -> Up', 'TRACK-6-STATE: 200 ip sla 2 reachability Down -> Up', ] },
  { warning = [ ] },
  { critical = [ 'TRACK-6-STATE: 100 ip sla 1 reachability Up -> Down', 'TRACK-6-STATE: 200 ip sla 2 reachability Up -> Down', ] },
  { unknown = [ '.*' ] },
]

## For didactic purposes, all our regex groups below explicitly define
## an all-inclusive match pattern for category "unknown" at the end of
## the group, so setting this option is not required in this sample
## configuration.  Without that, we would have wanted to set this
## option here.
# default_category = 'unknown'

## The set of result categories that should be counted in the
## aggregate statistics but otherwise have their input data points
## dropped and not output from this plugin.  May be specified
## either as a single 'string' (for a single such category) or as an
## [ 'array', 'of', 'strings' ].  If all matched data points should
## be passed through and no categories should be dropped, either
## leave this option undefined, or define it as an empty array.
drop_categories = 'ignore'

## How to label the end result of the classification processing.
result_field = 'status'

## Define these options as desired, to enable corresponding types
## of aggregated-statistics output.  If the full set of options
## needed for a particular type of aggregation reporting is not
## defined, or not defined with sufficient information, that
## aggregation type will be suppressed.
aggregation_period = '10m'
aggregation_measurement = 'aggregated_status'
aggregation_dropped_field = 'dropped'
aggregation_total_field = 'total'
aggregation_summary_tag = 'summary'
aggregation_summary_value = 'full'
aggregation_summary_fields = [
  'ignore', 'okay', 'warning', 'critical', 'unknown',
  'dropped', 'total'
]
aggregation_group_tag = 'host_type'
aggregation_group_fields = [
  'ignore', 'okay', 'warning', 'critical', 'unknown',
]
aggregation_selector_tag = 'host'
aggregation_selector_fields = [
  'ignore', 'okay', 'warning', 'critical', 'unknown',
]
aggregation_includes_zeroes = false
```
