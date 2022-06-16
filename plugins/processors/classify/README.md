# Classify Processor Plugin

[//]: # (
NOTE:
Viewing this text with any formatter other than the GitHub API for turning
this text into HTML might not produce correct results.
That's because Markdown is a horrifyingly bad formatting language.
We did the best we could with the features that are available,
and then we had to move outside of raw Markdown and use HTML in some places.
That gets us reasonable formatting of some things like definition lists,
that are not possible to format sensibly in raw Markdown.
So, don't blame us for the formatting you see; it's the tool that's at fault.
)

The `classify` plugin makes decisions based on comparison of designated string tag and field values against groups of
regular-expression patterns.  The result of those decisions is either that an input data point is dropped, or a new tag
or field is created showing the classification of the input data.  The original input data is always left undisturbed;
no replacement substitutions are carried out by the regular expression matching.

While performing input data classification, the `classify` plugin supports limited transient internal tag/field value
mapping.  This allows consolidation of duplicate sets of regular expressions in the configuration.  The original tag/field
value is left undisturbed, and the mapped value does not appear in the plugin output.

The tag/field mapping is included in the `classify` plugin because otherwise one would need to invoke the `enum` plugin
beforehand to perform value mapping, and the `fielddrop` or `tagexclude` option would need to be applied to the `classify`
plugin to strip out the mapped values needed only on a transient basis for the classification logic.  The extra parsing and
re-serialization of the data that would be invoked by using the `enum` plugin in the pipeline seems like pointless overhead.

In addition to those actions, the `classify` plugin supports accumulation of classification statistics and periodic output
of those statistics.  In this sense, the plugin acts as an aggregator, but only appending extra data to the stream without
altering the output of classification results.

The statistics aggregation is included in the `classify` plugin because of the following bug in Telegraf 1.x:

>  aggregators should not re-run processors
>  <br>
>  https://github.com/influxdata/telegraf/issues/7993

That issue is slated to be addressed in Telegraf 2.0, but there is no timeline for its release.  In the meantime, we
don't want the `classify` plugin to be invoked a second time after classification statistics are collected.

## Motivation

The text below describes the processing model in some detail, which upon initial reading may be a bit confusing.  It may
help to know that this plugin was developed in support of processing syslog messages from multiple hosts that may or
may not share the same computational purpose and therefore may or may not generate the same types of syslog messages.
We want to support commonality of configuration where that is appropriate, while allowing distinct configuration where
that is needed.  Hence there is a bit of indirection at the beginning of the processing stream, to identify the kind of
regular expressions to be applied to a given input data point.

Beyond that initial identification, we wanted to be able to recognize a variety of different messages that might appear
from a given source, classifying them into just a few result categories.  So the overall logical structure of the regular
expressions that can be configured is:

```
regex_group_name
    category_name
        regex
        regex
        regex
    category_name
        regex
        regex
regex_group_name
    category_name
        regex
        regex
        regex
    category_name
        regex
        regex
```

with as many items at each level of the hierarchy (`regex_group_name`, `category_name`, `regex`) as needed.
To that end, we wanted to make the specification of that hierarchy as compact as possible in the config file,
so the administrator does not get lost in the surrounding syntax and can concentrate on the task at hand.

Given this model of processing, it seemed sensible to implement it in a general way, not specifically tied to syslog messages.
So the abstract model of processing is couched in other terms, in the hope that this plugin may find use in other contexts.
The sample setup at the end may help to clarify actual practical application of this plugin.

## Comparison to the `regex` processor:

* An initial calculation is run to dynamically select amongst multiple sets of regular expressions.
* Groups of related regular expressions are easier to specify, instead of one-at-a-time setup.
* A single tag or field is created, representing the result of the classification.
* No tags or fields are overwritten, unless the configured result tag/field is already present in the incoming data.
* Aggregate processing statistics are emitted on a regular basis, separate from passing through or dropping classified
  data points.

## Abstract processing model

Here is a top-level view of how this plugin works, showing its essential simplicity.

![Processing Model](processing_model.png)

* Identify a selector item (tag or field), to discriminate between possible groups of regular expressions that might be
  applied to an input data point.

* Map the selector item's value to the name of a group of sets of regular expressions to match against an input data point.

* Identify a match item (tag or field) that is to be matched against the regular expressions chosen by the mapped selector
  item's value.

* Identify a result item (tag or field) that is to be added to the data point and contain the classification result.

