package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

type JSONStore struct {
	Filename string
	Indent   int

	states map[string]interface{}
}

func (s *JSONStore) Init() error {
	s.states = make(map[string]interface{})
	return nil
}

func (s *JSONStore) SetState(id string, state interface{}) error {
	s.states[id] = state
	return nil
}

func (s *JSONStore) GetState(id string) (interface{}, bool) {
	state, found := s.states[id]
	return state, found
}

func (s *JSONStore) GetStates() map[string]interface{} {
	return s.states
}

func (s *JSONStore) Read() error {
	// Read the states from disk
	in, err := ioutil.ReadFile(s.Filename)
	if err != nil {
		return err
	}

	// Unmarshal the id-states map
	states := make(map[string]string)
	if err := json.Unmarshal(in, &states); err != nil {
		return fmt.Errorf("unmarshalling id-states mapping failed: %v", err)
	}

	// Get the initialized state as blueprint for unmarshalling
	for id, serialized := range states {
		state, found := s.states[id]
		if !found {
			return fmt.Errorf("state not found for %v", id)
		}
		// Create a new empty state of the "state"-type. As we need a pointer
		// of the state, we cannot dereference it here due to the unknown
		// nature of the state-type.
		newState := reflect.New(reflect.TypeOf(state)).Interface()
		if err := json.Unmarshal([]byte(serialized), newState); err != nil {
			return fmt.Errorf("unmarshalling state of %v failed: %v", id, err)
		}
		// Dereference the pointer of the new state and set the plugin's state
		s.states[id] = reflect.ValueOf(newState).Elem().Interface()
	}

	return nil
}

func (s *JSONStore) Write() error {
	states := make(map[string]string)

	// Serialize the single states to JSON
	for id, state := range s.states {
		serialized, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("marshalling state %v failed: %v", id, err)
		}
		states[id] = string(serialized)
	}

	// Marshall the serialized id-state map
	out, err := json.Marshal(states)
	if err != nil {
		return fmt.Errorf("marshalling id-states mapping failed: %v", err)
	}

	// Indent if requested
	if s.Indent > 0 {
		var buf bytes.Buffer
		if err := json.Indent(&buf, out, "", strings.Repeat(" ", s.Indent)); err != nil {
			return fmt.Errorf("indentation failed: %v", err)
		}
		out = buf.Bytes()
	}

	// Write the states to disk
	f, err := os.Create(s.Filename)
	if err != nil {
		return fmt.Errorf("creating states file failed: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(out); err != nil {
		return fmt.Errorf("writing states file failed: %v", err)
	}

	return nil
}
