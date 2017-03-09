package icecast

import (
	"fmt"
	"net/http"
	"encoding/xml"
	"io/ioutil"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Icecast is an icecast plugin
type Icecast struct {
	Host				string
	Username		string
	Password		string
	Alias				string
}

// SourceListmounts holds single listener elements from the Icecast XML
type SourceListmounts struct {
	Mount					string		`xml:"mount,attr"`
	Listeners			int32			`xml:"listeners"`
	Connected			int32			`xml:"connected"`
	ContentType		string		`xml:"content-type"`
}

// Listmounts main structure of the icecast XML
type Listmounts struct {
	Sources []SourceListmounts `xml:"source"`
}

var sampleConfig = `
  ## Specify the IP adress to where the 'admin/listmounts' can be found. You can include port if needed.
  host = "localhost"

  ## The username/password combination needed to read the listmounts page.
  ## These must be equal to the admin login details specified in your Icecast configuration
  username = "admin"
  password = "hackme"

  ## If you wish your host name to be different then the one specified under host, you can change it here
  alias = ""
`

// Description returns description of Icecast plugin
func (ice *Icecast) Description() string {
	return "Read listeners from an Icecast instance per mount"
}

// The list of metrics that should be sent
var sendMetrics = []string{
	"listeners",
	"host",
	"mount",
}

// SampleConfig returns sample configuration message
func (ice *Icecast) SampleConfig() string {
	return sampleConfig
}

// Gather reads stats from all configured servers mount stats
func (ice *Icecast) Gather(acc telegraf.Accumulator) error {
  errChan := errchan.New(len(ice.Host))

  // Check to see if the needed fields are filled in, and if so, connect.
  if len(ice.Host) != 0 && len(ice.Username) != 0 && len(ice.Password) != 0 {
    errChan.C <- ice.gatherServer(ice.Host, ice.Username , ice.Password, ice.Alias, acc)
  }

	return errChan.Error()
}

// Main gather function
func (ice *Icecast) gatherServer(
	host string,
	username string,
  password string,
  alias string,
	acc telegraf.Accumulator,
) error {
	var err error
  var listmounts Listmounts
	var total int32

  // Create HTTP client to fetch the listmounts stats
  httpClientIcecast := &http.Client{}
  req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/admin/listmounts", host), nil)
  if err != nil {
    fmt.Errorf("HTTP request error: %s", err)
  }
  req.SetBasicAuth(username, password)

  // Starting the HTTP request
  icecastResponse, err := httpClientIcecast.Do(req)
  if icecastResponse == nil {
    fmt.Errorf("No response: %s", err)
  }
  if err != nil {
    fmt.Errorf("HTTP request error: %s", err)
  }
  defer icecastResponse.Body.Close()

  // Processing XML response
  if body, err := ioutil.ReadAll(icecastResponse.Body); err == nil {
    if err := xml.Unmarshal(body, &listmounts); err != nil {
      fmt.Errorf("XML error: %s", err)
    }
  } else {
    fmt.Errorf("Read error: %s", err)
  }

	// Settings alias as host if one is given
	if len(alias) != 0 {
		host = alias
	}

  // Run trough each mountpoint
  for _, sources := range listmounts.Sources {

    tags := map[string]string{
      "host": host,
      "mount": strings.Trim(sources.Mount,"/"),
    }
    fields := map[string]interface{}{
      "listeners":   sources.Listeners,
    }
		acc.AddFields("icecast", fields, tags)
    total += sources.Listeners
  }

  // Report total listeners as well
  tags_total := map[string]string{
    "host": host,
    "mount": "Total",
  }
  fields_total := map[string]interface{}{
    "listeners":   total,
  }
	acc.AddFields("icecast", fields_total, tags_total)

	return nil
}

func init() {
	inputs.Add("icecast", func() telegraf.Input {
		return &Icecast{}
	})
}