* Step through each set ("category") of regular expressions in the selected group of such sets, matching each regular
  expression in turn against the match item's value.  The first match that succeeds determines the final classification.
  The result value will be the name of the category that the matching regex belongs to.

### Simplified pseudo-code

_At plugin startup:_
* analyze the configuration
* create some corresponding aggregation counters (others will depend on details of the input data points, and be dynamically
  created as data is processed)
* set all those counters to zero
* initialize the required thread-synchronization objects
* start the aggregation thread
* start the processing thread

_Processing thread, executed once for each input data point:_
```
read data point
regex_group_name = {map selector value}
foreach regex_category in regexes[regex_group_name] {
    foreach regex in regexes[regex_group_name][regex_category] {
        if match_item matches regex {
            result = regex_category
            add result to this data point
            update aggregation statistics (if configured), in a
                manner synchronized with the aggregation thread
            exit all loops for matching this data point
        }
    }
}
write out this data point
```

_Aggregation thread, operating as a background task, at the end of each configured period:_
* synchronize access to the aggregration counters with the processing thread to prevent race conditions
* spill out measurements containing all of the configured aggregation counters
* zero out all the aggregation counters

_At plugin shutdown:_
* synchronize access to the aggregration counters with the processing thread to prevent race conditions
* spill out all aggregation counts collected since the last spill action
* zero out all the aggregation counters (pro forma)

## Aggregated classification statistics

Aggregation statistics are an invention of this plugin, meaning the details of their construction must be specified by
the configuration so these measurements have a form which is acceptable to whatever output plugin(s) you use.  As part of
that setup, the configuration must specify the measurement name as the principal component of the aggregated statistics,
since it might be different from the measurement name used by the input data points.

There are three types of aggregation-statistics output that can be produced, depending on how you configure this plugin.
For any of these types, if all of the fields to be reported in a given individual aggregation-data point are zero, that
data point will be suppressed.

* Full-volume ("summary") statistics, not sliced into smaller portions.  Nominally, there can be one such aggregated-data
  point emitted at the end of each aggregation period.

* Per-regex-group statistics.  Nominally, there can be one such aggregated-data point emitted at the end of each
  aggregation period for each regex group that was mapped to during that period.

* Per-selector-value statistics.  Nominally, there can be one such aggregated-data point emitted at the end of each
  aggregation period for each distinct selector value that was seen during that period.

The specific tag names used for the aggregation-data points are configurable, as are the sets of fields included in such
data points, to adapt to your local needs.

An example might help to show the utility of such constructions.  Suppose the selector is the hostname from a syslog
message, and the category represents the level of severity of that message (ignore, ok, warning, critical, unknown).
The grouping might use the hostname to identify the nature of the host (firewall node, compute node, network switch,
database machine, file server, etc.), and apply regexes tailored to that type of host.  With those ideas in mind, we
implement the sample configuration shown later in this document.

Let's show what the aggregated-data output would look like, using a simple example.  Suppose we have five hosts:

* `fire123`, a firewall node
* `ora456`, an Oracle database node
* `pg789`, a PostgreSQL database node
* `rout237`, a router
* `rout846`, a router

Let us further suppose we map hostnames into host-types as the regex group names.  For purposes of this mapping, we
assume that all the routers are running the same software, so the nature of their syslog messages will be the same and the
matching of those messages need only be configured once.  Also, for didactic purposes, two different kinds of databases
are folded into the same group.  In practice, that would only be useful if either we had an unspecified abstraction layer
running on both of those machines to map database-type-specific messages into some common format, or we simply combined
all of the regexes for either type of database into the one group of regexes configured in this plugin for such machines.

* `fire123` => `firewall`
* `ora456` => `database`
* `pg789` => `database`
* `rout237` => `router`
* `rout846` => `router`

In addition, let's suppose we have the following regex categories:

* `ignore`, for messages that should be dropped (but counted in this match category in aggregation statistics;
  see the `drop_categories` configuration option)
* `okay`, for messages that represent an operating-normally state
* `warning`, for messages indicating conditions of possible concern
* `critical`, for messages indicating conditions of definite concern
* `unknown`, for messages that we have not yet analyzed enough to develop specific regexes to match and otherwise classify

For the example aggregated-data output shown just below, we have the following options in play, along with definitions
of `aggregation_summary_fields`, `aggregation_group_fields`, and `aggregation_selector_fields` not shown here.  The
`aggregation_includes_zeroes` option is enabled here for didactic purposes, so you can see all the aggregation-point
category fields involved; most commonly, that would be left disabled.

