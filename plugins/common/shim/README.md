# Telegraf Execd Go Shim

The goal of this _shim_ is to make it trivial to extract an internal input,
processor, or output plugin from the main Telegraf repo out to a stand-alone
repo. This allows anyone to build and run it as a separate app using one of the
execd plugins:

- [inputs.execd](/plugins/inputs/execd)
- [processors.execd](/plugins/processors/execd)
- [outputs.execd](/plugins/outputs/execd)

## Steps to externalize a plugin

1. Move the project to an external repo, it's recommended to preserve the path
  structure, (but not strictly necessary). eg if your plugin was at
  `plugins/inputs/cpu`, it's recommended that it also be under `plugins/inputs/cpu`
  in the new repo. For a further example of what this might look like, take a
  look at [ssoroka/rand](https://github.com/ssoroka/rand) or
  [danielnelson/telegraf-plugins](https://github.com/danielnelson/telegraf-plugins)
1. Copy [main.go](./example/cmd/main.go) into your project under the `cmd` folder.
  This will be the entrypoint to the plugin when run as a stand-alone program, and
  it will call the shim code for you to make that happen. It's recommended to
  have only one plugin per repo, as the shim is not designed to run multiple
  plugins at the same time (it would vastly complicate things).
1. Edit the main.go file to import your plugin. Within Telegraf this would have
  been done in an all.go file, but here we don't split the two apart, and the change
  just goes in the top of main.go. If you skip this step, your plugin will do nothing.
  eg: `_ "github.com/me/my-plugin-telegraf/plugins/inputs/cpu"`
1. Optionally add a [plugin.conf](./example/cmd/plugin.conf) for configuration
  specific to your plugin. Note that this config file **must be separate from the
  rest of the config for Telegraf, and must not be in a shared directory where
  Telegraf is expecting to load all configs**. If Telegraf reads this config file
  it will not know which plugin it relates to. Telegraf instead uses an execd config
  block to look for this plugin.

## Steps to build and run your plugin

1. Build the cmd/main.go. For my rand project this looks like `go build -o rand cmd/main.go`
1. If you're building an input, you can test out the binary just by running it.
  eg `./rand -config plugin.conf`
  Depending on your polling settings and whether you implemented a service plugin or
  an input gathering plugin, you may see data right away, or you may have to hit enter
  first, or wait for your poll duration to elapse, but the metrics will be written to
  STDOUT. Ctrl-C to end your test.
  If you're testig a processor or output manually, you can still do this but you
  will need to feed valid metrics in on STDIN to verify that it is doing what you
  want. This can be a very valuable debugging technique before hooking it up to
  Telegraf.
1. Configure Telegraf to call your new plugin binary. For an input, this would
  look something like:

```toml
[[inputs.execd]]
  command = ["/path/to/rand", "-config", "/path/to/plugin.conf"]
  signal = "none"
```

  Refer to the execd plugin readmes for more information.

## Congratulations

You've done it! Consider publishing your plugin to github and open a Pull Request
back to the Telegraf repo letting us know about the availability of your
[external plugin](https://github.com/influxdata/telegraf/blob/master/EXTERNAL_PLUGINS.md).
