package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var buildTargets = []string{"build"}

var categories = []string{
	"aggregators",
	"inputs",
	"outputs",
	"parsers",
	"processors",
	"secretstores",
	"serializers",
}

const description = `
This is a tool build Telegraf with a custom set of plugins. The plugins are
select according to the specified Telegraf configuration files. This allows
to shrink the binary size by only selecting the plugins you really need.
A more detailed documentation is available at
http://github.com/influxdata/telegraf/tools/custom_builder/README.md
`

const examples = `
The following command with customize Telegraf to fit the configuration found
at the default locations

  custom_builder --config /etc/telegraf/telegraf.conf --config-dir /etc/telegraf/telegraf.d

You can the --config and --config-dir multiple times

  custom_builder --config global.conf --config myinputs.conf --config myoutputs.conf

or use one or more remote address(es) to load the config

  custom_builder --config global.conf --config http://myserver/plugins.conf

Combinations of local and remote config as well as config directories are
possible.
`

func usage() {
	fmt.Fprint(flag.CommandLine.Output(), description)
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
	fmt.Fprintln(flag.CommandLine.Output(), "  custom_builder [flags]")
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
	flag.PrintDefaults()
	fmt.Fprintln(flag.CommandLine.Output(), "")
	fmt.Fprintln(flag.CommandLine.Output(), "Examples:")
	fmt.Fprint(flag.CommandLine.Output(), examples)
	fmt.Fprintln(flag.CommandLine.Output(), "")
}

type cmdConfig struct {
	dryrun      bool
	showtags    bool
	migrations  bool
	quiet       bool
	root        string
	configFiles []string
	configDirs  []string
}

func main() {
	var cfg cmdConfig
	flag.Func("config",
		"Import plugins from configuration file (can be used multiple times)",
		func(s string) error {
			cfg.configFiles = append(cfg.configFiles, s)
			return nil
		},
	)
	flag.Func("config-dir",
		"Import plugins from configs in the given directory (can be used multiple times)",
		func(s string) error {
			cfg.configDirs = append(cfg.configDirs, s)
			return nil
		},
	)
	flag.BoolVar(&cfg.dryrun, "dry-run", false, "Skip the actual building step")
	flag.BoolVar(&cfg.quiet, "quiet", false, "Print fewer log messages")
	flag.BoolVar(&cfg.migrations, "migrations", false, "Include configuration migrations")
	flag.BoolVar(&cfg.showtags, "tags", false, "Show build-tags used")

	flag.Usage = usage
	flag.Parse()

	tagset, err := process(&cfg)
	if err != nil {
		log.Fatalln(err)
	}
	if len(tagset) == 0 {
		log.Fatalln("Nothing selected!")
	}
	tags := "custom,"
	if cfg.migrations {
		tags += "migrations,"
	}
	tags += strings.Join(tagset, ",")
	if cfg.showtags {
		fmt.Printf("Build tags: %s\n", tags)
	}

	if !cfg.dryrun {
		// Perform the build
		var out bytes.Buffer
		makeCmd := exec.Command("make", buildTargets...)
		makeCmd.Env = append(os.Environ(), "BUILDTAGS="+tags)
		makeCmd.Stdout = &out
		makeCmd.Stderr = &out

		if !cfg.quiet {
			log.Println("Running build...")
		}
		if err := makeCmd.Run(); err != nil {
			fmt.Println(out.String())
			log.Fatalf("Running make failed: %v", err)
		}
		if !cfg.quiet {
			fmt.Println(out.String())
		}
	} else if !cfg.quiet {
		log.Println("DRY-RUN: Skipping build.")
	}
}

func process(cmdcfg *cmdConfig) ([]string, error) {
	// Check configuration options
	if len(cmdcfg.configFiles) == 0 && len(cmdcfg.configDirs) == 0 {
		return nil, errors.New("no configuration specified")
	}

	// Collect all available plugins
	packages := packageCollection{root: cmdcfg.root}
	if err := packages.CollectAvailable(); err != nil {
		return nil, fmt.Errorf("collecting plugins failed: %w", err)
	}

	// Import the plugin list from Telegraf configuration files
	log.Println("Importing configuration file(s)...")
	cfg, nfiles, err := ImportConfigurations(cmdcfg.configFiles, cmdcfg.configDirs)
	if err != nil {
		return nil, fmt.Errorf("importing configuration(s) failed: %w", err)
	}
	if !cmdcfg.quiet {
		log.Printf("Found %d configuration files...", nfiles)
	}

	// Check if we do have a config
	if nfiles == 0 {
		return nil, errors.New("no configuration files loaded")
	}

	// Process the plugin list with the given config. This will
	// only keep the plugins that adhere to the filtering criteria.
	enabled, err := cfg.Filter(packages)
	if err != nil {
		return nil, fmt.Errorf("filtering packages failed: %w", err)
	}
	if !cmdcfg.quiet {
		enabled.Print()
	}

	// Extract the build-tags
	return enabled.ExtractTags(), nil
}