```
aggregation_measurement     = 'status'
aggregation_summary_tag     = 'summary'
aggregation_summary_value   = 'full'
aggregation_dropped_field   = 'dropped'
aggregation_total_field     = 'total'
aggregation_group_tag       = 'host_type'
aggregation_selector_tag    = 'host'
aggregation_includes_zeroes = true
```

With all of that (and a bit more) in play, the available aggregation counts might represent answers to the following kinds
of questions.  We show the corresponding output for this example setup in InfluxDB Line Protocol format.  Recall that
said format names the measurement, concatenated directly with all the (optional) tag data, followed by all the field data,
followed by an optional timestamp (which we do not show here).

_What is the total volume of messages across my infrastructure?_
```
## Show the overall distribution of classification states and plugin
## activity.  Bin the counts of data points reported here by {category}
## only, without discriminating by {group} or {selector}.
status,summary=full ignore=5,okay=3,warning=8,critical=2,unknown=1,dropped=5,total=19
```

_What kinds of machines/devices are generating lots of messages I might care about?_
```
## Show a coarse distribution of incoming data, based on the mapping
## of {selector} values to {group} values.  Bin the counts reported
## here by {group} and {category} only, ignoring the particular
## {selector} values involved.  For this example, we have chosen to
## report only known problem states and the total activity for each
## kind of host.
# aggregation_group_fields = [ 'warning', 'critical', 'total' ]
status,host_type=firewall warning=2,critical=1,total=4
status,host_type=database warning=5,critical=0,total=8
status,host_type=router warning=1,critical=1,total=7
```

_What kinds of services in my infrastructure are in good or bad shape?_
```
## Show a coarse distribution of incoming data, based on the mapping
## of {selector} values to {group} values.  Bin the counts reported
## here by {group} and {category} only, ignoring the particular
## {selector} values involved.  For this example, we have chosen to
## report the full set of calculated {category} states and nothing
## else, because that is all we care to graph.
# aggregation_group_fields = [
#   'ignore', 'okay', 'warning', 'critical', 'unknown'
# ]
status,host_type=firewall ignore=0,okay=1,warning=2,critical=1,unknown=0
status,host_type=database ignore=1,okay=2,warning=5,critical=0,unknown=0
status,host_type=router ignore=4,okay=0,warning=1,critical=1,unknown=1
```

_What is the bird's-eye view of how my infrastructure is running?_
```
## Show a coarse distribution of incoming data, based on the mapping
## of {selector} values to {group} values.  Bin the counts reported
## here by {group} and {category} only.  For this example, we not only
## collect the data needed for later per-service-type reporting, we
## also include total-traffic counts, because the NOC manager watches
## those numbers in a dashboard as a proxy for overall trouble to see
## if he needs to call in extra help for particular kinds of services.
# aggregation_group_fields = [
#   'ignore', 'okay', 'warning', 'critical', 'unknown', 'total'
# ]
status,host_type=firewall ignore=0,okay=1,warning=2,critical=1,unknown=0,total=4
status,host_type=database ignore=1,okay=2,warning=5,critical=0,unknown=0,total=8
status,host_type=router ignore=4,okay=0,warning=1,critical=1,unknown=1,total=7
```

_What particular machines are generating lots of messages, regardless of severity?_
```
## Show just the volume of incoming data across selector values,
## regardless of utility or classification.  This might be useful,
## for instance, if we have set up upstream filtering so only
## serious-state messages are forwarded to where the Telegraf plugin
## sees them, and all we need to graph is the total traffic for each
## host.  There would be no need to waste resources by generating
## and storing other categories of counts.
# aggregation_selector_fields = [ 'total' ]
status,host=fire123 total=4
status,host=ora456 total=2
status,host=pg789 total=6
status,host=rout237 total=7
```

_What were the overall states of particular machines during each reporting period?_
```
## Show detailed counts of classifications on a per-selector-value basis.
## Bin the counts by the combination of {selector} and {category}.  We
## don't care about the other available data (dropped-data-point and
## total-data-point counts), because with the full set of category counts,
## we will have all the information we need to act upon for everyday
## system administration and troubleshooting.
# aggregation_selector_fields = [
#   'ignore', 'okay', 'warning', 'critical', 'unknown'
# ]
status,host=fire123 ignore=0,okay=1,warning=2,critical=1,unknown=0
status,host=ora456 ignore=0,okay=1,warning=1,critical=0,unknown=0
status,host=pg789 ignore=1,okay=1,warning=4,critical=0,unknown=0
status,host=rout237 ignore=4,okay=0,warning=1,critical=1,unknown=1
```

