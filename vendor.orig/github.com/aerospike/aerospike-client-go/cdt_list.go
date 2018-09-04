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
	. "github.com/aerospike/aerospike-client-go/types"
)

// List bin operations. Create list operations used by client.Operate command.
// List operations support negative indexing.  If the index is negative, the
// resolved index starts backwards from end of list.
//
// Index/Range examples:
//
//    Index 0: First item in list.
//    Index 4: Fifth item in list.
//    Index -1: Last item in list.
//    Index -3: Third to last item in list.
//    Index 1 Count 2: Second and third items in list.
//    Index -3 Count 3: Last three items in list.
//    Index -5 Count 4: Range between fifth to last item to second to last item inclusive.
//
// If an index is out of bounds, a parameter error will be returned. If a range is partially
// out of bounds, the valid part of the range will be returned.

const (
	_CDT_LIST_APPEND       = 1
	_CDT_LIST_APPEND_ITEMS = 2
	_CDT_LIST_INSERT       = 3
	_CDT_LIST_INSERT_ITEMS = 4
	_CDT_LIST_POP          = 5
	_CDT_LIST_POP_RANGE    = 6
	_CDT_LIST_REMOVE       = 7
	_CDT_LIST_REMOVE_RANGE = 8
	_CDT_LIST_SET          = 9
	_CDT_LIST_TRIM         = 10
	_CDT_LIST_CLEAR        = 11
	_CDT_LIST_SIZE         = 16
	_CDT_LIST_GET          = 17
	_CDT_LIST_GET_RANGE    = 18
)

func packCDTParamsAsArray(packer BufferEx, opType int16, params ...Value) (int, error) {
	size := 0
	n, err := __PackShortRaw(packer, opType)
	if err != nil {
		return n, err
	}
	size += n

	if len(params) > 0 {
		if n, err = __PackArrayBegin(packer, len(params)); err != nil {
			return size + n, err
		}
		size += n

		for i := range params {
			if n, err = params[i].pack(packer); err != nil {
				return size + n, err
			}
			size += n
		}
	}
	return size, nil
}

func packCDTIfcParamsAsArray(packer BufferEx, opType int16, params ListValue) (int, error) {
	return packCDTIfcVarParamsAsArray(packer, opType, []interface{}(params)...)
}

func packCDTIfcVarParamsAsArray(packer BufferEx, opType int16, params ...interface{}) (int, error) {
	size := 0
	n, err := __PackShortRaw(packer, opType)
	if err != nil {
		return n, err
	}
	size += n

	if len(params) > 0 {
		if n, err = __PackArrayBegin(packer, len(params)); err != nil {
			return size + n, err
		}
		size += n

		for i := range params {
			if n, err = __PackObject(packer, params[i], false); err != nil {
				return size + n, err
			}
			size += n
		}
	}
	return size, nil
}

func listAppendOpEncoder(op *Operation, packer BufferEx) (int, error) {
	params := op.binValue.(ListValue)
	if len(params) == 1 {
		return packCDTIfcVarParamsAsArray(packer, _CDT_LIST_APPEND, params[0])
	} else if len(params) > 1 {
		return packCDTParamsAsArray(packer, _CDT_LIST_APPEND_ITEMS, params)
	}

	return -1, NewAerospikeError(PARAMETER_ERROR, "At least one value must be provided for ListAppendOp")
}

// ListAppendOp creates a list append operation.
// Server appends values to end of list bin.
// Server returns list size on bin name.
// It will panic is no values have been passed.
func ListAppendOp(binName string, values ...interface{}) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: ListValue(values), encoder: listAppendOpEncoder}
}

func listInsertOpEncoder(op *Operation, packer BufferEx) (int, error) {
	args := op.binValue.(ValueArray)
	params := args[1].(ListValue)
	if len(params) == 1 {
		return packCDTIfcVarParamsAsArray(packer, _CDT_LIST_INSERT, args[0], params[0])
	} else if len(params) > 1 {
		return packCDTParamsAsArray(packer, _CDT_LIST_INSERT_ITEMS, args[0], params)
	}

	return -1, NewAerospikeError(PARAMETER_ERROR, "At least one value must be provided for ListInsertOp")
}

// ListInsertOp creates a list insert operation.
// Server inserts value to specified index of list bin.
// Server returns list size on bin name.
// It will panic is no values have been passed.
func ListInsertOp(binName string, index int, values ...interface{}) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: ValueArray([]Value{IntegerValue(index), ListValue(values)}), encoder: listInsertOpEncoder}
}

func listPopOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_POP, op.binValue)
}

// ListPopOp creates list pop operation.
// Server returns item at specified index and removes item from list bin.
func ListPopOp(binName string, index int) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: IntegerValue(index), encoder: listPopOpEncoder}
}

func listPopRangeOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_POP_RANGE, op.binValue.(ValueArray)...)
}

// ListPopRangeOp creates a list pop range operation.
// Server returns items starting at specified index and removes items from list bin.
func ListPopRangeOp(binName string, index int, count int) *Operation {
	if count == 1 {
		return ListPopOp(binName, index)
	}

	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: ValueArray([]Value{IntegerValue(index), IntegerValue(count)}), encoder: listPopRangeOpEncoder}
}

func listPopRangeFromOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_POP_RANGE, op.binValue)
}

// ListPopRangeFromOp creates a list pop range operation.
// Server returns items starting at specified index to the end of list and removes items from list bin.
func ListPopRangeFromOp(binName string, index int) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: IntegerValue(index), encoder: listPopRangeFromOpEncoder}
}

func listRemoveOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_REMOVE, op.binValue)
}

// ListRemoveOp creates a list remove operation.
// Server removes item at specified index from list bin.
// Server returns number of items removed.
func ListRemoveOp(binName string, index int) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: IntegerValue(index), encoder: listRemoveOpEncoder}
}

func listRemoveRangeOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_REMOVE_RANGE, op.binValue.(ValueArray)...)
}

// ListRemoveRangeOp creates a list remove range operation.
// Server removes "count" items starting at specified index from list bin.
// Server returns number of items removed.
func ListRemoveRangeOp(binName string, index int, count int) *Operation {
	if count == 1 {
		return ListRemoveOp(binName, index)
	}

	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: ValueArray([]Value{IntegerValue(index), IntegerValue(count)}), encoder: listRemoveRangeOpEncoder}
}

func listRemoveRangeFromOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_REMOVE_RANGE, op.binValue)
}

// ListRemoveRangeFromOp creates a list remove range operation.
// Server removes all items starting at specified index to the end of list.
// Server returns number of items removed.
func ListRemoveRangeFromOp(binName string, index int) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: IntegerValue(index), encoder: listRemoveRangeFromOpEncoder}
}

func listSetOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTIfcParamsAsArray(packer, _CDT_LIST_SET, op.binValue.(ListValue))
}

// ListSetOp creates a list set operation.
// Server sets item value at specified index in list bin.
// Server does not return a result by default.
func ListSetOp(binName string, index int, value interface{}) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: ListValue([]interface{}{IntegerValue(index), value}), encoder: listSetOpEncoder}
}

func listTrimOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_TRIM, op.binValue.(ValueArray)...)
}

// ListTrimOp creates a list trim operation.
// Server removes "count" items in list bin that do not fall into range specified
// by index and count range.  If the range is out of bounds, then all items will be removed.
// Server returns number of elemts that were removed.
func ListTrimOp(binName string, index int, count int) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: ValueArray([]Value{IntegerValue(index), IntegerValue(count)}), encoder: listTrimOpEncoder}
}

func listClearOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_CLEAR)
}

// ListClearOp creates a list clear operation.
// Server removes all items in list bin.
// Server does not return a result by default.
func ListClearOp(binName string) *Operation {
	return &Operation{opType: CDT_MODIFY, binName: binName, binValue: NewNullValue(), encoder: listClearOpEncoder}
}

func listSizeOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_SIZE)
}

// ListSizeOp creates a list size operation.
// Server returns size of list on bin name.
func ListSizeOp(binName string) *Operation {
	return &Operation{opType: CDT_READ, binName: binName, binValue: NewNullValue(), encoder: listSizeOpEncoder}
}

func listGetOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_GET, op.binValue)
}

// ListGetOp creates a list get operation.
// Server returns item at specified index in list bin.
func ListGetOp(binName string, index int) *Operation {
	return &Operation{opType: CDT_READ, binName: binName, binValue: IntegerValue(index), encoder: listGetOpEncoder}
}

func listGetRangeOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_GET_RANGE, op.binValue.(ValueArray)...)
}

// ListGetRangeOp creates a list get range operation.
// Server returns "count" items starting at specified index in list bin.
func ListGetRangeOp(binName string, index int, count int) *Operation {
	return &Operation{opType: CDT_READ, binName: binName, binValue: ValueArray([]Value{IntegerValue(index), IntegerValue(count)}), encoder: listGetRangeOpEncoder}
}

func listGetRangeFromOpEncoder(op *Operation, packer BufferEx) (int, error) {
	return packCDTParamsAsArray(packer, _CDT_LIST_GET_RANGE, op.binValue)
}

// ListGetRangeFromOp creates a list get range operation.
// Server returns items starting at specified index to the end of list.
func ListGetRangeFromOp(binName string, index int) *Operation {
	return &Operation{opType: CDT_READ, binName: binName, binValue: IntegerValue(index), encoder: listGetRangeFromOpEncoder}
}
