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

/////////////////////////////////////////////////////////////
//
// NOTICE:
// 			THIS FEATURE HAS BEEN DEPRECATED ON SERVER.
//			THE API WILL BE REMOVED FROM THE CLIENT IN THE FUTURE.
//
/////////////////////////////////////////////////////////////

package aerospike

// LargeMap encapsulates a map within a single bin.
type LargeMap struct {
	*baseLargeObject
}

// NewLargeMap initializes a large map operator.
func NewLargeMap(client *Client, policy *WritePolicy, key *Key, binName string, userModule string) *LargeMap {
	return &LargeMap{
		baseLargeObject: newLargeObject(client, policy, key, binName, userModule, "lmap"),
	}
}

// Put adds an entry to the map.
// If the map does not exist, create it using specified userModule configuration.
func (lm *LargeMap) Put(name interface{}, value interface{}) error {
	_, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "put", lm.binName, NewValue(name), NewValue(value), lm.userModule)
	return err
}

// PutMap adds map values to the map.
// If the map does not exist, create it using specified userModule configuration.
func (lm *LargeMap) PutMap(theMap map[interface{}]interface{}) error {
	_, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "put_all", lm.binName, NewMapValue(theMap), lm.userModule)
	return err
}

// Exists checks existence of key in the map.
func (lm *LargeMap) Exists(keyValue interface{}) (bool, error) {
	res, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "exists", lm.binName, NewValue(keyValue))

	if err != nil {
		return false, err
	}

	if res == nil {
		return false, nil
	}
	return (res.(int) != 0), err
}

// Get returns value from map corresponding with the provided key.
func (lm *LargeMap) Get(name interface{}) (map[interface{}]interface{}, error) {
	res, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "get", lm.binName, NewValue(name))

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.(map[interface{}]interface{}), err
}

// Remove deletes a value from map given a key.
func (lm *LargeMap) Remove(name interface{}) error {
	_, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "remove", lm.binName, NewValue(name))
	return err
}

// Scan returns all objects in the map.
func (lm *LargeMap) Scan() (map[interface{}]interface{}, error) {
	res, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "scan", lm.binName)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.(map[interface{}]interface{}), err
}

// Filter selects items from the map.
func (lm *LargeMap) Filter(filterName string, filterArgs ...interface{}) (map[interface{}]interface{}, error) {
	res, err := lm.client.Execute(lm.policy, lm.key, lm.packageName, "filter", lm.binName, lm.userModule, NewStringValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.(map[interface{}]interface{}), err
}

// Destroy deletes the bin containing the map.
func (lm *LargeMap) Destroy() error {
	return lm.destroy(lm)
}

// Size returns size of the map.
func (lm *LargeMap) Size() (int, error) {
	return lm.size(lm)
}

// GetConfig returns map of map configuration parameters.
func (lm *LargeMap) GetConfig() (map[interface{}]interface{}, error) {
	return lm.getConfig(lm)
}
