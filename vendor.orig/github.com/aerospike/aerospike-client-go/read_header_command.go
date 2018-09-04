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
	Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

type readHeaderCommand struct {
	singleCommand

	policy *BasePolicy
	record *Record
}

func newReadHeaderCommand(cluster *Cluster, policy *BasePolicy, key *Key) *readHeaderCommand {
	newReadHeaderCmd := &readHeaderCommand{
		singleCommand: newSingleCommand(cluster, key),
		policy:        policy,
	}

	return newReadHeaderCmd
}

func (cmd *readHeaderCommand) getPolicy(ifc command) Policy {
	return cmd.policy
}

func (cmd *readHeaderCommand) writeBuffer(ifc command) error {
	return cmd.setReadHeader(cmd.policy, cmd.key)
}

func (cmd *readHeaderCommand) getNode(ifc command) (*Node, error) {
	return cmd.cluster.getReadNode(&cmd.partition, cmd.policy.ReplicaPolicy)
}

func (cmd *readHeaderCommand) parseResult(ifc command, conn *Connection) error {
	// Read header.
	if _, err := conn.Read(cmd.dataBuffer, int(_MSG_TOTAL_HEADER_SIZE)); err != nil {
		return err
	}

	header := Buffer.BytesToInt64(cmd.dataBuffer, 0)

	// Validate header to make sure we are at the beginning of a message
	if err := cmd.validateHeader(header); err != nil {
		return err
	}

	resultCode := cmd.dataBuffer[13] & 0xFF

	if resultCode == 0 {
		generation := Buffer.BytesToUint32(cmd.dataBuffer, 14)
		expiration := TTL(Buffer.BytesToUint32(cmd.dataBuffer, 18))
		cmd.record = newRecord(cmd.node, cmd.key, nil, generation, expiration)
	} else {
		if ResultCode(resultCode) == KEY_NOT_FOUND_ERROR {
			cmd.record = nil
		} else {
			return NewAerospikeError(ResultCode(resultCode))
		}
	}
	if err := cmd.emptySocket(conn); err != nil {
		return err
	}
	return nil
}

func (cmd *readHeaderCommand) GetRecord() *Record {
	return cmd.record
}

func (cmd *readHeaderCommand) Execute() error {
	return cmd.execute(cmd)
}
