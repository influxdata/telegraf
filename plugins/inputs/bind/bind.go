package bind

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Bind struct {
	Urls                 []string
	GatherMemoryContexts bool
	GatherViews          bool
	Timeout              config.Duration `toml:"timeout"`

	client http.Client
}

func (b *Bind) Init() error {
	b.client = http.Client{
		Timeout: time.Duration(b.Timeout),
	}

	return nil
}

func (b *Bind) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(b.Urls) == 0 {
		b.Urls = []string{"http://localhost:8053/xml/v3"}
	}

	for _, u := range b.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(b.gatherURL(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (b *Bind) gatherURL(addr *url.URL, acc telegraf.Accumulator) error {
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
		return fmt.Errorf("provided URL %s is ambiguous, please check plugin documentation for supported URL formats",
			addr)
	}
}

func init() {
	inputs.Add("bind", func() telegraf.Input { return &Bind{} })
}
