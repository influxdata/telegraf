package bind

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Bind struct {
	Urls []string
}

var sampleConfig = `
  ## An array of BIND XML statistics URI to gather stats.
  ## Default is "http://localhost:8053/".
  urls = ["http://localhost:8053/"]
`

func (p *Bind) Description() string {
	return "Read BIND nameserver XML statistics"
}

func (p *Bind) SampleConfig() string {
	return sampleConfig
}

func (p *Bind) Gather(acc telegraf.Accumulator) error {
	var outerr error
	var errch = make(chan error)

	for _, u := range p.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}

		go func(addr *url.URL) {
			errch <- p.GatherUrl(addr, acc)
		}(addr)
	}

	// Drain channel, waiting for all requests to finish and save last error
	for range p.Urls {
		if err := <-errch; err != nil {
			outerr = err
		}
	}

	return outerr
}

func (p *Bind) GatherUrl(addr *url.URL, acc telegraf.Accumulator) error {
	resp, err := http.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}

	// Wrap reader in a buffered reader so that we can peek ahead to determine schema version
	br := bufio.NewReader(resp.Body)

	if p, err := br.Peek(256); err != nil {
		return fmt.Errorf("Unable to peek ahead in stream to determine statistics version: %s", err)
	} else {
		var xmlRoot struct {
			XMLName xml.Name
			Version string `xml:"version,attr"`
		}

		err := xml.Unmarshal(p, &xmlRoot)

		if err != nil {
			// We expect an EOF error since we only fed the decoder a small fragment
			if _, ok := err.(*xml.SyntaxError); !ok {
				return fmt.Errorf("XML syntax error: %s", err)
			}
		}

		if xmlRoot.XMLName.Local == "statistics" && strings.HasPrefix(xmlRoot.Version, "3.") {
			// TODO: copy parsed stats into struct to feed into telegraf.Accumulator
			readStatsV3(br)
		} else {
			return readStatsV2(br, acc)
		}
	}

	return nil
}

func init() {
	inputs.Add("bind", func() telegraf.Input { return &Bind{} })
}
