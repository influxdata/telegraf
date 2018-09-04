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
	"fmt"

	Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

// Partition encapsulates partition information.
type Partition struct {
	Namespace   string
	PartitionId int
}

// NewPartitionByKey initializes a partition and determines the Partition Id
// from key digest automatically.
func NewPartitionByKey(key *Key) *Partition {
	partition := newPartitionByKey(key)
	return &partition
}

// newPartitionByKey initializes a partition and determines the Partition Id
// from key digest automatically. It return the struct itself, and not the address
func newPartitionByKey(key *Key) Partition {
	return Partition{
		Namespace: key.namespace,

		// CAN'T USE MOD directly - mod will give negative numbers.
		// First AND makes positive and negative correctly, then mod.
		// For any x, y : x % 2^y = x & (2^y - 1); the second method is twice as fast
		PartitionId: int(Buffer.LittleBytesToInt32(key.digest[:], 0)&0xFFFF) & (_PARTITIONS - 1),
	}
}

// NewPartition generates a partition instance.
func NewPartition(namespace string, partitionId int) *Partition {
	return &Partition{
		Namespace:   namespace,
		PartitionId: partitionId,
	}
}

// String implements the Stringer interface.
func (ptn *Partition) String() string {
	return fmt.Sprintf("%s:%d", ptn.Namespace, ptn.PartitionId)
}

// Equals checks equality of two partitions.
func (ptn *Partition) Equals(other *Partition) bool {
	return ptn.PartitionId == other.PartitionId && ptn.Namespace == other.Namespace
}
