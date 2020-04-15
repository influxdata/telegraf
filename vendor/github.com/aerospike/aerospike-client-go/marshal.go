// +build !as_performance

// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	. "github.com/aerospike/aerospike-client-go/types"
)

var aerospikeTag = "as"

const (
	aerospikeMetaTag = "asm"
	keyTag           = "key"
)

// SetAerospikeTag sets the bin tag to the specified tag.
// This will be useful for when a user wants to use the same tag name for two different concerns.
// For example, one will be able to use the same tag name for both json and aerospike bin name.
func SetAerospikeTag(tag string) {
	aerospikeTag = tag
}

func valueToInterface(f reflect.Value, clusterSupportsFloat bool) interface{} {
	// get to the core value
	for f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil
		}
		f = reflect.Indirect(f)
	}

	switch f.Kind() {
	case reflect.Uint64:
		return int64(f.Uint())
	case reflect.Float64, reflect.Float32:
		// support floats through integer encoding if
		// server doesn't support floats
		if clusterSupportsFloat {
			return f.Float()
		}
		return int(math.Float64bits(f.Float()))

	case reflect.Struct:
		if f.Type().PkgPath() == "time" && f.Type().Name() == "Time" {
			return f.Interface().(time.Time).UTC().UnixNano()
		}
		return structToMap(f, clusterSupportsFloat)
	case reflect.Bool:
		if f.Bool() {
			return int64(1)
		}
		return int64(0)
	case reflect.Map:
		if f.IsNil() {
			return nil
		}

		newMap := make(map[interface{}]interface{}, f.Len())
		for _, mk := range f.MapKeys() {
			newMap[valueToInterface(mk, clusterSupportsFloat)] = valueToInterface(f.MapIndex(mk), clusterSupportsFloat)
		}

		return newMap
	case reflect.Slice, reflect.Array:
		if f.Kind() == reflect.Slice && f.IsNil() {
			return nil
		}
		if f.Kind() == reflect.Slice && reflect.TypeOf(f.Interface()).Elem().Kind() == reflect.Uint8 {
			// handle blobs
			return f.Interface().([]byte)
		}
		// convert to primitives recursively
		newSlice := make([]interface{}, f.Len(), f.Cap())
		for i := 0; i < len(newSlice); i++ {
			newSlice[i] = valueToInterface(f.Index(i), clusterSupportsFloat)
		}
		return newSlice
	case reflect.Interface:
		if f.IsNil() {
			return nil
		}
		return f.Interface()
	default:
		return f.Interface()
	}
}

func fieldIsMetadata(f reflect.StructField) bool {
	meta := f.Tag.Get(aerospikeMetaTag)
	return strings.Trim(meta, " ") != ""
}

func fieldAlias(f reflect.StructField) string {
	alias := f.Tag.Get(aerospikeTag)
	if alias != "" {
		alias = strings.Trim(alias, " ")

		// if tag is -, the field should not be persisted
		if alias == "-" {
			return ""
		}
		return alias
	}
	return f.Name
}

func structToMap(s reflect.Value, clusterSupportsFloat bool) map[string]interface{} {
	if !s.IsValid() {
		return nil
	}

	typeOfT := s.Type()
	numFields := s.NumField()

	var binMap map[string]interface{}
	for i := 0; i < numFields; i++ {
		// skip unexported fields
		if typeOfT.Field(i).PkgPath != "" {
			continue
		}

		if fieldIsMetadata(typeOfT.Field(i)) {
			continue
		}

		// skip transiet fields tagged `-`
		alias := fieldAlias(typeOfT.Field(i))
		if alias == "" {
			continue
		}

		binValue := valueToInterface(s.Field(i), clusterSupportsFloat)

		if binMap == nil {
			binMap = make(map[string]interface{}, numFields)
		}

		binMap[alias] = binValue
	}

	return binMap
}

func marshal(v interface{}, clusterSupportsFloat bool) []*Bin {
	s := indirect(reflect.ValueOf(v))
	numFields := s.NumField()
	bins := binPool.Get(numFields).([]*Bin)

	binCount := 0
	n := structToMap(s, clusterSupportsFloat)
	for k, v := range n {
		bins[binCount].Name = k

		bins[binCount].Value = NewValue(v)
		binCount++
	}

	return bins[:binCount]
}

type syncMap struct {
	objectMappings map[reflect.Type]map[string]string
	objectFields   map[reflect.Type][]string
	objectTTLs     map[reflect.Type][]string
	objectGen      map[reflect.Type][]string
	mutex          sync.RWMutex
}

