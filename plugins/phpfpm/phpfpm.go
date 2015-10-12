package phpfpm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdb/telegraf/plugins"
)

const (
	PF_POOL                 = "pool"
	PF_PROCESS_MANAGER      = "process manager"
	PF_ACCEPTED_CONN        = "accepted conn"
	PF_LISTEN_QUEUE         = "listen queue"
	PF_MAX_LISTEN_QUEUE     = "max listen queue"
	PF_LISTEN_QUEUE_LEN     = "listen queue len"
	PF_IDLE_PROCESSES       = "idle processes"
	PF_ACTIVE_PROCESSES     = "active processes"
	PF_TOTAL_PROCESSES      = "total processes"
	PF_MAX_ACTIVE_PROCESSES = "max active processes"
	PF_MAX_CHILDREN_REACHED = "max children reached"
	PF_SLOW_REQUESTS        = "slow requests"
)

type metric map[string]int64
type poolStat map[string]metric

type phpfpm struct {
	Urls []string

	client *http.Client
}

var sampleConfig = `
	# An array of addresses to gather stats about. Specify an ip or hostname
	# with optional port and path.
	#
	# Plugin can be configured in two modes (both can be used):
	#   - http: the URL must start with http:// or https://, ex:
	#		"http://localhost/status"
	#		"http://192.168.130.1/status?full"
	#   - unixsocket: path to fpm socket, ex:
	#		"/var/run/php5-fpm.sock"
	#		"192.168.10.10:/var/run/php5-fpm-www2.sock"
	#
	# If no servers are specified, then default to 127.0.0.1/server-status
	urls = ["http://localhost/status"]
`

func (r *phpfpm) SampleConfig() string {
	return sampleConfig
}

func (r *phpfpm) Description() string {
	return "Read metrics of phpfpm, via HTTP status page or socket(pending)"
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *phpfpm) Gather(acc plugins.Accumulator) error {
	if len(g.Urls) == 0 {
		return g.gatherServer("http://127.0.0.1/status", acc)
	}

	var wg sync.WaitGroup

	var outerr error

	for _, serv := range g.Urls {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = g.gatherServer(serv, acc)
		}(serv)
	}

	wg.Wait()

	return outerr
}

// Request status page to get stat raw data
func (g *phpfpm) gatherServer(addr string, acc plugins.Accumulator) error {
	if g.client == nil {

		client := &http.Client{}
		g.client = client
	}

	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		u, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("Unable parse server address '%s': %s", addr, err)
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s%s", u.Scheme,
			u.Host, u.Path), nil)
		res, err := g.client.Do(req)
		if err != nil {
			return fmt.Errorf("Unable to connect to phpfpm status page '%s': %v",
				addr, err)
		}

		if res.StatusCode != 200 {
			return fmt.Errorf("Unable to get valid stat result from '%s': %v",
				addr, err)
		}

		importMetric(res.Body, acc, u.Host)
	} else {
		socketAddr := strings.Split(addr, ":")

		fcgi, _ := NewClient("unix", socketAddr[1])
		resOut, resErr, err := fcgi.Request(map[string]string{
			"SCRIPT_NAME":     "/status",
			"SCRIPT_FILENAME": "status",
			"REQUEST_METHOD":  "GET",
		}, "")

		if len(resErr) == 0 && err == nil {
			importMetric(bytes.NewReader(resOut), acc, socketAddr[0])
		}

	}

	return nil
}

// Import HTTP stat data into Telegraf system
func importMetric(r io.Reader, acc plugins.Accumulator, host string) (poolStat, error) {
	stats := make(poolStat)
	var currentPool string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		statLine := scanner.Text()
		keyvalue := strings.Split(statLine, ":")

		if len(keyvalue) < 2 {
			continue
		}
		fieldName := strings.Trim(keyvalue[0], " ")
		// We start to gather data for a new pool here
		if fieldName == PF_POOL {
			currentPool = strings.Trim(keyvalue[1], " ")
			stats[currentPool] = make(metric)
			continue
		}

		// Start to parse metric for current pool
		switch fieldName {
		case PF_ACCEPTED_CONN,
			PF_LISTEN_QUEUE,
			PF_MAX_LISTEN_QUEUE,
			PF_LISTEN_QUEUE_LEN,
			PF_IDLE_PROCESSES,
			PF_ACTIVE_PROCESSES,
			PF_TOTAL_PROCESSES,
			PF_MAX_ACTIVE_PROCESSES,
			PF_MAX_CHILDREN_REACHED,
			PF_SLOW_REQUESTS:
			fieldValue, err := strconv.ParseInt(strings.Trim(keyvalue[1], " "), 10, 64)
			if err == nil {
				stats[currentPool][fieldName] = fieldValue
			}
		}
	}

	// Finally, we push the pool metric
	for pool := range stats {
		tags := map[string]string{
			"url":  host,
			"pool": pool,
		}
		for k, v := range stats[pool] {
			acc.Add(strings.Replace(k, " ", "_", -1), v, tags)
		}
	}

	return stats, nil
}

func init() {
	plugins.Add("phpfpm", func() plugins.Plugin {
		return &phpfpm{}
	})
}
