/*
 * Copyright 2013-2017 Aerospike, Inc.
 *
 * Portions may be licensed to Aerospike, Inc. under one or more contributor
 * license agreements WHICH ARE COMPATIBLE WITH THE APACHE LICENSE, VERSION 2.0.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */

package aerospike

import (
	"encoding/base64"
	"fmt"
	"strconv"

	. "github.com/aerospike/aerospike-client-go/logger"
	. "github.com/aerospike/aerospike-client-go/types"
)

const (
	_PartitionGeneration = "partition-generation"
	_ReplicasMaster      = "replicas-master"
	_ReplicasAll         = "replicas-all"
)

// Parse node's master (and optionally prole) partitions.
type partitionParser struct {
	pmap           partitionMap
	buffer         []byte
	partitionCount int
	generation     int
	length         int
	offset         int
}

func newPartitionParser(node *Node, partitionCount int, requestProleReplicas bool) (*partitionParser, error) {
	newPartitionParser := &partitionParser{
		partitionCount: partitionCount,
	}

	// Send format 1:  partition-generation\nreplicas-master\n
	// Send format 2:  partition-generation\nreplicas-all\n
	command := _ReplicasMaster
	if requestProleReplicas {
		command = _ReplicasAll
	}
	info, err := node.requestRawInfo(_PartitionGeneration, command)
	if err != nil {
		return nil, err
	}

	newPartitionParser.buffer = info.msg.Data
	newPartitionParser.length = len(info.msg.Data)
	if newPartitionParser.length == 0 {
		return nil, NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Partition info is empty"))
	}

	newPartitionParser.generation, err = newPartitionParser.parseGeneration()
	if err != nil {
		return nil, err
	}

	newPartitionParser.pmap = make(partitionMap)

	if requestProleReplicas {
		err = newPartitionParser.parseReplicasAll(node)
	} else {
		err = newPartitionParser.parseReplicasMaster(node)
	}

	if err != nil {
		return nil, err
	}

	return newPartitionParser, nil
}

func (pp *partitionParser) getGeneration() int {
	return pp.generation
}

func (pp *partitionParser) getPartitionMap() partitionMap {
	return pp.pmap
}

func (pp *partitionParser) parseGeneration() (int, error) {
	if err := pp.expectName(_PartitionGeneration); err != nil {
		return -1, err
	}

	begin := pp.offset
	for pp.offset < pp.length {
		if pp.buffer[pp.offset] == '\n' {
			s := string(pp.buffer[begin:pp.offset])
			pp.offset++
			return strconv.Atoi(s)
		}
		pp.offset++
	}
	return -1, NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Failed to find partition-generation value"))
}

func (pp *partitionParser) parseReplicasMaster(node *Node) error {
	// Use low-level info methods and parse byte array directly for maximum performance.
	// Receive format: replicas-master\t<ns1>:<base 64 encoded bitmap1>;<ns2>:<base 64 encoded bitmap2>...\n
	if err := pp.expectName(_ReplicasMaster); err != nil {
		return err
	}

	begin := pp.offset

	for pp.offset < pp.length {
		if pp.buffer[pp.offset] == ':' {
			// Parse namespace.
			namespace := string(pp.buffer[begin:pp.offset])

			if len(namespace) <= 0 || len(namespace) >= 32 {
				response := pp.getTruncatedResponse()
				return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid partition namespace `%s` response: `%s`", namespace, response))
			}
			pp.offset++
			begin = pp.offset

			// Parse partition bitmap.
			for pp.offset < pp.length {
				b := pp.buffer[pp.offset]

				if b == ';' || b == '\n' {
					break
				}
				pp.offset++
			}

			if pp.offset == begin {
				response := pp.getTruncatedResponse()
				return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Empty partition id for namespace `%s` response: `%s`", namespace, response))
			}

			replicaArray := pp.pmap[namespace]

			if replicaArray == nil {
				replicaArray = make([][]*Node, 1)
				replicaArray[0] = make([]*Node, pp.partitionCount)
				pp.pmap[namespace] = replicaArray
			}

			if err := pp.decodeBitmap(node, replicaArray[0], begin); err != nil {
				return err
			}
			pp.offset++
			begin = pp.offset
		} else {
			pp.offset++
		}
	}

	return nil
}

