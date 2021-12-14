# External Plugins

[External plugins](/EXTERNAL_PLUGINS.md) are external programs that are built outside
of Telegraf that can run through an `execd` plugin. These external plugins allow for
more flexibility compared to internal Telegraf plugins.

- External plugins can be written in any language (internal Telegraf plugins can only written in Go)
- External plugins can access to libraries not written in Go
- Utilize licensed software that isn't available to the open source community
- Can include large dependencies that would otherwise bloat Telegraf
- You don't need to wait on the Telegraf team to publish your plugin and start working with it.
- using the [shim](/plugins/common/shim) you can easily convert plugins between internal and external use

## External Plugin Guidelines

The guidelines of writing external plugins would follow those for our general [input](/docs/INPUTS.md),
[output](/docs/OUTPUTS.md), [processor](/docs/PROCESSORS.md), and [aggregator](/docs/AGGREGATORS.md) plugins.
Please reference the documentation on how to create these plugins written in Go.

_For listed [external plugins](/EXTERNAL_PLUGINS.md), the author of the external plugin is also responsible for the maintenance
and feature development of external plugins. Expect to have users open plugin issues on its respective GitHub repository._

### Execd Go Shim

For Go plugins, there is a [Execd Go Shim](/plugins/common/shim/) that will make it trivial to extract an internal input, processor, or output plugin from the main Telegraf repo out to a stand-alone repo. This shim allows anyone to build and run it as a separate app using one of the `execd`plugins:

- [inputs.execd](/plugins/inputs/execd)
- [processors.execd](/plugins/processors/execd)
- [outputs.execd](/plugins/outputs/execd)

Follow the [Steps to externalize a plugin](/plugins/common/shim#steps-to-externalize-a-plugin) and [Steps to build and run your plugin](/plugins/common/shim#steps-to-build-and-run-your-plugin) to properly with the Execd Go Shim

### Step-by-Step guidelines

This is a guide to help you set up your plugin to use it with `execd`:

1. Write your Telegraf plugin. Depending on the plugin, follow the guidelines on how to create the plugin itself using InfluxData's best practices:
   - [Input Plugins](/docs/INPUTS.md)
   - [Processor Plugins](/docs/PROCESSORS.md)
   - [Aggregator Plugins](/docs/AGGREGATORS.md)
   - [Output Plugins](/docs/OUTPUTS.md)
2. If your plugin is written in Go, include the steps for the [Execd Go Shim](/plugins/common/shim#steps-to-build-and-run-your-plugin)
   - Move the project to an external repo, it's recommended to preserve the path
   structure, (but not strictly necessary). eg if your plugin was at
   `plugins/inputs/cpu`, it's recommended that it also be under `plugins/inputs/cpu`
   in the new repo. For a further example of what this might look like, take a
   look at [ssoroka/rand](https://github.com/ssoroka/rand) or
   [danielnelson/telegraf-execd-openvpn](https://github.com/danielnelson//telegraf-execd-openvpn)
   - Copy [main.go](/plugins/common/shim/example/cmd/main.go) into your project under the `cmd` folder.
   This will be the entrypoint to the plugin when run as a stand-alone program, and
   it will call the shim code for you to make that happen. It's recommended to
   have only one plugin per repo, as the shim is not designed to run multiple
   plugins at the same time (it would vastly complicate things).
   - Edit the main.go file to import your plugin. Within Telegraf this would have
   been done in an all.go file, but here we don't split the two apart, and the change
   just goes in the top of main.go. If you skip this step, your plugin will do nothing.
   eg: `_ "github.com/me/my-plugin-telegraf/plugins/inputs/cpu"`
   - Optionally add a [plugin.conf](./example/cmd/plugin.conf) for configuration
   specific to your plugin. Note that this config file **must be separate from the
   rest of the config for Telegraf, and must not be in a shared directory where
   Telegraf is expecting to load all configs**. If Telegraf reads this config file
   it will not know which plugin it relates to. Telegraf instead uses an execd config
   block to look for this plugin.
   - Add usage and development instructions in the homepage of your repository for running
   your plugin with its respective `execd` plugin. Please refer to
   [openvpn](https://github.com/danielnelson/telegraf-execd-openvpn#usage) and [awsalarms](https://github.com/vipinvkmenon/awsalarms#installation)
   for examples. Include the following steps:
     1. How to download the release package for your platform or how to clone the binary for your external plugin
     1. The commands to build your binary
     1. Location to edit your `telegraf.conf`
     1. Configuration to run your external plugin with [inputs.execd](/plugins/inputs/execd),
     [processors.execd](/plugins/processors/execd) or [outputs.execd](/plugins/outputs/execd)
   - Submit your plugin by opening a PR to add your external plugin to the [/EXTERNAL_PLUGINS.md](/EXTERNAL_PLUGINS.md)
   list. Please include the plugin name, link to the plugin repository and a short description of the plugin.
