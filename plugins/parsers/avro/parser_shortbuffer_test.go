package avro

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A short or empty message must be rejected with an error rather than
// panicking on an out-of-range slice. Consumer inputs (MQTT, Kafka, NATS,
// AMQP) hand remote payloads straight to Parse, so a panic here crashes the
// agent.
func TestParseShortBuffer(t *testing.T) {
	bare := &Parser{Schema: `{"type":"record","name":"r","fields":[{"name":"a","type":"long"}]}`}
	require.NoError(t, bare.Init())

	// Empty bare-Avro payload: too short for the single-object-encoding marker
	// and not decodable against the record schema.
	_, err := bare.Parse([]byte{})
	require.Error(t, err)

	// Single-object-encoding marker present but the buffer is truncated before
	// the eight-byte fingerprint completes.
	_, err = bare.Parse([]byte{'c', '3', 0x00, 0x00, 0x00})
	require.ErrorContains(t, err, "single-object encoding")

	// Confluent Wire Protocol expects a zero byte plus a four-byte schema ID.
	reg := &Parser{SchemaRegistry: "http://localhost:0"}
	require.NoError(t, reg.Init())
	_, err = reg.Parse([]byte{0x00, 0x01})
	require.ErrorContains(t, err, "Confluent Wire Protocol")
}
