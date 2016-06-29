package bind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Bind struct {
	Url string
}

func (_ *Bind) Description() string {
	return "Grab statistics from BIND servers"
}

var bindSampleConfig = `
  ## url of the statistics server
  url = "http://localhost:8053/json"
`

func (_ *Bind) SampleConfig() string {
	return bindSampleConfig
}

func (b *Bind) Gather(acc telegraf.Accumulator) error {
	response, err := http.Get(b.Url)

	if err != nil {
		return fmt.Errorf("error getting json statistics from bind: %s", err)
	}

	buffer := bytes.NewBuffer([]byte{})

	if _, err = io.Copy(buffer, response.Body); err != nil {
		return fmt.Errorf("error reading json statistics from bind: %s", err)
	}

	stats := statistics{}

	if err = json.Unmarshal(buffer.Bytes(), &stats); err != nil {
		return fmt.Errorf("error getting json statistics from bind: %s", err)
	}

	memory := map[string]int64{
		"MemoryTotalUse":    stats.Memory.TotalUse,
		"MemoryInUse":       stats.Memory.InUse,
		"MemoryBlockSize":   stats.Memory.BlockSize,
		"MemoryContextSize": stats.Memory.ContextSize,
		"MemoryLost":        stats.Memory.Lost,
	}

	tags := map[string]string{}
	accumulateFromMap(acc, "memory", memory, tags)
	accumulateFromMap(acc, "opcodes", stats.Opcodes, tags)
	accumulateFromMap(acc, "qtypes", stats.Qtypes, tags)
	accumulateFromMap(acc, "nsstats", stats.Nsstats, tags)
	accumulateFromMap(acc, "sockstats", stats.Sockstats, tags)

	for viewName, view := range stats.Views {
		tags["view"] = viewName

		accumulateFromMap(acc, "stats", view.Resolver.Stats, tags)
		accumulateFromMap(acc, "qtypes", view.Resolver.Qtypes, tags)
		accumulateFromMap(acc, "cache", view.Resolver.Cache, tags)
		accumulateFromMap(acc, "cachestats", view.Resolver.Cachestats, tags)
		accumulateFromMap(acc, "adb", view.Resolver.Adb, tags)
	}

	return nil
}

func accumulateFromMap(acc telegraf.Accumulator, context string, m map[string]int64, tags map[string]string) {
	tags["context"] = context

	fields := map[string]interface{}{}

	for key, value := range m {
		fields[key] = value
	}

	acc.AddFields("bind", fields, tags)
}

func init() {
	inputs.Add("bind", func() telegraf.Input {
		return &Bind{
			Url: "http://localhost:8053/json",
		}
	})
}
