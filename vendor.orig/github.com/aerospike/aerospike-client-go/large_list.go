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

// LargeList encapsulates a list within a single bin.
type LargeList struct {
	*baseLargeObject
}

// NewLargeList initializes a large list operator.
func NewLargeList(client *Client, policy *WritePolicy, key *Key, binName string, userModule string) *LargeList {
	return &LargeList{
		baseLargeObject: newLargeObject(client, policy, key, binName, userModule, "llist"),
	}
}

// Add adds values to the list.
// If the list does not exist, create it
func (ll *LargeList) Add(values ...interface{}) (err error) {
	_, err = ll.client.Execute(ll.policy, ll.key, ll.packageName, "add", ll.binName, ToValueArray(values))
	return err
}

// Update updates/adds each value in values list depending if key exists or not.
func (ll *LargeList) Update(values ...interface{}) (err error) {
	_, err = ll.client.Execute(ll.policy, ll.key, ll.packageName, "update", ll.binName, ToValueArray(values))
	return err
}

// Remove deletes value from list.
func (ll *LargeList) Remove(values ...interface{}) (err error) {
	_, err = ll.client.Execute(ll.policy, ll.key, ll.packageName, "remove", ll.binName, ToValueArray(values))
	return err
}

// Find selects values from list.
func (ll *LargeList) Find(value interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find", ll.binName, NewValue(value))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

//  Do key/values exist?  Return list of results in one batch call.
func (ll *LargeList) Exist(values ...interface{}) ([]bool, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "exists", ll.binName, NewValue(values))
	if err != nil {
		return nil, err
	}

	var ret []bool
	if res == nil {
		return make([]bool, len(values)), nil
	} else {
		ret = make([]bool, len(values))
		resTyped := res.([]interface{})
		for i := range resTyped {
			ret[i] = resTyped[i].(int) != 0
		}
	}

	return ret, err
}

// FindThenFilter selects values from list and applies specified Lua filter.
func (ll *LargeList) FindThenFilter(value interface{}, filterModule, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find", ll.binName, NewValue(value), NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// FindFirst selects values from the beginning of list up to a maximum count.
func (ll *LargeList) FindFirst(count int) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_first", ll.binName, NewValue(count))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// FFilterThenindFirst selects values from the beginning of list up to a maximum count after applying lua filter.
func (ll *LargeList) FFilterThenindFirst(count int, filterModule, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_first", ll.binName, NewValue(count), NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// FindLast selects values from the end of list up to a maximum count.
func (ll *LargeList) FindLast(count int) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_last", ll.binName, NewValue(count))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// FilterThenFindLast selects values from the end of list up to a maximum count after applying lua filter.
func (ll *LargeList) FilterThenFindLast(count int, filterModule, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_last", ll.binName, NewValue(count), NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// FindFrom selects values from the begin key up to a maximum count.
func (ll *LargeList) FindFrom(begin interface{}, count int) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_from", ll.binName, NewValue(begin), NewValue(count))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// FilterThenFindFrom selects values from the begin key up to a maximum count after applying lua filter.
func (ll *LargeList) FilterThenFindFrom(begin interface{}, count int, filterModule, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_from", ll.binName, NewValue(begin), NewValue(count), NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// Range selects a range of values from the large list.
func (ll *LargeList) Range(begin, end interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_range", ll.binName, NewValue(begin), NewValue(end))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// RangeN selects a range of values up to a maximum count from the large list.
func (ll *LargeList) RangeN(begin, end interface{}, count int) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "find_range", ll.binName, NewValue(begin), NewValue(end), NewValue(count))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// RangeThenFilter selects a range of values from the large list then apply filter.
func (ll *LargeList) RangeThenFilter(begin, end interface{}, filterModule string, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "range", ll.binName, NewValue(begin), NewValue(end), NewValue(0), NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// RangeNThenFilter selects a range of values up to a maximum count from the large list then apply filter.
func (ll *LargeList) RangeNThenFilter(begin, end interface{}, count int, filterModule string, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "range", ll.binName, NewValue(begin), NewValue(end), NewValue(count), NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// Scan returns all objects in the list.
func (ll *LargeList) Scan() ([]interface{}, error) {
	return ll.scan(ll)
}

// Filter selects values from list and apply specified Lua filter.
func (ll *LargeList) Filter(filterModule, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "scan", ll.binName, NewValue(filterModule), NewValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return []interface{}{}, nil
	}
	return res.([]interface{}), err
}

// Destroy deletes the bin containing the list.
func (ll *LargeList) Destroy() error {
	return ll.destroy(ll)
}

// Size returns size of list.
func (ll *LargeList) Size() (int, error) {
	return ll.size(ll)
}

// SetPageSize sets the LDT page size.
func (ll *LargeList) SetPageSize(pageSize int) error {
	_, err := ll.client.Execute(ll.policy, ll.key, ll.packageName, "setPageSize", ll.binName, NewValue(pageSize))
	return err
}

// GetConfig returns map of list configuration parameters.
func (ll *LargeList) GetConfig() (map[interface{}]interface{}, error) {
	return ll.getConfig(ll)
}
