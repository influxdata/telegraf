package tail

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/hpcloud/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"

	_ "github.com/influxdata/telegraf/internal/encoding/graphite"
	_ "github.com/influxdata/telegraf/internal/encoding/influx"
)

// Tail represents a tail service to
// use line-protocol to read metrics from the file given
type Tail struct {
	Files []string

	DataFormat string

	Separator string
	Templates []string

	mu sync.Mutex

	encodingParser encoding.Parser

	logger *log.Logger

	tailPointers []*tail.Tail

	wg   sync.WaitGroup
	done chan struct{}

	// channel for all incoming parsed points
	metricC chan telegraf.Metric
}

var sampleConfig = `
  ### The file to be monited by this tail plugin
  files = ["/tmp/test","/tmp/test2"]

  # Data format to consume. This can be "influx" or "graphite" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "graphite"

  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."
  
  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### 1. filter + template
  ### 2. filter + template + extra tag
  ### 3. filter + template with field key
  ### 4. default template
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
 ]
`

func (t *Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) Description() string {
	return "Tail read line-protocol metrics from the file given!"
}

// Open starts the Graphite input processing data.
func (t *Tail) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	configs := make(map[string]interface{})
	configs["Separator"] = t.Separator
	configs["Templates"] = t.Templates

	var err error
	t.encodingParser, err = encoding.NewParser(t.DataFormat, configs)

	if err != nil {
		return fmt.Errorf("Tail input configuration is error: %s ", err.Error())
	}

	t.done = make(chan struct{})
	t.metricC = make(chan telegraf.Metric, 50000)
	t.tailPointers = make([]*tail.Tail, len(t.Files))

	for i, fileName := range t.Files {
		t.tailPointers[i], err = t.tailFile(fileName)
		if err != nil {
			return fmt.Errorf("Can not open the file: %s to tail", fileName)
		} else {
			t.logger.Printf("Openning the file: %s to tail", fileName)
		}
	}

	return nil
}

func (t *Tail) tailFile(fileName string) (*tail.Tail, error) {
	tailPointer, err := tail.TailFile(fileName, tail.Config{Follow: true})
	if err != nil {
		return nil, err
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for line := range tailPointer.Lines {
			t.handleLine(strings.TrimSpace(line.Text))
		}
		tailPointer.Wait()
	}()

	return tailPointer, nil

}

func (t *Tail) handleLine(line string) {
	if line == "" {
		return
	}

	// Parse it.
	metric, err := t.encodingParser.ParseLine(line)
	if err != nil {
		t.logger.Printf("unable to parse line: %s: %s", line, err)
		return
	}
	if metric != nil {
		t.metricC <- metric
	}

}

// Close stops all data processing on the Graphite input.
func (t *Tail) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, tailPointer := range t.tailPointers {
		tailPointer.Cleanup()
		tailPointer.Stop()
	}

	close(t.done)
	t.wg.Wait()
	t.done = nil
}

func (t *Tail) Gather(acc telegraf.Accumulator) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	npoints := len(t.metricC)
	for i := 0; i < npoints; i++ {
		metric := <-t.metricC
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}
	return nil
}

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return &Tail{logger: log.New(os.Stderr, "[tail] ", log.LstdFlags)}
	})
}
