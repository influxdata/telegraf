package phpfpm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
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
  ## An array of addresses to gather stats about. Specify an ip or hostname
  ## with optional port and path
  ##
  ## Plugin can be configured in three modes (either can be used):
  ##   - http: the URL must start with http:// or https://, ie:
  ##       "http://localhost/status"
  ##       "http://192.168.130.1/status?full"
  ##
  ##   - unixsocket: path to fpm socket, ie:
  ##       "/var/run/php5-fpm.sock"
  ##      or using a custom fpm status path:
  ##       "/var/run/php5-fpm.sock:fpm-custom-status-path"
  ##
  ##   - fcgi: the URL must start with fcgi:// or cgi://, and port must be present, ie:
  ##       "fcgi://10.0.0.12:9000/status"
  ##       "cgi://10.0.10.12:9001/status"
  ##
  ## Example of multiple gathering from local socket and remove host
  ## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
  urls = ["http://localhost/status"]
`

func (r *phpfpm) SampleConfig() string {
	return sampleConfig
}

func (r *phpfpm) Description() string {
	return "Read metrics of phpfpm, via HTTP status page or socket"
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *phpfpm) Gather(acc telegraf.Accumulator) error {
	if len(g.Urls) == 0 {
		return g.gatherServer("http://127.0.0.1/status", acc)
	}

	var wg sync.WaitGroup

	for _, serv := range g.Urls {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			acc.AddError(g.gatherServer(serv, acc))
		}(serv)
	}

	wg.Wait()

	return nil
}

// Request status page to get stat raw data and import it
func (g *phpfpm) gatherServer(addr string, acc telegraf.Accumulator) error {
	if g.client == nil {
		client := &http.Client{}
		g.client = client
	}

	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return g.gatherHttp(addr, acc)
	}

	var (
		fcgi       *conn
		socketPath string
		statusPath string
	)

	var err error
	if strings.HasPrefix(addr, "fcgi://") || strings.HasPrefix(addr, "cgi://") {
		u, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("Unable parse server address '%s': %s", addr, err)
		}
		socketAddr := strings.Split(u.Host, ":")
		fcgiIp := socketAddr[0]
		fcgiPort, _ := strconv.Atoi(socketAddr[1])
		fcgi, err = newFcgiClient(fcgiIp, fcgiPort)
		if err != nil {
			return err
		}
		if len(u.Path) > 1 {
			statusPath = strings.Trim(u.Path, "/")
		} else {
			statusPath = "status"
		}
	} else {
		socketAddr := strings.Split(addr, ":")
		if len(socketAddr) >= 2 {
			socketPath = socketAddr[0]
			statusPath = socketAddr[1]
		} else {
			socketPath = socketAddr[0]
			statusPath = "status"
		}

		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			return fmt.Errorf("Socket doesn't exist  '%s': %s", socketPath, err)
		}
		fcgi, err = newFcgiClient("unix", socketPath)
	}

	if err != nil {
		return err
	}

	return g.gatherFcgi(fcgi, statusPath, acc, addr)
}

// Gather stat using fcgi protocol
func (g *phpfpm) gatherFcgi(fcgi *conn, statusPath string, acc telegraf.Accumulator, addr string) error {
	fpmOutput, fpmErr, err := fcgi.Request(map[string]string{
		"SCRIPT_NAME":     "/" + statusPath,
		"SCRIPT_FILENAME": statusPath,
		"REQUEST_METHOD":  "GET",
		"CONTENT_LENGTH":  "0",
		"SERVER_PROTOCOL": "HTTP/1.0",
		"SERVER_SOFTWARE": "go / fcgiclient ",
		"REMOTE_ADDR":     "127.0.0.1",
	}, "/"+statusPath)

	if len(fpmErr) == 0 && err == nil {
		importMetric(bytes.NewReader(fpmOutput), acc, addr)
		return nil
	} else {
		return fmt.Errorf("Unable parse phpfpm status. Error: %v %v", string(fpmErr), err)
	}
}

// Gather stat using http protocol
func (g *phpfpm) gatherHttp(addr string, acc telegraf.Accumulator) error {
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
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Unable to get valid stat result from '%s': %v",
			addr, err)
	}

	importMetric(res.Body, acc, addr)
	return nil
}

// Import stat data into Telegraf system
func importMetric(r io.Reader, acc telegraf.Accumulator, addr string) (poolStat, error) {
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
			"pool": pool,
			"url":  addr,
		}
		fields := make(map[string]interface{})
		for k, v := range stats[pool] {
			fields[strings.Replace(k, " ", "_", -1)] = v
		}
		acc.AddFields("phpfpm", fields, tags)
	}

	return stats, nil
}

func init() {
	inputs.Add("phpfpm", func() telegraf.Input {
		return &phpfpm{}
	})
}
