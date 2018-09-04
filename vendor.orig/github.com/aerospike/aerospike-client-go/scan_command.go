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

import . "github.com/aerospike/aerospike-client-go/types"

type scanCommand struct {
	baseMultiCommand

	policy    *ScanPolicy
	namespace string
	setName   string
	binNames  []string
	taskId    uint64
}

func newScanCommand(
	node *Node,
	policy *ScanPolicy,
	namespace string,
	setName string,
	binNames []string,
	recordset *Recordset,
	taskId uint64,
) *scanCommand {
	cmd := &scanCommand{
		baseMultiCommand: *newMultiCommand(node, recordset),
		policy:           policy,
		namespace:        namespace,
		setName:          setName,
		binNames:         binNames,
		taskId:           taskId,
	}

	cmd.terminationErrorType = SCAN_TERMINATED

	return cmd
}

func (cmd *scanCommand) getPolicy(ifc command) Policy {
	return cmd.policy
}

func (cmd *scanCommand) writeBuffer(ifc command) error {
	return cmd.setScan(cmd.policy, &cmd.namespace, &cmd.setName, cmd.binNames, cmd.taskId)
}

func (cmd *scanCommand) parseResult(ifc command, conn *Connection) error {
	return cmd.baseMultiCommand.parseResult(cmd, conn)
}

func (cmd *scanCommand) Execute() error {
	defer cmd.recordset.signalEnd()
	err := cmd.execute(cmd)
	if err != nil {
		cmd.recordset.sendError(err)
	}
	return err
}
