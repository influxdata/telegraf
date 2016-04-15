package newrelic

import (
	"testing"
	// "github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/newrelic"
	// "github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
  dm := newrelic.DemoMetric{MyName: "Foo", TagList: newrelic.DemoTagList()}
  component := newrelic.NewRelicComponent{TMetric: dm}
  require.EqualValues(t, component.Name(), "Foo")
}

func TestGuid(t *testing.T) {
  dm := newrelic.DemoMetric{MyName: "Lulu", TagList: newrelic.DemoTagList()}
  component := newrelic.NewRelicComponent{TMetric: dm, GuidBase: "org.betterplace.telegraf-agent"}
  require.EqualValues(t, component.Guid(), "org.betterplace.telegraf-agent-lulu")
}

func TestTags(t *testing.T) {
  dm := newrelic.DemoMetric{MyName: "Lulu", TagList: newrelic.DemoTagList()}
  component := newrelic.NewRelicComponent{TMetric: dm}
  component.Tags()
}

func TestHostname(t *testing.T) {
  tagList := newrelic.DemoTagList()
  tagList["host"] = "baba"
  dm := newrelic.DemoMetric{MyName: "Lulu", TagList: tagList}
  component := newrelic.NewRelicComponent{TMetric: dm}
  require.EqualValues(t, component.Hostname(), "baba")
}

func TestMetricName(t *testing.T) {
  tagList := newrelic.DemoTagList()
  dm := newrelic.DemoMetric{MyName: "Lulu", TagList: tagList}
  component := newrelic.NewRelicComponent{TMetric: dm}
	require.EqualValues(t, component.MetricName("fnord"), "Component/Lulu/Fnord/Fluff-naa/Hoof-bar/Zoo-goo[Units]")
}
