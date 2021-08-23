package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log" // nolint:revive
	"os"
	"path/filepath"

	"github.com/influxdata/toml"
)

const usageText = `Bob-the-builder, The configuration tool for generating stripped-down telegraf binaries.

Usage:

  bob [flags] [config file]

The flags are:

  --allyesconfig                 generate a configuration with all plugins enabled
  --allnoconfig                  generate a configuration with all plugins disabled
	--fallback										 generate and save an all-yes-config if loading the config file fails
	--list                         list all configurable plugins
  --save                         save configuration to file

  config file                    file to read the configuration from (default: build.conf)


Examples:

  # generate telegraf with all plugins enabled:
  go run buildconfig/bob.go --allyesconfig
  make

  # generate config with all plugins disabled for manual selection of few plugins
  go run buildconfig/bob.go --allnoconfig --save

  # build with custom config file
  go run buildconfig/bob.go mybuildsetup.conf
  make
`

var fConfigYes = flag.Bool("allyesconfig", false, "generate a configuration with all plugins enabled")
var fConfigNo = flag.Bool("allnoconfig", false, "generate a configuration with all plugins disabled")
var fFallback = flag.Bool("fallback", false, "generate and save an all-yes-config if loading the config file fails")
var fList = flag.Bool("list", false, "list all configurable plugins")
var fSave = flag.Bool("save", false, "save configuration to file")

type buildConfig struct {
	Inputs      map[string]bool
	Outputs     map[string]bool
	Processors  map[string]bool
	Aggregators map[string]bool
}

func isExcluded(name string, excludes ...string) bool {
	for _, e := range excludes {
		if name == e {
			return true
		}
	}
	return false
}

func getPlugins(category string, excludes ...string) ([]string, error) {
	dirname := filepath.Join("plugins", category)
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	plugins := []string{}
	for _, file := range files {
		if !file.IsDir() || isExcluded(file.Name(), excludes...) {
			continue
		}

		if matches, _ := filepath.Glob(filepath.Join(dirname, file.Name(), "*.go")); len(matches) > 0 {
			// No subprojects expected
			plugins = append(plugins, file.Name())
		} else {
			// No go files found, check subprojects
			subprojects, err := getPlugins(filepath.Join(category, file.Name()), excludes...)
			if err != nil {
				return nil, err
			}
			for _, subproject := range subprojects {
				plugins = append(plugins, filepath.Join(file.Name(), subproject))
			}
		}
	}

	return plugins, nil
}

func generatePluginsFile(category string, plugins []string) error {
	path := filepath.Join("plugins", category, "all")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return fmt.Errorf("creating dir %q failed: %v", path, err)
		}
	}

	filename := filepath.Join(path, "all.go")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("package all\n\nimport (\n"); err != nil {
		return err
	}
	for _, plugin := range plugins {
		fn := fmt.Sprintf("\t_ \"github.com/influxdata/telegraf/plugins/%s/%s\"\n", category, plugin)
		if _, err := f.WriteString(fn); err != nil {
			return err
		}
	}
	if _, err := f.WriteString(")\n"); err != nil {
		return err
	}
	return f.Sync()
}

func getAllPlugins(categories []string) (map[string][]string, error) {
	plugins := make(map[string][]string)

	for _, category := range categories {
		plugins[category] = make([]string, 0)
		pluginslist, err := getPlugins(category, "all")
		if err != nil {
			return nil, err
		}
		plugins[category] = append(plugins[category], pluginslist...)
	}

	return plugins, nil
}

func getConfiguredPlugins(cfg buildConfig, allPlugins map[string][]string) map[string][]string {
	config := make(map[string][]string)

	categories := map[string](map[string]bool){
		"inputs":      cfg.Inputs,
		"outputs":     cfg.Outputs,
		"processors":  cfg.Processors,
		"aggregators": cfg.Aggregators,
	}

	for category, list := range categories {
		config[category] = make([]string, 0)
		for k, v := range list {
			if !v {
				continue
			}
			if k == "*" {
				if list, ok := allPlugins[category]; ok {
					config[category] = list
				} else {
					log.Printf("W! [bob] No all-plugins list for category %q", category)
				}
				break
			}
			config[category] = append(config[category], k)
		}
	}

	return config
}

