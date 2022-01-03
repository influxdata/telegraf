//go:build !windows
// +build !windows

package ping

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf"
)

func (p *Ping) pingToURL(u string, acc telegraf.Accumulator) {
	tags := map[string]string{"url": u}
	fields := map[string]interface{}{"result_code": 0}

	out, err := p.pingHost(p.Binary, 60.0, p.args(u, runtime.GOOS)...)
	if err != nil {
		// Some implementations of ping return a non-zero exit code on
		// timeout, if this occurs we will not exit and try to parse
		// the output.
		// Linux iputils-ping returns 1, BSD-derived ping returns 2.
		status := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
				status = ws.ExitStatus()
				fields["result_code"] = status
			}
		}

		var timeoutExitCode int
		switch runtime.GOOS {
		case "freebsd", "netbsd", "openbsd", "darwin":
			timeoutExitCode = 2
		case "linux":
			timeoutExitCode = 1
		default:
			timeoutExitCode = 1
		}

		if status != timeoutExitCode {
			// Combine go err + stderr output
			out = strings.TrimSpace(out)
			if len(out) > 0 {
				acc.AddError(fmt.Errorf("host %s: %s, %s", u, out, err))
			} else {
				acc.AddError(fmt.Errorf("host %s: %s", u, err))
			}
			fields["result_code"] = 2
			acc.AddFields("ping", fields, tags)
			return
		}
	}
	stats, err := processPingOutput(out)
	if err != nil {
		// fatal error
		acc.AddError(fmt.Errorf("%s: %s", err, u))
		fields["result_code"] = 2
		acc.AddFields("ping", fields, tags)
		return
	}

	// Calculate packet loss percentage
	loss := float64(stats.trans-stats.recv) / float64(stats.trans) * 100.0

	fields["packets_transmitted"] = stats.trans
	fields["packets_received"] = stats.recv
	fields["percent_packet_loss"] = loss
	if stats.ttl >= 0 {
		fields["ttl"] = stats.ttl
	}
	if stats.min >= 0 {
		fields["minimum_response_ms"] = stats.min
	}
	if stats.avg >= 0 {
		fields["average_response_ms"] = stats.avg
	}
	if stats.max >= 0 {
		fields["maximum_response_ms"] = stats.max
	}
	if stats.stddev >= 0 {
		fields["standard_deviation_ms"] = stats.stddev
	}
	acc.AddFields("ping", fields, tags)
}

// args returns the arguments for the 'ping' executable
func (p *Ping) args(url string, system string) []string {
	if len(p.Arguments) > 0 {
		return append(p.Arguments, url)
	}

	// build the ping command args based on toml config
	args := []string{"-c", strconv.Itoa(p.Count), "-n", "-s", "16"}
	if p.PingInterval > 0 {
		args = append(args, "-i", strconv.FormatFloat(p.PingInterval, 'f', -1, 64))
	}
	if p.Timeout > 0 {
		switch system {
		case "darwin":
			args = append(args, "-W", strconv.FormatFloat(p.Timeout*1000, 'f', -1, 64))
		case "freebsd":
			if strings.Contains(p.Binary, "ping6") {
				args = append(args, "-x", strconv.FormatFloat(p.Timeout*1000, 'f', -1, 64))
			} else {
				args = append(args, "-W", strconv.FormatFloat(p.Timeout*1000, 'f', -1, 64))
			}
		case "netbsd", "openbsd":
			args = append(args, "-W", strconv.FormatFloat(p.Timeout*1000, 'f', -1, 64))
		case "linux":
			args = append(args, "-W", strconv.FormatFloat(p.Timeout, 'f', -1, 64))
		default:
			// Not sure the best option here, just assume GNU ping?
			args = append(args, "-W", strconv.FormatFloat(p.Timeout, 'f', -1, 64))
		}
	}
	if p.Deadline > 0 {
		switch system {
		case "freebsd":
			if strings.Contains(p.Binary, "ping6") {
				args = append(args, "-X", strconv.Itoa(p.Deadline))
			} else {
				args = append(args, "-t", strconv.Itoa(p.Deadline))
			}
		case "darwin", "netbsd", "openbsd":
			args = append(args, "-t", strconv.Itoa(p.Deadline))
		case "linux":
			args = append(args, "-w", strconv.Itoa(p.Deadline))
		default:
			// not sure the best option here, just assume gnu ping?
			args = append(args, "-w", strconv.Itoa(p.Deadline))
		}
	}
	if p.Interface != "" {
		switch system {
		case "darwin":
			args = append(args, "-I", p.Interface)
		case "freebsd", "netbsd", "openbsd":
			args = append(args, "-S", p.Interface)
		case "linux":
			args = append(args, "-I", p.Interface)
		default:
			// not sure the best option here, just assume gnu ping?
			args = append(args, "-i", p.Interface)
		}
	}
	args = append(args, url)
	return args
}

