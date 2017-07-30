package bind

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Bind struct {
	Urls                 []string
	GatherMemoryContexts bool
	GatherViews          bool
}

var sampleConfig = `
  ## An array of BIND XML statistics URI to gather stats.
  ## Default is "http://localhost:8053/xml/v3".
  # urls = ["http://localhost:8053/xml/v3"]
  # gather_memory_contexts = false
  # gather_views = false
`

var client = &http.Client{
	Timeout: time.Duration(4 * time.Second),
}

func (b *Bind) Description() string {
	return "Read BIND nameserver XML statistics"
}

func (b *Bind) SampleConfig() string {
	return sampleConfig
}

func (b *Bind) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(b.Urls) == 0 {
		b.Urls = []string{"http://localhost:8053/xml/v3"}
	}

	for _, u := range b.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(b.gatherUrl(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (b *Bind) gatherUrl(addr *url.URL, acc telegraf.Accumulator) error {
	switch addr.Path {
	case "":
		// BIND 9.6 - 9.8
		return b.readStatsXMLv2(addr, acc)
	case "/json/v1":
		// BIND 9.10+
		return b.readStatsJSON(addr, acc)
	case "/xml/v2":
		// BIND 9.9
		return b.readStatsXMLv2(addr, acc)
	case "/xml/v3":
		// BIND 9.9+
		return b.readStatsXMLv3(addr, acc)
	default:
		return fmt.Errorf("URL %s is ambiguous. Please check plugin documentation for supported URL formats.",
			addr)
	}
}

func init() {
	inputs.Add("bind", func() telegraf.Input { return &Bind{} })
}
