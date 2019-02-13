package filescan

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/rjeczalik/notify"
)

type Filescan struct {
	Files []string

	filechan chan string
	parser   parsers.Parser
	acc      telegraf.Accumulator
}

const sampleConfig = `
	## Directories to monitor
	directories = ["D:\Data\InputData\DCInputData\Incoming"]
	## Data format to consume. Only influx is supported
	data_format = "influx"
`

func (dm *Filescan) SampleConfig() string {
	return sampleConfig
}

func (dm *Filescan) Description() string {
	return "Monitor a directory for DL Files"
}

func (fs *Filescan) Monitor() error {
	var eventChan = make(chan notify.EventInfo, 10)

	if err := notify.Watch(dir, eventChan, notify.Rename|notify.Create); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(eventChan)

	// Handle event channel. Queue up items if we are not ready.
	for true {
		eventName := <-eventChan
		fileName := strings.Replace(eventName.Path(), "\\", "/", -1)

		if ddo.IsFileMatch(fileName) {
			go ddo.AddToRtQueue(fileName)
		}
	}

	return nil
}

func (fs *Filescan) Start(acc telegraf.Accumulator) error {
	fs.acc = acc
	// Create a monitor for each directory
	for _, entry := range fs.Files {

	}

	return nil
}

func (fs *Filescan) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (fs *Filescan) Stop() {
}

func init() {
	fmt.Println("dirmon starting")
	inputs.Add("dirmon", func() telegraf.Input {
		return &Filescan{}
	})
	fmt.Println("dirmon finished")
}
