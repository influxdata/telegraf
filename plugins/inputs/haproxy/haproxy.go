package haproxy

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	typeNames    = []string{"frontend", "backend", "server", "listener"}
	fieldRenames = map[string]string{
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
)

// CSV format: https://cbonte.github.io/haproxy-dconv/1.5/configuration.html#9.1

type HAProxy struct {
	Servers        []string `toml:"servers"`
	KeepFieldNames bool     `toml:"keep_field_names"`
	Username       string   `toml:"username"`
	Password       string   `toml:"password"`
	MasterSocket   string   `toml:"master_socket"`
	UseMaster      bool     `toml:"use_master"`
	AggregateWorkers bool   `toml:"aggregate_workers"`
	AddSourceTag   bool     `toml:"add_source_tag"`
	Concurrency    int      `toml:"concurrency"`
	tls.ClientConfig

	client *http.Client
}

func (*HAProxy) SampleConfig() string {
	return sampleConfig
}

func (h *HAProxy) Gather(acc telegraf.Accumulator) error {
	if len(h.Servers) == 0 {
		return h.gatherServer("http://127.0.0.1:1936/haproxy?stats", acc)
	}

	endpoints := make([]string, 0, len(h.Servers))

	for _, endpoint := range h.Servers {
		if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") || strings.HasPrefix(endpoint, "tcp://") {
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

	// If configured, try to gather via master socket(s) first (to obtain worker stats)
	if h.UseMaster {
		masterCandidates := []string{}
		if h.MasterSocket != "" {
			masterCandidates = append(masterCandidates, h.MasterSocket)
		} else {
			masterCandidates = append(masterCandidates,
				"/run/haproxy-master.sock",
				"/run/haproxy/master.sock",
				"/run/haproxy/haproxy-master.sock",
			)
		}

		// try each candidate until one succeeds
		for _, m := range masterCandidates {
			if _, err := os.Stat(m); err != nil {
				// try dialing anyway - some environments may not show file but socket accessible
			}
			err := h.gatherFromMaster(m, acc)
			if err == nil {
				// success; do not fail, continue to gather other endpoints as well
				break
			}
			// record error and continue trying other candidates
			acc.AddError(fmt.Errorf("master socket %s: %w", m, err))
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	for _, server := range endpoints {
		go func(serv string) {
			defer wg.Done()
			// If use_master was enabled we already attempted master collection; avoid double gathering
			// when the endpoint equals one of the master candidates and UseMaster succeeded we still allow gatherSocket
			if err := h.gatherServer(serv, acc); err != nil {
				acc.AddError(err)
			}
		}(server)
	}

	wg.Wait()
	return nil
}

func (h *HAProxy) gatherFromMaster(masterSocket string, acc telegraf.Accumulator) error {
	// Dial unix master socket
	if masterSocket == "" {
		return fmt.Errorf("empty master socket")
	}

	// set concurrency default
	concurrency := h.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	dialer := &net.Dialer{Timeout: 3 * time.Second}
	c, err := dialer.Dial("unix", masterSocket)
	if err != nil {
		return fmt.Errorf("could not connect to master socket %s: %w", masterSocket, err)
	}
	defer c.Close()

	// helper for writing command and reading response
	writeAndRead := func(cmd string) (string, error) {
		c.SetDeadline(time.Now().Add(4 * time.Second))
		if _, err := c.Write([]byte(cmd)); err != nil {
			return "", err
		}
		// read until timeout or EOF
		buf := make([]byte, 0, 32*1024)
		tmp := make([]byte, 1024)
		for {
			n, err := c.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				// timeout or other error
				break
			}
			// if less than buffer, continue trying until deadline
		}
		return string(buf), nil
	}

	procOut, err := writeAndRead("show proc\n")
	if err != nil {
		return fmt.Errorf("failed show proc on master socket %s: %w", masterSocket, err)
	}

	pids := []string{}
	for _, line := range strings.Split(procOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// first token
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if _, err := strconv.Atoi(fields[0]); err == nil {
			// ignore lines that contain the word 'master'
			if strings.Contains(line, "master") {
				continue
			}
			pids = append(pids, fields[0])
		}
	}

	if len(pids) == 0 {
		return fmt.Errorf("no worker pids found in master response")
	}

	// For each pid, ask master for @!pid show stat
	// We will collect responses and either import individually or aggregate
	responses := make([]string, len(pids))
	var gerr error
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, pid := range pids {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, pid string) {
			defer wg.Done()
			defer func() { <-sem }()

			cmd := fmt.Sprintf("@!%s show stat\n", pid)
			c, err := dialer.Dial("unix", masterSocket)
			if err != nil {
				gerr = err
				return
			}
			defer c.Close()
			c.SetDeadline(time.Now().Add(4 * time.Second))
			if _, err := c.Write([]byte(cmd)); err != nil {
				gerr = err
				return
			}
			buf := make([]byte, 0, 32*1024)
			tmp := make([]byte, 1024)
			for {
				n, err := c.Read(tmp)
				if n > 0 {
					buf = append(buf, tmp[:n]...)
				}
				if err != nil {
					break
				}
			}
			responses[idx] = string(buf)
		}(i, pid)
	}
	wg.Wait()
	if gerr != nil {
		return fmt.Errorf("error gathering stats from workers: %w", gerr)
	}

	if h.AggregateWorkers {
		// simple aggregation by pxname+svname summing numeric columns
		agg := map[string]map[string]uint64{} // key -> field -> sum
		statusMap := map[string]string{}
		for _, resp := range responses {
			if strings.TrimSpace(resp) == "" {
				continue
			}
			r := csv.NewReader(strings.NewReader(resp))
			head, err := r.Read()
			if err != nil {
				continue
			}
			if len(head) > 0 && len(head[0]) > 2 && head[0][:2] == "# " {
				head[0] = head[0][2:]
			}
			for {
				row, err := r.Read()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					break
				}
				if len(row) != len(head) {
					continue
				}
				// find pxname and svname
				var px, sv, st string
				for i, v := range row {
					if v == "" {
						continue
					}
					col := head[i]
					switch col {
					case "pxname":
						px = v
					case "svname":
						sv = v
					case "status":
						st = v
					}
				}
				if px == "" || sv == "" {
					continue
				}
				key := px + "," + sv
				if _, ok := agg[key]; !ok {
					agg[key] = map[string]uint64{}
				}
				// accumulate
				for i, v := range row {
					if v == "" {
						continue
					}
					col := head[i]
					// ignore string columns
					switch col {
					case "pxname", "svname", "status", "check_status", "last_chk", "mode", "tracked", "agent_status", "last_agt", "addr", "cookie":
						continue
					}
					vi, err := strconv.ParseUint(v, 10, 64)
					if err != nil {
						continue
					}
					fieldName := col
					if !h.KeepFieldNames {
						if fn, ok := fieldRenames[col]; ok {
							fieldName = fn
						}
					}
					agg[key][fieldName] += vi
				}
				// status handling: prefer UP/OPEN if any worker has it
				if prev, ok := statusMap[key]; !ok {
					statusMap[key] = st
				} else {
					if st == "UP" || st == "OPEN" {
						statusMap[key] = st
					}
				}
			}
		}
		// publish aggregated metrics
		now := time.Now()
		for key, fields := range agg {
			parts := strings.Split(key, ",")
			px := parts[0]
			sv := parts[1]
			tags := map[string]string{
				"server": masterSocket,
				"proxy":  px,
				"sv":     sv,
				"type":   "server",
			}
			if h.AddSourceTag {
				tags["source"] = "master:aggregated"
			}
			// add string fields like status if available
			if st, ok := statusMap[key]; ok {
				fields["status"] = 0 // placeholder: status is a string, we'll add separately
				// We'll add status as a string field below
				// Build a copy of numeric fields
				numFields := map[string]interface{}{}
				for k, v := range fields {
					numFields[k] = v
				}
				// add status as string
				numFields["status"] = st
				acc.AddFields("haproxy", numFields, tags, now)
			} else {
				numFields := map[string]interface{}{}
				for k, v := range fields {
					numFields[k] = v
				}
				acc.AddFields("haproxy", numFields, tags, now)
			}
		}
		return nil
	}

	// not aggregating: import each response separately
	for i, resp := range responses {
		if strings.TrimSpace(resp) == "" {
			continue
		}
		extra := map[string]string{}
		if h.AddSourceTag {
			extra["source"] = fmt.Sprintf("master:pid=%s", pids[i])
		}
		if err := h.importCsvResult(strings.NewReader(resp), acc, masterSocket, extra); err != nil {
			// continue collecting others, but record error
			acc.AddError(fmt.Errorf("master %s pid %s: %w", masterSocket, pids[i], err))
		}
	}

	return nil
}

