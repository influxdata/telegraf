package statsd

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventGather(t *testing.T) {
	acc := &testutil.Accumulator{}

	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text", "default-hostname")
	require.Nil(t, err)
	err = s.parseEventMessage(now.Add(1), "_e{10,24}:test title|test\\line1\\nline2\\nline3", "default-hostname")
	require.Nil(t, err)
	err = s.parseEventMessage(now.Add(2), "_e{10,9}:test title|test text|d:21", "default-hostname")
	require.Nil(t, err)

	err = s.Gather(acc)
	require.Nil(t, err)

	assert.Equal(t, acc.NMetrics(), uint64(3))

	assert.Equal(t, "test title", acc.Metrics[0].Measurement)
	assert.Equal(t, "test title", acc.Metrics[1].Measurement)
	assert.Equal(t, "test title", acc.Metrics[2].Measurement)

	assert.Equal(t, map[string]string{"source": "default-hostname"}, acc.Metrics[0].Tags)
	assert.Equal(t, map[string]string{"source": "default-hostname"}, acc.Metrics[1].Tags)
	assert.Equal(t, map[string]string{"source": "default-hostname"}, acc.Metrics[2].Tags)

	assert.Equal(t,
		map[string]interface{}{
			"priority":   priorityNormal,
			"alert-type": "info",
			"text":       "test text",
		},
		acc.Metrics[0].Fields)
	assert.Equal(t, map[string]interface{}{
		"priority":   priorityNormal,
		"alert-type": "info",
		"text":       "test\\line1\nline2\nline3",
	}, acc.Metrics[1].Fields)
	assert.Equal(t, map[string]interface{}{
		"priority":   priorityNormal,
		"alert-type": "info",
		"text":       "test text",
		"ts":         int64(21),
	}, acc.Metrics[2].Fields)
}

// These tests adapted from tests in
// https://github.com/DataDog/datadog-agent/blob/master/pkg/dogstatsd/parser_test.go
// to ensure compatibility with the datadog-agent parser
func TestEventMinimal(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text", "default-hostname")
	require.Nil(t, err)
	e := s.events[0]

	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, now, e.ts)
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, priorityNormal, e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
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
	assert.Len(t, e.tags, 1)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
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

func TestEventMetadataTimestamp(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|d:21", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(21), e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataPriority(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|p:low", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "low", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataHostname(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|h:localhost", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "localhost", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataHostnameInTag(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|#host:localhost", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "localhost", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataEmptyHostTag(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|#host:,other:tag", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "", e.tags["source"])
	assert.Equal(t, map[string]string{"other": "tag", "source": ""}, e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataAlertType(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|t:warning", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "warning", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])

}

func TestEventMetadataAggregatioKey(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|k:some aggregation key", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "some aggregation key", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataSourceType(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|s:this is the source", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, "this is the source", e.fields["source-type-name"])
}

func TestEventMetadataTags(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|#tag1,tag2:test", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, nil, e.fields["ts"])
	assert.Equal(t, "normal", e.fields["priority"])
	assert.Equal(t, "default-hostname", e.tags["source"])
	assert.Equal(t, map[string]string{"tag1": "", "tag2": "test", "source": "default-hostname"}, e.tags)
	assert.Equal(t, "info", e.fields["alert-type"])
	assert.Equal(t, "", e.tags["aggregation-key"])
	assert.Equal(t, nil, e.fields["source-type-name"])
}

func TestEventMetadataMultiple(t *testing.T) {
	now := time.Now()
	s := NewTestStatsd()
	err := s.parseEventMessage(now, "_e{10,9}:test title|test text|t:warning|d:12345|p:low|h:some.host|k:aggKey|s:source test|#tag1,tag2:test", "default-hostname")

	require.Nil(t, err)
	e := s.events[0]
	assert.Equal(t, "test title", e.name)
	assert.Equal(t, "test text", e.fields["text"])
	assert.Equal(t, int64(12345), e.fields["ts"])
	assert.Equal(t, "low", e.fields["priority"])
	assert.Equal(t, "some.host", e.tags["source"])
	assert.Equal(t, map[string]string{"aggregation-key": "aggKey", "tag1": "", "tag2": "test", "source": "some.host"}, e.tags)
	assert.Equal(t, "warning", e.fields["alert-type"])
	assert.Equal(t, "aggKey", e.tags["aggregation-key"])
	assert.Equal(t, "source test", e.fields["source-type-name"])
}