func (sm *syncMap) setMapping(objType reflect.Type, mapping map[string]string, fields, ttl, gen []string) {
	sm.mutex.Lock()
	sm.objectMappings[objType] = mapping
	sm.objectFields[objType] = fields
	sm.objectTTLs[objType] = ttl
	sm.objectGen[objType] = gen
	sm.mutex.Unlock()
}

func indirect(obj reflect.Value) reflect.Value {
	for obj.Kind() == reflect.Ptr {
		if obj.IsNil() {
			return obj
		}
		obj = obj.Elem()
	}
	return obj
}

func indirectT(objType reflect.Type) reflect.Type {
	for objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	return objType
}

func (sm *syncMap) mappingExists(objType reflect.Type) (map[string]string, bool) {
	sm.mutex.RLock()
	mapping, exists := sm.objectMappings[objType]
	sm.mutex.RUnlock()
	return mapping, exists
}

func (sm *syncMap) getMapping(objType reflect.Type) map[string]string {
	objType = indirectT(objType)
	mapping, exists := sm.mappingExists(objType)
	if !exists {
		cacheObjectTags(objType)
		mapping, _ = sm.mappingExists(objType)
	}

	return mapping
}

func (sm *syncMap) getMetaMappings(objType reflect.Type) (ttl, gen []string) {
	objType = indirectT(objType)
	if _, exists := sm.mappingExists(objType); !exists {
		cacheObjectTags(objType)
	}

	sm.mutex.RLock()
	ttl = sm.objectTTLs[objType]
	gen = sm.objectGen[objType]
	sm.mutex.RUnlock()
	return ttl, gen
}

func (sm *syncMap) fieldsExists(objType reflect.Type) ([]string, bool) {
	sm.mutex.RLock()
	mapping, exists := sm.objectFields[objType]
	sm.mutex.RUnlock()
	return mapping, exists
}

func (sm *syncMap) getFields(objType reflect.Type) []string {
	objType = indirectT(objType)
	fields, exists := sm.fieldsExists(objType)
	if !exists {
		cacheObjectTags(objType)
		fields, _ = sm.fieldsExists(objType)
	}

	return fields
}

var objectMappings = &syncMap{
	objectMappings: map[reflect.Type]map[string]string{},
	objectFields:   map[reflect.Type][]string{},
	objectTTLs:     map[reflect.Type][]string{},
	objectGen:      map[reflect.Type][]string{},
}

func cacheObjectTags(objType reflect.Type) {
	mapping := map[string]string{}
	fields := []string{}
	ttl := []string{}
	gen := []string{}

	numFields := objType.NumField()
	for i := 0; i < numFields; i++ {
		f := objType.Field(i)
		// skip unexported fields
		if f.PkgPath != "" {
			continue
		}

		tag := strings.Trim(f.Tag.Get(aerospikeTag), " ")
		tagM := strings.Trim(f.Tag.Get(aerospikeMetaTag), " ")

		if tag != "" && tagM != "" {
			panic(fmt.Sprintf("Cannot accept both data and metadata tags on the same attribute on struct: %s.%s", objType.Name(), f.Name))
		}

		if tag != "-" && tagM == "" {
			if tag != "" {
				mapping[tag] = f.Name
				fields = append(fields, tag)
			} else {
				fields = append(fields, f.Name)
			}
		}

		if tagM == "ttl" {
			ttl = append(ttl, f.Name)
		} else if tagM == "gen" {
			gen = append(gen, f.Name)
		} else if tagM != "" {
			panic(fmt.Sprintf("Invalid metadata tag `%s` on struct attribute: %s.%s", tagM, objType.Name(), f.Name))
		}
	}

	objectMappings.setMapping(objType, mapping, fields, ttl, gen)
}

func binMapToBins(bins []*Bin, binMap BinMap) []*Bin {
	i := 0
	for k, v := range binMap {
		bins[i].Name = k
		bins[i].Value = NewValue(v)
		i++
	}

	return bins
}

// pool Bins so that we won't have to allocate them every time
var binPool = NewPool(512)

func init() {
	binPool.New = func(params ...interface{}) interface{} {
		size := params[0].(int)
		bins := make([]*Bin, size, size)
		for i := range bins {
			bins[i] = &Bin{}
		}
		return bins
	}

	binPool.IsUsable = func(obj interface{}, params ...interface{}) bool {
		return len(obj.([]*Bin)) >= params[0].(int)
	}
}
