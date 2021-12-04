package socketstat

import (
	"bufio"
	"bytes"
        "errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Socketstat is a telegraf plugin to gather indicators from established connections, using iproute2's  `ss` command.
type Socketstat struct {
        BeginsWithBlank  *regexp.Regexp
        CmdName          string
	Lister           socketLister
        Log              telegraf.Logger
        Measurement      string
	SocketProto      []string
	Timeout          config.Duration
        ValidValues      *regexp.Regexp
}

type socketLister func(CmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error)

// Description returns a short description of the plugin
func (ss *Socketstat) Description() string {
	return "Gather indicators from established connections, using iproute2's `ss` command."
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
	// best effort : we continue through the protocols even if an error is encountered,
	// but we keep track of the last error.
	for _, proto := range ss.SocketProto {
		out, e := ss.Lister(ss.CmdName, proto, ss.Timeout)
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

func socketList(CmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
	// Run ss for the given protocol, return the output as bytes.Buffer
        args := []string{"-in", "--"+proto}
	cmd := exec.Command(CmdName, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, time.Duration(timeout))
	if err != nil {
		return &out, fmt.Errorf("error running ss -in --%s: ", err)
	}
	return &out, nil
}

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
		if !ss.BeginsWithBlank.MatchString(line) {
			if flushData {
				acc.AddFields(ss.Measurement, fields, tags)
				flushData = false
			}
			// Delegate the real parsing to getTagsAndState, which manages various
			// formats depending on the protocol
			tags, fields = getTagsAndState(proto, words, ss.Log)
			flushData = true
		} else {
			for _, word := range words {
				if ss.ValidValues.MatchString(word) {
                                        // kv will have 2 fields because it matched the regexp
					kv := strings.Split(word, ":")
					fields[kv[0]], err = strconv.ParseUint(kv[1], 10, 64)
					if err != nil {
                                                ss.Log.Infof("Couldn't parse metric: %s", word)
						continue
					}
				}
			}
			acc.AddFields(ss.Measurement, fields, tags)
			flushData = false
		}
	}
	if flushData {
		acc.AddFields(ss.Measurement, fields, tags)
	}
	return nil
}

func getTagsAndState(proto string, words []string, log telegraf.Logger) (map[string]string, map[string]interface{}) {
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
		// Local and remote addresses are fields 3 and 4
		// Separate addresses and ports with the last ':'
		local_index := strings.LastIndex(words[3], ":")
		remote_index := strings.LastIndex(words[4], ":")
		tags["local_addr"] = words[3][:local_index]
		tags["local_port"] = words[3][local_index+1:]
		tags["remote_addr"] = words[4][:remote_index]
		tags["remote_port"] = words[4][remote_index+1:]
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
                log.Infof("Couldn't read recv_q and send_q in: %s", words)
	}
	return tags, fields
}

func (ss *Socketstat) Init() error {

        if len(ss.SocketProto) == 0 {
            return errors.New("Error: no protocol specified for the socketstat input plugin.")
        }

        // Check that ss is installed, get its path
        ssPath, err := exec.LookPath("ss")
        if err != nil {
                return err
        }
        ss.CmdName = ssPath

        ss.Measurement = "socketstat"

        if ss.Timeout < config.Duration(time.Second) {
                ss.Timeout = config.Duration(time.Second)
        }

        // Initialize regexps to validate input data
        validFields := "(bytes_acked|bytes_received|segs_out|segs_in|data_segs_in|data_segs_out)"
        ss.ValidValues = regexp.MustCompile("^" + validFields + ":[0-9]+$")
        ss.BeginsWithBlank = regexp.MustCompile("^\\s+.*$")

        return nil

}

func init() {
	inputs.Add("socketstat", func() telegraf.Input {
		return &Socketstat{
			Lister:  socketList,
		}
	})
}
