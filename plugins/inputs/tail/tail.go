package tail

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/hpcloud/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Tail represents a tail service to
// use line-protocol to read metrics from the file given
type Tail struct {
	Files     []string
	Separator string
	Tags      []string
	Templates []string

	mu sync.Mutex

	parser       *Parser
	logger       *log.Logger
	config       *Config
	tailPointers []*tail.Tail

	wg   sync.WaitGroup
	done chan struct{}

	// channel for all incoming parsed points
	metricC chan telegraf.Metric
}

var sampleConfig = `
  ### The file to be monited by this tail plugin
  files = ["/tmp/test","/tmp/test2"]

  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."
  
  ### Default tags that will be added to all metrics.  These can be overridden at the template level
  ### or by tags extracted from metric
  tags = ["region=north-china", "zone=1c"]
  
  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### filter + template
  ### filter + template + extra tag
  ### filter + template with field key
  ### default template. Ignore the first graphite component "servers"
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

	c := &Config{
		Files:     t.Files,
		Separator: t.Separator,
		Tags:      t.Tags,
		Templates: t.Templates,
	}
	c.WithDefaults()
	if err := c.Validate(); err != nil {
		return fmt.Errorf("Graphite input configuration is error! ", err.Error())
	}
	t.config = c

	parser, err := NewParserWithOptions(Options{
		Templates:   t.config.Templates,
		DefaultTags: t.config.DefaultTags(),
		Separator:   t.config.Separator})
	if err != nil {
		return fmt.Errorf("Graphite input parser config is error! ", err.Error())
	}
	t.parser = parser

	t.done = make(chan struct{})
	t.metricC = make(chan telegraf.Metric, 10000)
	t.tailPointers = make([]*tail.Tail, len(t.Files))

	for i, fileName := range t.Files {
		t.tailPointers[i], err = t.tailFile(fileName)
		if err != nil {
			fmt.Errorf("Can not open the file: %s to tail", fileName)
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
	mertic, err := t.parser.Parse(line)
	if err != nil {
		switch err := err.(type) {
		case *UnsupposedValueError:
			// Graphite ignores NaN values with no error.
			if math.IsNaN(err.Value) {
				return
			}
		}
		t.logger.Printf("unable to parse line: %s: %s", line, err)
		return
	}

	t.metricC <- mertic
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
