// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "encoding/json"

// jsonOrString returns its input as a jsonString if it is valid JSON, and as a
// string otherwise.
func jsonOrString(d []byte) interface{} {
	var js json.RawMessage
	if err := json.Unmarshal(d, &js); err == nil {
		return jsonString(d)
	}
	return string(d)
}

// jsonString assists in debug logging: The debug map could be marshalled as
// JSON or just printed directly.
type jsonString string

// MarshalJSON returns the JSONString unmodified without any escaping.
func (js jsonString) MarshalJSON() ([]byte, error) {
	if "" == js {
		return []byte("null"), nil
	}
	return []byte(js), nil
}
