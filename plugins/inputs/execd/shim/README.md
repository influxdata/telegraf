# Telegraf Execd Go Shim

The goal of this _shim_ is to make it trivial to extract an internal input plugin
out to a stand-alone repo for the purpose of compiling it as a separate app and
running it from the inputs.execd plugin.

The execd-shim is still experimental and the interface may change in the future.
Especially as the concept expands to processors, aggregators, and outputs.

## Steps to externalize a plugin

1. Move the project to an external repo, optionally preserving the
  _plugins/inputs/plugin_name_ folder structure. For an example of what this might
  look at, take a look at [ssoroka/rand](https://github.com/ssoroka/rand) or
  [danielnelson/telegraf-plugins](https://github.com/danielnelson/telegraf-plugins)
1. Copy [main.go](./example/cmd/main.go) into your project under the cmd folder.
  This will be the entrypoint to the plugin when run as a stand-alone program, and
  it will call the shim code for you to make that happen.
1. Edit the main.go file to import your plugin. Within Telegraf this would have
  been done in an all.go file, but here we don't split the two apart, and the change
  just goes in the top of main.go. If you skip this step, your plugin will do nothing.
1. Optionally add a [plugin.conf](./example/cmd/plugin.conf) for configuration
  specific to your plugin. Note that this config file **must be separate from the
  rest of the config for Telegraf, and must not be in a shared directory where
  Telegraf is expecting to load all configs**. If Telegraf reads this config file
  it will not know which plugin it relates to.

## Steps to build and run your plugin

1. Build the cmd/main.go. For my rand project this looks like `go build -o rand cmd/main.go`
1. Test out the binary if you haven't done this yet. eg `./rand -config plugin.conf`
  Depending on your polling settings and whether you implemented a service plugin or
  an input gathering plugin, you may see data right away, or you may have to hit enter
  first, or wait for your poll duration to elapse, but the metrics will be written to
  STDOUT. Ctrl-C to end your test.
1. Configure Telegraf to call your new plugin binary. eg:

```
[[inputs.execd]]
  command = ["/path/to/rand", "-config", "/path/to/plugin.conf"]
  signal = "none"
```

## Congratulations!

You've done it! Consider publishing your plugin to github and open a Pull Request
back to the Telegraf repo letting us know about the availability of your
[external plugin](https://github.com/influxdata/telegraf/blob/master/EXTERNAL_PLUGINS.md).