_How healthy are my servers, in the eyes of everyone who cares in one way or another?_
```
## Show detailed counts of classifications on a per-selector-value basis,
## and also capture the total-traffic count for each host.  The latter
## will be used not for the NOC operators, but for a management report to
## direct attention to equipment that may need to be upgraded or replaced.
# aggregation_selector_fields = [
#   'ignore', 'okay', 'warning', 'critical', 'unknown', 'total'
# ]
status,host=fire123 ignore=0,okay=1,warning=2,critical=1,unknown=0,total=4
status,host=ora456 ignore=0,okay=1,warning=1,critical=0,unknown=0,total=2
status,host=pg789 ignore=1,okay=1,warning=4,critical=0,unknown=0,total=6
status,host=rout237 ignore=4,okay=0,warning=1,critical=1,unknown=1,total=7
```

Which of these statistics need to be aggregated and output depends on your own use case, so this work is all configurable.
Each kind of statistic can be individually enabled by configuration.  If none of them are enabled, no aggregation counting
will occur and no aggregated measurements will be generated.

Notice that no aggregation output appeared for host `rout846`.  That's because it sent no messages during this period.
The aggregation does not manufacture and send out total=0 or equivalent data points in this case.  That has an effect
on how you set up valid graphing for aggregation data.  (You shouldn't be connecting successive available non-zero data
points with lines, as that would be misleading.  If your graphing tool has the ability to treat missing values in a graph
interval as zero values, then you could reasonably connect successive data points with lines.  Beyond such simple advice,
the topic of metric graphing is outside the scope of this plugin documentation.)

## Configuration options

The `classify` plugin needs complex data for its configuration, including TOML literal strings as hash keys and arrays
of hashes.  However, the Telegraf 1.x TOML parser is limited in its functionality and does not support the full TOML
v1.0.0 specification.  Hopefully, that will be addressed in Telegraf 2.0.  In the meantime, we are forced to move the
configuration for this plugin out to a separate file, and we leave behind only a single plugin-specific option to be
processed by the Telegraf-internal parser.

<dl>
<dt><tt>classify_config_file</tt></dt>
<dd>
The path to a file containing the full plugin configuration data as described below.  By convention, we use a <tt>.toml</tt>
filename extension on that file, so we can park it alongside the other Telegraf config files but have Telegraf itself
ignore the file.

```
classify_config_file = '/etc/telegraf/telegraf.d/classify.toml'
```
</dd>
<dd>
At the point where Telegraf itself is able to parse full TOML v1.0.0, this config option will be deprecated, and the
standard recommended setup will have the plugin configuration in the usual location.
</dd>
</dl>

The rest of this section describes what is in the pointed-to config file.  Additional information about how to set these
options is given in the comments in the sample configuration later in this document.

<dl>

<dt><tt>selector_tag</tt>
<br><tt>selector_field</tt></dt>
<dd>
The name of the tag or field ("selector item") to be used to determine the <tt>mapped_selector_regexes</tt> group to use
for regex matching.
</dd>
<dd>
These two options are mutually exclusive.  If neither one is defined, the <tt>selector_mapping</tt> is irrelevant, and the
separately configured <tt>default_regex_group</tt> value will be used unconditionally as the <tt>mapped_selector_regexes</tt>
group to use for regex matching.
</dd>

<dt><tt>selector_mapping</tt></dt>
<dd>
The mapping of selector values to regex-group names.  It is applied according to the following rules:

* If the selector item is missing from the input data point, this data point will be dropped silently (except for being
  counted as such in aggregation statistics).

* If this mapping is not defined at all, or if the selector item's value matches a key in the mapping and the mapped
  value is the special string `*`, the selector item's value will be used unmodified as the mapped value.  That means it
  will generally be the chosen `mapped_selector_regexes` group name, unless there is no such group name; see the special
  handling described below for that corner case.
  ```
  selector_mapping.'unmodified_value' = '*'
  ```

* If this mapping is defined, the selector item's value matches a key in the mapping, and the mapped value is an empty
  string, this data point will be dropped silently (except for being counted as such in aggregation statistics).
  ```
  selector_mapping.'dropped_value' = ''
  ```
  Speaking of dropping input data points, if there is any possibility of having the selector item be present in an input
  data point but with an empty string as its value, you should consider what you want to do in that situation.

  * You can explicitly declare that input data points with an empty string as the selector item value should be dropped
    silently (except for being counted as such in aggregation statistics).  This would probably be the most common choice.
    ```
    selector_mapping.'^$' = ''
    ```

  * You can explicitly declare that empty selector item values should map to some particular non-empty regex-group name:
    ```
    selector_mapping.'^$' = 'my_regex_group'
    ```

  * You can allow such an input data point to be defaulted as not matching any key in the `selector_mapping`, and have
    the later rule kick in that will use the value of the `default_regex_group` option as the selector mapping value.

