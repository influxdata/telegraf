package haproxy

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//CSV format: https://cbonte.github.io/haproxy-dconv/1.5/configuration.html#9.1

type haproxy struct {
	Servers        []string
	KeepFieldNames bool
	Username       string
	Password       string
	tls.ClientConfig

	client *http.Client
}

var sampleConfig = `
  ## An array of address to gather stats about. Specify an ip on hostname
  ## with optional port. ie localhost, 10.10.3.33:1936, etc.
  ## Make sure you specify the complete path to the stats endpoint
  ## including the protocol, ie http://10.10.3.33:1936/haproxy?stats

  ## If no servers are specified, then default to 127.0.0.1:1936/haproxy?stats
  servers = ["http://myhaproxy.com:1936/haproxy?stats"]

  ## Credentials for basic HTTP authentication
  # username = "admin"
  # password = "admin"

  ## You can also use local socket with standard wildcard globbing.
  ## Server address not starting with 'http' will be treated as a possible
  ## socket, so both examples below are valid.
  # servers = ["socket:/run/haproxy/admin.sock", "/run/haproxy/*.sock"]

  ## By default, some of the fields are renamed from what haproxy calls them.
  ## Setting this option to true results in the plugin keeping the original
  ## field names.
  # keep_field_names = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (r *haproxy) SampleConfig() string {
	return sampleConfig
}

func (r *haproxy) Description() string {
	return "Read metrics of haproxy, via socket or csv stats page"
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *haproxy) Gather(acc telegraf.Accumulator) error {
	if len(g.Servers) == 0 {
		return g.gatherServer("http://127.0.0.1:1936/haproxy?stats", acc)
	}

	endpoints := make([]string, 0, len(g.Servers))

	for _, endpoint := range g.Servers {

		if strings.HasPrefix(endpoint, "http") {
			endpoints = append(endpoints, endpoint)
			continue
		}

		socketPath := getSocketAddr(endpoint)

		matches, err := filepath.Glob(socketPath)

		if err != nil {
			return err
		}

		if len(matches) == 0 {
			endpoints = append(endpoints, socketPath)
		} else {
			for _, match := range matches {
				endpoints = append(endpoints, match)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, server := range endpoints {
		go func(serv string) {
			defer wg.Done()
			if err := g.gatherServer(serv, acc); err != nil {
				acc.AddError(err)
			}
		}(server)
	}

	wg.Wait()
	return nil
}

func (g *haproxy) gatherServerSocket(addr string, acc telegraf.Accumulator) error {
	socketPath := getSocketAddr(addr)

	c, err := net.Dial("unix", socketPath)

	if err != nil {
		return fmt.Errorf("Could not connect to socket '%s': %s", addr, err)
	}

	_, errw := c.Write([]byte("show stat\n"))

	if errw != nil {
		return fmt.Errorf("Could not write to socket '%s': %s", addr, errw)
	}

	return g.importCsvResult(c, acc, socketPath)
}

func (g *haproxy) gatherServer(addr string, acc telegraf.Accumulator) error {
	if !strings.HasPrefix(addr, "http") {
		return g.gatherServerSocket(addr, acc)
	}

	if g.client == nil {
		tlsCfg, err := g.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}
		g.client = client
	}

	if !strings.HasSuffix(addr, ";csv") {
		addr += "/;csv"
	}

	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("unable parse server address '%s': %s", addr, err)
	}

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return fmt.Errorf("unable to create new request '%s': %s", addr, err)
	}
	if u.User != nil {
		p, _ := u.User.Password()
		req.SetBasicAuth(u.User.Username(), p)
		u.User = &url.Userinfo{}
		addr = u.String()
	}

	if g.Username != "" || g.Password != "" {
		req.SetBasicAuth(g.Username, g.Password)
	}

	res, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect to haproxy server '%s': %s", addr, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to get valid stat result from '%s', http response code : %d", addr, res.StatusCode)
	}

	if err := g.importCsvResult(res.Body, acc, u.Host); err != nil {
		return fmt.Errorf("unable to parse stat result from '%s': %s", addr, err)
	}

	return nil
}

func getSocketAddr(sock string) string {
	socketAddr := strings.Split(sock, ":")

	if len(socketAddr) >= 2 {
		return socketAddr[1]
	} else {
		return socketAddr[0]
	}
}

var typeNames = []string{"frontend", "backend", "server", "listener"}
var fieldRenames = map[string]string{
	"pxname":     "proxy",
	"svname":     "sv",
	"act":        "active_servers",
	"bck":        "backup_servers",
	"cli_abrt":   "cli_abort",
	"srv_abrt":   "srv_abort",
	"hrsp_1xx":   "http_response.1xx",
	"hrsp_2xx":   "http_response.2xx",
	"hrsp_3xx":   "http_response.3xx",
	"hrsp_4xx":   "http_response.4xx",
	"hrsp_5xx":   "http_response.5xx",
	"hrsp_other": "http_response.other",
}

func (g *haproxy) importCsvResult(r io.Reader, acc telegraf.Accumulator, host string) error {
	csvr := csv.NewReader(r)
	now := time.Now()

	headers, err := csvr.Read()
	if err != nil {
		return err
	}
	if len(headers[0]) <= 2 || headers[0][:2] != "# " {
		return fmt.Errorf("did not receive standard haproxy headers")
	}
	headers[0] = headers[0][2:]

	for {
		row, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fields := make(map[string]interface{})
		tags := map[string]string{
			"server": host,
		}

		if len(row) != len(headers) {
			return fmt.Errorf("number of columns does not match number of headers. headers=%d columns=%d", len(headers), len(row))
		}
		for i, v := range row {
			if v == "" {
				continue
			}

			colName := headers[i]
			fieldName := colName
			if !g.KeepFieldNames {
				if fieldRename, ok := fieldRenames[colName]; ok {
					fieldName = fieldRename
				}
			}

			switch colName {
			case "pxname", "svname":
				tags[fieldName] = v
			case "type":
				vi, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("unable to parse type value '%s'", v)
				}
				if vi >= int64(len(typeNames)) {
					return fmt.Errorf("received unknown type value: %d", vi)
				}
				tags[fieldName] = typeNames[vi]
			case "check_desc", "agent_desc":
				// do nothing. These fields are just a more verbose description of the check_status & agent_status fields
			case "status", "check_status", "last_chk", "mode", "tracked", "agent_status", "last_agt", "addr", "cookie":
				// these are string fields
				fields[fieldName] = v
			case "lastsess":
				vi, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					//TODO log the error. And just once (per column) so we don't spam the log
					continue
				}
				fields[fieldName] = vi
			default:
				vi, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					//TODO log the error. And just once (per column) so we don't spam the log
					continue
				}
				fields[fieldName] = vi
			}
		}
		acc.AddFields("haproxy", fields, tags, now)
	}
	return err
}

func init() {
	inputs.Add("haproxy", func() telegraf.Input {
		return &haproxy{}
	})
}
