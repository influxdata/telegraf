package nsd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type runner func(cmdName string, Timeout internal.Duration, UseSudo bool, Server string, ThreadAsTag bool) (*bytes.Buffer, error)

type Nsd struct {
	Binary      string
	Timeout     internal.Duration
	UseSudo     bool
	Server      string
	ThreadAsTag bool

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

  ## When set to true, thread metrics are tagged with the thread id.
  ##
  ## The default is false for backwards compatibility, and will be change to
  ## true in a future version.  It is recommended to set to true on new
  ## deployments.
  thread_as_tag = false
`

func (s *Nsd) SampleConfig() string {
	return sampleConfig
}

func (s *Nsd) Description() string {
	return "A plugin to collect stats from the NSD DNS Server"
}

func nsdRunner(cmdName string, Timeout internal.Duration, UseSudo bool, Server string, ThreadAsTag bool) (*bytes.Buffer, error) {
	cmdArgs := []string{"stats_noreset"}

	if Server != "" {
		host, port, err := net.SplitHostPort(Server)
		if err != nil {
			host = Server
			port = ""
		}

		resolver := net.Resolver{}
		ctx, lookUpCancel := context.WithTimeout(context.Background(), Timeout.Duration)
		defer lookUpCancel()
		serverIps, err := resolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("error looking up ip for server: %s: %s", Server, err)
		}
		if len(serverIps) == 0 {
			return nil, fmt.Errorf("error no ip for server: %s: %s", Server, err)
		}
		server := serverIps[0].IP.String()
		if port != "" {
			server = server + "@" + port
		}

		cmdArgs = append([]string{"-s", server}, cmdArgs...)
	}

	cmd := exec.Command(cmdName, cmdArgs...)

	if UseSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running nsd-control: %s (%s %v)", err, cmdName, cmdArgs)
	}

	return &out, nil
}

func (s *Nsd) Gather(acc telegraf.Accumulator) error {
	out, err := s.run(s.Binary, s.Timeout, s.UseSudo, s.Server, s.ThreadAsTag)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	fields := make(map[string]interface{})
	fieldsThreads := make(map[string]map[string]interface{})

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "=")

		if len(cols) != 2 {
			continue
		}

		stat := cols[0]
		value := cols[1]

		fieldValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("Expected a numerical value for %s = %v",
				stat, value))
			continue
		}

		if s.ThreadAsTag && strings.HasPrefix(stat, "server") {
			statTokens := strings.Split(stat, ".")
			if len(statTokens) > 1 {
				threadID := strings.TrimPrefix(statTokens[0], "server")
				if _, err = strconv.Atoi(threadID); err == nil {
					threadTokens := statTokens[1:]
					field := strings.Join(threadTokens[:], "_")
					if fieldsThreads[threadID] == nil {
						fieldsThreads[threadID] = make(map[string]interface{})
					}
					fieldsThreads[threadID][field] = fieldValue
				}
			}
		} else {
			field := strings.Replace(stat, ".", "_", -1)
			fields[field] = fieldValue
		}
	}

	acc.AddFields("nsd", fields, nil)

	if s.ThreadAsTag && len(fieldsThreads) > 0 {
		for thisThreadID, thisThreadFields := range fieldsThreads {
			thisThreadTag := map[string]string{"thread": thisThreadID}
			acc.AddFields("nsd_threads", thisThreadFields, thisThreadTag)
		}
	}

	return nil
}

func init() {
	inputs.Add("nsd", func() telegraf.Input {
		return &Nsd{
			run:         nsdRunner,
			Binary:      defaultBinary,
			Timeout:     defaultTimeout,
			UseSudo:     false,
			Server:      "",
			ThreadAsTag: false,
		}
	})
}
