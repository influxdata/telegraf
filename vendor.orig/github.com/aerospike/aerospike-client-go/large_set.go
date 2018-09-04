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

// LargeSet encapsulates a set within a single bin.
type LargeSet struct {
	*baseLargeObject
}

// NewLargeSet initializes a large set operator.
func NewLargeSet(client *Client, policy *WritePolicy, key *Key, binName string, userModule string) *LargeSet {
	return &LargeSet{
		baseLargeObject: newLargeObject(client, policy, key, binName, userModule, "lset"),
	}
}

// Add adds values to the set.
// If the set does not exist, create it using specified userModule configuration.
func (ls *LargeSet) Add(values ...interface{}) error {
	var err error
	if len(values) == 1 {
		_, err = ls.client.Execute(ls.policy, ls.key, ls.packageName, "add", ls.binName, NewValue(values[0]), ls.userModule)
	} else {
		_, err = ls.client.Execute(ls.policy, ls.key, ls.packageName, "add_all", ls.binName, ToValueArray(values), ls.userModule)
	}

	return err
}

// Remove delete value from set.
func (ls *LargeSet) Remove(value interface{}) error {
	_, err := ls.client.Execute(ls.policy, ls.key, ls.packageName, "remove", ls.binName, NewValue(value))
	return err
}

// Get selects a value from set.
func (ls *LargeSet) Get(value interface{}) (interface{}, error) {
	return ls.client.Execute(ls.policy, ls.key, ls.packageName, "get", ls.binName, NewValue(value))
}

// Exists checks existence of value in the set.
func (ls *LargeSet) Exists(value interface{}) (bool, error) {
	ret, err := ls.client.Execute(ls.policy, ls.key, ls.packageName, "exists", ls.binName, NewValue(value))
	if err != nil {
		return false, err
	}
	return (ret == 1), nil
}

// Scan returns all objects in the set.
func (ls *LargeSet) Scan() ([]interface{}, error) {
	return ls.scan(ls)
}

// Filter select values from set and applies specified Lua filter.
func (ls *LargeSet) Filter(filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ls.client.Execute(ls.policy, ls.key, ls.packageName, "filter", ls.binName, ls.userModule, NewStringValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.([]interface{}), err
}

// Destroy deletes the bin containing the set.
func (ls *LargeSet) Destroy() error {
	return ls.destroy(ls)
}

// Size returns size of the set.
func (ls *LargeSet) Size() (int, error) {
	return ls.size(ls)
}

// GetConfig returns map of set configuration parameters.
func (ls *LargeSet) GetConfig() (map[interface{}]interface{}, error) {
	return ls.getConfig(ls)
}
