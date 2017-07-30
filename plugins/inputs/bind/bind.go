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
  ## Default is "http://localhost:8053/".
  # urls = ["http://localhost:8053/"]
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
		b.Urls = []string{"http://localhost:8053/"}
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
	case "/json/v1":
		return b.readStatsJSON(addr, acc)
	case "/xml/v2":
		return b.readStatsXMLv2(addr, acc)
	case "/xml/v3":
		return b.readStatsXMLv3(addr, acc)
	default:
		return fmt.Errorf("URL %s is ambiguous. Please include a /json/v1, /xml/v2, or /xml/v3 path component.",
			addr)
	}
}

func init() {
	inputs.Add("bind", func() telegraf.Input { return &Bind{} })
}
