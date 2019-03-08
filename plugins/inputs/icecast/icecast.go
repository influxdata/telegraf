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
	URLs            [][]string        `toml:"urls"`
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	Username        string            `toml:"username"`
	Password        string            `toml:"password"`
	Slash           bool              `toml:"slash"`
}

var sampleConfig = `
  ## Specify the IP adress/hostname to where the '/admin/listmounts' can be found. You can include port if needed.
  ## If you'd like to report under an alias, specify it in the second field (optional)
  ## You can use multiple hosts who use the same login credentials
  ## For example:
  ## urls = [ [ "http://localhost", "Server 1" ],
  			  [ "http://example.org" ] ]
  urls = [ [ "http://localhost", "Server 1" ] ]

  ## Timeout to the complete conection and reponse time in seconds. Default (5 seconds)
  # response_timeout = "25s"

  ## The username/password combination needed to read the listmounts page.
  ## These must be equal to the admin login details specified in your Icecast configuration
  username = "admin"
  password = "hackme"

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
	if len(n.URLs) == 0 {
		return fmt.Errorf("No hostname/IP given")
	}
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	var outerr error
	var errch = make(chan error)

	for _, u := range n.URLs {
		// Default admin listmounts page of Icecast
		adminPageURL := "/admin/listmounts"

		// Check to see if there is an alias
		var alias string
		if len(u) > 1 {
			alias = u[1]
		}

		// Parsing the URL to see if it's ok
		addr, err := url.Parse(u[0])
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u[0], err)
		}
		addr.Path = path.Join(addr.Path, adminPageURL)

		go func(addr *url.URL) {
			errch <- n.gatherURL(addr, alias, acc)
		}(addr)
	}

	// Drain channel, waiting for all requests to finish and save last error.
	for range n.URLs {
		if err := <-errch; err != nil {
			outerr = err
		}
	}

	return outerr
}

func (n *Icecast) gatherURL(
	addr *url.URL,
	alias string,
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
		return fmt.Errorf("error on new request to %s : %s", addr.String(), err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s", addr.String(), err)
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
	if len(alias) != 0 {
		host = alias
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
