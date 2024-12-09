package persister

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/influxdata/telegraf"
)

type Persister struct {
	Filename string

	register map[string]telegraf.StatefulPlugin
}

func (p *Persister) Init() error {
	p.register = make(map[string]telegraf.StatefulPlugin)

	return nil
}

func (p *Persister) Register(id string, plugin telegraf.StatefulPlugin) error {
	if _, found := p.register[id]; found {
		return fmt.Errorf("plugin with ID %q already registered", id)
	}
	p.register[id] = plugin

	return nil
}

func (p *Persister) Load() error {
	// Read the states from disk
	in, err := os.ReadFile(p.Filename)
	if err != nil {
		return fmt.Errorf("reading states file failed: %w", err)
	}

	// Unmarshal the id to serialized states map
	var states map[string][]byte
	if err := json.Unmarshal(in, &states); err != nil {
		return fmt.Errorf("unmarshalling states failed: %w", err)
	}

	// Get the initialized state as blueprint for unmarshalling
	for id, serialized := range states {
		// Check if we have a plugin with that ID
		plugin, found := p.register[id]
		if !found {
			continue
		}

		// Create a new empty state of the "state"-type. As we need a pointer
		// of the state, we cannot dereference it here due to the unknown
		// nature of the state-type.
		nstate := reflect.New(reflect.TypeOf(plugin.GetState())).Interface()
		if err := json.Unmarshal(serialized, &nstate); err != nil {
			return fmt.Errorf("unmarshalling state for %q failed: %w", id, err)
		}
		state := reflect.ValueOf(nstate).Elem().Interface()

		// Set the state in the plugin
		if err := plugin.SetState(state); err != nil {
			return fmt.Errorf("setting state of %q failed: %w", id, err)
		}
	}

	return nil
}

func (p *Persister) Store() error {
	states := make(map[string][]byte)

	// Collect the states and serialize the individual data chunks
	// to later serialize all items in the id / serialized-states map
	for id, plugin := range p.register {
		state, err := json.Marshal(plugin.GetState())
		if err != nil {
			return fmt.Errorf("marshalling state for id %q failed: %w", id, err)
		}
		states[id] = state
	}

	// Serialize the states
	serialized, err := json.Marshal(states)
	if err != nil {
		return fmt.Errorf("marshalling states failed: %w", err)
	}

	// Write the states to disk
	f, err := os.Create(p.Filename)
	if err != nil {
		return fmt.Errorf("creating states file %q failed: %w", p.Filename, err)
	}
	defer f.Close()

	if _, err := f.Write(serialized); err != nil {
		return fmt.Errorf("writing states failed: %w", err)
	}

	return nil
}
