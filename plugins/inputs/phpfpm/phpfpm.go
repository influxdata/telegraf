//go:generate ../../../tools/readme_config_includer/generator
package phpfpm

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	PfPool               = "pool"
	PfProcessManager     = "process manager"
	PfStartSince         = "start since"
	PfAcceptedConn       = "accepted conn"
	PfListenQueue        = "listen queue"
	PfMaxListenQueue     = "max listen queue"
	PfListenQueueLen     = "listen queue len"
	PfIdleProcesses      = "idle processes"
	PfActiveProcesses    = "active processes"
	PfTotalProcesses     = "total processes"
	PfMaxActiveProcesses = "max active processes"
	PfMaxChildrenReached = "max children reached"
	PfSlowRequests       = "slow requests"
)

type JSONMetrics struct {
	Pool               string `json:"pool"`
	ProcessManager     string `json:"process manager"`
	StartTime          int    `json:"start time"`
	StartSince         int    `json:"start since"`
	AcceptedConn       int    `json:"accepted conn"`
	ListenQueue        int    `json:"listen queue"`
	MaxListenQueue     int    `json:"max listen queue"`
	ListenQueueLen     int    `json:"listen queue len"`
	IdleProcesses      int    `json:"idle processes"`
	ActiveProcesses    int    `json:"active processes"`
	TotalProcesses     int    `json:"total processes"`
	MaxActiveProcesses int    `json:"max active processes"`
	MaxChildrenReached int    `json:"max children reached"`
	SlowRequests       int    `json:"slow requests"`
	Processes          []struct {
		Pid               int     `json:"pid"`
		State             string  `json:"state"`
		StartTime         int     `json:"start time"`
		StartSince        int     `json:"start since"`
		Requests          int     `json:"requests"`
		RequestDuration   int     `json:"request duration"`
		RequestMethod     string  `json:"request method"`
		RequestURI        string  `json:"request uri"`
		ContentLength     int     `json:"content length"`
		User              string  `json:"user"`
		Script            string  `json:"script"`
		LastRequestCPU    float64 `json:"last request cpu"`
		LastRequestMemory float64 `json:"last request memory"`
	} `json:"processes"`
}

type metric map[string]int64
type poolStat map[string]metric

type phpfpm struct {
	Format  string          `toml:"format"`
	Timeout config.Duration `toml:"timeout"`
	Urls    []string        `toml:"urls"`

	tls.ClientConfig
	client *http.Client
	Log    telegraf.Logger
}

func (*phpfpm) SampleConfig() string {
	return sampleConfig
}

func (p *phpfpm) Init() error {
	tlsCfg, err := p.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	switch p.Format {
	case "":
		p.Format = "status"
	case "status", "json":
		// both valid
	default:
		return fmt.Errorf("invalid format: %s", p.Format)
	}

	p.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(p.Timeout),
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
		return p.gatherHTTP(addr, acc)
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
			return fmt.Errorf("unable parse server address %q: %w", addr, err)
		}
		socketAddr := strings.Split(u.Host, ":")
		if len(socketAddr) < 2 {
			return fmt.Errorf("url does not follow required 'address:port' format: %s", u.Host)
		}
		fcgiIP := socketAddr[0]
		fcgiPort, _ := strconv.Atoi(socketAddr[1])
		fcgi, err = newFcgiClient(fcgiIP, fcgiPort)
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
		p.importMetric(bytes.NewReader(fpmOutput), acc, addr)
		return nil
	}
	return fmt.Errorf("unable parse phpfpm status, error: %s; %w", string(fpmErr), err)
}

// Gather stat using http protocol
func (p *phpfpm) gatherHTTP(addr string, acc telegraf.Accumulator) error {
	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("unable parse server address %q: %w", addr, err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create new request %q: %w", addr, err)
	}

	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect to phpfpm status page %q: %w", addr, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to get valid stat result from %q: %w", addr, err)
	}

	p.importMetric(res.Body, acc, addr)
	return nil
}

