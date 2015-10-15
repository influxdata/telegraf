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

	"github.com/influxdb/telegraf/plugins"
)

type Apache struct {
	Urls []string
}

var sampleConfig = `
  # An array of Apache status URI to gather stats.
  urls = ["http://localhost/server-status?auto"]
`

func (n *Apache) SampleConfig() string {
	return sampleConfig
}

func (n *Apache) Description() string {
	return "Read Apache status information (mod_status)"
}

func (n *Apache) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup
	var outerr error

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			outerr = n.gatherUrl(addr, acc)
		}(addr)
	}

	wg.Wait()

	return outerr
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{Transport: tr}

func (n *Apache) gatherUrl(addr *url.URL, acc plugins.Accumulator) error {
	resp, err := client.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}

	tags := getTags(addr)

	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, ":") {

			parts := strings.SplitN(line, ":", 2)
			key, part := strings.Replace(parts[0], " ", "", -1), strings.TrimSpace(parts[1])

			switch key {

			case "Scoreboard":
				n.gatherScores(part, acc, tags)
			default:
				value, err := strconv.ParseFloat(part, 32)
				if err != nil {
					continue
				}
				acc.Add(key, value, tags)
			}
		}
	}

	return nil
}

func (n *Apache) gatherScores(data string, acc plugins.Accumulator, tags map[string]string) {

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

	acc.Add("scboard_waiting", float64(waiting), tags)
	acc.Add("scboard_starting", float64(S), tags)
	acc.Add("scboard_reading", float64(R), tags)
	acc.Add("scboard_sending", float64(W), tags)
	acc.Add("scboard_keepalive", float64(K), tags)
	acc.Add("scboard_dnslookup", float64(D), tags)
	acc.Add("scboard_closing", float64(C), tags)
	acc.Add("scboard_logging", float64(L), tags)
	acc.Add("scboard_finishing", float64(G), tags)
	acc.Add("scboard_idle_cleanup", float64(I), tags)
	acc.Add("scboard_open", float64(open), tags)
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
	plugins.Add("apache", func() plugins.Plugin {
		return &Apache{}
	})
}