func (pp *partitionParser) parseReplicasAll(node *Node) error {
	// Use low-level info methods and parse byte array directly for maximum performance.
	// Receive format: replicas-all\t
	//                 <ns1>:<count>,<base 64 encoded bitmap1>,<base 64 encoded bitmap2>...;
	//                 <ns2>:<count>,<base 64 encoded bitmap1>,<base 64 encoded bitmap2>...;\n
	if err := pp.expectName(_ReplicasAll); err != nil {
		return err
	}

	begin := pp.offset

	for pp.offset < pp.length {
		if pp.buffer[pp.offset] == ':' {
			// Parse namespace.
			namespace := string(pp.buffer[begin:pp.offset])

			if len(namespace) <= 0 || len(namespace) >= 32 {
				response := pp.getTruncatedResponse()
				return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid partition namespace `%s` response: `%s`", namespace, response))
			}
			pp.offset++
			begin = pp.offset

			// Parse replica count.
			for pp.offset < pp.length {
				b := pp.buffer[pp.offset]

				if b == ',' {
					break
				}
				pp.offset++
			}

			replicaCount, err := strconv.Atoi(string(pp.buffer[begin:pp.offset]))
			if err != nil {
				return err
			}

			// Ensure replicaArray is correct size.
			replicaArray := pp.pmap[namespace]

			if replicaArray == nil {
				// Create new replica array.
				replicaArray = make([][]*Node, replicaCount)

				for i := 0; i < replicaCount; i++ {
					replicaArray[i] = make([]*Node, pp.partitionCount)
				}

				pp.pmap[namespace] = replicaArray
			} else if len(replicaArray) != replicaCount {
				Logger.Info("Namespace `%s` replication factor changed from `%d` to `%d` ", namespace, len(replicaArray), replicaCount)

				// Resize replica array.
				replicaTarget := make([][]*Node, replicaCount)

				if len(replicaArray) < replicaCount {
					i := 0

					// Copy existing entries.
					for ; i < len(replicaArray); i++ {
						replicaTarget[i] = replicaArray[i]
					}

					// Create new entries.
					for ; i < replicaCount; i++ {
						replicaTarget[i] = make([]*Node, pp.partitionCount)
					}
				} else {
					// Copy existing entries.
					for i := 0; i < replicaCount; i++ {
						replicaTarget[i] = replicaArray[i]
					}
				}

				replicaArray = replicaTarget
				pp.pmap[namespace] = replicaArray
			}

			// Parse partition bitmaps.
			for i := 0; i < replicaCount; i++ {
				pp.offset++
				begin = pp.offset

				// Find bitmap endpoint
				for pp.offset < pp.length {
					b := pp.buffer[pp.offset]

					if b == ',' || b == ';' {
						break
					}
					pp.offset++
				}

				if pp.offset == begin {
					response := pp.getTruncatedResponse()
					return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Empty partition id for namespace `%s` response: `%s`", namespace, response))
				}

				if err := pp.decodeBitmap(node, replicaArray[i], begin); err != nil {
					return err
				}
			}
			pp.offset++
			begin = pp.offset
		} else {
			pp.offset++
		}
	}

	return nil
}

func (pp *partitionParser) decodeBitmap(node *Node, nodeArray []*Node, begin int) error {
	restoreBuffer, err := base64.StdEncoding.DecodeString(string(pp.buffer[begin:pp.offset]))
	if err != nil {
		return err
	}

	for i := 0; i < pp.partitionCount; i++ {
		nodeOld := nodeArray[i]

		if (restoreBuffer[i>>3] & (0x80 >> uint(i&7))) != 0 {
			// Node owns this partition.
			if nodeOld != nil && nodeOld != node {
				// Force previously mapped node to refresh it's partition map on next cluster tend.
				nodeOld.partitionGeneration.Set(-1)
			}

			// Use lazy set because there is only one producer thread. In addition,
			// there is a one second delay due to the cluster tend polling interval.
			// An extra millisecond for a node change will not make a difference and
			// overall performance is improved.
			nodeArray[i] = node
		} else {
			// Node does not own partition.
			if node == nodeOld {
				// Must erase previous map.
				nodeArray[i] = nil
			}
		}
	}

	return nil
}

func (pp *partitionParser) expectName(name string) error {
	begin := pp.offset

	for pp.offset < pp.length {
		if pp.buffer[pp.offset] == '\t' {
			s := string(pp.buffer[begin:pp.offset])
			if name == s {
				pp.offset++
				return nil
			}
			break
		}
		pp.offset++
	}

	return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Failed to find `%s`", name))
}

func (pp *partitionParser) getTruncatedResponse() string {
	max := pp.length
	if max > 200 {
		max = 200
	}
	return string(pp.buffer[0:max])
}
