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
	"strconv"
	"strings"

	. "github.com/aerospike/aerospike-client-go/types"
)

// ExecuteTask is used to poll for long running server execute job completion.
type ExecuteTask struct {
	*baseTask

	taskId uint64
	scan   bool
}

// NewExecuteTask initializes task with fields needed to query server nodes.
func NewExecuteTask(cluster *Cluster, statement *Statement) *ExecuteTask {
	return &ExecuteTask{
		baseTask: newTask(cluster, false),
		taskId:   statement.TaskId,
		scan:     statement.IsScan(),
	}
}

// IsDone queries all nodes for task completion status.
func (etsk *ExecuteTask) IsDone() (bool, error) {
	var module string
	if etsk.scan {
		module = "scan"
	} else {
		module = "query"
	}

	command := "jobs:module=" + module + ";cmd=get-job;trid=" + strconv.FormatUint(etsk.taskId, 10)

	nodes := etsk.cluster.GetNodes()

	for _, node := range nodes {
		responseMap, err := node.RequestInfo(command)
		if err != nil {
			return false, err
		}
		response := responseMap[command]

		if strings.HasPrefix(response, "ERROR:2") {
			// Task not found. This could mean task already completed or
			// task not started yet.  We are going to have to assume that
			// the task already completed...
			continue
		}

		if strings.HasPrefix(response, "ERROR:") {
			// Mark done and quit immediately.
			return false, NewAerospikeError(UDF_BAD_RESPONSE, response)
		}

		find := "status="
		index := strings.Index(response, find)

		if index < 0 {
			return false, nil
		}

		begin := index + len(find)
		response = response[begin:]
		find = ":"
		index = strings.Index(response, find)

		if index < 0 {
			continue
		}

		status := strings.ToLower(response[:index])
		if !strings.HasPrefix(status, "done") {
			return false, nil
		}
	}

	return true, nil
}

// OnComplete returns a channel which will be closed when the task is
// completed.
// If an error is encountered while performing the task, an error
// will be sent on the channel.
func (etsk *ExecuteTask) OnComplete() chan error {
	return etsk.onComplete(etsk)
}