* If this mapping is defined, the selector item's value matches a key in the mapping, and the mapped value is present in
  the configuration as the name of a `mapped_selector_regexes` group, that group will be used for regex matching.

* Under any of the following conditions, the value of the `default_regex_group` option will be used as the selector
  mapping value.

  * Some `selector_mapping` mappings are defined, but the selector item's value does not match any key in the mapping.

  * The mapped selector item value, however that was calculated, does not match the name of any `mapped_selector_regexes`
    group.

  This design decision (using the value of the `default_regex_group` option) is intended mostly to avoid situations where
  an unintended misconfiguration (e.g., a typo) might otherwise go unnoticed because the data point would be dropped.
  If you don't want that behavior, you have several choices:

  * Extend the set of mapping keys to cover the cases you are concerned with.

  * Define the `default_regex_group` option to be an empty string. or equivalently just leave it undefined, so those data
    points will be intentionally dropped.

  * Put in place a general wildcard rule at the end of the selector_mapping`` that maps all other selector values to an
    empty string, so those data points will be intentionally dropped.

    ```
    selector_mapping = [
      { 'fire\d{3}' = 'firewall' },
      { 'pg789'     = 'database' },
      { 'rout846'   = 'router'   },
      { '.*'        = ''         },
    ]
    ```

* If the value of the `default_regex_group` option is needed under the previous rule but that option is not defined, it
  will be treated as though it were defined as an empty string.  Which means, this data point will be dropped silently
  (except for being counted as such in aggregation statistics).
</dd>
<dd>
The <tt>selector_mapping</tt> keys are regular expressions, not just literal strings.  Thus for example if you are using
hostnames as selectors, and you have cleverly chosen the names of all your firewall machines from the set <tt>fire000</tt>
through <tt>fire999</tt>, you could use the following as one of your mapping elements and not have to deal with the
individual machine names.

```
selector_mapping.'fire\d{3}' = 'firewall'
```
</dd>
<dd>

Because regexes are allowed as mapping keys, not just fixed literal strings, there is the possibility that a given
selector item value might match more than one key in the mapping.  To eliminate that ambiguity, the format of the
<tt>selector_mapping</tt> option has been chosen to specify a particular order in which the entries will be used at
runtime to attempt to match a given input data point.  The <tt>classify</tt> plugin will always attempt selector item
value matching using patterns in order as you list them.  This gives you complete control over which pattern will match
first and thereby provide the regex group name for the next phase of processing.

