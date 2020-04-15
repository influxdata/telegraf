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

import "time"

// AdminPolicy contains attributes used for user administration commands.
type AdminPolicy struct {

	// User administration command socket timeout in milliseconds.
	// Default is one second timeout.
	Timeout time.Duration
}

// NewAdminPolicy generates a new AdminPolicy with default values.
func NewAdminPolicy() *AdminPolicy {
	return &AdminPolicy{
		Timeout: 1 * time.Second,
	}
}