func (h *HAProxy) gatherServerSocket(addr string, acc telegraf.Accumulator) error {
	var network, address string
	if strings.HasPrefix(addr, "tcp://") {
		network = "tcp"
		address = strings.TrimPrefix(addr, "tcp://")
	} else {
		network = "unix"
		address = getSocketAddr(addr)
	}

	c, err := net.Dial(network, address)
	if err != nil {
		return fmt.Errorf("could not connect to '%s://%s': %w", network, address, err)
	}

	_, errw := c.Write([]byte("show stat\n"))
	if errw != nil {
		return fmt.Errorf("could not write to socket '%s://%s': %w", network, address, errw)
	}

	return h.importCsvResult(c, acc, address, nil)
}

func (h *HAProxy) gatherServer(addr string, acc telegraf.Accumulator) error {
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
		return fmt.Errorf("unable parse server address %q: %w", addr, err)
	}

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return fmt.Errorf("unable to create new request %q: %w", addr, err)
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
		return fmt.Errorf("unable to connect to haproxy server %q: %w", addr, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to get valid stat result from %q, http response code : %d", addr, res.StatusCode)
	}

	if err := h.importCsvResult(res.Body, acc, u.Host, nil); err != nil {
		return fmt.Errorf("unable to parse stat result from %q: %w", addr, err)
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

// importCsvResult now accepts optional extraTags to be merged with tags for each row
func (h *HAProxy) importCsvResult(r io.Reader, acc telegraf.Accumulator, host string, extraTags map[string]string) error {
	csvr := csv.NewReader(r)
	now := time.Now()

	headers, err := csvr.Read()
	if err != nil {
		return err
	}
	if len(headers[0]) <= 2 || headers[0][:2] != "# " {
		return errors.New("did not receive standard haproxy headers")
	}
	headers[0] = headers[0][2:]

	for {
		row, err := csvr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		fields := make(map[string]interface{})
		tags := map[string]string{
			"server": host,
		}
		for k, v := range extraTags {
			tags[k] = v
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
					return fmt.Errorf("unable to parse type value %q", v)
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
					// TODO log the error. And just once (per column) so we don't spam the log
					continue
				}
				fields[fieldName] = vi
			default:
				vi, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					// TODO log the error. And just once (per column) so we don't spam the log
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
		return &HAProxy{}
	})
}