```
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

This ordering allows you to provide a final wildcard entry.  So for instance if you want all otherwise unmapped selector
item values to have the input data point be matched against the <tt>compute</tt> group of regexes, without depending
on the rule noted above that the value of the <tt>default_regex_group</tt> option will be used in this case, specify a
corresponding wildcard match as the last <tt>selector_mapping</tt> entry in the configuration:

```
selector_mapping = [
  { 'desk\d{3}' = 'desktop' },
  { 'file\d{3}' = 'storage' },
  { 'tape\d{3}' = 'storage' },
  { 'back\d{3}' = 'backups' },
  { 'wire\d{3}' = 'network' },
  { '.*'        = 'compute' },
]
```
</dd>

<dt><tt>default_regex_group</tt></dt>
<dd>
A regex group name to be used as the value of applying the <tt>selector_mapping</tt> under any of the following conditions:

* Neither the <tt>selector_tag</tt> option nor the <tt>selector_field</tt> option is defined.

* No <tt>selector_mapping</tt> match to an input data point's selector item value is found.  This includes the case where the
  selector item value is an empty string and you do not have any <tt>selector_mapping</tt> entry that matches an empty string.

* The mapped selector item value does not match the name of any <tt>mapped_selector_regexes</tt> group.

If the <tt>default_regex_group</tt> option is needed for a given input data point, that data point will be dropped (and
counted as such in aggregation totals) under any of the following conditions:

* The <tt>default_regex_group</tt> option is not defined.

* The <tt>default_regex_group</tt> option is defined as an empty string.

* The <tt>default_regex_group</tt> option is defined as a non-empty string, but its value does not match the name of any
  <tt>mapped_selector_regexes</tt> group.
</dd>

<dt><tt>match_tag</tt>
<br><tt>match_field</tt></dt>
<dd>
The name of the tag or field ("match item") within the input data point to be used to match against regular expressions.
</dd>
<dd>
These two options are mutually exclusive.  One of them must be specified, as otherwise there will be no way to calculate the classification of input data points.
</dd>

<dt><tt>mapped_selector_regexes</tt></dt>
<dd>
An ordered group of category-to-regexes mappings, each of which defines an ordered array of regular expressions.  You will
define one or more such groups in the plugin configuration.  Names of the groups are chosen by the user, and are not
predefined by this plugin.  Those names correspond to the values specified in the <tt>selector_mapping</tt> mapping.
</dd>
<dd>
All the mappings for a given group, taken together, represent successive patterns to be used for matching if the mapped
selector value is the name of that group.  Each group can contain zero or more category-to-regexes mappings, from which
the regexes are pattern-matched against the match item (<tt>match_tag</tt> or <tt>match_field</tt>) in order as seen in
the configuration of the group.
</dd>
<dd>
Each category-to-regexes mapping can contain zero or more regular expressions, expressed in any of several possible
forms as shown in the sample setup toward the end of this document.  A match on any one of the regular expressions
in this set will end the classification logic for this data point, with the result being the name of this category.
Names of the categories are chosen by the user, and are not predefined by this plugin.  Typically, you will use the same
set of categories for every <tt>mapped_selector_regexes</tt> group.  Failure to match any of the regular expressions in
a given category-to-regexes mapping means the logic will proceed to attempt matching of the next such mapping defined
in the same <tt>mapped_selector_regexes</tt> group.  If matching attempts reach the end of all the category-to-regexes
mappings in the selected <tt>mapped_selector_regexes</tt> group and no match has been achieved, this data point will
be dropped silently (except for being counted as such in aggregation statistics) unless the <tt>default_category</tt>
is defined as a non-empty string and the <tt>default_category</tt> value is not listed in the <tt>drop_categories</tt>.
</dd>
<dd>
The config-file syntax for specifying each category-to-regexes mapping is a bit touchy, due to limitations of the TOML
config-file language.  In particular, no newlines are allowed anywhere between the opening and closing braces that enclose
each individual mapping, except for those that you can include within a set of regexes in the multi-line format, which
is shown just below.  In this regard, follow the examples carefully.
</dd>
<dd>
In the multi-line format for specifying a set of category regexes, all leading and trailing whitespace in each regex line
will be automatically trimmed and not be part of the regex, and blank lines within a given set of regexes will be ignored.
That allows you to visually separate distinct clusters of related regexes, making for easier understanding and maintenance.
Here is an example:

```
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
</dd>

<dt><tt>default_category</tt></dt>
<dd>
A category name to be used as the value of the result item if regular-expression matching was in fact attempted but no
pattern match succeeded.  If the <tt>default_category</tt> is not defined, or is defined as an empty string, data points
having no regex match will be dropped (and counted as such in aggregation totals).  The category named by this option is
subject to filtering by being mentioned in <tt>drop_categories</tt>, just as any successful-regex-match category would be.
</dd>
<dd>
Note that the <tt>default_category</tt> is not used as the value of the result item if the classification logic decided
that no regular-expression matching could be attempted and the input data point was dropped.  So for example if the
<tt>selector_mapping</tt> does not end up generating a regex group name in the <tt>mapped_selector_regexes</tt> and the
<tt>default_regex_group</tt> also has the same fate, the input data point is doomed.
</dd>
<dd>
One use for this option would be to name a "normal", "don't care", or other anodyne category, to be applied if all unmatched
data points are to be treated as unexceptional.  The contrary use is to have this option name a serious-state category,
so you can be notified when something happens that you have never seen before.
</dd>

<dt><tt>drop_categories</tt></dt>
<dd>
Zero or more regex-match category names wherein if the regex matching result is one of those categories, however that matching
result was calculated, that data point will be dropped from the plugin output.  Such a data point will be counted in the
aggregation statistics both under the respective match category, and under the named <tt>aggregation_dropped_field</tt>
if that is defined.

```
drop_categories = 'ignore'
```
</dd>

