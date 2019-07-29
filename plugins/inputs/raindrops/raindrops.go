package raindrops

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
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Raindrops struct {
	Urls        []string
	http_client *http.Client
}

var sampleConfig = `
  ## An array of raindrops middleware URI to gather stats.
  urls = ["http://localhost:8080/_raindrops"]
`

func (r *Raindrops) SampleConfig() string {
	return sampleConfig
}

func (r *Raindrops) Description() string {
	return "Read raindrops stats (raindrops - real-time stats for preforking Rack servers)"
}

func (r *Raindrops) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, u := range r.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(r.gatherUrl(addr, acc))
		}(addr)
	}

	wg.Wait()

	return nil
}

func (r *Raindrops) gatherUrl(addr *url.URL, acc telegraf.Accumulator) error {
	resp, err := r.http_client.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}
	buf := bufio.NewReader(resp.Body)

	// Calling
	_, err = buf.ReadString(':')
	if err != nil {
		return err
	}
	line, err := buf.ReadString('\n')
	if err != nil {
		return err
	}
	calling, err := strconv.ParseUint(strings.TrimSpace(line), 10, 64)
	if err != nil {
		return err
	}

	// Writing
	_, err = buf.ReadString(':')
	if err != nil {
		return err
	}
	line, err = buf.ReadString('\n')
	if err != nil {
		return err
	}
	writing, err := strconv.ParseUint(strings.TrimSpace(line), 10, 64)
	if err != nil {
		return err
	}
	tags := r.getTags(addr)
	fields := map[string]interface{}{
		"calling": calling,
		"writing": writing,
	}
	acc.AddFields("raindrops", fields, tags)

	iterate := true
	var queued_line_str string
	var active_line_str string
	var active_err error
	var queued_err error

	for iterate {
		// Listen
		var tags map[string]string

		lis := map[string]interface{}{
			"active": 0,
			"queued": 0,
		}
		active_line_str, active_err = buf.ReadString('\n')
		if active_err != nil {
			iterate = false
			break
		}
		if strings.Compare(active_line_str, "\n") == 0 {
			break
		}
		queued_line_str, queued_err = buf.ReadString('\n')
		if queued_err != nil {
			iterate = false
		}
		active_line := strings.Split(active_line_str, " ")
		listen_name := active_line[0]

		active, err := strconv.ParseUint(strings.TrimSpace(active_line[2]), 10, 64)
		if err != nil {
			active = 0
		}
		lis["active"] = active

		queued_line := strings.Split(queued_line_str, " ")
		queued, err := strconv.ParseUint(strings.TrimSpace(queued_line[2]), 10, 64)
		if err != nil {
			queued = 0
		}
		lis["queued"] = queued
		if strings.Contains(listen_name, ":") {
			listener := strings.Split(listen_name, ":")
			tags = map[string]string{
				"ip":   listener[0],
				"port": listener[1],
			}

		} else {
			tags = map[string]string{
				"socket": listen_name,
			}
		}
		acc.AddFields("raindrops_listen", lis, tags)
	}
	return nil
}

// Get tag(s) for the raindrops calling/writing plugin
func (r *Raindrops) getTags(addr *url.URL) map[string]string {
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
	inputs.Add("raindrops", func() telegraf.Input {
		return &Raindrops{http_client: &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(3 * time.Second),
			},
			Timeout: time.Duration(4 * time.Second),
		}}
	})
}
