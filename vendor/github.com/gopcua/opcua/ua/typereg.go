// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import (
	"reflect"
	"sync"

	"github.com/gopcua/opcua/errors"
)

// TypeRegistry provides a registry for Go types.
//
// Each type is registered with a unique identifier
// which cannot be changed for the lifetime of the component.
//
// Types can be registered multiple times under different
// identifiers.
//
// The implementation is safe for concurrent use.
type TypeRegistry struct {
	mu    sync.RWMutex
	types map[string]reflect.Type
	ids   map[reflect.Type]string
}

// NewTypeRegistry returns a new type registry.
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]reflect.Type),
		ids:   make(map[reflect.Type]string),
	}
}

// New returns a new instance of the type with the given id.
//
// If the id is not known the function returns nil.
func (r *TypeRegistry) New(id string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typ, ok := r.types[id]
	if !ok {
		return nil
	}
	return reflect.New(typ.Elem()).Interface()
}

// Lookup returns the id of the type of v or an empty string if
// the type is not registered.
//
// If the type was registered multiple times the first
// registered id for this type is returned.
func (r *TypeRegistry) Lookup(v interface{}) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ids[reflect.TypeOf(v)]
}

// Register adds a new type to the registry.
//
// If the id is already registered the function returns an error.
func (r *TypeRegistry) Register(id string, v interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	typ := reflect.TypeOf(v)

	if r.types[id] != nil {
		return errors.Errorf("%s is already registered", id)
	}
	r.types[id] = typ

	if _, exists := r.ids[typ]; !exists {
		r.ids[typ] = id
	}
	return nil
}
