package icecast

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

type Icecast struct {
	Urls               []string          `toml:"urls"`
	ResponseTimeout    internal.Duration `toml:"response_timeout"`
	Username           string            `toml:"username"`
	Password           string            `toml:"password"`
	Slash              bool              `toml:"slash"`
	SSLCA              string            `toml:"ssl_ca"`   // Path to CA file
	SSLCert            string            `toml:"ssl_cert"` // Path to host cert file
	SSLKey             string            `toml:"ssl_key"`  // Path to cert key file
	InsecureSkipVerify bool              // Use SSL but skip chain & host verification
}

var sampleConfig = `
  ## Specify the IP adress to where the '/admin/listmounts' can be found. You can include port if needed.
  ## If you'd like to report under an alias, use ; (e.g. https://localhost;Server 1)
  ## You can use multiple hosts who use the same login credentials by dividing with , (e.g. "http://localhost","https://127.0.0.1")
  urls = ["http://localhost"]

  ## Timeout to the complete conection and reponse time in seconds. Default (5 seconds)
  # response_timeout = "25s"

  ## The username/password combination needed to read the listmounts page.
  ## These must be equal to the admin login details specified in your Icecast configuration
  username = "admin"
  password = "hackme"

  ## Include the slash in mountpoint names or not
  slash = false

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
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
	if len(n.Urls) == 0 {
		n.Urls = []string{"http://localhost"}
	}
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	var outerr error
	var errch = make(chan error)

	for _, u := range n.Urls {
		// Default admin listmounts page of Icecast
		adminPageURL := "/admin/listmounts"

		// Check to see if there is an alias
		var alias string
		if strings.Contains(u, ";") {
			urlAlais := strings.Split(u, ";")
			alias = urlAlais[1]
			u = urlAlais[0]
		}

		// Check to see if the user isn't adding a / at the end
		if u[len(u)-1:] == "/" {
			if last := len(u) - 1; last >= 0 && u[last] == '/' {
				u = u[:last]
			}
		}

		// Parsing the URL to see if it's ok
		hosturl := u + adminPageURL
		addr, err := url.Parse(hosturl)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}
		tempURL := u

		go func(addr *url.URL) {
			errch <- n.gatherURL(addr, tempURL, alias, acc)
		}(addr)
	}

	// Drain channel, waiting for all requests to finish and save last error.
	for range n.Urls {
		if err := <-errch; err != nil {
			outerr = err
		}
	}

	return outerr
}

func (n *Icecast) gatherURL(
	addr *url.URL,
	host string,
	alias string,
	acc telegraf.Accumulator,
) error {
	var tr *http.Transport
	var listmounts Listmounts
	var total int32

	if addr.Scheme == "https" {
		tlsCfg, err := internal.GetTLSConfig(
			n.SSLCert, n.SSLKey, n.SSLCA, n.InsecureSkipVerify)
		if err != nil {
			return err
		}
		tr = &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
			TLSClientConfig:       tlsCfg,
		}
	} else {
		tr = &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   n.ResponseTimeout.Duration,
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

	// Setting alias if availible
	if len(alias) != 0 {
		host = alias
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
