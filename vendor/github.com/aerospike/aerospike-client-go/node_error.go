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

// NodeError is a type to encapsulate the node that the error occurred in.
type NodeError struct {
	error

	node *Node
}

func newNodeError(node *Node, err error) *NodeError {
	return &NodeError{
		error: err,
		node:  node,
	}
}

func newAerospikeNodeError(node *Node, code ResultCode, messages ...string) *NodeError {
	return &NodeError{
		error: NewAerospikeError(code, messages...),
		node:  node,
	}
}

// Node returns the node where the error occurred.
func (ne *NodeError) Node() *Node { return ne.node }

// Err returns the error
func (ne *NodeError) Err() error { return ne.error }
