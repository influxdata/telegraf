package nsd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type runner func(cmdName string, timeout internal.Duration, useSudo bool, server string) (*bytes.Buffer, error)

type Nsd struct {
	Binary  string
	Timeout internal.Duration
	UseSudo bool
	Server  string

	host string
	port string
	ip net.IP

	filter filter.Filter
	run    runner
}

var defaultBinary = "/usr/sbin/nsd-control"
var defaultTimeout = internal.Duration{Duration: time.Second}

var sampleConfig = `
  ## Address of server to connect to, read from nsd conf default, optionally ':port'
  ## Will lookup IP if given a hostname
  server = "127.0.0.1:8952"

  ## If running as a restricted user you can prepend sudo for additional access:
  # use_sudo = false

  ## The default location of the nsd-control binary can be overridden with:
  # binary = "/usr/sbin/nsd-control"

  ## The default timeout of 1s can be overridden with:
  # timeout = "1s"
`

const defaultPort = "8952"
var localhostNet *net.IPNet
var localhostNetv6 *net.IPNet

func (s *Nsd) SampleConfig() string {
	return sampleConfig
}

func (s *Nsd) Description() string {
	return "A plugin to collect stats from the NSD DNS Server"
}

func nsdRunner(cmdName string, timeout internal.Duration, useSudo bool, server string) (*bytes.Buffer, error) {
	cmdArgs := []string{"stats_noreset"}

	if server != "" {
		host, port, err := net.SplitHostPort(server)
		if err != nil {
			host = server
			port = ""
		}

		resolver := net.Resolver{}
		ctx, lookUpCancel := context.WithTimeout(context.Background(), timeout.Duration)
		defer lookUpCancel()
		serverIps, err := resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("error looking up ip for server: %s: %s", server, err)
		}
		if len(serverIps) == 0 {
			return nil, fmt.Errorf("error no ip for server: %s: %s", server, err)
		}
		server := serverIps[0].IP.String()
		if port != "" {
			server = server + "@" + port
		}

		cmdArgs = append([]string{"-s", server}, cmdArgs...)
	}

	cmd := exec.Command(cmdName, cmdArgs...)

	if useSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running nsd-control: %s (%s %v)", err, cmdName, cmdArgs)
	}

	return &out, nil
}

func (s *Nsd) Gather(acc telegraf.Accumulator) error {
	out, err := s.run(s.Binary, s.Timeout, s.UseSudo, s.Server)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	fields := make(map[string]interface{})
	fieldsServers := make(map[string]map[string]interface{})

	if s.host == "" || s.port == "" {
		s.host, s.port, err = net.SplitHostPort(s.Server)
		// the port is nil since we already checked for other errors in the nsdRunner
		if err != nil {
			s.port = defaultPort
		}
		s.ip = net.ParseIP(s.host)
		if s.ip == nil {
			return fmt.Errorf("invalid server ip for host: %s", s.host)
		}

		if localhostNet.Contains(s.ip) || localhostNetv6.Contains(s.ip) {
			hostname, err := os.Hostname()
			if err != nil {
				s.host = "localhost"
			} else {
				s.host = hostname
			}
		}
	}

	tags := map[string]string{"source": s.host, "port": s.port}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "=")

		if len(cols) != 2 {
			continue
		}

		stat := cols[0]
		value := cols[1]

		var fieldValue interface{}
		// only two values are floats
		if stat == "time.boot" || stat == "time.elapsed" {
			fieldValue, err = strconv.ParseFloat(value, 64)
		} else {
			fieldValue, err = strconv.ParseUint(value, 10, 64)
		}
		if err != nil {
			acc.AddError(fmt.Errorf("Expected a numerical value for %s = %v",
				stat, value))
			continue
		}

		if strings.HasPrefix(stat, "server") {
			statTokens := strings.Split(stat, ".")
			if len(statTokens) > 1 {
				serverID := strings.TrimPrefix(statTokens[0], "server")
				if _, err = strconv.Atoi(serverID); err == nil {
					serverTokens := statTokens[1:]
					field := strings.Join(serverTokens[:], "_")
					if fieldsServers[serverID] == nil {
						fieldsServers[serverID] = make(map[string]interface{})
					}
					fieldsServers[serverID][field] = fieldValue
				}
			}
		} else {
			field := strings.Replace(stat, ".", "_", -1)
			fields[field] = fieldValue
		}
	}

	acc.AddFields("nsd", fields, tags)

	if len(fieldsServers) > 0 {
		for thisServerID, thisServerFields := range fieldsServers {
			tags["server"] = thisServerID
			acc.AddFields("nsd", thisServerFields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("nsd", func() telegraf.Input {
		return &Nsd{
			run:     nsdRunner,
			Binary:  defaultBinary,
			Timeout: defaultTimeout,
			UseSudo: false,
			Server:  "",
		}
	})
	_, localhostNet, _ = net.ParseCIDR("127.0.0.1/8")
	_, localhostNetv6, _ = net.ParseCIDR("::1/128")
}
