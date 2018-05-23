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

type runner func(cmdName string, Timeout internal.Duration, UseSudo bool, Server string, ThreadAsTag bool) (*bytes.Buffer, error)

// Unbound is used to store configuration values
type Unbound struct {
	Binary      string
	Timeout     internal.Duration
	UseSudo     bool
	Server      string
	ThreadAsTag bool

	filter filter.Filter
	run    runner
}

var defaultBinary = "/usr/sbin/unbound-control"
var defaultTimeout = internal.Duration{Duration: time.Second}

var sampleConfig = `
  ## Address of server to connect to, read from unbound conf default, optionally ':port'
  ## Will lookup IP if given a hostname
  server = "127.0.0.1:8953"

  ## If running as a restricted user you can prepend sudo for additional access:
  # use_sudo = false

  ## The default location of the unbound-control binary can be overridden with:
  # binary = "/usr/sbin/unbound-control"

  ## The default timeout of 1s can be overriden with:
  # timeout = "1s"

  ## When set to true, thread metrics are tagged with the thread id.
  ##
  ## The default is false for backwards compatibility, and will be change to
  ## true in a future version.  It is recommended to set to true on new
  ## deployments.
  thread_as_tag = false
`

// Description displays what this plugin is about
func (s *Unbound) Description() string {
	return "A plugin to collect stats from the Unbound DNS resolver"
}

// SampleConfig displays configuration instructions
func (s *Unbound) SampleConfig() string {
	return sampleConfig
}

// Shell out to unbound_stat and return the output
func unboundRunner(cmdName string, Timeout internal.Duration, UseSudo bool, Server string, ThreadAsTag bool) (*bytes.Buffer, error) {
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
	statExcluded := []string{"histogram.*"}
	filterExcluded, err := filter.Compile(statExcluded)
	if err != nil {
		return err
	}

	out, err := s.run(s.Binary, s.Timeout, s.UseSudo, s.Server, s.ThreadAsTag)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	// Process values
	fields := make(map[string]interface{})
	fieldsThreads := make(map[string]map[string]interface{})

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
		if filterExcluded.Match(stat) {
			continue
		}

		fieldValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("Expected a numerical value for %s = %v",
				stat, value))
			continue
		}

		// is this a thread related value?
		if s.ThreadAsTag && strings.HasPrefix(stat, "thread") {
			// split the stat
			statTokens := strings.Split(stat, ".")
			// make sure we split something
			if len(statTokens) > 1 {
				// set the thread identifier
				threadID := strings.TrimPrefix(statTokens[0], "thread")
				// make sure we have a proper thread ID
				if _, err = strconv.Atoi(threadID); err == nil {
					// create new slice without the thread identifier (skip first token)
					threadTokens := statTokens[1:]
					// re-define stat
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

	acc.AddFields("unbound", fields, nil)

	if s.ThreadAsTag && len(fieldsThreads) > 0 {
		for thisThreadID, thisThreadFields := range fieldsThreads {
			thisThreadTag := map[string]string{"thread": thisThreadID}
			acc.AddFields("unbound_threads", thisThreadFields, thisThreadTag)
		}
	}

	return nil
}

func init() {
	inputs.Add("unbound", func() telegraf.Input {
		return &Unbound{
			run:         unboundRunner,
			Binary:      defaultBinary,
			Timeout:     defaultTimeout,
			UseSudo:     false,
			Server:      "",
			ThreadAsTag: false,
		}
	})
}
