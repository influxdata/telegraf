package icecast

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SourceListmounts holds single listener elements from the Icecast XML
type SourceListmounts struct {
	Mount       string `xml:"mount,attr"`
	Listeners   int32  `xml:"listeners"`
	Connected   int32  `xml:"connected"`
	ContentType string `xml:"content-type"`
}

// Listmounts main structure of the icecast XML
type Listmounts struct {
	Sources []SourceListmounts `xml:"source"`
}

// Icecast contains all the required details to connect to
type Icecast struct {
	Servers         map[string]server `toml:"servers"`
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	Slash           bool              `toml:"slash"`
}

type server struct {
	URL      string
	Alias    string
	Username string
	Password string
}

var sampleConfig = `
  ## Specify the IP adress/hostname to where the '/admin/listmounts' can be found. You can include port if needed.
  ## You can also specify an alias. If none is given, the hostname will be used. Multiple servers can also be specified
  [servers]
	[server.1]
	url = "http://localhost"
	alias = "Server 1"
	username = "telegraf"
	password = "passwd"	

  ## Timeout to the complete conection and reponse time in seconds. Default (5 seconds)
  # response_timeout = "25s"

  ## Include the slash in mountpoint names or not
  slash = false
`

// SampleConfig is called upon when auto-creating a configuration
func (n *Icecast) SampleConfig() string {
	return sampleConfig
}

// Description is used to describe the input module
func (n *Icecast) Description() string {
	return "Read listeners from an Icecast instance per mount"
}

// Gather will fetch the metrics from Icecast
func (n *Icecast) Gather(acc telegraf.Accumulator) error {
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	var outerr error
	var errch = make(chan error)

	for _, s := range n.Servers {
		server := s

		// Default admin listmounts page of Icecast
		adminPageURL := "/admin/listmounts"

		// Parsing the URL to see if it's ok
		addr, err := url.Parse(server.URL)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", server.URL, err)
		}
		addr.Path = path.Join(addr.Path, adminPageURL)
		go func(addr *url.URL) {
			errch <- n.gatherURL(addr, server, acc)
		}(addr)
	}

	// Drain channel, waiting for all requests to finish and save last error.
	for range n.Servers {
		if err := <-errch; err != nil {
			outerr = err
		}
	}

	return outerr
}

func (n *Icecast) gatherURL(
	addr *url.URL,
	s server,
	acc telegraf.Accumulator,
) error {
	var listmounts Listmounts
	var total int32

	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	req, err := http.NewRequest("GET", addr.String(), nil)
	if err != nil {
		return fmt.Errorf("Error on new request to %s : %s", addr.String(), err)
	}

	if len(s.Username) != 0 && len(s.Password) != 0 {
		req.SetBasicAuth(s.Username, s.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error on request to %s : %s", addr.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}

	// Processing XML response
	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		if err := xml.Unmarshal(body, &listmounts); err != nil {
			return fmt.Errorf("XML error: %s", err)
		}
	} else {
		return fmt.Errorf("Read error: %s", err)
	}

	var host string
	// Setting alias if available
	if len(s.Alias) != 0 {
		host = s.Alias
	} else {
		host = addr.Hostname()
	}

	// Run trough each mountpoint
	for _, sources := range listmounts.Sources {
		var mountname string

		if n.Slash == false {
			mountname = strings.Trim(sources.Mount, "/")
		} else {
			mountname = sources.Mount
		}

		tags := map[string]string{
			"host":  host,
			"mount": mountname,
		}
		fields := map[string]interface{}{
			"listeners": sources.Listeners,
		}
		acc.AddFields("icecast", fields, tags)
		total += sources.Listeners
	}

	// Report total listeners as well
	tagsTotal := map[string]string{
		"host":  host,
		"mount": "Total",
	}
	fieldsTotal := map[string]interface{}{
		"listeners": total,
	}
	acc.AddFields("icecast", fieldsTotal, tagsTotal)

	return nil
}

func init() {
	inputs.Add("icecast", func() telegraf.Input {
		return &Icecast{}
	})
}
