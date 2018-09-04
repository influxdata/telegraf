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

// MultiPolicy contains parameters for policy attributes used in
// query and scan operations.
type MultiPolicy struct {
	*BasePolicy

	// Maximum number of concurrent requests to server nodes at any poin int time.
	// If there are 16 nodes in the cluster and maxConcurrentNodes is 8, then queries
	// will be made to 8 nodes in parallel.  When a query completes, a new query will
	// be issued until all 16 nodes have been queried.
	// Default (0) is to issue requests to all server nodes in parallel.
	MaxConcurrentNodes int

	// Number of records to place in queue before blocking.
	// Records received from multiple server nodes will be placed in a queue.
	// A separate goroutine consumes these records in parallel.
	// If the queue is full, the producer goroutines will block until records are consumed.
	RecordQueueSize int //= 5000

	// Blocks until on-going migrations are over
	WaitUntilMigrationsAreOver bool //=false
}

// NewMultiPolicy initializes a MultiPolicy instance with default values.
func NewMultiPolicy() *MultiPolicy {
	return &MultiPolicy{
		BasePolicy:                 NewPolicy(),
		MaxConcurrentNodes:         0,
		RecordQueueSize:            50,
		WaitUntilMigrationsAreOver: false,
	}
}
