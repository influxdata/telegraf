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

type batchNode struct {
	Node            *Node
	BatchNamespaces []*batchNamespace
	KeyCapacity     int
}

func newBatchNodeList(cluster *Cluster, policy *BasePolicy, keys []*Key) ([]*batchNode, error) {
	nodes := cluster.GetNodes()

	if len(nodes) == 0 {
		return nil, NewAerospikeError(SERVER_NOT_AVAILABLE, "command failed because cluster is empty.")
	}

	nodeCount := len(nodes)
	keysPerNode := len(keys)/nodeCount + 10

	// Split keys by server node.
	batchNodes := make([]*batchNode, 0, nodeCount)

	for i, key := range keys {
		partition := NewPartitionByKey(key)

		// error not required
		node, err := cluster.getReadNode(partition, policy.ReplicaPolicy)
		if err != nil {
			return nil, err
		}
		batchNode := findBatchNode(batchNodes, node)

		if batchNode == nil {
			batchNodes = append(batchNodes, newBatchNode(node, keysPerNode, key.Namespace(), i))
		} else {
			batchNode.AddKey(key.Namespace(), i)
		}
	}
	return batchNodes, nil
}

func newBatchNode(node *Node, keyCapacity int, namespace string, offset int) *batchNode {
	return &batchNode{
		Node:            node,
		KeyCapacity:     keyCapacity,
		BatchNamespaces: []*batchNamespace{newBatchNamespace(&namespace, keyCapacity, offset)},
	}
}

func (bn *batchNode) AddKey(namespace string, offset int) {
	batchNamespace := bn.findNamespace(&namespace)

	if batchNamespace == nil {
		bn.BatchNamespaces = append(bn.BatchNamespaces, newBatchNamespace(&namespace, bn.KeyCapacity, offset))
	} else {
		batchNamespace.add(offset)
	}
}

func (bn *batchNode) findNamespace(ns *string) *batchNamespace {
	for _, batchNamespace := range bn.BatchNamespaces {
		// Note: use both pointer equality and equals.
		if batchNamespace.namespace == ns || *batchNamespace.namespace == *ns {
			return batchNamespace
		}
	}
	return nil
}

func findBatchNode(nodes []*batchNode, node *Node) *batchNode {
	for i := range nodes {
		// Note: using pointer equality for performance.
		if nodes[i].Node == node {
			return nodes[i]
		}
	}
	return nil
}

type batchNamespace struct {
	namespace  *string
	offsets    []int
	offsetSize int
}

func newBatchNamespace(namespace *string, capacity, offset int) *batchNamespace {
	res := &batchNamespace{
		namespace:  namespace,
		offsets:    make([]int, 0, capacity),
		offsetSize: 1,
	}
	res.offsets = append(res.offsets, offset)

	return res
}

func (bn *batchNamespace) add(offset int) {
	bn.offsets = append(bn.offsets, offset)
	bn.offsetSize++
}
