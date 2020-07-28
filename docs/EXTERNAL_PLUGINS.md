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
To-be-added




