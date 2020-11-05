package shim

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors"
)

type config struct {
	Inputs     map[string][]toml.Primitive
	Processors map[string][]toml.Primitive
	Outputs    map[string][]toml.Primitive
}

type loadedConfig struct {
	Input     telegraf.Input
	Processor telegraf.StreamingProcessor
	Output    telegraf.Output
}

// LoadConfig Adds plugins to the shim
func (s *Shim) LoadConfig(filePath *string) error {
	conf, err := LoadConfig(filePath)
	if err != nil {
		return err
	}
	if conf.Input != nil {
		if err = s.AddInput(conf.Input); err != nil {
			return fmt.Errorf("Failed to add Input: %w", err)
		}
	} else if conf.Processor != nil {
		if err = s.AddStreamingProcessor(conf.Processor); err != nil {
			return fmt.Errorf("Failed to add Processor: %w", err)
		}
	} else if conf.Output != nil {
		if err = s.AddOutput(conf.Output); err != nil {
			return fmt.Errorf("Failed to add Output: %w", err)
		}
	}
	return nil
}

// LoadConfig loads the config and returns inputs that later need to be loaded.
func LoadConfig(filePath *string) (loaded loadedConfig, err error) {
	var data string
	conf := config{}
	if filePath != nil && *filePath != "" {

		b, err := ioutil.ReadFile(*filePath)
		if err != nil {
			return loadedConfig{}, err
		}

		data = expandEnvVars(b)

	} else {
		conf, err = DefaultImportedPlugins()
		if err != nil {
			return loadedConfig{}, err
		}
	}

	md, err := toml.Decode(data, &conf)
	if err != nil {
		return loadedConfig{}, err
	}

	return createPluginsWithTomlConfig(md, conf)
}

func expandEnvVars(contents []byte) string {
	return os.Expand(string(contents), getEnv)
}

func getEnv(key string) string {
	v := os.Getenv(key)

	return envVarEscaper.Replace(v)
}

func createPluginsWithTomlConfig(md toml.MetaData, conf config) (loadedConfig, error) {
	loadedConf := loadedConfig{}

	for name, primitives := range conf.Inputs {
		creator, ok := inputs.Inputs[name]
		if !ok {
			return loadedConf, errors.New("unknown input " + name)
		}

		plugin := creator()
		if len(primitives) > 0 {
			primitive := primitives[0]
			if err := md.PrimitiveDecode(primitive, plugin); err != nil {
				return loadedConf, err
			}
		}

		loadedConf.Input = plugin
		break
	}

	for name, primitives := range conf.Processors {
		creator, ok := processors.Processors[name]
		if !ok {
			return loadedConf, errors.New("unknown processor " + name)
		}

		plugin := creator()
		if len(primitives) > 0 {
			primitive := primitives[0]
			var p telegraf.PluginDescriber = plugin
			if processor, ok := plugin.(unwrappable); ok {
				p = processor.Unwrap()
			}
			if err := md.PrimitiveDecode(primitive, p); err != nil {
				return loadedConf, err
			}
		}
		loadedConf.Processor = plugin
		break
	}

	for name, primitives := range conf.Outputs {
		creator, ok := outputs.Outputs[name]
		if !ok {
			return loadedConf, errors.New("unknown output " + name)
		}

		plugin := creator()
		if len(primitives) > 0 {
			primitive := primitives[0]
			if err := md.PrimitiveDecode(primitive, plugin); err != nil {
				return loadedConf, err
			}
		}
		loadedConf.Output = plugin
		break
	}
	return loadedConf, nil
}

// DefaultImportedPlugins defaults to whatever plugins happen to be loaded and
// have registered themselves with the registry. This makes loading plugins
// without having to define a config dead easy.
func DefaultImportedPlugins() (config, error) {
	conf := config{
		Inputs:     map[string][]toml.Primitive{},
		Processors: map[string][]toml.Primitive{},
		Outputs:    map[string][]toml.Primitive{},
	}
	for name := range inputs.Inputs {
		log.Println("No config found. Loading default config for plugin", name)
		conf.Inputs[name] = []toml.Primitive{}
		return conf, nil
	}
	for name := range processors.Processors {
		log.Println("No config found. Loading default config for plugin", name)
		conf.Processors[name] = []toml.Primitive{}
		return conf, nil
	}
	for name := range outputs.Outputs {
		log.Println("No config found. Loading default config for plugin", name)
		conf.Outputs[name] = []toml.Primitive{}
		return conf, nil
	}
	return conf, nil
}

type unwrappable interface {
	Unwrap() telegraf.Processor
}