func readConfig(filename string) (buildConfig, error) {
	log.Printf("I! [bob] Reading configuration from %q...", filename)
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return buildConfig{}, err
	}
	var config buildConfig
	if err := toml.Unmarshal(buf, &config); err != nil {
		return buildConfig{}, err
	}
	return config, nil
}

func toConfig(plugins map[string][]string, val bool) buildConfig {
	config := buildConfig{}

	// Convert to config format
	config.Inputs = make(map[string]bool)
	config.Outputs = make(map[string]bool)
	config.Processors = make(map[string]bool)
	config.Aggregators = make(map[string]bool)
	if list, ok := plugins["inputs"]; ok {
		for _, plugin := range list {
			config.Inputs[plugin] = val
		}
	}
	if list, ok := plugins["outputs"]; ok {
		for _, plugin := range list {
			config.Outputs[plugin] = val
		}
	}
	if list, ok := plugins["processors"]; ok {
		for _, plugin := range list {
			config.Processors[plugin] = val
		}
	}
	if list, ok := plugins["aggregators"]; ok {
		for _, plugin := range list {
			config.Aggregators[plugin] = val
		}
	}

	return config
}

func writeConfig(filename string, config buildConfig) error {
	log.Printf("I! [bob] Writing configuration to %q...", filename)
	buf, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0644)
}

func printPlugins(plugins map[string][]string) {
	for name, list := range plugins {
		//nolint:revive
		fmt.Printf("%s:\n", name)
		for _, plugin := range list {
			//nolint:revive
			fmt.Printf("\t%s\n", plugin)
		}
	}
}

func usage() {
	//nolint:revive
	fmt.Println(usageText)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	// Check the command-line options
	if len(args) > 1 {
		log.Fatalf("E! [bob] Too many command line arguments!")
	}

	// Determine config filename
	configFile := "build.conf"
	if len(args) > 0 {
		configFile = args[0]
	}

	// Trigger the fallback
	if *fFallback {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			*fConfigYes = true
			*fConfigNo = false
			*fSave = true
		}
	}

	// Define the categories we can handle
	categories := []string{"inputs", "outputs", "processors", "aggregators"}

	// Get the list of all configurable plugins
	allPlugins, err := getAllPlugins(categories)
	if err != nil {
		log.Fatalf("E! [bob] Cannot determine list of configurable plugins: %v", err)
	}

	// Handle listing request
	if *fList {
		if *fSave {
			config := toConfig(allPlugins, false)
			err = writeConfig(configFile, config)
			if err != nil {
				log.Fatalf("E! [bob] Cannot write list of plugins to %q: %v", configFile, err)
			}
		} else {
			printPlugins(allPlugins)
		}
		return
	}

	// Generate all yes/no configurations if specified
	var config buildConfig
	if *fConfigYes && *fConfigNo {
		log.Fatalf("E! [bob] Cannot use 'allyesconfig' and 'allnoconfig' at the same time!")
	} else if *fConfigYes {
		config = toConfig(allPlugins, true)
		if *fSave {
			err = writeConfig(configFile, config)
			if err != nil {
				log.Fatalf("E! [bob] Cannot write list of plugins to %q: %v", configFile, err)
			}
		}
	} else if *fConfigNo {
		config = toConfig(allPlugins, false)
		if *fSave {
			err = writeConfig(configFile, config)
			if err != nil {
				log.Fatalf("E! [bob] Cannot write list of plugins to %q: %v", configFile, err)
			}
		}
	} else {
		// Read the build configuration file
		config, err = readConfig(configFile)
		if err != nil {
			log.Fatalf("E! [bob] Cannot read configuration file %q: %v", configFile, err)
		}
	}

	// Filter only activated plugins and generate the include files
	configuredPlugins := getConfiguredPlugins(config, allPlugins)
	for category, plugins := range configuredPlugins {
		log.Printf("I! [bob] Generating %s-plugins file with %d entries...", category, len(plugins))
		err = generatePluginsFile(category, plugins)
		if err != nil {
			log.Fatalf("E! [bob] Cannot generate %s-plugins file: %v", category, err)
		}
	}
}
