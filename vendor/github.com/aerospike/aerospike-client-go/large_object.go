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

// LargeObject interface defines methods to work with LDTs.
type LargeObject interface {
	// Destroy the bin containing LDT.
	Destroy() error
	// Size returns the size of the LDT.
	Size() (int, error)
	// GetConfig returns a map containing LDT config values.
	GetConfig() (map[interface{}]interface{}, error)
}

// Create and manage a large object within a single bin. A large object is last in/first out (LIFO).
type baseLargeObject struct {
	client      *Client
	policy      *WritePolicy
	key         *Key
	binName     Value
	userModule  Value
	packageName string
}

// Initialize large large object operator.
//
// client        client
// policy        generic configuration parameters, pass in nil for defaults
// key         unique record identifier
// binName       bin name
// userModule      Lua function name that initializes list configuration parameters, pass nil for default large object
func newLargeObject(client *Client, policy *WritePolicy, key *Key, binName, userModule string, packageName string) *baseLargeObject {
	r := &baseLargeObject{
		client:      client,
		policy:      policy,
		key:         key,
		binName:     NewStringValue(binName),
		packageName: packageName,
	}

	if userModule == "" {
		r.userModule = NewNullValue()
	} else {
		r.userModule = NewStringValue(userModule)
	}

	return r
}

// Delete bin containing the object.
func (lo *baseLargeObject) destroy(ifc LargeObject) error {
	_, err := lo.client.Execute(lo.policy, lo.key, lo.packageName, "destroy", lo.binName)
	return err
}

// Return size of object.
func (lo *baseLargeObject) size(ifc LargeObject) (int, error) {
	ret, err := lo.client.Execute(lo.policy, lo.key, lo.packageName, "size", lo.binName)
	if err != nil {
		return -1, err
	}

	if ret != nil {
		return ret.(int), nil
	}
	return 0, nil
}

// Return map of object configuration parameters.
func (lo *baseLargeObject) getConfig(ifc LargeObject) (map[interface{}]interface{}, error) {
	res, err := lo.client.Execute(lo.policy, lo.key, lo.packageName, "get_config", lo.binName)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.(map[interface{}]interface{}), err
}

// Return list of all objects on the large object.
func (lo *baseLargeObject) scan(ifc LargeObject) ([]interface{}, error) {
	ret, err := lo.client.Execute(lo.policy, lo.key, lo.packageName, "scan", lo.binName)
	if err != nil {
		return nil, err
	}

	if ret == nil {
		return []interface{}{}, nil
	}
	return ret.([]interface{}), nil
}
