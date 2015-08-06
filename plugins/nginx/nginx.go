package nginx

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

type Nginx struct {
	Urls []string
}

var sampleConfig = `
# An array of Nginx stub_status URI to gather stats.
urls = ["localhost/status"]`

func (n *Nginx) SampleConfig() string {
	return sampleConfig
}

func (n *Nginx) Description() string {
	return "Read Nginx's basic status information (ngx_http_stub_status_module)"
}

func (n *Nginx) Gather(acc plugins.Accumulator) error {
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

func (n *Nginx) gatherUrl(addr *url.URL, acc plugins.Accumulator) error {
	resp, err := client.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}
	r := bufio.NewReader(resp.Body)

	// Active connections
	_, err = r.ReadString(':')
	if err != nil {
		return err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	active, err := strconv.ParseUint(strings.TrimSpace(line), 10, 64)
	if err != nil {
		return err
	}

	// Server accepts handled requests
	_, err = r.ReadString('\n')
	if err != nil {
		return err
	}
	line, err = r.ReadString('\n')
	if err != nil {
		return err
	}
	data := strings.SplitN(strings.TrimSpace(line), " ", 3)
	accepts, err := strconv.ParseUint(data[0], 10, 64)
	if err != nil {
		return err
	}
	handled, err := strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		return err
	}
	requests, err := strconv.ParseUint(data[2], 10, 64)
	if err != nil {
		return err
	}

	// Reading/Writing/Waiting
	line, err = r.ReadString('\n')
	if err != nil {
		return err
	}
	data = strings.SplitN(strings.TrimSpace(line), " ", 6)
	reading, err := strconv.ParseUint(data[1], 10, 64)
	if err != nil {
		return err
	}
	writing, err := strconv.ParseUint(data[3], 10, 64)
	if err != nil {
		return err
	}
	waiting, err := strconv.ParseUint(data[5], 10, 64)
	if err != nil {
		return err
	}

	host, _, _ := net.SplitHostPort(addr.Host)
	tags := map[string]string{"server": host}
	acc.Add("active", active, tags)
	acc.Add("accepts", accepts, tags)
	acc.Add("handled", handled, tags)
	acc.Add("requests", requests, tags)
	acc.Add("reading", reading, tags)
	acc.Add("writing", writing, tags)
	acc.Add("waiting", waiting, tags)

	return nil
}

func init() {
	plugins.Add("nginx", func() plugins.Plugin {
		return &Nginx{}
	})
}
