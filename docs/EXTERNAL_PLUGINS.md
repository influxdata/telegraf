### External Plugins

External plugins are external programs that are built outside of Telegraf that
can run through an `execd` plugin. These external plugins allow for more flexibility
compared to internal Telegraf plugins.  

- External plugins can be written in any language (internal Telegraf plugins can only written in Go)
- External plugins can access to libraries not written in Go
- Utilize licensed software that isn't available to the open source community
- Can include large dependencies that would otherwise bloat Telegraf

### External Plugin Guidelines
The guidelines of writing external plugins would follow those for our general [input](docs/INPUTS.md), 
[output](docs/OUTPUTS.md), [processor](docs/PROCESSORS.md), and [aggregator](docs/AGGREGATOR.md) plugins. 
Please reference the documentation on how to create these plugins written in Go.


#### Execd Go Shim
For Go plugins, there is a [Execd Go Shim](plugins/common/shim) that will make it trivial to extract an internal input, processor, or output plugin from the main Telegraf repo out to a stand-alone repo.  This shim This allows anyone to build and run it as a separate app using one of the `execd`plugins:
- [inputs.execd](/plugins/inputs/execd)
- [processors.execd](/plugins/processors/execd)
- [outputs.execd](/plugins/outputs/execd)

Follow the [Steps to externalize a plugin](plugins/common/shim#steps-to-externalize-a-plugin) and [Steps to build and run your plugin](plugins/common/shim#steps-to-build-and-run-your-plugin) to properly with the Execd Go Shim

#### Step-by-Step guidelines
This is a guide to help you set up your plugin to use it with `execd`
1. Write your Telegraf plugin.  Depending on the plugin, follow the guidelines on how to create the plugin itself using InfluxData's best practices:
   - [Input Plugins](/docs/INPUTS.md)
   - [Processor Plugins](/docs/PROCESSORS.md)
   - [Aggregator Plugins](/docs/AGGREGATORS.md)
   - [Output Plugins](docs/OUTPUTS.md)
2. If your plugin is written in Go, include the steps for the [Execd Go Shim](plugins/common/shim#steps-to-build-and-run-your-plugin)
  1. Move the project to an external repo, it's recommended to preserve the path
  structure, (but not strictly necessary). eg if your plugin was at
  `plugins/inputs/cpu`, it's recommended that it also be under `plugins/inputs/cpu`
  in the new repo. For a further example of what this might look like, take a
  look at [ssoroka/rand](https://github.com/ssoroka/rand) or
  [danielnelson/telegraf-execd-openvpn](https://github.com/danielnelson//telegraf-execd-openvpn)
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
 





