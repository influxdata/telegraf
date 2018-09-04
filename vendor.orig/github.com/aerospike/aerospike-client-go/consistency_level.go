/*
 * Copyright 2013-2017 Aerospike, Inc.
 *
 * Portions may be licensed to Aerospike, Inc. under one or more contributor
 * license agreements.
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

// ConsistencyLevel indicates how replicas should be consulted in a read
// operation to provide the desired consistency guarantee.
type ConsistencyLevel int

const (
	// CONSISTENCY_ONE indicates only a single replica should be consulted in
	// the read operation.
	CONSISTENCY_ONE = iota

	// CONSISTENCY_ALL indicates that all replicas should be consulted in
	// the read operation.
	CONSISTENCY_ALL
)
