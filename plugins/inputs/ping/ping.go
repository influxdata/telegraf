// +build !windows

package ping

import (
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HostPinger is a function that runs the "ping" function using a list of
// passed arguments. This can be easily switched with a mocked ping function
// for unit test purposes (see ping_test.go)
type HostPinger func(timeout float64, args ...string) (string, error)

type Ping struct {
	// Interval at which to ping (ping -i <INTERVAL>)
	PingInterval float64 `toml:"ping_interval"`

	// Number of pings to send (ping -c <COUNT>)
	Count int

	// Ping timeout, in seconds. 0 means no timeout (ping -W <TIMEOUT>)
	Timeout float64

	// Interface to send ping from (ping -I <INTERFACE>)
	Interface string

	// URLs to ping
	Urls []string

	// host ping function
	pingHost HostPinger

	IPVersion int
}

func (_ *Ping) Description() string {
	return "Ping given url(s) and return statistics"
}

const sampleConfig = `
  ## NOTE: this plugin forks the ping command. You may need to set capabilities
  ## via setcap cap_net_raw+p /bin/ping
  #
  ## urls to ping
  urls = ["www.google.com"] # required
  ## number of pings to send per collection (ping -c <COUNT>)
  count = 1 # required
  ## interval, in s, at which to ping. 0 == default (ping -i <PING_INTERVAL>)
  ping_interval = 0.0
  ## per-ping timeout, in s. 0 == no timeout (ping -W <TIMEOUT>)
  timeout = 1.0
  ## interface to send ping from (ping -I <INTERFACE>)
  interface = ""
  ## force which IP version to use. (ping -6 or ping -4)
  IPVersion = 4
`

func (_ *Ping) SampleConfig() string {
	return sampleConfig
}

func (p *Ping) Gather(acc telegraf.Accumulator) error {

	var wg sync.WaitGroup
	errorChannel := make(chan error, len(p.Urls)*2)

	// Spin off a go routine for each url to ping
	for _, url := range p.Urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			args := p.args(u)
			totalTimeout := float64(p.Count)*p.Timeout + float64(p.Count-1)*p.PingInterval
			out, err := p.pingHost(totalTimeout, args...)
			if err != nil {
				// Combine go err + stderr output
				errorChannel <- errors.New(
					strings.TrimSpace(out) + ", " + err.Error())
			}
			tags := map[string]string{"url": u}

			trans, rec, avg, ipversion, err := processPingOutput(out)
			if err != nil {
				// fatal error
				errorChannel <- err
				return
			}
			tags["ipversion"] = strconv.Itoa(ipversion)
			// Calculate packet loss percentage
			loss := float64(trans-rec) / float64(trans) * 100.0
			fields := map[string]interface{}{
				"packets_transmitted": trans,
				"packets_received":    rec,
				"percent_packet_loss": loss,
			}
			if avg > 0 {
				fields["average_response_ms"] = avg
			}
			acc.AddFields("ping", fields, tags)
		}(url)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

func hostPinger(timeout float64, args ...string) (string, error) {
	bin, err := exec.LookPath("ping")
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)
	out, err := internal.CombinedOutputTimeout(c,
		time.Second*time.Duration(timeout+1))
	return string(out), err
}

// args returns the arguments for the 'ping' executable
func (p *Ping) args(url string) []string {
	// Build the ping command args based on toml config
	args := []string{"-c", strconv.Itoa(p.Count), "-n", "-s", "16"}
	if p.PingInterval > 0 {
		args = append(args, "-i", strconv.FormatFloat(p.PingInterval, 'f', 1, 64))
	}
	if p.Timeout > 0 {
		switch runtime.GOOS {
		case "darwin":
			args = append(args, "-W", strconv.FormatFloat(p.Timeout*1000, 'f', 1, 64))
		case "linux":
			args = append(args, "-W", strconv.FormatFloat(p.Timeout, 'f', 1, 64))
		default:
			// Not sure the best option here, just assume GNU ping?
			args = append(args, "-W", strconv.FormatFloat(p.Timeout, 'f', 1, 64))
		}
	}
	if p.Interface != "" {
		args = append(args, "-I", p.Interface)
	}
	args = append(args, url)
	switch p.IPVersion {
	case 4: args = append(args, "-4")
	case 6: args = append(args, "-6")
	}

	return args
}

// processPingOutput takes in a string output from the ping command, like:
//
//     PING www.google.com (173.194.115.84): 56 data bytes
//     64 bytes from 173.194.115.84: icmp_seq=0 ttl=54 time=52.172 ms
//     64 bytes from 173.194.115.84: icmp_seq=1 ttl=54 time=34.843 ms
//
//     --- www.google.com ping statistics ---
//     2 packets transmitted, 2 packets received, 0.0% packet loss
//     round-trip min/avg/max/stddev = 34.843/43.508/52.172/8.664 ms
//
// It returns (<transmitted packets>, <received packets>, <average response>)

// for IPv6 address the ping looks like
//    PING google.com(fra16s08-in-x0e.1e100.net (2a00:1450:4001:817::200e)) 56 data bytes
func processPingOutput(out string) (int, int, float64, int, error) {
	var trans, recv, ipversion int
	var avg float64
	// Set this error to nil if we find a 'transmitted' line
	err := errors.New("Fatal error processing ping output")
	re := regexp.MustCompile(`(?i)^ping.+\(([a-f0-9\.:]+)\)`)  // re[1] is the IP
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		str := re.FindStringSubmatch(line)
		if len(str) != 0 {
			if strings.Contains(str[1],":") {
				ipversion = 6
			} else {
				ipversion = 4
			}
		} else if strings.Contains(line, "transmitted") &&
			strings.Contains(line, "received") {
			err = nil
			stats := strings.Split(line, ", ")
			// Transmitted packets
			trans, err = strconv.Atoi(strings.Split(stats[0], " ")[0])
			if err != nil {
				return trans, recv, avg, ipversion, err
			}
			// Received packets
			recv, err = strconv.Atoi(strings.Split(stats[1], " ")[0])
			if err != nil {
				return trans, recv, avg, ipversion, err
			}
		} else if strings.Contains(line, "min/avg/max") {
			stats := strings.Split(line, " = ")[1]
			avg, err = strconv.ParseFloat(strings.Split(stats, "/")[1], 64)
			if err != nil {
				return trans, recv, avg, ipversion, err
			}
		}
	}
	return trans, recv, avg, ipversion, err
}

func init() {
	inputs.Add("ping", func() telegraf.Input {
		return &Ping{pingHost: hostPinger}
	})
}
