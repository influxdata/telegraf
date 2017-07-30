package bind

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
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
			acc.AddError(b.GatherUrl(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (b *Bind) GatherUrl(addr *url.URL, acc telegraf.Accumulator) error {
	resp, err := client.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status: %s", addr, resp.Status)
	}

	log.Printf("D! Response content length: %d", resp.ContentLength)

	contentType := resp.Header.Get("Content-Type")

	if contentType == "text/xml" {
		// Wrap reader in a buffered reader so that we can peek ahead to determine schema version
		br := bufio.NewReader(resp.Body)

		if p, err := br.Peek(256); err != nil {
			return fmt.Errorf("Unable to peek ahead in stream to determine statistics version: %s", err)
		} else {
			var xmlRoot struct {
				XMLName xml.Name
				Version float64 `xml:"version,attr"`
			}

			err := xml.Unmarshal(p, &xmlRoot)

			if err != nil {
				// We expect an EOF error since we only fed the decoder a small fragment
				if _, ok := err.(*xml.SyntaxError); !ok {
					return fmt.Errorf("XML syntax error: %s", err)
				}
			}

			if (xmlRoot.XMLName.Local == "statistics") && (int(xmlRoot.Version) == 3) {
				return b.readStatsV3(br, acc, addr.Host)
			} else {
				return b.readStatsV2(br, acc, addr.Host)
			}
		}
	} else if contentType == "application/json" {
		return b.readStatsJson(resp.Body, acc, addr.Host)
	} else {
		return fmt.Errorf("Unsupported Content-Type in response: %#v", contentType)
	}
}

func init() {
	inputs.Add("bind", func() telegraf.Input { return &Bind{} })
}
