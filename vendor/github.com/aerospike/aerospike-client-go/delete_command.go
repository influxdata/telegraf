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

// guarantee deleteCommand implements command interface
var _ command = &deleteCommand{}

type deleteCommand struct {
	singleCommand

	policy  *WritePolicy
	existed bool
}

func newDeleteCommand(cluster *Cluster, policy *WritePolicy, key *Key) *deleteCommand {
	newDeleteCmd := &deleteCommand{
		singleCommand: newSingleCommand(cluster, key),
		policy:        policy,
	}

	return newDeleteCmd
}

func (cmd *deleteCommand) getPolicy(ifc command) Policy {
	return cmd.policy
}

func (cmd *deleteCommand) writeBuffer(ifc command) error {
	return cmd.setDelete(cmd.policy, cmd.key)
}

func (cmd *deleteCommand) getNode(ifc command) (*Node, error) {
	return cmd.cluster.getMasterNode(&cmd.partition)
}

func (cmd *deleteCommand) parseResult(ifc command, conn *Connection) error {
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

	if resultCode != 0 && ResultCode(resultCode) != KEY_NOT_FOUND_ERROR {
		return NewAerospikeError(ResultCode(resultCode))
	}
	cmd.existed = resultCode == 0

	return cmd.emptySocket(conn)
}

func (cmd *deleteCommand) Existed() bool {
	return cmd.existed
}

func (cmd *deleteCommand) Execute() error {
	return cmd.execute(cmd)
}
