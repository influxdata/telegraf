package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	// TODO: import your plugins
	// _ "github.com/my_github_user/my_plugin_repo/plugins/inputs/mypluginname"

	"github.com/influxdata/telegraf/plugins/inputs/execd/shim"
)

var pollInterval = flag.Duration("poll_interval", 1*time.Second, "how often to send metrics")
var pollIntervalDisabled = flag.Bool("poll_interval_disabled", false, "how often to send metrics")
var configFile = flag.String("config", "", "path to the config file for this plugin")
var err error

// This is designed to be simple; Just change the import above and you're good.
//
// However, if you want to do all your config in code, you can like so:
//
// // initialize your plugin with any settngs you want
// myInput := &mypluginname.MyPlugin{
// 	DefaultSettingHere: 3,
// }
//
// shim := shim.New()
//
// shim.AddInput(myInput)
//
// // now the shim.Run() call as below.
//
func main() {
	// parse command line options
	flag.Parse()
	if *pollIntervalDisabled {
		*pollInterval = shim.PollIntervalDisabled
	}

	// create the shim. This is what will run your plugins.
	shim := shim.New()

	// If no config is specified, all imported plugins are loaded.
	// otherwise follow what the config asks for.
	// Check for settings from a config toml file,
	// (or just use whatever plugins were imported above)
	err = shim.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err loading input: %s\n", err)
		os.Exit(1)
	}

	// run the input plugin(s) until stdin closes or we receive a termination signal
	if err := shim.Run(*pollInterval); err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
		os.Exit(1)
	}
}
