package statsd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventMinimal(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text", "default-hostname")
	require.Nil(t, err)
	e := s.events[0]

	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, now, e.fields["ts"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, priorityNormal, e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, nil, e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
}

func TestEventMultilinesText(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,24}:test title|test\\line1\\nline2\\nline3", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test\\line1\nline2\nline3", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
}

func TestEventPipeInTitle(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,24}:test|title|test\\line1\\nline2\\nline3", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test|title", e.name)
	assert.Equal(t, "test\\line1\nline2\nline3", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Len(t, e.tags, 1)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventError(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	// missing length header
	err := s.parseEventMessage(now, "_e:title|text", "default-hostname")
	assert.Error(t, err)

	// greater length than packet
	err = s.parseEventMessage(now, "_e{10,10}:title|text", "default-hostname")
	assert.Error(t, err)

	// zero length
	err = s.parseEventMessage(now, "_e{0,0}:a|a", "default-hostname")
	assert.Error(t, err)

	// missing title or text length
	err = s.parseEventMessage(now, "_e{5555:title|text", "default-hostname")
	assert.Error(t, err)

	// missing wrong len format
	err = s.parseEventMessage(now, "_e{a,1}:title|text", "default-hostname")
	assert.Error(t, err)

	err = s.parseEventMessage(now, "_e{1,a}:title|text", "default-hostname")
	assert.Error(t, err)

	// missing title or text length
	err = s.parseEventMessage(now, "_e{5,}:title|text", "default-hostname")
	assert.Error(t, err)

	err = s.parseEventMessage(now, "_e{,4}:title|text", "default-hostname")
	assert.Error(t, err)

	err = s.parseEventMessage(now, "_e{}:title|text", "default-hostname")
	assert.Error(t, err)

	err = s.parseEventMessage(now, "_e{,}:title|text", "default-hostname")
	assert.Error(t, err)

	// not enough information
	err = s.parseEventMessage(now, "_e|text", "default-hostname")
	assert.Error(t, err)

	err = s.parseEventMessage(now, "_e:|text", "default-hostname")
	assert.Error(t, err)

	// invalid timestamp
	err = s.parseEventMessage(now, "_e{5,4}:title|text|d:abc", "default-hostname")
	assert.NoError(t, err)

	// invalid priority
	err = s.parseEventMessage(now, "_e{5,4}:title|text|p:urgent", "default-hostname")
	assert.NoError(t, err)

	// invalid priority
	err = s.parseEventMessage(now, "_e{5,4}:title|text|p:urgent", "default-hostname")
	assert.NoError(t, err)

	// invalid alert type
	err = s.parseEventMessage(now, "_e{5,4}:title|text|t:test", "default-hostname")
	assert.NoError(t, err)

	// unknown metadata
	err = s.parseEventMessage(now, "_e{5,4}:title|text|x:1234", "default-hostname")
	assert.NoError(t, err)
}

/*
func TestEventMetadataTimestamp(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|d:21"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(21), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataPriority(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|p:low"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, metrics.EventPriorityLow, e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataHostname(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|h:localhost"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "localhost", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataHostnameInTag(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|#host:localhost"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "localhost", e.tags["source"])
	assert.Equal(t, []string{}, e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataEmptyHostTag(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|#host:,other:tag"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "", e.tags["source"])
	assert.Equal(t, []string{"other:tag"}, e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataAlertType(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|t:warning"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, metrics.EventAlertTypeWarning, e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataAggregatioKey(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|k:some aggregation key"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "some aggregation key", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataSourceType(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|s:this is the source"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string(nil), e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "this is the source", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataTags(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|#tag1,tag2:test"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(0), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, []string{"tag1", "tag2:test"}, e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.fields["aggregation-key"])
	assert.Equal(t, "", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}

func TestEventMetadataMultiple(t *testing.T) {
	e, err := parseEventMessage([]byte("_e{10,9}:test title|test text|t:warning|d:12345|p:low|h:some.host|k:aggKey|s:source test|#tag1,tag2:test"), "default-hostname")

	require.Nil(t, err)
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(12345), e.fields["ts"])
	assert.Equal(t, metrics.EventPriorityLow, e.fields["priority"])
	assert.Equal(t, "some.host", e.tags["source"])
	assert.Equal(t, []string{"tag1", "tag2:test"}, e.tags)
	assert.Equal(t, metrics.EventAlertTypeWarning, e.fields["alert-type"])
	assert.Equal(t, "aggKey", e.fields["aggregation-key"])
	assert.Equal(t, "source test", e.tags["source-type-name"])
	assert.Equal(t, "", e.EventType)
}
*/
