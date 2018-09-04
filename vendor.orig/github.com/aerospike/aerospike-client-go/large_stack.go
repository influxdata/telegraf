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

// LargeStack encapsulates a stack within a single bin.
// A stack is last in/first out (LIFO) data structure.
type LargeStack struct {
	*baseLargeObject
}

// NewLargeStack initializes a large stack operator.
func NewLargeStack(client *Client, policy *WritePolicy, key *Key, binName string, userModule string) *LargeStack {
	return &LargeStack{
		baseLargeObject: newLargeObject(client, policy, key, binName, userModule, "lstack"),
	}
}

// Push pushes values onto stack.
// If the stack does not exist, create it using specified userModule configuration.
func (lstk *LargeStack) Push(values ...interface{}) error {
	var err error
	if len(values) == 1 {
		_, err = lstk.client.Execute(lstk.policy, lstk.key, lstk.packageName, "push", lstk.binName, NewValue(values[0]), lstk.userModule)
	} else {
		_, err = lstk.client.Execute(lstk.policy, lstk.key, lstk.packageName, "push_all", lstk.binName, ToValueArray(values), lstk.userModule)
	}
	return err
}

// Peek select items from top of stack, without removing them
func (lstk *LargeStack) Peek(peekCount int) ([]interface{}, error) {
	res, err := lstk.client.Execute(lstk.policy, lstk.key, lstk.packageName, "peek", lstk.binName, NewIntegerValue(peekCount))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.([]interface{}), nil
}

// Pop selects items from top of stack and then removes them.
func (lstk *LargeStack) Pop(count int) ([]interface{}, error) {
	res, err := lstk.client.Execute(lstk.policy, lstk.key, lstk.packageName, "pop", lstk.binName, NewIntegerValue(count))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.([]interface{}), nil
}

// Scan returns all objects in the stack.
func (lstk *LargeStack) Scan() ([]interface{}, error) {
	return lstk.scan(lstk)
}

// Filter selects items from top of stack.
func (lstk *LargeStack) Filter(peekCount int, filterName string, filterArgs ...interface{}) ([]interface{}, error) {
	res, err := lstk.client.Execute(lstk.policy, lstk.key, lstk.packageName, "filter", lstk.binName, NewIntegerValue(peekCount), lstk.userModule, NewStringValue(filterName), ToValueArray(filterArgs))
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.([]interface{}), nil
}

// Destroy deletes the bin containing the stack.
func (lstk *LargeStack) Destroy() error {
	return lstk.destroy(lstk)
}

// Size returns size of the stack.
func (lstk *LargeStack) Size() (int, error) {
	return lstk.size(lstk)
}

// GetConfig returns map of stack configuration parameters.
func (lstk *LargeStack) GetConfig() (map[interface{}]interface{}, error) {
	return lstk.getConfig(lstk)
}
