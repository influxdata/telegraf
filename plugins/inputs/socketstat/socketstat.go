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
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const measurement = "socketstat"

// Socketstat is a telegraf plugin to gather indicators from established connections, using iproute2's  `ss` command.
type Socketstat struct {
	SocketProto []string        `toml:"protocols"`
	Timeout     config.Duration `toml:"timeout"`
	Log         telegraf.Logger `toml:"-"`

	isNewConnection *regexp.Regexp
	validValues     *regexp.Regexp
	cmdName         string
	lister          socketLister
}

type socketLister func(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error)

// Description returns a short description of the plugin
func (ss *Socketstat) Description() string {
	return "Gather indicators from established connections, using iproute2's `ss` command."
}

// SampleConfig returns sample configuration options
func (ss *Socketstat) SampleConfig() string {
	return `
  ## ss can display information about tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## List of protocol types to collect
  # protocols = [ "tcp", "udp" ]
  ## The default timeout of 1s for ss execution can be overridden here:
  # timeout = "1s"
`
}

// Gather gathers indicators from established connections
func (ss *Socketstat) Gather(acc telegraf.Accumulator) error {
	// best effort : we continue through the protocols even if an error is encountered,
	// but we keep track of the last error.
	for _, proto := range ss.SocketProto {
		out, e := ss.lister(ss.cmdName, proto, ss.Timeout)
		if e != nil {
			acc.AddError(e)
			continue
		}
		e = ss.parseAndGather(acc, out, proto)
		if e != nil {
			acc.AddError(e)
			continue
		}
	}
	return nil
}

func socketList(cmdName string, proto string, timeout config.Duration) (*bytes.Buffer, error) {
	// Run ss for the given protocol, return the output as bytes.Buffer
	args := []string{"-in", "--" + proto}
	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, time.Duration(timeout))
	if err != nil {
		return &out, fmt.Errorf("error running ss -in --%s: %v", proto, err)
	}
	return &out, nil
}

func (ss *Socketstat) parseAndGather(acc telegraf.Accumulator, data *bytes.Buffer, proto string) error {
	scanner := bufio.NewScanner(data)
	tags := map[string]string{}
	fields := make(map[string]interface{})

	// ss output can have blank lines, and/or socket basic info lines and more advanced
	// statistics lines, in turns.
	// In all non-empty lines, we can have metrics, so we need to group those relevant to
	// the same connection.
	// To achieve this, we're using the flushData variable which indicates if we should add
	// a new measurement or postpone it to a later line.

	// The first line is only headers
	scanner.Scan()

	flushData := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		words := strings.Fields(line)

		if ss.isNewConnection.MatchString(line) {
			// A line with starting whitespace means metrics about the current connection.
			// We should never get 2 consecutive such lines. If we do, log a warning and in
			// a best effort, extend the metrics from the 1st line with the metrics of the 2nd
			// one, possibly overwriting.
			for _, word := range words {
				if !ss.validValues.MatchString(word) {
					continue
				}
				// kv will have 2 fields because it matched the regexp
				kv := strings.Split(word, ":")
				v, err := strconv.ParseUint(kv[1], 10, 64)
				if err != nil {
					ss.Log.Infof("Couldn't parse metric %q: %v", word, err)
					continue
				}
				fields[kv[0]] = v
			}
			if !flushData {
				ss.Log.Warnf("Found orphaned metrics: %s", words)
				ss.Log.Warnf("Added them to the last known connection.")
			}
			acc.AddFields(measurement, fields, tags)
			flushData = false
			continue
		}
		// A line with no starting whitespace means we're going to parse a new connection.
		// Flush what we gathered about the previous one, if any.
		if flushData {
			acc.AddFields(measurement, fields, tags)
		}

		// Delegate the real parsing to getTagsAndState, which manages various
		// formats depending on the protocol.
		tags, fields = getTagsAndState(proto, words, ss.Log)

		// This line containted metrics, so record that.
		flushData = true
	}
	if flushData {
		acc.AddFields(measurement, fields, tags)
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
		localIndex := strings.LastIndex(words[3], ":")
		remoteIndex := strings.LastIndex(words[4], ":")
		tags["local_addr"] = words[3][:localIndex]
		tags["local_port"] = words[3][localIndex+1:]
		tags["remote_addr"] = words[4][:remoteIndex]
		tags["remote_port"] = words[4][remoteIndex+1:]
	case "unix", "packet":
		fields["netid"] = words[0]
		tags["local_addr"] = words[4]
		tags["local_port"] = words[5]
		tags["remote_addr"] = words[6]
		tags["remote_port"] = words[7]
	}
	var err error
	fields["recv_q"], err = strconv.ParseUint(words[1], 10, 64)
	if err != nil {
		log.Warnf("Couldn't read recv_q in: %s", words)
	}
	fields["send_q"], err = strconv.ParseUint(words[2], 10, 64)
	if err != nil {
		log.Warnf("Couldn't read send_q in: %s", words)
	}
	return tags, fields
}

func (ss *Socketstat) Init() error {
	if len(ss.SocketProto) == 0 {
		ss.SocketProto = []string{"tcp", "udp"}
	}

	// Check that ss is installed, get its path
	ssPath, err := exec.LookPath("ss")
	if err != nil {
		return err
	}
	ss.cmdName = ssPath

	if ss.Timeout < config.Duration(time.Second) {
		ss.Timeout = config.Duration(time.Second)
	}

	// Initialize regexps to validate input data
	validFields := "(bytes_acked|bytes_received|segs_out|segs_in|data_segs_in|data_segs_out)"
	ss.validValues = regexp.MustCompile("^" + validFields + ":[0-9]+$")
	ss.isNewConnection = regexp.MustCompile(`^\s+.*$`)

	ss.lister = socketList

	return nil
}

func init() {
	inputs.Add("socketstat", func() telegraf.Input {
		return &Socketstat{}
	})
}
