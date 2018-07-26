package apache

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Apache struct {
	Urls            []string
	Username        string
	Password        string
	ResponseTimeout internal.Duration
	tls.ClientConfig

	client *http.Client
}

var sampleConfig = `
  ## An array of URLs to gather from, must be directed at the machine
  ## readable version of the mod_status page including the auto query string.
  ## Default is "http://localhost/server-status?auto".
  urls = ["http://localhost/server-status?auto"]

  ## Credentials for basic HTTP authentication.
  # username = "myuser"
  # password = "mypassword"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (n *Apache) SampleConfig() string {
	return sampleConfig
}

func (n *Apache) Description() string {
	return "Read Apache status information (mod_status)"
}

func (n *Apache) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(n.Urls) == 0 {
		n.Urls = []string{"http://localhost/server-status?auto"}
	}
	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	if n.client == nil {
		client, err := n.createHttpClient()
		if err != nil {
			return err
		}
		n.client = client
	}

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(n.gatherUrl(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (n *Apache) createHttpClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (n *Apache) gatherUrl(addr *url.URL, acc telegraf.Accumulator) error {
	req, err := http.NewRequest("GET", addr.String(), nil)
	if err != nil {
		return fmt.Errorf("error on new request to %s : %s\n", addr.String(), err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := n.client.Do(req)
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
