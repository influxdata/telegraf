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
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	PF_POOL                 = "pool"
	PF_PROCESS_MANAGER      = "process manager"
	PF_START_SINCE          = "start since"
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
	Urls    []string
	Timeout internal.Duration
	tls.ClientConfig

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
  ## Example of multiple gathering from local socket and remote host
  ## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
  urls = ["http://localhost/status"]

  ## Duration allowed to complete HTTP requests.
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (p *phpfpm) SampleConfig() string {
	return sampleConfig
}

func (p *phpfpm) Description() string {
	return "Read metrics of phpfpm, via HTTP status page or socket"
}

func (p *phpfpm) Init() error {
	tlsCfg, err := p.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	p.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: p.Timeout.Duration,
	}
	return nil
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (p *phpfpm) Gather(acc telegraf.Accumulator) error {
	if len(p.Urls) == 0 {
		return p.gatherServer("http://127.0.0.1/status", acc)
	}

	var wg sync.WaitGroup

	urls, err := expandUrls(p.Urls)
	if err != nil {
		return err
	}

	for _, serv := range urls {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			acc.AddError(p.gatherServer(serv, acc))
		}(serv)
	}

	wg.Wait()

	return nil
}

// Request status page to get stat raw data and import it
func (p *phpfpm) gatherServer(addr string, acc telegraf.Accumulator) error {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return p.gatherHttp(addr, acc)
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
		socketPath, statusPath = unixSocketPaths(addr)
		if statusPath == "" {
			statusPath = "status"
		}
		fcgi, err = newFcgiClient("unix", socketPath)
	}

	if err != nil {
		return err
	}

	return p.gatherFcgi(fcgi, statusPath, acc, addr)
}

// Gather stat using fcgi protocol
func (p *phpfpm) gatherFcgi(fcgi *conn, statusPath string, acc telegraf.Accumulator, addr string) error {
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
func (p *phpfpm) gatherHttp(addr string, acc telegraf.Accumulator) error {
	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("unable parse server address '%s': %v", addr, err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path), nil)
	if err != nil {
		return fmt.Errorf("unable to create new request '%s': %v", addr, err)
	}

	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect to phpfpm status page '%s': %v", addr, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to get valid stat result from '%s': %v", addr, err)
	}

	importMetric(res.Body, acc, addr)
	return nil
}

// Import stat data into Telegraf system
func importMetric(r io.Reader, acc telegraf.Accumulator, addr string) poolStat {
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
		case PF_START_SINCE,
			PF_ACCEPTED_CONN,
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

	return stats
}

func expandUrls(urls []string) ([]string, error) {
	addrs := make([]string, 0, len(urls))
	for _, url := range urls {
		if isNetworkURL(url) {
			addrs = append(addrs, url)
			continue
		}
		paths, err := globUnixSocket(url)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, paths...)
	}
	return addrs, nil
}

func globUnixSocket(url string) ([]string, error) {
	pattern, status := unixSocketPaths(url)
	glob, err := globpath.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("could not compile glob %q: %v", pattern, err)
	}
	paths := glob.Match()
	if len(paths) == 0 {
		if _, err := os.Stat(paths[0]); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("Socket doesn't exist  '%s': %s", pattern, err)
			}
			return nil, err
		}
		return nil, nil
	}

	addrs := make([]string, 0, len(paths))

	for _, path := range paths {
		if status != "" {
			path = path + ":" + status
		}
		addrs = append(addrs, path)
	}

	return addrs, nil
}

func unixSocketPaths(addr string) (string, string) {
	var socketPath, statusPath string

	socketAddr := strings.Split(addr, ":")
	if len(socketAddr) >= 2 {
		socketPath = socketAddr[0]
		statusPath = socketAddr[1]
	} else {
		socketPath = socketAddr[0]
		statusPath = ""
	}

	return socketPath, statusPath
}

func isNetworkURL(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") || strings.HasPrefix(addr, "fcgi://") || strings.HasPrefix(addr, "cgi://")
}

func init() {
	inputs.Add("phpfpm", func() telegraf.Input {
		return &phpfpm{}
	})
}
