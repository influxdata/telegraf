// +build windows

package ping

import (
	"errors"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HostPinger is a function that runs the "ping" function using a list of
// passed arguments. This can be easily switched with a mocked ping function
// for unit test purposes (see ping_test.go)
type HostPinger func(timeout float64, args ...string) (string, error)

type Ping struct {
	// Number of pings to send (ping -c <COUNT>)
	Count int

	// Ping timeout, in seconds. 0 means no timeout (ping -W <TIMEOUT>)
	Timeout float64

	// URLs to ping
	Urls []string

	// host ping function
	pingHost HostPinger
}

func (s *Ping) Description() string {
	return "Ping given url(s) and return statistics"
}

const sampleConfig = `
	## List of urls to ping
	urls = ["www.google.com"]

	## number of pings to send per collection (ping -n <COUNT>)
	# count = 1

	## Ping timeout, in seconds. 0.0 means default timeout (ping -w <TIMEOUT>)
	# timeout = 0.0
`

func (s *Ping) SampleConfig() string {
	return sampleConfig
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

// processPingOutput takes in a string output from the ping command
// based on linux implementation but using regex ( multilanguage support )
// It returns (<transmitted packets>, <received reply>, <received packet>, <average response>, <min response>, <max response>)
func processPingOutput(out string) (int, int, int, int, int, int, error) {
	// So find a line contain 3 numbers except reply lines
	var stats, aproxs []string = nil, nil
	err := errors.New("Fatal error processing ping output")
	stat := regexp.MustCompile(`=\W*(\d+)\D*=\W*(\d+)\D*=\W*(\d+)`)
	aprox := regexp.MustCompile(`=\W*(\d+)\D*ms\D*=\W*(\d+)\D*ms\D*=\W*(\d+)\D*ms`)
	tttLine := regexp.MustCompile(`TTL=\d+`)
	lines := strings.Split(out, "\n")
	var receivedReply int = 0
	for _, line := range lines {
		if tttLine.MatchString(line) {
			receivedReply++
		} else {
			if stats == nil {
				stats = stat.FindStringSubmatch(line)
			}
			if stats != nil && aproxs == nil {
				aproxs = aprox.FindStringSubmatch(line)
			}
		}
	}

	// stats data should contain 4 members: entireExpression + ( Send, Receive, Lost )
	if len(stats) != 4 {
		return 0, 0, 0, -1, -1, -1, err
	}
	trans, err := strconv.Atoi(stats[1])
	if err != nil {
		return 0, 0, 0, -1, -1, -1, err
	}
	receivedPacket, err := strconv.Atoi(stats[2])
	if err != nil {
		return 0, 0, 0, -1, -1, -1, err
	}

	// aproxs data should contain 4 members: entireExpression + ( min, max, avg )
	if len(aproxs) != 4 {
		return trans, receivedReply, receivedPacket, -1, -1, -1, err
	}
	min, err := strconv.Atoi(aproxs[1])
	if err != nil {
		return trans, receivedReply, receivedPacket, -1, -1, -1, err
	}
	max, err := strconv.Atoi(aproxs[2])
	if err != nil {
		return trans, receivedReply, receivedPacket, -1, -1, -1, err
	}
	avg, err := strconv.Atoi(aproxs[3])
	if err != nil {
		return 0, 0, 0, -1, -1, -1, err
	}

	return trans, receivedReply, receivedPacket, avg, min, max, err
}

func (p *Ping) timeout() float64 {
	// According to MSDN, default ping timeout for windows is 4 second
	// Add also one second interval

	if p.Timeout > 0 {
		return p.Timeout + 1
	}
	return 4 + 1
}

// args returns the arguments for the 'ping' executable
func (p *Ping) args(url string) []string {
	args := []string{"-n", strconv.Itoa(p.Count)}

	if p.Timeout > 0 {
		args = append(args, "-w", strconv.FormatFloat(p.Timeout*1000, 'f', 0, 64))
	}

	args = append(args, url)

	return args
}

func (p *Ping) Gather(acc telegraf.Accumulator) error {
	if p.Count < 1 {
		p.Count = 1
	}
	var wg sync.WaitGroup
	errorChannel := make(chan error, len(p.Urls)*2)
	var pendingError error = nil
	// Spin off a go routine for each url to ping
	for _, url := range p.Urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			tags := map[string]string{"url": u}
			fields := map[string]interface{}{"result_code": 0}

			_, err := net.LookupHost(u)
			if err != nil {
				errorChannel <- err
				fields["result_code"] = 1
				acc.AddFields("ping", fields, tags)
				return
			}

			args := p.args(u)
			totalTimeout := p.timeout() * float64(p.Count)
			out, err := p.pingHost(totalTimeout, args...)
			// ping host return exitcode != 0 also when there was no response from host
			// but command was execute successfully
			if err != nil {
				// Combine go err + stderr output
				pendingError = errors.New(strings.TrimSpace(out) + ", " + err.Error())
			}
			trans, recReply, receivePacket, avg, min, max, err := processPingOutput(out)
			if err != nil {
				// fatal error
				if pendingError != nil {
					errorChannel <- pendingError
				}
				errorChannel <- err

				fields["errors"] = 100.0
				acc.AddFields("ping", fields, tags)
				return
			}
			// Calculate packet loss percentage
			lossReply := float64(trans-recReply) / float64(trans) * 100.0
			lossPackets := float64(trans-receivePacket) / float64(trans) * 100.0

			fields["packets_transmitted"] = trans
			fields["reply_received"] = recReply
			fields["packets_received"] = receivePacket
			fields["percent_packet_loss"] = lossPackets
			fields["percent_reply_loss"] = lossReply
			if avg >= 0 {
				fields["average_response_ms"] = float64(avg)
			}
			if min >= 0 {
				fields["minimum_response_ms"] = float64(min)
			}
			if max >= 0 {
				fields["maximum_response_ms"] = float64(max)
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

func init() {
	inputs.Add("ping", func() telegraf.Input {
		return &Ping{
			pingHost: hostPinger,
			Count:    1,
		}
	})
}
