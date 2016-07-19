package newrelic

import (
	"testing"
	"github.com/influxdata/telegraf/plugins/outputs/newrelic"
	// "github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"

)

func TestTagValue(t *testing.T) {
	require.EqualValues(t, newrelic.TagValue("Hello/World"), "hello-world")
	require.EqualValues(t, newrelic.TagValue("/Hello/World"), "hello-world")
	require.EqualValues(t, newrelic.TagValue(""), "root")
}

func TestFillSimple(t *testing.T) {
	tags := &newrelic.NewRelicTags{}
	tags.Fill(newrelic.DemoTagList())
	require.EqualValues(t, tags.SortedKeys, []string{"Fluff", "Hoof", "Zoo"})
}

func TestFillWithHost(t *testing.T) {
	tags := &newrelic.NewRelicTags{}
	demoList := newrelic.DemoTagList()
	demoList["host"] = "hulu"
	tags.Fill(demoList)
	require.EqualValues(t, tags.SortedKeys, []string{"Fluff", "Hoof", "Zoo"})
	require.EqualValues(t, tags.Hostname, "hulu")
}

func TestFillAndGetTag(t *testing.T) {
	tags := &newrelic.NewRelicTags{}
	tags.Fill(newrelic.DemoTagList())
	require.EqualValues(t, tags.GetTag("Zoo"), "goo")
}
