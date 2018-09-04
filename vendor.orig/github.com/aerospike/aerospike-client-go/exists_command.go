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

// guarantee existsCommand implements command interface
var _ command = &existsCommand{}

type existsCommand struct {
	singleCommand

	policy *BasePolicy
	exists bool
}

func newExistsCommand(cluster *Cluster, policy *BasePolicy, key *Key) *existsCommand {
	return &existsCommand{
		singleCommand: newSingleCommand(cluster, key),
		policy:        policy,
	}
}

func (cmd *existsCommand) getPolicy(ifc command) Policy {
	return cmd.policy
}

func (cmd *existsCommand) writeBuffer(ifc command) error {
	return cmd.setExists(cmd.policy, cmd.key)
}

func (cmd *existsCommand) getNode(ifc command) (*Node, error) {
	return cmd.cluster.getReadNode(&cmd.partition, cmd.policy.ReplicaPolicy)
}

func (cmd *existsCommand) parseResult(ifc command, conn *Connection) error {
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
	cmd.exists = resultCode == 0
	return cmd.emptySocket(conn)
}

func (cmd *existsCommand) Exists() bool {
	return cmd.exists
}

func (cmd *existsCommand) Execute() error {
	return cmd.execute(cmd)
}
