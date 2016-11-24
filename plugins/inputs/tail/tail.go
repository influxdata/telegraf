package tail

import (
	"fmt"
	"log"
	"sync"

	"github.com/hpcloud/tail"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Tail struct {
	Files         []string
	FromBeginning bool

	tailers []*tail.Tail
	parser  parsers.Parser
	wg      sync.WaitGroup
	acc     telegraf.Accumulator

	sync.Mutex
}

func NewTail() *Tail {
	return &Tail{
		FromBeginning: false,
	}
}

const sampleConfig = `
  ## files to tail.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
  ##   "/var/log/apache.log" -> just tail the apache log file
  ##
  ## See https://github.com/gobwas/glob for more examples
  ##
  files = ["/var/mymetrics.out"]
  ## Read file from beginning.
  from_beginning = false

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (t *Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) Description() string {
	return "Stream a log file, like the tail -f command"
}

func (t *Tail) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (t *Tail) Start(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	t.acc = acc

	var seek tail.SeekInfo
	if !t.FromBeginning {
		seek.Whence = 2
		seek.Offset = 0
	}

	var errS string
	// Create a "tailer" for each file
	for _, filepath := range t.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			log.Printf("E! Error Glob %s failed to compile, %s", filepath, err)
		}
		for file, _ := range g.Match() {
			tailer, err := tail.TailFile(file,
				tail.Config{
					ReOpen:    true,
					Follow:    true,
					Location:  &seek,
					MustExist: true,
				})
			if err != nil {
				errS += err.Error() + " "
				continue
			}
			// create a goroutine for each "tailer"
			t.wg.Add(1)
			go t.receiver(tailer)
			t.tailers = append(t.tailers, tailer)
		}
	}

	if errS != "" {
		return fmt.Errorf(errS)
	}
	return nil
}

// this is launched as a goroutine to continuously watch a tailed logfile
// for changes, parse any incoming msgs, and add to the accumulator.
func (t *Tail) receiver(tailer *tail.Tail) {
	defer t.wg.Done()

	var m telegraf.Metric
	var err error
	var line *tail.Line
	for line = range tailer.Lines {
		if line.Err != nil {
			log.Printf("E! Error tailing file %s, Error: %s\n",
				tailer.Filename, err)
			continue
		}
		m, err = t.parser.ParseLine(line.Text)
		if err == nil {
			t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		} else {
			log.Printf("E! Malformed log line in %s: [%s], Error: %s\n",
				tailer.Filename, line.Text, err)
		}
	}
}

func (t *Tail) Stop() {
	t.Lock()
	defer t.Unlock()

	for _, t := range t.tailers {
		err := t.Stop()
		if err != nil {
			log.Printf("E! Error stopping tail on file %s\n", t.Filename)
		}
		t.Cleanup()
	}
	t.wg.Wait()
}

func (t *Tail) SetParser(parser parsers.Parser) {
	t.parser = parser
}

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return NewTail()
	})
}
