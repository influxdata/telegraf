package apache

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Apache struct {
	Urls            []string
	Username        string
	Password        string
	ResponseTimeout internal.Duration
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool
}

var sampleConfig = `
  ## An array of Apache status URI to gather stats.
  ## Default is "http://localhost/server-status?auto".
  urls = ["http://localhost/server-status?auto"]
  ## user credentials for basic HTTP authentication
  username = "myuser"
  password = "mypassword"

  ## Timeout to the complete conection and reponse time in seconds
  response_timeout = "25s" ## default to 5 seconds

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

func (n *Apache) SampleConfig() string {
	return sampleConfig
}

func (n *Apache) Description() string {
	return "Read Apache status information (mod_status)"
}

func (n *Apache) Gather(acc telegraf.Accumulator) error {
	if len(n.Urls) == 0 {
		n.Urls = []string{"http://localhost/server-status?auto"}
	}
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	var outerr error
	var errch = make(chan error)

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}

		go func(addr *url.URL) {
			errch <- n.gatherUrl(addr, acc)
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

func (n *Apache) gatherUrl(addr *url.URL, acc telegraf.Accumulator) error {

	var tr *http.Transport

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
		return fmt.Errorf("error on new request to %s : %s\n", addr.String(), err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", addr.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}

	tags := getTags(addr)

	sc := bufio.NewScanner(resp.Body)
	fields := make(map[string]interface{})
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key, part := strings.Replace(parts[0], " ", "", -1), strings.TrimSpace(parts[1])

			switch key {
			case "Scoreboard":
				for field, value := range n.gatherScores(part) {
					fields[field] = value
				}
			default:
				value, err := strconv.ParseFloat(part, 64)
				if err != nil {
					continue
				}
				fields[key] = value
			}
		}
	}
	acc.AddFields("apache", fields, tags)

	return nil
}

func (n *Apache) gatherScores(data string) map[string]interface{} {
	var waiting, open int = 0, 0
	var S, R, W, K, D, C, L, G, I int = 0, 0, 0, 0, 0, 0, 0, 0, 0

	for _, s := range strings.Split(data, "") {

		switch s {
		case "_":
			waiting++
		case "S":
			S++
		case "R":
			R++
		case "W":
			W++
		case "K":
			K++
		case "D":
			D++
		case "C":
			C++
		case "L":
			L++
		case "G":
			G++
		case "I":
			I++
		case ".":
			open++
		}
	}

	fields := map[string]interface{}{
		"scboard_waiting":      float64(waiting),
		"scboard_starting":     float64(S),
		"scboard_reading":      float64(R),
		"scboard_sending":      float64(W),
		"scboard_keepalive":    float64(K),
		"scboard_dnslookup":    float64(D),
		"scboard_closing":      float64(C),
		"scboard_logging":      float64(L),
		"scboard_finishing":    float64(G),
		"scboard_idle_cleanup": float64(I),
		"scboard_open":         float64(open),
	}
	return fields
}

// Get tag(s) for the apache plugin
func getTags(addr *url.URL) map[string]string {
	h := addr.Host
	host, port, err := net.SplitHostPort(h)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}
	return map[string]string{"server": host, "port": port}
}

func init() {
	inputs.Add("apache", func() telegraf.Input {
		return &Apache{}
	})
}
