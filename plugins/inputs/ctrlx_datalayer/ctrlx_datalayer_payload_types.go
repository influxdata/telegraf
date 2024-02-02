package ctrlx_datalayer

// Once a subscription is created, the server will send event notifications to this plugin.
// This file contains the different event types and the included event payload.

// sseEventData represents the json structure send by the ctrlX CORE
// server on an "update" event.
type sseEventData struct {
	Node      string      `json:"node"`
	Timestamp int64       `json:"timestamp"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
}

// sseEventError represents the json structure send by the ctrlX CORE
// server on an "error" event.
type sseEventError struct {
	Instance string `json:"instance"`
}
