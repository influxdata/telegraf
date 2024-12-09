# Customization

You can build customized versions of Telegraf with a specific plugin set using
the [custom builder](/tools/custom_builder) tool or
[build-tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints).
For build tags, the plugins can be selected either category-wise, i.e.
`inputs`, `outputs`,`processors`, `aggregators`, `parsers`, `secretstores`
and `serializers` or individually, e.g. `inputs.modbus` or `outputs.influxdb`.

Usually the build tags correspond to the plugin names used in the Telegraf
configuration. To be sure, check the files in the corresponding
`plugin/<category>/all` directory. Make sure to include all parsers you intend
to use.

__Note:__ You _always_ need to include the `custom` tag when customizing the
build as otherwise _all_ plugins will be selected regardless of other tags.

## Via make

When using the project's makefile, the build can be customized via the
`BUILDTAGS` environment variable containing a __comma-separated__ list of the
selected plugins (or categories) __and__ the `custom` tag.

For example

```shell
BUILDTAGS="custom,inputs,outputs.influxdb_v2,parsers.json" make
```

will build a customized Telegraf including _all_ `inputs`, the InfluxDB v2
`output` and the `json` parser.

## Via `go build`

If you wish to build Telegraf using native go tools, you can use the `go build`
command with the `-tags` option. Specify  a __comma-separated__ list of the
selected plugins (or categories) __and__ the `custom` tag as argument.

For example

```shell
go build -tags "custom,inputs,outputs.influxdb_v2,parsers.json" ./cmd/telegraf
```

will build a customized Telegraf including _all_ `inputs`, the InfluxDB v2
`output` and the `json` parser.
