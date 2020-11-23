package loki

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

type tuple struct {
	key, value string
}

func generateLabelsAndTag(tt ...tuple) (map[string]string, []*telegraf.Tag) {
	labels := map[string]string{}
	var tags []*telegraf.Tag

	for _, t := range tt {
		labels[t.key] = t.value
		tags = append(tags, &telegraf.Tag{Key: t.key, Value: t.value})
	}

	return labels, tags
}

func TestGenerateLabelsAndTag(t *testing.T) {
	labels, tags := generateLabelsAndTag(
		tuple{key: "key1", value: "value1"},
		tuple{key: "key2", value: "value2"},
		tuple{key: "key3", value: "value3"},
	)

	expectedTags := []*telegraf.Tag{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
		{Key: "key3", Value: "value3"},
	}

	require.Len(t, labels, 3)
	require.Len(t, tags, 3)
	require.Equal(t, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}, labels)
	require.Equal(t, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}, labels)
	require.Equal(t, expectedTags, tags)
}

func TestStream_insertLog(t *testing.T) {
	s := Streams{Streams: make([]Stream, 0)}
	log1 := Log{"123", "this log isn't useful"}
	log2 := Log{"124", "this log isn't useful neither"}
	log3 := Log{"122", "again"}

	labels1, tags1 := generateLabelsAndTag(
		tuple{key: "key1", value: "value1"},
		tuple{key: "key2", value: "value2"},
		tuple{key: "key3", value: "value3"},
	)

	labels2, tags2 := generateLabelsAndTag(
		tuple{key: "key2", value: "value2"},
	)

	s.insertLog(tags1, log1)

	require.Len(t, s.Streams, 1)
	require.Equal(t, "key1value1key2value2key3value3", s.Streams[0].key)
	require.Len(t, s.Streams[0].Logs, 1)
	require.Equal(t, labels1, s.Streams[0].Labels)
	require.Equal(t, "123", s.Streams[0].Logs[0][0])
	require.Equal(t, "this log isn't useful", s.Streams[0].Logs[0][1])

	s.insertLog(tags1, log2)

	require.Len(t, s.Streams, 1)
	require.Len(t, s.Streams[0].Logs, 2)
	require.Equal(t, "124", s.Streams[0].Logs[1][0])
	require.Equal(t, "this log isn't useful neither", s.Streams[0].Logs[1][1])

	s.insertLog(tags2, log3)

	require.Len(t, s.Streams, 2)
	require.Equal(t, "key2value2", s.Streams[1].key)
	require.Len(t, s.Streams[1].Logs, 1)
	require.Equal(t, labels2, s.Streams[1].Labels)
	require.Equal(t, "122", s.Streams[1].Logs[0][0])
	require.Equal(t, "again", s.Streams[1].Logs[0][1])
}

func TestUniqKeyFromTagList(t *testing.T) {
	tests := []struct {
		in  []*telegraf.Tag
		out string
	}{
		{
			in: []*telegraf.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
				{Key: "key3", Value: "value3"},
			},
			out: "key1value1key2value2key3value3",
		},
		{
			in: []*telegraf.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key3", Value: "value3"},
				{Key: "key4", Value: "value4"},
			},
			out: "key1value1key3value3key4value4",
		},
	}

	for _, test := range tests {
		require.Equal(t, test.out, uniqKeyFromTagList(test.in))
	}
}

func Test_newStream(t *testing.T) {
	labels, tags := generateLabelsAndTag(
		tuple{key: "key1", value: "value1"},
		tuple{key: "key2", value: "value2"},
		tuple{key: "key3", value: "value3"},
	)

	s := newStream("uniqKey", tags)

	require.Equal(t, s.key, "uniqKey")
	require.Empty(t, s.Logs)
	require.Equal(t, s.Labels, labels)
}

func Test_newStream_noTag(t *testing.T) {
	s := newStream("uniqKey-empty", nil)

	require.Equal(t, s.key, "uniqKey-empty")
	require.Empty(t, s.Logs)
	require.Empty(t, s.Labels)
}
