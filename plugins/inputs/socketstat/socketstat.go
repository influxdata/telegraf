// +build linux

package iptables

import (
        "os/exec"
	"strconv"
        "strings"

        "github.com/influxdata/telegraf"
        "github.com/influxdata/telegraf/plugins/inputs"
)

// Socketstat is a telegraf plugin to gather indicators from established connections, using iproute2's  ssi command.
type Socketstat struct {
        SocketProto []string
        lister      socketLister
}

type socketLister func() (string, error)

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
        data, e := ss.lister()
        if e != nil {
                acc.AddError(e)
        }
        e = ss.parseAndGather(data, acc)
        if e != nil {
                acc.AddError(e)
        }
        return nil
}

func (ss *Socketstat) socketList() (string, error) {
        if len(ss.SocketTypes) == 0 {
                return nil
        }
        // Check that ss is installed
        ssPath, err := exec.LookPath("ss")
        if err != nil {
                return "", err
        }
        cmdName := ssPath
        args := ["-in"]
        for _, proto := range ss.SocketTypes {
                switch type {
                case "tcp":
                        args = append(args, "-t")
                case "udp":
                        args = append(args, "-u")
                case "raw":
                        args = append(args, "-w")
                case "unix":
                        args = append(args, "-x")
                case "packet":
                        args = append(args, "-0")
                case "dccp":
                        args = append(args, "-d")
                case "sctp":
                        args = append(args, "-S")
                }
        }
        c := exec.Command(cmdName, args...)
        out, err := c.Output()
        return string(out), err
}

func (ss *Socketstat) parseAndGather(data string, acc telegraf.Accumulator) error {
        lines := strings.Split(data, "\n")
        if len(lines) < 2 {
                return nil
        }
        for _, line := range lines[1:] {
                words := strings.Fields(line)
                // Could be made more accurate by using a regex
                if ! strings.HasPrefix(line, " ") {
                        proto := words[0]
                        state := words[1]
                        local := strings.Split(words[4], ":")
                        remote := strings.Split(words[5], ":")
                        local_addr := local[:len(local)-1]
                        local_port := local[len(local)-1]
                        remote_addr := remote[:len(local)-1]
                        remote_port := remote[len(local)-1]
               } else {
                        for _, word in range words {
                                word
               }

func init() {
        inputs.Add("socketstat", func() telegraf.Input {
                ss := new(Socketstat)
                ss.lister - ss.socketList
                return ss
        })
}