// Import stat data into Telegraf system
func (p *phpfpm) importMetric(r io.Reader, acc telegraf.Accumulator, addr string) {
	if p.Format == "json" {
		p.parseJSON(r, acc, addr)
	} else {
		parseLines(r, acc, addr)
	}
}

func parseLines(r io.Reader, acc telegraf.Accumulator, addr string) {
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
		if fieldName == PfPool {
			currentPool = strings.Trim(keyvalue[1], " ")
			stats[currentPool] = make(metric)
			continue
		}

		// Start to parse metric for current pool
		switch fieldName {
		case PfStartSince,
			PfAcceptedConn,
			PfListenQueue,
			PfMaxListenQueue,
			PfListenQueueLen,
			PfIdleProcesses,
			PfActiveProcesses,
			PfTotalProcesses,
			PfMaxActiveProcesses,
			PfMaxChildrenReached,
			PfSlowRequests:
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
			fields[strings.ReplaceAll(k, " ", "_")] = v
		}
		acc.AddFields("phpfpm", fields, tags)
	}
}

func (p *phpfpm) parseJSON(r io.Reader, acc telegraf.Accumulator, addr string) {
	var metrics JSONMetrics
	if err := json.NewDecoder(r).Decode(&metrics); err != nil {
		p.Log.Errorf("Unable to decode JSON response: %s", err)
		return
	}
	timestamp := time.Now()

	tags := map[string]string{
		"pool": metrics.Pool,
		"url":  addr,
	}
	fields := map[string]any{
		"start_since":          metrics.StartSince,
		"accepted_conn":        metrics.AcceptedConn,
		"listen_queue":         metrics.ListenQueue,
		"max_listen_queue":     metrics.MaxListenQueue,
		"listen_queue_len":     metrics.ListenQueueLen,
		"idle_processes":       metrics.IdleProcesses,
		"active_processes":     metrics.ActiveProcesses,
		"total_processes":      metrics.TotalProcesses,
		"max_active_processes": metrics.MaxActiveProcesses,
		"max_children_reached": metrics.MaxChildrenReached,
		"slow_requests":        metrics.SlowRequests,
	}
	acc.AddFields("phpfpm", fields, tags, timestamp)

	for _, process := range metrics.Processes {
		tags := map[string]string{
			"pool":           metrics.Pool,
			"url":            addr,
			"user":           process.User,
			"request_uri":    process.RequestURI,
			"request_method": process.RequestMethod,
			"script":         process.Script,
		}
		fields := map[string]any{
			"pid":                 process.Pid,
			"state":               process.State,
			"start_time":          process.StartTime,
			"requests":            process.Requests,
			"request_duration":    process.RequestDuration,
			"content_length":      process.ContentLength,
			"last_request_cpu":    process.LastRequestCPU,
			"last_request_memory": process.LastRequestMemory,
		}
		acc.AddFields("phpfpm_process", fields, tags, timestamp)
	}
}

func expandUrls(urls []string) ([]string, error) {
	addrs := make([]string, 0, len(urls))
	for _, address := range urls {
		if isNetworkURL(address) {
			addrs = append(addrs, address)
			continue
		}
		paths, err := globUnixSocket(address)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, paths...)
	}
	return addrs, nil
}

func globUnixSocket(address string) ([]string, error) {
	pattern, status := unixSocketPaths(address)
	glob, err := globpath.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("could not compile glob %q: %w", pattern, err)
	}
	paths := glob.Match()
	if len(paths) == 0 {
		return nil, fmt.Errorf("socket doesn't exist %q", pattern)
	}

	addresses := make([]string, 0, len(paths))
	for _, path := range paths {
		if status != "" {
			path = path + ":" + status
		}
		addresses = append(addresses, path)
	}

	return addresses, nil
}

func unixSocketPaths(addr string) (socketPath string, statusPath string) {
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
