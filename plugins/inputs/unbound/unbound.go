package unbound

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type runner func(cmdName string, Timeout internal.Duration, UseSudo bool, Server string) (*bytes.Buffer, error)

// Unbound is used to store configuration values
type Unbound struct {
	Binary  string
	Timeout internal.Duration
	UseSudo bool
	Server  string

	filter filter.Filter
	run    runner
}

var defaultBinary = "/usr/sbin/unbound-control"
var defaultTimeout = internal.Duration{Duration: time.Second}

var sampleConfig = `
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the unbound-control binary can be overridden with:
  binary = "/usr/sbin/unbound-control"

  ## The default timeout of 1s can be overriden with:
  timeout = "1s"

  ## Use the builtin fielddrop/fieldpass telegraf filters in order to keep/remove specific fields
  fieldpass = ["total_*", "num_*","time_up", "mem_*"]
  
  ## IP of server to connect to, read from unbound conf default, optionally ':port'
  ## Will lookup IP if given a hostname
  server = "127.0.0.1:8953"
`

func (s *Unbound) Description() string {
	return "A plugin to collect stats from Unbound - a validating, recursive, and caching DNS resolver"
}

// SampleConfig displays configuration instructions
func (s *Unbound) SampleConfig() string {
	return sampleConfig
}

// Shell out to unbound_stat and return the output
func unboundRunner(cmdName string, Timeout internal.Duration, UseSudo bool, Server string) (*bytes.Buffer, error) {
	cmdArgs := []string{"stats_noreset"}

	if Server != "" {
		host, port, err := net.SplitHostPort(Server)
		if err != nil { // No port was specified
			host = Server
			port = ""
		}

		// Unbound control requires an IP address, and we want to be nice to the user
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
		return &out, fmt.Errorf("error running unbound-control: %s (%s %v)", err, cmdName, cmdArgs)
	}

	return &out, nil
}

// Gather collects stats from unbound-control and adds them to the Accumulator
//
// All the dots in stat name will replaced by underscores. Histogram statistics will not be collected.
func (s *Unbound) Gather(acc telegraf.Accumulator) error {

	// Always exclude histrogram statistics
	stat_excluded := []string{"histogram.*"}
	filter_excluded, err := filter.Compile(stat_excluded)
	if err != nil {
		return err
	}

	out, err := s.run(s.Binary, s.Timeout, s.UseSudo, s.Server)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	// Process values
	fields := make(map[string]interface{})
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {

		cols := strings.Split(scanner.Text(), "=")

		// Check split correctness
		if len(cols) != 2 {
			continue
		}

		stat := cols[0]
		value := cols[1]

		// Filter value
		if filter_excluded.Match(stat) {
			continue
		}

		field := strings.Replace(stat, ".", "_", -1)

		fields[field], err = strconv.ParseFloat(value, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("Expected a numerical value for %s = %v\n",
				stat, value))
		}
	}

	acc.AddFields("unbound", fields, nil)

	return nil
}

func init() {
	inputs.Add("unbound", func() telegraf.Input {
		return &Unbound{
			run:     unboundRunner,
			Binary:  defaultBinary,
			Timeout: defaultTimeout,
			UseSudo: false,
			Server:  "",
		}
	})
}
