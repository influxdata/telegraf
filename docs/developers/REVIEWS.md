# Reviews

Pull-requests require two approvals before being merged. Expect several rounds of back and forth on
reviews, non-trivial changes are rarely accepted on the first pass. It might take some time
until you see a first review so please be patient.

All pull requests should follow the style and best practices in the
[CONTRIBUTING.md](https://github.com/influxdata/telegraf/blob/master/CONTRIBUTING.md)
document.

## Process

The review process is roughly structured as follows:

1. Submit a pull request.
Please check that you signed the [CLA](https://www.influxdata.com/legal/cla/) (and [Corporate CLA](https://www.influxdata.com/legal/ccla/) if you are contributing code on as an employee of your company). Provide a short description of your submission and reference issues that you potentially close. Make sure the CI tests are all green and there are no linter-issues.
1. Get feedback from a first reviewer and a `ready for final review` tag.
Please constructively work with the reviewer to get your code into a mergeable state (see also [below](#reviewing-plugin-code)).
1. Get a final review by one of the InfluxData maintainers.
Please fix any issue raised.
1. Wait for the pull-request to be merged.
It might take some time until your PR gets merged, depending on the release cycle and the type of
your pull-request (bugfix, enhancement of existing code, new plugin, etc). Remember, it might be necessary to rebase your code before merge to resolve conflicts.

Please read the review comments carefully, fix the related part of the code and/or respond in case there is anything unclear. Maintainers will add the `waiting for response` tag to PRs to make it clear we are waiting on the submitter for updates. __Once the tag is added, if there is no activity on a pull request or the contributor does not respond, our bot will automatically close the PR after two weeks!__ If you expect a longer period of inactivity or you want to abandon a pull request, please let us know.

In case you still want to continue with the PR, feel free to reopen it.

## Reviewing Plugin Code

- Avoid variables scoped to the package. Everything should be scoped to the plugin struct, since multiple instances of the same plugin are allowed and package-level variables will cause race conditions.
- SampleConfig must match the readme, but not include the plugin name.
- structs should include toml tags for fields that are expected to be editable from the config. eg `toml:"command"` (snake_case)
- plugins that want to log should declare the Telegraf logger, not use the log package. eg:

```Go
  Log telegraf.Logger `toml:"-"`
```

(in tests, you can do `myPlugin.Log = testutil.Logger{}`)

- Initialization and config checking should be done on the `Init() error` function, not in the Connect, Gather, or Start functions.
- `Init() error` should not contain connections to external services. If anything fails in Init, Telegraf will consider it a configuration error and refuse to start.
- plugins should avoid synchronization code if they are not starting goroutines. Plugin functions are never called in parallel.
- avoid goroutines when you don't need them and removing them would simplify the code
- errors should almost always be checked.
- avoid boolean fields when a string or enumerated type would be better for future extension. Lots of boolean fields also make the code difficult to maintain.
- use config.Duration instead of internal.Duration
- compose tls.ClientConfig as opposed to specifying all the TLS fields manually
- http.Client should be declared once on `Init() error` and reused, (or better yet, on the package if there's no client-specific configuration). http.Client has built-in concurrency protection and reuses connections transparently when possible.
- avoid doing network calls in loops where possible, as this has a large performance cost. This isn't always possible to avoid.
- when processing batches of records with multiple network requests (some outputs that need to partition writes do this), return an error when you want the whole batch to be retried, log the error when you want the batch to continue without the record
- consider using the StreamingProcessor interface instead of the (legacy) Processor interface
- avoid network calls in processors when at all possible. If it's necessary, it's possible, but complicated (see processor.reversedns).
- avoid dependencies when:
  - they require cgo
  - they pull in massive projects instead of small libraries
  - they could be replaced by a simple http call
  - they seem unnecessary, superfluous, or gratuitous
- consider adding build tags if plugins have OS-specific considerations
- use the right logger log levels so that Telegraf is normally quiet eg `plugin.Log.Debugf()` only shows up when running Telegraf with `--debug`
- consistent field types: dynamically setting the type of a field should be strongly avoided as it causes problems that are difficult to solve later, made worse by having to worry about backwards compatibility in future changes. For example, if an numeric value comes from a string field and it is not clear if the field can sometimes be a float, the author should pick either a float or an int, and parse that field consistently every time. Better to sometimes truncate a float, or to always store ints as floats, rather than changing the field type, which causes downstream problems with output databases.
- backwards compatibility: We work hard not to break existing configurations during new changes. Upgrading Telegraf should be a seamless transition. Possible tools to make this transition smooth are:
  - enumerable type fields that allow you to customize behavior (avoid boolean feature flags)
  - version fields that can be used to opt in to newer changed behavior without breaking old (see inputs.mysql for example)
  - a new version of the plugin if it has changed significantly (eg outputs.influxdb and outputs.influxdb_v2)
  - Logger and README deprecation warnings
  - changing the default value of a field can be okay, but will affect users who have not specified the field and should be approached cautiously.
  - The general rule here is "don't surprise me": users should not be caught off-guard by unexpected or breaking changes.

## Linting

Each pull request will have the appropriate linters checking the files for any common mistakes. The github action Super Linter is used: [super-linter](https://github.com/github/super-linter). If it is failing you can click on the action and read the logs to figure out the issue. You can also run the github action locally by following these instructions: [run-linter-locally.md](https://github.com/github/super-linter/blob/main/docs/run-linter-locally.md). You can find more information on each of the linters in the super linter readme.

## Testing

Sufficient unit tests must be created.  New plugins must always contain
some unit tests.  Bug fixes and enhancements should include new tests, but
they can be allowed if the reviewer thinks it would not be worth the effort.

[Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests) are
encouraged to reduce boiler plate in unit tests.

The [stretchr/testify](https://github.com/stretchr/testify) library should be
used for assertions within the tests when possible, with preference towards
github.com/stretchr/testify/require.

Primarily use the require package to avoid cascading errors:

```go
assert.Equal(t, lhs, rhs) # avoid
require.Equal(t, lhs, rhs) # good
```

## Configuration

The config file is the primary interface and should be carefully scrutinized.

Ensure the [[SampleConfig]] and
[README](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/EXAMPLE_README.md)
match with the current standards.

READMEs should:

- be spaces, not tabs
- be indented consistently, matching other READMEs
- have two `#` for comments
- have one `#` for defaults, which should always match the default value of the plugin
- include all appropriate types as a list for enumerable field types
- include a useful example, avoiding "example", "test", etc.
- include tips for any common problems
- include example output from the plugin, if input/processor/aggregator/parser/serializer

## Metric Schema

Telegraf metrics are heavily based on InfluxDB points, but have some
extensions to support other outputs and metadata.

New metrics must follow the recommended
[schema design](https://docs.influxdata.com/influxdb/latest/concepts/schema_and_data_layout/).
Each metric should be evaluated for _series cardinality_, proper use of tags vs
fields, and should use existing patterns for encoding metrics.

Metrics use `snake_case` naming style.

### Enumerations

Generally enumeration data should be encoded as a tag.  In some cases it may
be desirable to also include the data as an integer field:

```shell
net_response,result=success result_code=0i
```

### Histograms

Use tags for each range with the `le` tag, and `+Inf` for the values out of
range.  This format is inspired by the Prometheus project:

```shell
cpu,le=0.0 usage_idle_bucket=0i 1486998330000000000
cpu,le=50.0 usage_idle_bucket=2i 1486998330000000000
cpu,le=100.0 usage_idle_bucket=2i 1486998330000000000
cpu,le=+Inf usage_idle_bucket=2i 1486998330000000000
```

### Lists

Lists are tricky, but the general technique is to encode using a tag, creating
one series be item in the list.

### Counters

Counters retrieved from other projects often are in one of two styles,
monotonically increasing without reset and reset on each interval.  No attempt
should be made to switch between these two styles but if given the option it
is preferred to use the non-resetting variant.  This style is more resilient in
the face of downtime and does not contain a fixed time element.

### Source tag

When metrics are gathered from another host, the metric schema should have a tag
named "source" that contains the other host's name. See [this feature
request](https://github.com/influxdata/telegraf/issues/4413) for details.

The metric schema doesn't need to have a tag for the host running
telegraf. Telegraf agent code can add a tag named "host" and by default
containing the hostname reported by the kernel. This can be configured through
the "hostname" and "omit_hostname" agent settings.

## Go Best Practices

In general code should follow best practice describe in [Code Review
Comments](https://github.com/golang/go/wiki/CodeReviewComments).

### Networking

All network operations should have appropriate timeouts.  The ability to
cancel the option, preferably using a context, is desirable but not always
worth the implementation complexity.

### Channels

Channels should be used in judiciously as they often complicate the design and
can easily be used improperly.  Only use them when they are needed.
