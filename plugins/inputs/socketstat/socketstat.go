package socketstat

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Socketstat is a telegraf plugin to gather indicators from established connections, using iproute2's  `ss` command.
type Socketstat struct {
	lister      socketLister
	SocketProto []string
	Timeout     internal.Duration
}

type socketLister func(proto string, Timeout internal.Duration) (*bytes.Buffer, error)

const measurement = "socketstat"

var defaultTimeout = internal.Duration{Duration: time.Second}

// Description returns a short description of the plugin
func (ss *Socketstat) Description() string {
	return "Gather indicators from established connections, using iproute2's  `ss` command."
}

// SampleConfig returns sample configuration options
func (ss *Socketstat) SampleConfig() string {
	return `
  ## ss can display information about tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## Specify here the types you want to gather
  socket_proto = [ "tcp", "udp" ]
  ## The default timeout of 1s for ss execution can be overridden here:
  # timeout = "1s"
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
		out, e := ss.lister(proto, ss.Timeout)
		if e != nil {
			acc.AddError(e)
			continue
		}
		e = ss.parseAndGather(out, proto, acc)
		if e != nil {
			acc.AddError(e)
			continue
		}
	}
	return nil
}

func socketList(proto string, Timeout internal.Duration) (*bytes.Buffer, error) {
	// Check that ss is installed
	ssPath, err := exec.LookPath("ss")
	if err != nil {
		return new(bytes.Buffer), err
	}

	// Add needed args
	cmdName := ssPath
	var args []string
	args = append(args, "-in")
	args = append(args, "--"+proto)

	// Run ss, return the output as bytes.Buffer
	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running ss -in --%s: ", err)
	}
	return &out, nil
}

var validFields = "(bytes_acked|bytes_received|segs_out|segs_in|data_segs_in|data_segs_out)"
var validValues = regexp.MustCompile("^" + validFields + ":[0-9]+$")
var beginsWithBlank = regexp.MustCompile("^\\s+.*$")

func (ss *Socketstat) parseAndGather(data *bytes.Buffer, proto string, acc telegraf.Accumulator) error {
	scanner := bufio.NewScanner(data)
	tags := map[string]string{}
	fields := make(map[string]interface{})

	// ss output can have blank lines, and/or socket basic info lines and more advanced
	// statistics lines, in turns.
	// We're using the flushData variable to determine if we should add a new measurement
	// or postpone it to a later line

	// The first line is only headers
	scanner.Scan()

	flushData := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		words := strings.Fields(line)

		var err error
		if !beginsWithBlank.MatchString(line) {
			if flushData {
				acc.AddFields(measurement, fields, tags)
				flushData = false
			}
			// Delegate the real parsing to getTagsAndState, which manages various
			// formats depending on the protocol
			tags, fields = getTagsAndState(proto, words)
			flushData = true
		} else {
			for _, word := range words {
				if validValues.MatchString(word) {
					kv := strings.Split(word, ":")
					fields[kv[0]], err = strconv.ParseUint(kv[1], 10, 64)
					if err != nil {
						continue
					}
				}
			}
			acc.AddFields(measurement, fields, tags)
			flushData = false
		}
	}
	if flushData {
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

func getTagsAndState(proto string, words []string) (map[string]string, map[string]interface{}) {
	tags := map[string]string{}
	fields := make(map[string]interface{})
	tags["proto"] = proto
	switch proto {
	case "udp", "raw":
		words = append([]string{"dummy"}, words...)
	case "tcp", "dccp", "sctp":
		fields["state"] = words[0]
	}
	switch proto {
	case "tcp", "udp", "raw", "dccp", "sctp":
		local := strings.Split(words[3], ":")
		remote := strings.Split(words[4], ":")
		tags["local_addr"] = strings.Join(local[:len(local)-1], ":")
		tags["local_port"] = local[len(local)-1]
		tags["remote_addr"] = strings.Join(remote[:len(remote)-1], ":")
		tags["remote_port"] = remote[len(remote)-1]
	case "unix", "packet":
		fields["netid"] = words[0]
		tags["local_addr"] = words[4]
		tags["local_port"] = words[5]
		tags["remote_addr"] = words[6]
		tags["remote_port"] = words[7]
	}
	var err error
	fields["recv_q"], err = strconv.ParseUint(words[1], 10, 64)
	fields["send_q"], err = strconv.ParseUint(words[2], 10, 64)
	if err != nil {
	}
	return tags, fields
}

func init() {
	inputs.Add("socketstat", func() telegraf.Input {
		return &Socketstat{
			lister:  socketList,
			Timeout: defaultTimeout,
		}
	})
}
