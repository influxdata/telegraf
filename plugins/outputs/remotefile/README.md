# Remote File Output Plugin

This plugin writes metrics to files in a remote location using the
[rclone library][rclone]. Currently the following backends are supported:

- `local`: [Local filesystem](https://rclone.org/local/)
- `s3`: [Amazon S3 storage providers](https://rclone.org/s3/)
- `sftp`: [Secure File Transfer Protocol](https://rclone.org/sftp/)

⭐ Telegraf v1.32.0
🏷️ datastore
💻 all

[rclone]: https://rclone.org

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `remote` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Send telegraf metrics to file(s) in a remote filesystem
[[outputs.remotefile]]
  ## Remote location according to https://rclone.org/#providers
  ## Check the backend configuration options and specify them in
  ##   <backend type>[,<param1>=<value1>[,...,<paramN>=<valueN>]]:[root]
  ## for example:
  ##   remote = 's3,provider=AWS,access_key_id=...,secret_access_key=...,session_token=...,region=us-east-1:mybucket'
  ## By default, remote is the local current directory
  # remote = "local:"

  ## Files to write in the remote location
  ## Each file can be a Golang template for generating the filename from metrics.
  ## See https://pkg.go.dev/text/template for a reference and use the metric
  ## name (`{{.Name}}`), tag values (`{{.Tag "name"}}`), field values
  ## (`{{.Field "name"}}`) or the metric time (`{{.Time}}) to derive the
  ## filename.
  ## The 'files' setting may contain directories relative to the root path
  ## defined in 'remote'.
  files = ['{{.Name}}-{{.Time.Format "2006-01-02"}}']

  ## Use batch serialization format instead of line based delimiting.
  ## The batch format allows for the production of non-line-based output formats
  ## and may more efficiently encode metrics.
  # use_batch_format = false

  ## Cache settings
  ## Time to wait for all writes to complete on shutdown of the plugin.
  # final_write_timeout = "10s"

  ## Time to wait between writing to a file and uploading to the remote location
  # cache_write_back = "5s"

  ## Maximum size of the cache on disk (infinite by default)
  # cache_max_size = -1

  ## Forget files after not being touched for longer than the given time
  ## This is useful to prevent memory leaks when using time-based filenames
  ## as it allows internal structures to be cleaned up.
  ## Note: When writing to a file after is has been forgotten, the file is
  ##       treated as a new file which might cause file-headers to be appended
  ##       again by certain serializers like CSV.
  ## By default files will be kept indefinitely.
  # forget_files_after = "0s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
  
  ## Compress output data with the specified algorithm.
  ## If empty, compression will be disabled and files will be plain text.
  ## Supported algorithms are "zstd", "gzip" and "zlib".
  # compression_algorithm = ""

  ## Compression level for the algorithm above.
  ## Please note that different algorithms support different levels:
  ##   zstd  -- supports levels 1, 3, 7 and 11.
  ##   gzip -- supports levels 0, 1 and 9.
  ##   zlib -- supports levels 0, 1, and 9.
  ## By default the default compression level for each algorithm is used.
  # compression_level = -1
```

## Available custom functions

The following functions can be used in the templates:

- `now`: returns the current time (example: `{{now.Format "2006-01-02"}}`)
