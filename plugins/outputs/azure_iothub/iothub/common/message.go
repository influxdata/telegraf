package common

import (
	"time"
)

// Message is a common message format for all device-facing protocols.
// This message format is used for both device-to-cloud and cloud-to-device messages.
// See: https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-devguide-messages-construct
type Message struct {
	// MessageID is a user-settable identifier for the message used for request-reply patterns.
	MessageID string `json:"MessageId,omitempty"`

	// To is a destination specified in cloud-to-device messages.
	To string `json:"To,omitempty"`

	// ExpiryTime is time of message expiration.
	ExpiryTime *time.Time `json:"ExpiryTimeUtc,omitempty"`

	// EnqueuedTime is time the Cloud-to-Device message was received by IoT Hub.
	EnqueuedTime *time.Time `json:"EnqueuedTime,omitempty"`

	// CorrelationID is a string property in a response message that typically
	// contains the MessageId of the request, in request-reply patterns.
	CorrelationID string `json:"CorrelationId,omitempty"`

	// UserID is an ID used to specify the origin of messages.
	UserID string `json:"UserId,omitempty"`

	// ConnectionDeviceID is an ID set by IoT Hub on device-to-cloud messages.
	// It contains the deviceId of the device that sent the message.
	ConnectionDeviceID string `json:"ConnectionDeviceId,omitempty"`

	// ConnectionDeviceGenerationID is an ID set by IoT Hub on device-to-cloud messages.
	// It contains the generationId (as per Device identity properties)
	// of the device that sent the message.
	ConnectionDeviceGenerationID string `json:"ConnectionDeviceGenerationId,omitempty"`

	// ConnectionAuthMethod is an authentication method set by IoT Hub on
	// device-to-cloud messages. This property contains information about
	// the authentication method used to authenticate the device sending the message.
	ConnectionAuthMethod *ConnectionAuthMethod `json:"ConnectionAuthMethod,omitempty"`

	// MessageSource determines a device-to-cloud message transport.
	MessageSource string `json:"MessageSource,omitempty"`

	// Payload is message data.
	Payload []byte `json:"Payload,omitempty"`

	// Properties are custom message properties (property bags).
	Properties map[string]string `json:"Properties,omitempty"`

	// TransportOptions transport specific options.
	TransportOptions map[string]interface{} `json:"-"`
}

// ConnectionAuthMethod is an authentication method of device-to-cloud communication.
type ConnectionAuthMethod struct {
	Scope  string `json:"scope"`
	Type   string `json:"type"`
	Issuer string `json:"issuer"`
}
