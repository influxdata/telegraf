package igloo

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// format of timestamps
const (
	rfcFormat string = "%s-%s-%sT%s:%s:%s.%sZ"
)

var (
	// regex for finding timestamps
	tRe = regexp.MustCompile(`Timestamp=((\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2}),(\d+))`)
)

type Tail struct {
	Files         []string
	FromBeginning bool
	TagKeys       []string
	Counters      []string
	NumFields     []string
	StrFields     []string

	numfieldsRe map[string]*regexp.Regexp
	strfieldsRe map[string]*regexp.Regexp
	countersRe  map[string]*regexp.Regexp
	tagsRe      map[string]*regexp.Regexp

	counters map[string]map[string]int64

	tailers []*tail.Tail
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
  ## logfiles to parse.
  ##
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
  ##   "/var/log/apache.log" -> just tail the apache log file
  ##
  ## See https://github.com/gobwas/glob for more examples
  ##
  files = ["$HOME/sample.log"]
  ## Read file from beginning.
  from_beginning = false

  ## Each log message is searched for these tag keys in TagKey=Value format.
  ## Any that are found will be tagged on the resulting influx measurements.
  tag_keys = [
    "HostLocal",
    "ProductName",
    "OperationName",
  ]

  ## counters are keys which are treated as counters.
  ##   so if counters = ["Result"], then this means that the following ocurrence
  ##   on a log line:
  ##     Result=Success
  ##   would be treated as a counter: Result_Success, and it will be incremented
  ##   for every occurrence, until Telegraf is restarted.
  counters   = ["Result"]
  ## num_fields are log line occurrences that are translated into numerical
  ## fields. ie:
  ##   Duration=1
  num_fields = ["Duration", "Attempt"]
  ## str_fields are log line occurences that are translated into string fields,
  ## ie:
  ##   ActivityGUID=0bb03bf4-ae1d-4487-bb6f-311653b35760
  str_fields = ["ActivityGUID"]
`

func (t *Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) Description() string {
	return "Stream an igloo file, like the tail -f command"
}

func (t *Tail) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (t *Tail) buildRegexes() error {
	t.numfieldsRe = make(map[string]*regexp.Regexp)
	t.strfieldsRe = make(map[string]*regexp.Regexp)
	t.tagsRe = make(map[string]*regexp.Regexp)
	t.countersRe = make(map[string]*regexp.Regexp)
	t.counters = make(map[string]map[string]int64)

	for _, field := range t.NumFields {
		re, err := regexp.Compile(field + `=([0-9\.]+)`)
		if err != nil {
			return err
		}
		t.numfieldsRe[field] = re
	}

	for _, field := range t.StrFields {
		re, err := regexp.Compile(field + `=([0-9a-zA-Z\.\-]+)`)
		if err != nil {
			return err
		}
		t.strfieldsRe[field] = re
	}

	for _, field := range t.TagKeys {
		re, err := regexp.Compile(field + `=([0-9a-zA-Z\.\-]+)`)
		if err != nil {
			return err
		}
		t.tagsRe[field] = re
	}

	for _, field := range t.Counters {
		re, err := regexp.Compile("(" + field + ")" + `=([0-9a-zA-Z\.\-]+)`)
		if err != nil {
			return err
		}
		t.countersRe[field] = re
	}

	return nil
}

func (t *Tail) Start(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	t.acc = acc
	if err := t.buildRegexes(); err != nil {
		return err
	}

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
			log.Printf("ERROR Glob %s failed to compile, %s", filepath, err)
		}
		for file, _ := range g.Match() {
			tailer, err := tail.TailFile(file,
				tail.Config{
					ReOpen:   true,
					Follow:   true,
					Location: &seek,
				})
			if err != nil {
				errS += err.Error() + " "
				continue
			}
			// create a goroutine for each "tailer"
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
	t.wg.Add(1)
	defer t.wg.Done()

	var err error
	var line *tail.Line
	for line = range tailer.Lines {
		if line.Err != nil {
			log.Printf("ERROR tailing file %s, Error: %s\n",
				tailer.Filename, err)
			continue
		}
		err = t.Parse(line.Text)
		if err != nil {
			log.Printf("ERROR: %s", err)
		}
	}
}

func (t *Tail) Parse(line string) error {
	// find the timestamp:
	match := tRe.FindAllStringSubmatch(line, -1)
	if len(match) < 1 {
		return nil
	}
	if len(match[0]) < 9 {
		return nil
	}
	// make an rfc3339 timestamp and parse it:
	ts, err := time.Parse(time.RFC3339Nano,
		fmt.Sprintf(rfcFormat, match[0][2], match[0][3], match[0][4], match[0][5], match[0][6], match[0][7], match[0][8]))
	if err != nil {
		return nil
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	// parse numerical fields:
	for name, re := range t.numfieldsRe {
		match := re.FindAllStringSubmatch(line, -1)
		if len(match) < 1 {
			continue
		}
		if len(match[0]) < 2 {
			continue
		}
		num, err := strconv.ParseFloat(match[0][1], 64)
		if err == nil {
			fields[name] = num
		}
	}

	// parse string fields:
	for name, re := range t.strfieldsRe {
		match := re.FindAllStringSubmatch(line, -1)
		if len(match) < 1 {
			continue
		}
		if len(match[0]) < 2 {
			continue
		}
		fields[name] = match[0][1]
	}

	// parse tags:
	for name, re := range t.tagsRe {
		match := re.FindAllStringSubmatch(line, -1)
		if len(match) < 1 {
			continue
		}
		if len(match[0]) < 2 {
			continue
		}
		tags[name] = match[0][1]
	}

	if len(t.countersRe) > 0 {
		// Make a unique key for the measurement name/tags
		var tg []string
		for k, v := range tags {
			tg = append(tg, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(tg)
		hash := fmt.Sprintf("%s%s", strings.Join(tg, ""), "igloo")

		// check if this hash already has a counter map
		_, ok := t.counters[hash]
		if !ok {
			// doesnt have counter map, so make one
			t.counters[hash] = make(map[string]int64)
		}

		// search for counter matches:
		for _, re := range t.countersRe {
			match := re.FindAllStringSubmatch(line, -1)
			if len(match) < 1 {
				continue
			}
			if len(match[0]) < 3 {
				continue
			}
			counterName := match[0][1] + "_" + match[0][2]
			// increment this counter
			t.counters[hash][counterName] += 1
			// add this counter to the output fields
			fields[counterName] = t.counters[hash][counterName]
		}
	}

	t.acc.AddFields("igloo", fields, tags, ts)
	return nil
}

func (t *Tail) Stop() {
	t.Lock()
	defer t.Unlock()

	for _, t := range t.tailers {
		err := t.Stop()
		if err != nil {
			log.Printf("ERROR stopping tail on file %s\n", t.Filename)
		}
		t.Cleanup()
	}
	t.wg.Wait()
}

func init() {
	inputs.Add("igloo", func() telegraf.Input {
		return NewTail()
	})
}
