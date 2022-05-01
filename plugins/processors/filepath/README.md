<!-- markdownlint-disable MD024 -->

# Filepath Processor Plugin

The `filepath` processor plugin maps certain go functions from [path/filepath](https://golang.org/pkg/path/filepath/)
onto tag and field values. Values can be modified in place or stored in another key.

Implemented functions are:

* [Base](https://golang.org/pkg/path/filepath/#Base) (accessible through `[[processors.filepath.basename]]`)
* [Rel](https://golang.org/pkg/path/filepath/#Rel) (accessible through `[[processors.filepath.rel]]`)
* [Dir](https://golang.org/pkg/path/filepath/#Dir) (accessible through `[[processors.filepath.dir]]`)
* [Clean](https://golang.org/pkg/path/filepath/#Clean) (accessible through `[[processors.filepath.clean]]`)
* [ToSlash](https://golang.org/pkg/path/filepath/#ToSlash) (accessible through `[[processors.filepath.toslash]]`)

 On top of that, the plugin provides an extra function to retrieve the final path component without its extension. This
 function is accessible through the `[[processors.filepath.stem]]` configuration item.

Please note that, in this implementation, these functions are processed in the order that they appear above( except for
`stem` that is applied in the first place).

Specify the `tag` and/or `field` that you want processed in each section and optionally a `dest` if you want the result
stored in a new tag or field.

If you plan to apply multiple transformations to the same `tag`/`field`, bear in mind the processing order stated above.

Telegraf minimum version: Telegraf 1.15.0

## Configuration

```toml
# Performs file path manipulations on tags and fields
[[processors.filepath]]
  ## Treat the tag value as a path and convert it to its last element, storing the result in a new tag
  # [[processors.filepath.basename]]
  #   tag = "path"
  #   dest = "basepath"

  ## Treat the field value as a path and keep all but the last element of path, typically the path's directory
  # [[processors.filepath.dirname]]
  #   field = "path"

  ## Treat the tag value as a path, converting it to its the last element without its suffix
  # [[processors.filepath.stem]]
  #   tag = "path"

  ## Treat the tag value as a path, converting it to the shortest path name equivalent
  ## to path by purely lexical processing
  # [[processors.filepath.clean]]
  #   tag = "path"

  ## Treat the tag value as a path, converting it to a relative path that is lexically
  ## equivalent to the source path when joined to 'base_path'
  # [[processors.filepath.rel]]
  #   tag = "path"
  #   base_path = "/var/log"

  ## Treat the tag value as a path, replacing each separator character in path with a '/' character. Has only
  ## effect on Windows
  # [[processors.filepath.toslash]]
  #   tag = "path"
```

## Considerations

### Clean

Even though `clean` is provided a standalone function, it is also invoked when using the `rel` and `dirname` functions,
so there is no need to use it along with them.

That is:

 ```toml
[[processors.filepath]]
   [[processors.filepath.dir]]
     tag = "path"
   [[processors.filepath.clean]]
     tag = "path"
 ```

Is equivalent to:

 ```toml
[[processors.filepath]]
   [[processors.filepath.dir]]
     tag = "path"
 ```

### ToSlash

The effects of this function are only noticeable on Windows platforms, because of the underlying golang implementation.

## Examples

### Basename

```toml
[[processors.filepath]]
  [[processors.filepath.basename]]
    tag = "path"
```

```diff
- my_metric,path="/var/log/batch/ajob.log" duration_seconds=134 1587920425000000000
+ my_metric,path="ajob.log" duration_seconds=134 1587920425000000000
```

### Dirname

```toml
[[processors.filepath]]
  [[processors.filepath.dirname]]
    field = "path"
    dest = "folder"
```

```diff
- my_metric path="/var/log/batch/ajob.log",duration_seconds=134 1587920425000000000
+ my_metric path="/var/log/batch/ajob.log",folder="/var/log/batch",duration_seconds=134 1587920425000000000
```

### Stem

```toml
[[processors.filepath]]
  [[processors.filepath.stem]]
    tag = "path"
```

```diff
- my_metric,path="/var/log/batch/ajob.log" duration_seconds=134 1587920425000000000
+ my_metric,path="ajob" duration_seconds=134 1587920425000000000
```

### Clean

```toml
[[processors.filepath]]
  [[processors.filepath.clean]]
    tag = "path"
```

```diff
- my_metric,path="/var/log/dummy/../batch//ajob.log" duration_seconds=134 1587920425000000000
+ my_metric,path="/var/log/batch/ajob.log" duration_seconds=134 1587920425000000000
```

### Rel

```toml
[[processors.filepath]]
  [[processors.filepath.rel]]
    tag = "path"
    base_path = "/var/log"
```

```diff
- my_metric,path="/var/log/batch/ajob.log" duration_seconds=134 1587920425000000000
+ my_metric,path="batch/ajob.log" duration_seconds=134 1587920425000000000
```

### ToSlash

```toml
[[processors.filepath]]
  [[processors.filepath.rel]]
    tag = "path"
```

```diff
- my_metric,path="\var\log\batch\ajob.log" duration_seconds=134 1587920425000000000
+ my_metric,path="/var/log/batch/ajob.log" duration_seconds=134 1587920425000000000
```

## Processing paths from tail plugin

This plugin can be used together with the
[tail input plugn](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/tail) to make modifications
to the `path` tag injected for every file.

Scenario:

* A log file `/var/log/myjobs/mysql_backup.log`, containing logs for a job execution. Whenever the job ends, a line is
written to the log file following this format: `2020-04-05 11:45:21 total time execution: 70 seconds`
* We want to generate a measurement that captures the duration of the script as a field and includes the `path` as a
tag
  * We are interested in the filename without its extensions, since it might be enough information for plotting our
    execution times in a dashboard
  * Just in case, we don't want to override the original path (if for some reason we end up having duplicates we might
    want this information)

For this purpose, we will use the `tail` input plugin, the `grok` parser plugin and the `filepath` processor.

```toml
# Performs file path manipulations on tags and fields
[[inputs.tail]]
  files = ["/var/log/myjobs/**.log"]
  data_format = "grok"
  grok_patterns = ['%{TIMESTAMP_ISO8601:timestamp:ts-"2006-01-02 15:04:05"} total time execution: %{NUMBER:duration_seconds:int}']
  name_override = "myjobs"

[[processors.filepath]]
   [[processors.filepath.stem]]
     tag = "path"
     dest = "stempath"
```

The resulting output for a job taking 70 seconds for the mentioned log file would look like:

```text
myjobs_duration_seconds,host="my-host",path="/var/log/myjobs/mysql_backup.log",stempath="mysql_backup" 70 1587920425000000000
```
