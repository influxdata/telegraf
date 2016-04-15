package newrelic

import (
	"testing"
  "time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/newrelic"
	// "github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"

)

func TestAddMetrics(t *testing.T) {
  dm := []telegraf.Metric{newrelic.DemoMetric{MyName: "Foo", TagList: newrelic.DemoTagList()}}
  data := newrelic.NewRelicData{
		LastWrite: time.Now(),
		Hosts: make(map[string][]newrelic.NewRelicComponent),
		GuidBase: "org.betterplace.test-foo"}
  data.AddMetrics(dm)
  require.EqualValues(t, len(data.Hosts), 1)
}

func TestDataSets(t *testing.T) {
  dm := []telegraf.Metric{newrelic.DemoMetric{MyName: "Foo", TagList: newrelic.DemoTagList()}}
  data := newrelic.NewRelicData{
		LastWrite: time.Now(),
		Hosts: make(map[string][]newrelic.NewRelicComponent),
		GuidBase: "org.betterplace.test-foo"}
  data.AddMetrics(dm)
  sets := data.DataSets()
  require.EqualValues(t, len(sets), 1)
  set, _ := sets[0].(map[string]interface{})
  agent := set["agent"].(map[string]string)
  require.EqualValues(t, agent["host"], "Hulu")
}
