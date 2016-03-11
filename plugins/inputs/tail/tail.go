package tail

import (
	"sync"

	tailfile "github.com/hpcloud/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const sampleConfig = `
  ## The measurement name
  name_override = "nginx_access"

  ## A LTSV formatted log file path.
  ## See http://ltsv.org/ for Labeled Tab-separated Values (LTSV)
  ## Here is an example config for nginx (http://nginx.org/en/).
  ##
  ##  log_format  ltsv  'time:$time_iso8601\t'
  ##                    'host:$host\t'
  ##                    'http_host:$http_host\t'
  ##                    'scheme:$scheme\t'
  ##                    'remote_addr:$remote_addr\t'
  ##                    'remote_user:$remote_user\t'
  ##                    'request:$request\t'
  ##                    'status:$status\t'
  ##                    'body_bytes_sent:$body_bytes_sent\t'
  ##                    'http_referer:$http_referer\t'
  ##                    'http_user_agent:$http_user_agent\t'
  ##                    'http_x_forwarded_for:$http_x_forwarded_for\t'
  ##                    'request_time:$request_time';
  ##  access_log  /var/log/nginx/access.ltsv.log  ltsv;
  ##
  filename = "/var/log/nginx/access.ltsv.log"

  ## Reopen recreated files (tail -F)
  re_open = true

  ## Fail early if the file does not exist
  must_exist = false

  ## Poll for file changes instead of using inotify
  poll = false

  ## Set this to true if the file is a named pipe (mkfifo)
  pipe = false

  ## If non-zero, split longer lines into multiple lines
  max_line_size = 0

  ## Set this false to enable logging to stderr, true to disable logging
  disable_logging = false

  ## Data format to consume. Currently only "ltsv" is supported.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "ltsv"

  ## Time label to be used to create a timestamp for a measurement.
  time_label = "time"

  ## Time format for parsing timestamps.
  ## Please see https://golang.org/pkg/time/#Parse for the format string.
  time_format = "2006-01-02T15:04:05Z07:00"

  ## Labels for string fields.
  str_field_labels = []

  ## Labels for integer (64bit signed decimal integer) fields.
  ## For acceptable integer values, please refer to:
  ## https://golang.org/pkg/strconv/#ParseInt
  int_field_labels = ["body_bytes_sent"]

  ## Labels for float (64bit float) fields.
  ## For acceptable float values, please refer to:
  ## https://golang.org/pkg/strconv/#ParseFloat
  float_field_labels = ["request_time"]

  ## Labels for boolean fields.
  ## For acceptable boolean values, please refer to:
  ## https://golang.org/pkg/strconv/#ParseBool
  bool_field_labels = []

  ## Labels for tags to be added
  tag_labels = ["host", "http_host", "scheme", "remote_addr", "remote_user", "request", "status", "http_referer", "http_user_agent", "http_x_forwarded_for"]

  ## Method to modify duplicated measurement points.
  ## Must be one of "add_uniq_tag", "increment_time", "no_op".
  ## This will be used to modify duplicated points.
  ## For detail, please see https://docs.influxdata.com/influxdb/v0.10/troubleshooting/frequently_encountered_issues/#writing-duplicate-points
  ## NOTE: For modifier methods other than "no_op" to work correctly, the log lines
  ## MUST be sorted by timestamps in ascending order.
  duplicate_points_modifier_method = "add_uniq_tag"

  ## When duplicate_points_modifier_method is "increment_time",
  ## this will be added to the time of the previous measurement
  ## if the time of current time is equal to or less than the
  ## time of the previous measurement.
  ##
  ## NOTE: You need to set this value equal to or greater than
  ## precisions of your output plugins. Otherwise the times will
  ## become the same value!
  ## For the precision of the InfluxDB plugin, please see
  ## https://github.com/influxdata/telegraf/blob/v0.10.1/plugins/outputs/influxdb/influxdb.go#L40-L42
  ## For the duration string format, please see
  ## https://golang.org/pkg/time/#ParseDuration
  duplicate_points_increment_duration = "1us"

  ## When duplicate_points_modifier_method is "add_uniq_tag",
  ## this will be the label of the tag to be added to ensure uniqueness of points.
  ## NOTE: The uniq tag will be only added to the successive points of duplicated
  ## points, it will not be added to the first point of duplicated points.
  ## If you want to always add the uniq tag, add a tag with the same name as
  ## duplicate_points_modifier_uniq_tag and the string value "0" to [inputs.tail.tags].
  duplicate_points_modifier_uniq_tag = "uniq"

  ## Defaults tags to be added to measurements.
  [inputs.tail.tags]
    log_host = "log.example.com"
`

type Tail struct {
	Filename string

	// File-specfic
	ReOpen    bool // Reopen recreated files (tail -F)
	MustExist bool // Fail early if the file does not exist
	Poll      bool // Poll for file changes instead of using inotify
	Pipe      bool // Is a named pipe (mkfifo)

	// Generic IO
	MaxLineSize int // If non-zero, split longer lines into multiple lines

	DisableLogging bool // If false, logs are printed to stderr

	sync.Mutex
	done chan struct{}

	acc    telegraf.Accumulator
	parser parsers.Parser
	tail   *tailfile.Tail
}

func (t *Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) Description() string {
	return "Read a log file like the BSD tail command"
}

func (t *Tail) SetParser(parser parsers.Parser) {
	t.parser = parser
}

// Start a tail log reader. Caller must call *Tail.Stop() to clean up.
func (t *Tail) Start(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	t.acc = acc
	t.done = make(chan struct{})

	config := tailfile.Config{
		ReOpen:      t.ReOpen,
		MustExist:   t.MustExist,
		Poll:        t.Poll,
		Pipe:        t.Pipe,
		Follow:      true,
		MaxLineSize: t.MaxLineSize,
	}
	if t.DisableLogging {
		config.Logger = tailfile.DiscardingLogger
	}
	tf, err := tailfile.TailFile(t.Filename, config)
	if err != nil {
		return err
	}
	t.tail = tf

	// Start the log file reader
	go t.receiver()
	t.tail.Logger.Printf("Started a tail log reader, filename: %s\n", t.Filename)

	return nil
}

func (t *Tail) receiver() {
	for {
		for line := range t.tail.Lines {
			if err := line.Err; err != nil {
				t.tail.Logger.Printf("error while reading from %s, error: %s\n", t.Filename, err.Error())
			} else {
				metric, err := t.parser.ParseLine(line.Text)
				if err != nil {
					t.tail.Logger.Printf("error while parsing from %s, error: %s\n", t.Filename, err.Error())
				}
				t.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
			}

			select {
			case <-t.done:
				t.tail.Done()
				return
			default:
				// Start reading lines again
			}
		}
	}
}

func (t *Tail) Stop() {
	t.Lock()
	close(t.done)
	t.Unlock()
}

// All the work is done in the Start() function, so this is just a dummy
// function.
func (t *Tail) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return &Tail{}
	})
}
