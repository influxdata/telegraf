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
	"bytes"

	. "github.com/aerospike/aerospike-client-go/types"
	Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

type batchCommandExists struct {
	baseMultiCommand

	batchNamespace *batchNamespace
	policy         *BasePolicy
	keys           []*Key
	existsArray    []bool
	index          int
}

func newBatchCommandExists(
	node *Node,
	batchNamespace *batchNamespace,
	policy *BasePolicy,
	keys []*Key,
	existsArray []bool,
) *batchCommandExists {
	res := &batchCommandExists{
		baseMultiCommand: *newMultiCommand(node, nil),
		batchNamespace:   batchNamespace,
		policy:           policy,
		keys:             keys,
		existsArray:      existsArray,
	}
	res.oneShot = false
	return res
}

func (cmd *batchCommandExists) getPolicy(ifc command) Policy {
	return cmd.policy
}

func (cmd *batchCommandExists) writeBuffer(ifc command) error {
	return cmd.setBatchExists(cmd.policy, cmd.keys, cmd.batchNamespace)
}

// Parse all results in the batch.  Add records to shared list.
// If the record was not found, the bins will be nil.
func (cmd *batchCommandExists) parseRecordResults(ifc command, receiveSize int) (bool, error) {
	//Parse each message response and add it to the result array
	cmd.dataOffset = 0

	for cmd.dataOffset < receiveSize {
		if err := cmd.readBytes(int(_MSG_REMAINING_HEADER_SIZE)); err != nil {
			return false, err
		}

		resultCode := ResultCode(cmd.dataBuffer[5] & 0xFF)

		// The only valid server return codes are "ok" and "not found".
		// If other return codes are received, then abort the batch.
		if resultCode != 0 && resultCode != KEY_NOT_FOUND_ERROR {
			return false, NewAerospikeError(resultCode)
		}

		info3 := cmd.dataBuffer[3]

		// If cmd is the end marker of the response, do not proceed further
		if (int(info3) & _INFO3_LAST) == _INFO3_LAST {
			return false, nil
		}

		fieldCount := int(Buffer.BytesToUint16(cmd.dataBuffer, 18))
		opCount := int(Buffer.BytesToUint16(cmd.dataBuffer, 20))

		if opCount > 0 {
			return false, NewAerospikeError(PARSE_ERROR, "Received bins that were not requested!")
		}

		key, err := cmd.parseKey(fieldCount)
		if err != nil {
			return false, err
		}

		offset := cmd.batchNamespace.offsets[cmd.index]
		cmd.index++

		if bytes.Equal(key.digest[:], cmd.keys[offset].digest[:]) {
			// only set the results to true; as a result, no synchronization is needed
			if resultCode == 0 {
				cmd.existsArray[offset] = true
			}
		} else {
			return false, NewAerospikeError(PARSE_ERROR, "Unexpected batch key returned: "+string(key.namespace)+","+Buffer.BytesToHexString(key.digest[:]))
		}
	}
	return true, nil
}

func (cmd *batchCommandExists) Execute() error {
	return cmd.execute(cmd)
}
