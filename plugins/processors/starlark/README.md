# Starlark Processor

The `starlark` processor calls a Starlark function for each matched metric, allowing for custom metric processing.

The Starlark language is a dialect of Python, and will be familier to those who
have experience with the Python language.  However, keep in mind that it is not
Python, and there are syntax [differences][#Python Differences].  Existing
Python code is unlikley to work unmodified.  The execution environment is
sandboxed, and it is not possible to do I/O.

The Starlark [specification][] has details about the syntax and available
functions.

### Configuration

```toml
[[processors.starlark]]
  source = """
  def apply(metric):
      for k, v in metric.fields.items():
          if type(v) == "float":
              metric.fields[k] = v * 10
      return metric
  """
```

### Usage

To get started the script should contain a function called `apply` that takes
one argument.  The function will be called with each metric, and can return a
single metric or a list of metrics:
```python
def apply(metric):
	return metric
```

In addition to the built-in functionality provided by the Starlark language, Telegraf exposes types and functions.

#### Metric object

class **Metric(name)**:

  Create a new metric with the given name.  The metric will have no tags or
  fields.  The timestamp defaults to the current time.

  **name**
  	The name is the metric name aka measurement.

  **tags**
  	A dict-like object containing the metric's tags.

  **fields**
  	A dict-like object containing the metric's fields.

  **time**
  	The timestamp of the metric as an integer in nanoseconds since the Unix
  	epoch.

#### Functions

**deepcopy(metric)**

  Make a copy of an existing metric.

### Python Differences

While Starlark is similar to Python, it is not the same language.

- Starlark has limited support for error handling and no exceptions.  If an
  error occurs the script will immediately exit.  Check the Telegraf logfile
  for details about the error.

- It is not possible to import other packages and the Python standard library
  is not available.  As such, it is not possible to open files or sockets.


### Gotchas

global scope is not modifiable.

don't return two references to the same metric.

error line number

### TODO

how to delete a metric?
- must call Drop?
- don't return: check returned values and autodrop

how to copy a metric?
- must call deepcopy()
- returning multiple references is an error

how to return multiple metrics?
- return a list of metric

if an error occurs do we drop the metric tracking?

how to create a new metric?

fastest way to iterate?

how to modify while iterating

how to retain metrics/modify globals
- global scope is froze

### Example

[specification]: https://github.com/google/starlark-go/blob/master/doc/spec.md
