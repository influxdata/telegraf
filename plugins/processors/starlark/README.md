# Starlark Processor Plugin

This plugin calls the provided Starlark function for each matched metric,
allowing for custom programmatic metric processing.

The Starlark language is a dialect of Python, and will be familiar to those who
have experience with the Python language. However, there are major
[differences](#python-differences). Existing Python code is unlikely to work
unmodified. The execution environment is sandboxed, and it is not possible to
do I/O operations such as reading from files or sockets.

The **[Starlark specification][spec]** has details about the syntax and
available functions.

⭐ Telegraf v1.15.0
🏷️ general purpose
💻 all

[spec]: https://github.com/google/starlark-go/blob/d1966c6b9fcd/doc/spec.md

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Process metrics using a Starlark script
[[processors.starlark]]
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.

  ## Source of the Starlark script.
  source = '''
def apply(metric):
  return metric
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"

  ## The constants of the Starlark script.
  # [processors.starlark.constants]
  #   max_size = 10
  #   threshold = 0.75
  #   default_name = "Julia"
  #   debug_mode = true
```

## Usage

The Starlark code should contain a function called `apply` that takes a metric
as its single argument.  The function will be called with each metric, and can
return `None`, a single metric, or a list of metrics.

```python
def apply(metric):
    return metric
```

For a list of available types and functions that can be used in the code, see
the [Starlark specification][spec].

In addition to these, the following InfluxDB-specific
types and functions are exposed to the script.

- **Metric(*name*)**:
Create a new metric with the given measurement name.  The metric will have no
tags or fields and defaults to the current time.

- **name**:
The name is a [string][string] containing the metric measurement name.

- **tags**:
A [dict-like][dict] object containing the metric's tags.

- **fields**:
A [dict-like][dict] object containing the metric's fields.  The values may be
of type int, float, string, or bool.

- **time**:
The timestamp of the metric as an integer in nanoseconds since the Unix
epoch.

- **deepcopy(*metric*, *track=false*)**:
Copy an existing metric with or without tracking information. If `track` is set
to `true`, the tracking information is copied.
**Caution:** Make sure to always return *all* metrics with tracking information!
Otherwise, the corresponding inputs will never receive the delivery information
and potentially overrun!

### Python Differences

While Starlark is similar to Python, there are important differences to note:

- Starlark has limited support for error handling and no exceptions.  If an
  error occurs the script will immediately end and Telegraf will drop the
  metric.  Check the Telegraf logfile for details about the error.

- It is not possible to import other packages and the Python standard library
  is not available.

- It is not possible to open files or sockets.

- These common keywords are **not supported** in the Starlark grammar:

  ```text
  as             finally        nonlocal
  assert         from           raise
  class          global         try
  del            import         with
  except         is             yield
  ```

### Libraries available

The ability to load external scripts other than your own is pretty limited. The
following libraries are available for loading:

- json: `load("json.star", "json")` provides the functions `json.encode()`,
        `json.decode()`, `json.indent()`. See [json.star](testdata/json.star)
        for an example. For more details about the functions, please refer to the
        [library documentation][json_lib].
- log:  `load("logging.star", "log")` provides the functions `log.debug()`,
        `log.info()`, `log.warn()`, `log.error()`. See
         [logging.star](testdata/logging.star) for an example.
- math: `load("math.star", "math")` provides the function
         [documented in the library][math_lib]. See
         [math.star](testdata/math.star) for an example.
- time: `load("time.star", "time")` provides the functions `time.from_timestamp()`,
        `time.is_valid_timezone()`, `time.now()`, `time.parse_duration()`,
        `time.parse_time()`, `time.time()`. See
         [time_date.star](testdata/time_date.star),
         [time_duration.star](testdata/time_duration.star) and
         [time_timestamp.star](testdata/time_timestamp.star) for examples. For
         more details about the functions, please refer to the
         [library documentation][time_lib].

If you would like to see support for something else here, please open an issue.

[json_lib]: https://pkg.go.dev/go.starlark.net/lib/json
[math_lib]: https://pkg.go.dev/go.starlark.net/lib/math
[time_lib]: https://pkg.go.dev/go.starlark.net/lib/time

### Common Questions

**What's the performance cost to using Starlark?**

In local tests, it takes about 1µs (1 microsecond) to run a modest script to
process one metric. This is going to vary with the size of your script, but the
total impact is minimal.  At this pace, it's likely not going to be the
bottleneck in your Telegraf setup.

**How can I drop/delete a metric?**

If you don't return the metric it will be deleted.  Usually this means the
function should `return None`.

**How should I make a copy of a metric?**

Use `deepcopy(metric)` to create a copy of the metric.

**How can I return multiple metrics?**

You can return a list of metrics:

```python
def apply(metric):
    m2 = deepcopy(metric)
    return [metric, m2]
```

**What happens to a tracking metric if an error occurs in the script?**

The metric is marked as undelivered.

**How do I create a new metric?**

Use the `Metric(name)` function and set at least one field.

**What is the fastest way to iterate over tags/fields?**

The fastest way to iterate is to use a for-loop on the tags or fields attribute:

```python
def apply(metric):
    for k in metric.tags:
        pass
    return metric
```

When you use this form, it is not possible to modify the tags inside the loop,
if this is needed you should use one of the `.keys()`, `.values()`, or
`.items()` methods:

```python
def apply(metric):
    for k, v in metric.tags.items():
        pass
    return metric
```

**How can I save values across multiple calls to the script?**

Telegraf freezes the global scope, which prevents it from being modified, except
for a special shared global dictionary named `state`, this can be used by the
`apply` function.  See an example of this in [compare with previous
metric](testdata/compare_metrics.star)

Other than the `state` variable, attempting to modify the global scope will fail
with an error.

**How to manage errors that occur in the apply function?**

In case you need to call some code that may return an error, you can delegate
the call to the built-in function `catch` which takes as argument a `Callable`
and returns the error that occurred if any, `None` otherwise.

So for example:

```python
load("json.star", "json")

def apply(metric):
    error = catch(lambda: failing(metric))
    if error != None:
        # Some code to execute in case of an error
        metric.fields["error"] = error
    return metric

def failing(metric):
    json.decode("non-json-content")
```

**How to reuse the same script but with different parameters?**

In case you have a generic script that you would like to reuse for different
instances of the plugin, you can use constants as input parameters of your
script.

So for example, assuming that you have the next configuration:

```toml
[[processors.starlark]]
  script = "/usr/local/bin/myscript.star"

  [processors.starlark.constants]
    somecustomnum = 10
    somecustomstr = "mycustomfield"
```

Your script could then use the constants defined in the configuration as
follows:

```python
def apply(metric):
    if metric.fields[somecustomstr] >= somecustomnum:
        metric.fields.clear()
    return metric
```

**What does `cannot represent integer ...` mean?**

The error occurs if an integer value in starlark exceeds the signed 64 bit
integer limit. This can occur if you are summing up large values in a starlark
integer value or convert an unsigned 64 bit integer to starlark and then create
a new metric field from it.

This is due to the fact that integer values in starlark are *always* signed and
can grow beyond the 64-bit size. Therefore converting the value back fails in
the cases mentioned above.

As a workaround you can either clip the field value at the signed 64-bit limit
or return the value as a floating-point number.

### Examples

- [drop fields containing string values](testdata/drop_string_fields.star)
- [drop fields with unexpected types](testdata/drop_fields_with_unexpected_type.star)
- [obtain IOPS for aggregation and computing max IOPS)](testdata/iops.star)
- [process JSON in a metric field](testdata/json.star) - see
  [library documentation][json_lib] for function documentation
- [use math function to compute a field value](testdata/math.star) - see
  [library documentation][math_lib] for function documentation
- [transform numerical values](testdata/number_logic.star)
- [pivot a key's value to be the key for another field](testdata/pivot.star)
- [compute the ratio of two integer fields](testdata/ratio.star)
- [rename tags or fields using a name mapping](testdata/rename.star)
- [scale field values](testdata/scale.star)
- [parse date and extract year, month and day](testdata/time_date.star) - see
  [library documentation][time_lib] for function documentation
- [parse duration and convert into seconds](testdata/time_duration.star)
- [filter metrics based on timestamp in seconds](testdata/time_timestamp.star)
- [filter metrics based on the timestamp with nanoseconds](testdata/time_timestamp_nanos.star)
- [setting metric timestamp to current/local time](testdata/time_set_timestamp.star)
- [filter metric based on field value](testdata/value_filter.star)
- [log messages with Telegraf logger](testdata/logging.star)
- [return multiple metrics using a list](testdata/multiple_metrics.star)
- [return multiple metrics from JSON array](testdata/multiple_metrics_with_json.star)
- [return custom error using `fail`](testdata/fail.star)
- [compare metric with previous metric using a shared state](testdata/compare_metrics.star)
- [rename prometheus remote-write measurement name](testdata/rename_prometheus_remote_write.star)

[All examples](testdata) are in the testdata folder.

Open a Pull Request to add any other useful Starlark examples.

[string]: https://github.com/google/starlark-go/blob/d1966c6b9fcd/doc/spec.md#strings
[dict]: https://github.com/google/starlark-go/blob/d1966c6b9fcd/doc/spec.md#dictionaries
