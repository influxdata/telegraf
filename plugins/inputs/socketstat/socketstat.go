// +build linux

package socketstat

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Socketstat is a telegraf plugin to gather indicators from established connections, using iproute2's  ssi command.
type Socketstat struct {
	SocketProto []string
	lister      socketLister
}

type socketLister func(proto string) (string, error)

const measurement = "socketstat"

// Description returns a short description of the plugin
func (ss *Socketstat) Description() string {
	return "Gather indicators from established connections, using iproute2's  ssi command."
}

// SampleConfig returns sample configuration options
func (ss *Socketstat) SampleConfig() string {
	return `
  ## ss can display information about tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## Specify here the types you want to gather
  socket_proto = [ "tcp", "udp", "raw" ]
`
}

// Gather gathers indicators from established connections
func (ss *Socketstat) Gather(acc telegraf.Accumulator) error {
	if len(ss.SocketProto) == 0 {
		return nil
	}
	// best effort : we continue through the protocols even if an error is encountered,
	// but we keep track of the last error.
	for _, proto := range ss.SocketProto {
		data, e := ss.lister(proto)
		if e != nil {
			acc.AddError(e)
			continue
		}
		e = ss.parseAndGather(data, proto, acc)
		if e != nil {
			acc.AddError(e)
			continue
		}
	}
	return nil
}

func (ss *Socketstat) socketList(proto string) (string, error) {
	// Check that ss is installed
	ssPath, err := exec.LookPath("ss")
	if err != nil {
		return "", err
	}
	cmdName := ssPath
	var args []string
	args = append(args, "-in")
	args = append(args, "--"+proto)
	c := exec.Command(cmdName, args...)
	out, err := c.Output()
	return string(out), err
}

var validFields = "(bytes_acked|bytes_received|segs_out|segs_in|data_segs_in|data_segs_out)"
var validValues = regexp.MustCompile("^" + validFields + ":[0-9]+$")
var beginsWithBlank = regexp.MustCompile("^\\s+.*$")

func (ss *Socketstat) parseAndGather(data, proto string, acc telegraf.Accumulator) error {
	lines := strings.Split(data, "\n")
	if len(lines) < 2 {
		return nil
	}
	tags := map[string]string{}
	fields := make(map[string]interface{})
	for _, line := range lines[1:] {
		words := strings.Fields(line)
		if (! beginsWithBlank.MatchString(line) && line != "") {
			fields["state"] = words[0]
			fields["recv_q"] = words[1]
			fields["send_q"] = words[2]
			local := strings.Split(words[3], ":")
			remote := strings.Split(words[4], ":")
			tags["proto"] = proto
			tags["local_addr"] = strings.Join(local[:len(local)-1], ":")
			tags["local_port"] = local[len(local)-1]
			tags["remote_addr"] = strings.Join(remote[:len(remote)-1], ":")
			tags["remote_port"] = remote[len(remote)-1]
		} else {
			for _, word := range words {
				if validValues.MatchString(word) {
					kv := strings.Split(word, ":")
					fields[kv[0]] = kv[1]
				}
			}
		}
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("socketstat", func() telegraf.Input {
		ss := new(Socketstat)
		ss.lister = ss.socketList
		return ss
	})
}
