// +build linux

package netstat

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/net/tcp_states.h#n12
var tcpStatus = map[string]string{
	"01": "tcp_established",
	"02": "tcp_syn_sent",
	"03": "tcp_syn_recv",
	"04": "tcp_fin_wait1",
	"05": "tcp_fin_wait2",
	"06": "tcp_time_wait",
	"07": "tcp_close",
	"08": "tcp_close_wait",
	"09": "tcp_last_ack",
	"0A": "tcp_listen",
	"0B": "tcp_closing",
}

type NetStats struct {
	Log telegraf.Logger
}

func (_ *NetStats) Description() string {
	return "Read TCP metrics such as established, time wait and sockets counts."
}

func (_ *NetStats) SampleConfig() string {
	return ""
}

func (s *NetStats) Gather(acc telegraf.Accumulator) error {
	s.gatherTCP(acc)
	s.gatherUDP(acc)
	return nil
}

func (s *NetStats) gatherTCP(acc telegraf.Accumulator) {

	counts := map[string]int{
		"tcp_none": 0,
	}
	tags := map[string]string{}
	for _, v := range tcpStatus {
		counts[v] = 0
	}

	for _, i := range []string{"/net/tcp", "/net/tcp6"} {
		path := filepath.Join(getHostProc(), i)
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				s.Log.Debugf("file [%s] not exist, ignoring", path)
			} else {
				s.Log.Warnf("open file [%s] error: %s", path, err.Error())
			}
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.Fields(scanner.Text())
			if line[3] == "st" {
				continue // skip first line
			}
			if v, ok := tcpStatus[line[3]]; ok {
				counts[v]++
			} else {
				counts["tcp_none"]++
			}
		}
	}

	fields := map[string]interface{}{}
	for k, v := range counts {
		fields[k] = v
	}

	acc.AddGauge("netstat", fields, tags)

}

func (s *NetStats) gatherUDP(acc telegraf.Accumulator) {
	count := 0
	tags := map[string]string{}

	for _, i := range []string{"/net/udp", "/net/udp6"} {
		path := filepath.Join(getHostProc(), i)
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				s.Log.Debugf("file [%s] not exist, ignoring", path)
			} else {
				s.Log.Warnf("open file [%s] error: %s", path, err.Error())
			}
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.Fields(scanner.Text())
			if line[3] == "st" {
				continue // skip first line
			}
			count++
		}
	}

	acc.AddGauge("netstat", map[string]interface{}{"udp_socket": count}, tags)
}

func getHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func init() {
	inputs.Add("netstat", func() telegraf.Input {
		return &NetStats{}
	})
}
