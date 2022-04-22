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

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (h *haproxy) Gather(acc telegraf.Accumulator) error {
	if len(h.Servers) == 0 {
		return h.gatherServer("http://127.0.0.1:1936/haproxy?stats", acc)
	}

	endpoints := make([]string, 0, len(h.Servers))

	for _, endpoint := range h.Servers {
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
			endpoints = append(endpoints, matches...)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, server := range endpoints {
		go func(serv string) {
			defer wg.Done()
			if err := h.gatherServer(serv, acc); err != nil {
				acc.AddError(err)
			}
		}(server)
	}

	wg.Wait()
	return nil
}

func (h *haproxy) gatherServerSocket(addr string, acc telegraf.Accumulator) error {
	socketPath := getSocketAddr(addr)

	c, err := net.Dial("unix", socketPath)

	if err != nil {
		return fmt.Errorf("could not connect to socket '%s': %s", addr, err)
	}

	_, errw := c.Write([]byte("show stat\n"))

	if errw != nil {
		return fmt.Errorf("could not write to socket '%s': %s", addr, errw)
	}

	return h.importCsvResult(c, acc, socketPath)
}

func (h *haproxy) gatherServer(addr string, acc telegraf.Accumulator) error {
	if !strings.HasPrefix(addr, "http") {
		return h.gatherServerSocket(addr, acc)
	}

	if h.client == nil {
		tlsCfg, err := h.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: 3 * time.Second,
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   4 * time.Second,
		}
		h.client = client
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

	if h.Username != "" || h.Password != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}

	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect to haproxy server '%s': %s", addr, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to get valid stat result from '%s', http response code : %d", addr, res.StatusCode)
	}

	if err := h.importCsvResult(res.Body, acc, u.Host); err != nil {
		return fmt.Errorf("unable to parse stat result from '%s': %s", addr, err)
	}

	return nil
}

func getSocketAddr(sock string) string {
	socketAddr := strings.Split(sock, ":")

	if len(socketAddr) >= 2 {
		return socketAddr[1]
	}
	return socketAddr[0]
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

func (h *haproxy) importCsvResult(r io.Reader, acc telegraf.Accumulator, host string) error {
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
			if !h.KeepFieldNames {
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
