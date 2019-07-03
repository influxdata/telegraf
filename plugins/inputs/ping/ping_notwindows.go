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
		// Some implementations of ping return a 1 exit code on
		// timeout, if this occurs we will not exit and try to parse
		// the output.
		status := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
				status = ws.ExitStatus()
				fields["result_code"] = status
			}
		}

		if status != 1 {
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
	trans, rec, ttl, min, avg, max, stddev, err := processPingOutput(out)
	if err != nil {
		// fatal error
		acc.AddError(fmt.Errorf("%s: %s", err, u))
		fields["result_code"] = 2
		acc.AddFields("ping", fields, tags)
		return
	}

	// Calculate packet loss percentage
	loss := float64(trans-rec) / float64(trans) * 100.0

	fields["packets_transmitted"] = trans
	fields["packets_received"] = rec
	fields["percent_packet_loss"] = loss
	if ttl >= 0 {
		fields["ttl"] = ttl
	}
	if min >= 0 {
		fields["minimum_response_ms"] = min
	}
	if avg >= 0 {
		fields["average_response_ms"] = avg
	}
	if max >= 0 {
		fields["maximum_response_ms"] = max
	}
	if stddev >= 0 {
		fields["standard_deviation_ms"] = stddev
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
		case "freebsd", "netbsd", "openbsd":
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
		case "darwin", "freebsd", "netbsd", "openbsd":
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
func processPingOutput(out string) (int, int, int, float64, float64, float64, float64, error) {
	var trans, recv, ttl int = 0, 0, -1
	var min, avg, max, stddev float64 = -1.0, -1.0, -1.0, -1.0
	// Set this error to nil if we find a 'transmitted' line
	err := errors.New("Fatal error processing ping output")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		// Reading only first TTL, ignoring other TTL messages
		if ttl == -1 && strings.Contains(line, "ttl=") {
			ttl, err = getTTL(line)
		} else if strings.Contains(line, "transmitted") &&
			strings.Contains(line, "received") {
			trans, recv, err = getPacketStats(line, trans, recv)
			if err != nil {
				return trans, recv, ttl, min, avg, max, stddev, err
			}
		} else if strings.Contains(line, "min/avg/max") {
			min, avg, max, stddev, err = checkRoundTripTimeStats(line, min, avg, max, stddev)
			if err != nil {
				return trans, recv, ttl, min, avg, max, stddev, err
			}
		}
	}
	return trans, recv, ttl, min, avg, max, stddev, err
}

func getPacketStats(line string, trans, recv int) (int, int, error) {
	stats := strings.Split(line, ", ")
	// Transmitted packets
	trans, err := strconv.Atoi(strings.Split(stats[0], " ")[0])
	if err != nil {
		return trans, recv, err
	}
	// Received packets
	recv, err = strconv.Atoi(strings.Split(stats[1], " ")[0])
	return trans, recv, err
}

func getTTL(line string) (int, error) {
	ttlLine := regexp.MustCompile(`ttl=(\d+)`)
	ttlMatch := ttlLine.FindStringSubmatch(line)
	return strconv.Atoi(ttlMatch[1])
}

func checkRoundTripTimeStats(line string, min, avg, max,
	stddev float64) (float64, float64, float64, float64, error) {
	stats := strings.Split(line, " ")[3]
	data := strings.Split(stats, "/")

	min, err := strconv.ParseFloat(data[0], 64)
	if err != nil {
		return min, avg, max, stddev, err
	}
	avg, err = strconv.ParseFloat(data[1], 64)
	if err != nil {
		return min, avg, max, stddev, err
	}
	max, err = strconv.ParseFloat(data[2], 64)
	if err != nil {
		return min, avg, max, stddev, err
	}
	if len(data) == 4 {
		stddev, err = strconv.ParseFloat(data[3], 64)
		if err != nil {
			return min, avg, max, stddev, err
		}
	}
	return min, avg, max, stddev, err
}
