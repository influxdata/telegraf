package persister

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/persister/store"
)

const SampleConfig = `
# Configuration for plugin state persistance
[persister]
	## Enables or disables state persistance.
	# enabled = false

  ## File for storing/loading the states to/from. If left empty, the persister is disabled.
	# filename = ""

	## Format of the state file. Can be
	##   "json" -- save the states in JSON format
	# file_format = "json"

	## Indentation level for nicely formatted (e.g. JSON) output.
	## If zero, no indentation is performed, saving diskspace but hampers readability.
	# indent = 0
`

// Persister is the instance for persisting states
type Persister struct {
	Enabled  bool   `toml:"enabled"`
	Filename string `toml:"filename"`
	Format   string `toml:"file_format"`
	Indent   int    `toml:"indent"`

	register map[string]telegraf.StatefulPlugin
	store    store.Store
	logger   telegraf.Logger
}

// Non-plugin facing (public) interface

func (p *Persister) Init() error {
	if !p.Enabled {
		return fmt.Errorf("init called on disabled persister")
	}

	p.logger = models.NewLogger("agent", "persister", "persister")

	switch p.Format {
	case "", "json":
		if p.Filename == "" {
			return fmt.Errorf("invalid filename for \"json\" store")
		}
		p.store = &store.JSONStore{
			Filename: p.Filename,
			Indent:   p.Indent,
		}
	default:
		return fmt.Errorf("unknown file-format %q for persister", p.Format)
	}
	if err := p.store.Init(); err != nil {
		return fmt.Errorf("initializing store failed: %v", err)
	}

	p.register = make(map[string]telegraf.StatefulPlugin)

	return nil
}

func (p *Persister) NewPluginWrapper(prefix string, plugin interface{}) (*PersisterPluginWrapper, error) {
	id, err := generatePluginID(prefix, plugin)
	if err != nil {
		return nil, err
	}

	wrapper := PersisterPluginWrapper{
		id:        id,
		persister: p,
	}

	return &wrapper, nil
}

func (p *Persister) Register(prefix string, plugin telegraf.StatefulPlugin) error {
	if p.register == nil {
		return fmt.Errorf("not initialized")
	}

	id, err := generatePluginID(prefix, plugin)
	if err != nil {
		return err
	}
	p.logger.Debugf("Registering plugin %q with id %q...", prefix, id)

	if _, found := p.register[id]; found {
		return fmt.Errorf("plugin with ID %q already registered", id)
	}
	p.register[id] = plugin
	return p.SetState(id, plugin.GetState())
}

func (p *Persister) GetState(id string) (interface{}, bool) {
	if p.store == nil {
		return nil, false
	}

	return p.store.GetState(id)
}

func (p *Persister) SetState(id string, state interface{}) error {
	if p.store == nil {
		return fmt.Errorf("not initialized")
	}
	p.logger.Debugf("Setting state of %q to %v", id, state)
	return p.store.SetState(id, state)
}

func (p *Persister) Load() error {
	if p.register == nil || p.store == nil {
		return fmt.Errorf("not initialized")
	}

	// Read the states from disk
	if err := p.store.Read(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p.logger.Infof("State file %q does not exist... Skip restoring states...", p.Filename)
			return nil
		}
		return fmt.Errorf("reading states file failed: %v", err)
	}

	// Get the initialized state as blueprint for unmarshalling
	for id, state := range p.store.GetStates() {
		// Check if we have a plugin with that ID
		plugin, found := p.register[id]
		if !found {
			p.logger.Info("No plugin for ID %q... Skipping.", id)
			continue
		}

		// Set the state in the plugin
		if err := plugin.SetState(state); err != nil {
			return fmt.Errorf("setting state of %v failed: %v", id, err)
		}
	}

	return nil
}

func (p *Persister) Store() error {
	if p.register == nil || p.store == nil {
		return fmt.Errorf("not initialized")
	}

	// Update the states before writing
	if err := p.collect(); err != nil {
		return fmt.Errorf("collection failed: %v", err)
	}

	// Write the states to disk
	return p.store.Write()
}

// Plugin manipulation

func (p *Persister) SetPersisterOnPlugin(prefix string, plugin interface{}) {
	value := reflect.Indirect(reflect.ValueOf(plugin))
	wrapper, err := p.NewPluginWrapper(prefix, plugin)
	if err != nil {
		return
	}

	field := value.FieldByName("Persister")
	if !field.IsValid() {
		return
	}

	switch field.Type().String() {
	case "telegraf.StatePersister":
		if field.CanSet() {
			field.Set(reflect.ValueOf(wrapper))
		}
	default:
		p.logger.Warnf("Plugin %q defines a 'Persister' field on its struct of an unexpected type %q. Expected telegraf.StatePersister",
			value.Type().Name(), field.Type().String())
	}
}

// Internal

func (p *Persister) collect() error {
	for id, plugin := range p.register {
		state := plugin.GetState()
		if err := p.SetState(id, state); err != nil {
			return fmt.Errorf("%v: %v", id, err)
		}
	}
	return nil
}