<dt><tt>result_tag</tt>
<br><tt>result_field</tt></dt>
<dd>
The name of the tag or field where the classification result ("result item") is stored, if the data point is not dropped.
The result item value will either be the name of the first category one of whose regexes pattern-matched the match item
value, or the value of <tt>default_category</tt> if no such match was found.  This item will be added to the data point
if not present, or overwritten if already present.
</dd>
<dd>
These two options are mutually exclusive.  One of them must be specified, as otherwise there will be no way to report out the classification of input data points.
</dd>
</dd>

<dt><tt>aggregation_period</tt></dt>
<dd>
How often to spill out the aggregation counters as a measurement separate from the data points which are streaming through
this plugin.  If left undefined or set to a non-positive value, aggregation will be disabled.  Specified as a number with
a trailing letter for time units (<tt>s</tt>, <tt>m</tt>, <tt>h</tt>).  This field is required if you wish to output any
kind of aggregated statistics.

```
aggregation_period = '10m'
```
</dd>

<dt><tt>aggregation_measurement</tt></dt>
<dd>
The name of the measurement that will be used to record aggregated classification statistics that are sent downstream.
This one measurement name will be used for all generated aggregation-summary, aggregation-by-group, and aggregation-by-selector
data points.  This field is required if you wish to output any kind of aggregated statistics.

```
aggregation_measurement = 'message_counts'
```
</dd>

<dt><tt>aggregation_dropped_field</tt></dt>
<dd>
The name of a field that may be mentioned in <tt>aggregation_summary_fields</tt> or <tt>aggregation_group_fields</tt> or
<tt>aggregation_selector_fields</tt> to report the number of data points that were dropped for any reason whatsoever during
their transit through this plugin.  There is only one such field name supported for all of those types of aggregation
statistics, though like category names, the count it represents varies slightly depending on which type of aggregation is
being reported.

```
aggregation_dropped_field = 'dropped'
```
</dd>

<dt><tt>aggregation_total_field</tt></dt>
<dd>
The name of a field that may be mentioned in <tt>aggregation_summary_fields</tt> or <tt>aggregation_group_fields</tt> or
<tt>aggregation_selector_fields</tt> to report the total number of data points that were processed through this plugin,
whether or not they were ultimately dropped or output.  There is only one such field name supported for all of those types
of aggregation statistics, though like category names, the count it represents varies slightly depending on which type of
aggregation is being reported.

```
aggregation_total_field = 'total'
```
</dd>

<dt><tt>aggregation_summary_tag</tt></dt>
<dd>
The name of a tag to be used when reporting aggregation-summary level statistics, representing the full volume of input
data points during each <tt>aggregation_period</tt> without any breakdown into finer granularity.  This field is required
if you wish to output that level of statistics.  Its purpose is to provide a handle for querying data and retrieving only
the summary-level statistics.

```
aggregation_summary_tag = 'severity'
```
</dd>

<dt><tt>aggregation_summary_value</tt></dt>
<dd>
The single fixed value of the <tt>aggregation_summary_tag</tt> to be used when reporting aggregation-summary level statistics.
This field is required if you wish to output that level of statistics.

```
aggregation_summary_value = 'all'
```
</dd>

<dt><tt>aggregation_summary_fields</tt></dt>
<dd>
A list of the fields that should appear in each aggregation-summary output data point.  This field is required if
you wish to output that level of statistics.  These fields may be the names of any of the regex categories defined
in <tt>mapped_selector_regexes</tt>, or the field named by <tt>aggregation_dropped_field</tt>, or the field named by
<tt>aggregation_total_field</tt>.

```
aggregation_summary_fields = [
    'ignore', 'okay', 'warning', 'critical', 'unknown',
    'dropped', 'total'
]
```
</dd>

<dt><tt>aggregation_group_tag</tt></dt>
<dd>
The name of the tag to attach to each aggregated-data output data point that bins input data points firstly by which
regex group the selector mapped to.  This field is required if you wish to output that level of statistics.  At the end
of each <tt>aggregation_period</tt>, one output data point is produced for each such group that had at least one input
data point mapped to that group during that period.  The value of this tag will be the name of the regex group whose
counts are summarized in that output data point.

```
aggregation_group_tag = 'host_type'
```
</dd>

<dt><tt>aggregation_group_fields</tt></dt>
<dd>
A list of the fields that should appear in each aggregation-by-group output data point.  This field is required if
you wish to output that level of statistics.  These fields may be the names of any of the regex categories defined
in <tt>mapped_selector_regexes</tt>, or the field named by <tt>aggregation_dropped_field</tt>, or the field named by
<tt>aggregation_total_field</tt>.

```
aggregation_group_fields = [
    'okay', 'warning', 'critical', 'unknown', 'dropped', 'total'
]
```
</dd>

