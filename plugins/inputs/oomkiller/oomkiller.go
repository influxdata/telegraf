package oomkiller

import (
	"log"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/tail"
)

type Oomkiller struct {
	Logfile     string `toml:"logfile"`
}

var sampleConfig = `
  ## Logfile where oom killer is reflected
  logfile = "/var/log/messages"
`

// SampleConfig returns a sample configuration block
func (a *Oomkiller) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the Mesos plugin
func (a *Oomkiller) Description() string {
	return "Telegraf plugin for gathering metrics from kernel Oomkiller"
}

func (a *Oomkiller) SetDefaults() {
	if a.Logfile == "" {
		log.Println("I! [oomkiller] Missing logfile value, setting default value (/var/log/messages)")
		a.Logfile = "/var/log/messages"
	}
}

func IsOomkillerEvent(line string) bool {
	matched, _ := regexp.MatchString("invoked oom-killer", line)
	return matched
}

// Gather() metrics from Oomkiller
func (a *Oomkiller) Gather(acc telegraf.Accumulator) error {
	a.SetDefaults()
	//This is to NOT start tailing log from the beginning
	var seek *tail.SeekInfo
	seek = &tail.SeekInfo{ Whence: 2, Offset: 0, }
	t, _ := tail.TailFile(a.Logfile, tail.Config{ Follow: true, Location:  seek, ReOpen: true})
	for line := range t.Lines {
		if IsOomkillerEvent(line.Text) && line.Text != "" {
			s := strings.Split(line.Text, " ")
			invokedBy := s[6]
			var metric string = "event"
			var value int = 1
			data := make(map[string]interface{})
			data[metric] = value
			tags := make(map[string]string)
			tags["invokedby"] = invokedBy
			acc.AddFields("oomkiller", data, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("oomkiller", func() telegraf.Input {
		return &Oomkiller{}
	})
}