// processPingOutput takes in a string output from the ping command, like:
//
//     ping www.google.com (173.194.115.84): 56 data bytes
//     64 bytes from 173.194.115.84: icmp_seq=0 ttl=54 time=52.172 ms
//     64 bytes from 173.194.115.84: icmp_seq=1 ttl=54 time=34.843 ms
//
//     --- www.google.com ping statistics ---
//     2 packets transmitted, 2 packets received, 0.0% packet loss
//     round-trip min/avg/max/stddev = 34.843/43.508/52.172/8.664 ms
//
// It returns (<transmitted packets>, <received packets>, <average response>)
func processPingOutput(out string) (stats, error) {
	stats := stats{
		trans: 0,
		recv:  0,
		ttl:   -1,
		roundTripTimeStats: roundTripTimeStats{
			min:    -1.0,
			avg:    -1.0,
			max:    -1.0,
			stddev: -1.0,
		},
	}

	// Set this error to nil if we find a 'transmitted' line
	err := errors.New("fatal error processing ping output")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		// Reading only first TTL, ignoring other TTL messages
		if stats.ttl == -1 && (strings.Contains(line, "ttl=") || strings.Contains(line, "hlim=")) {
			stats.ttl, err = getTTL(line)
		} else if strings.Contains(line, "transmitted") && strings.Contains(line, "received") {
			stats.trans, stats.recv, err = getPacketStats(line)
			if err != nil {
				return stats, err
			}
		} else if strings.Contains(line, "min/avg/max") {
			stats.roundTripTimeStats, err = checkRoundTripTimeStats(line)
			if err != nil {
				return stats, err
			}
		}
	}
	return stats, err
}

func getPacketStats(line string) (trans int, recv int, err error) {
	trans, recv = 0, 0

	stats := strings.Split(line, ", ")
	// Transmitted packets
	trans, err = strconv.Atoi(strings.Split(stats[0], " ")[0])
	if err != nil {
		return trans, recv, err
	}
	// Received packets
	recv, err = strconv.Atoi(strings.Split(stats[1], " ")[0])
	return trans, recv, err
}

func getTTL(line string) (int, error) {
	ttlLine := regexp.MustCompile(`(ttl|hlim)=(\d+)`)
	ttlMatch := ttlLine.FindStringSubmatch(line)
	return strconv.Atoi(ttlMatch[2])
}

func checkRoundTripTimeStats(line string) (roundTripTimeStats, error) {
	roundTripTimeStats := roundTripTimeStats{
		min:    -1.0,
		avg:    -1.0,
		max:    -1.0,
		stddev: -1.0,
	}

	stats := strings.Split(line, " ")[3]
	data := strings.Split(stats, "/")

	var err error
	roundTripTimeStats.min, err = strconv.ParseFloat(data[0], 64)
	if err != nil {
		return roundTripTimeStats, err
	}
	roundTripTimeStats.avg, err = strconv.ParseFloat(data[1], 64)
	if err != nil {
		return roundTripTimeStats, err
	}
	roundTripTimeStats.max, err = strconv.ParseFloat(data[2], 64)
	if err != nil {
		return roundTripTimeStats, err
	}
	if len(data) == 4 {
		roundTripTimeStats.stddev, err = strconv.ParseFloat(data[3], 64)
		if err != nil {
			return roundTripTimeStats, err
		}
	}
	return roundTripTimeStats, err
}