<dt><tt>aggregation_selector_tag</tt></dt>
<dd>
The name of the tag to attach to each aggregated-data output data point that bins input data points firstly by which selector
value was present in the input data point.  This field is required if you wish to output that level of statistics.  At the
end of each <tt>aggregation_period</tt>, one output data point is produced for each selector value that had at least one
input data point with that value during that period.  The value of this tag will be that selector value, corresponding
to the counts which are summarized in that output data point.

```
aggregation_selector_tag = 'host'
```
</dd>

<dt><tt>aggregation_selector_fields</tt></dt>
<dd>
A list of the fields that should appear in each aggregation-by-selector output data point.  This field is required
if you wish to output that level of statistics.  These fields may be the names of any of the regex categories defined
in <tt>mapped_selector_regexes</tt>, or the field named by <tt>aggregation_dropped_field</tt>, or the field named by
<tt>aggregation_total_field</tt>.

```
aggregation_selector_fields = [
    'okay', 'warning', 'critical', 'unknown', 'dropped', 'total'
]
```
</dd>

<dt><tt>aggregation_includes_zeroes</tt></dt>
<dd>
Whether or not to include fields for categories that have a zero count in an aggregation data point.  By default (if this
option is not defined), such fields are not included in aggregation data points, to reduce downstream load and storage
requirements.  If you want explicit zeroes to show up for such categories (though only when there is at least one non-zero
counter to force out the aggregation data point), define the <tt>aggregation_includes_zeroes</tt> option to be <tt>true</tt>.

```
aggregation_includes_zeroes = false
```
</dd>
</dl>

## Sample configuration

The `classify` plugin needs the full expressiveness of TOML v1.0.0 for its configuration data.  To make that available,
the detailed setup for this plugin is moved to a separate file, and the standard Telegraf configuration for this plugin
is just a single option.

```toml
# Classify Telegraf data points according to user-specified rules.
[[processors.classify]]
  ## The detailed configuration data for the classify plugin lives
  ## in a separate file, whose path is given here.
  classify_config_file = '/etc/telegraf/telegraf.d/classify.toml'
```

The following example setup (a sample `classify.toml` file) shows how you might classify syslog messages, as described
in the main text above.

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

## Bugs

The built-in TOML parser in Telegraf 1.x was written some time ago and does not handle the full syntax of TOML v1.0.0.
We are therefore forced to invoke a separate TOML parser from within the plugin to provide support for a compact and
readable configuration.  Hopefully this situation will be fixed in Telegraf 2.0, where we would expect Telegraf to adopt
the github.com/BurntSushi/toml package for TOML parsing, as the `classify` plugin has currently done.

The config-file formatting is complex, both because the underlying structure of the data to be represented is complex,
and because of limitations in the TOML config-file format (and its parsers) used by Telegraf.
(See https://toml.io/en/v1.0.0 for the TOML specification.)
In particular, we would like the as-listed ordering of hash keys in a single key/value table to be accessible
after parsing the config file, in addition to by-key lookups within the table.
Both forms of access to the keys are important, but we are not allowed to quash an array of tables into a single table
(thereby also avoiding potential duplicate keys) and retain numerically-indexed access to the keys.
This idea may blow your mind because you have been trained to think only in terms of simple arrays and simple
random-iteration-order hashes as separate data structures (and never the twain shall meet, except for nesting).
If that is the case, take a look at the Boost Multi-index Containers Library:

  https://www.boost.org/doc/libs/1_79_0/libs/multi_index/doc/index.html

The following package seems to be an implementation of the concept for Go, though it's hard to tell because as of this
writing the package is not well-documented:

  https://pkg.go.dev/github.com/eosspark/geos/libraries/multiindex

We should describe how to run this plugin in a mode that simply verifies that a sane configuration has been supplied,
without attempting to process any data points.

We should describe how to run this plugin in a mode that logs its internal decisions, and tell users where to find the log,
so users can debug the behavior of misconfigured setups.

There should be a means to control the level of logging detail, to allow finer granularity of logging while debugging
a configuration and coarser granularity during production use of the plugin.

## Possible future features

* Extend the model to support multiple independent classifications, possibly chained.  Chaining would mean that the
  `result_tag` or `result_field` would be used as the `selector_tag` or `selector_field` for a subsequent classification.
  Supporting multiple independent classifications would also mean adding another option to tell whether each intermediate
  result in a chained classification would be output as well as the final classification in that chain.  These ideas
  await a definitively useful use case.

