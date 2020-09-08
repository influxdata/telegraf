package execd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/process"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const sampleConfig = `
  ## Program to run as daemon
  command = ["telegraf-smartctl", "-d", "/dev/sda"]

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"
`

type Execd struct {
	Command      []string        `toml:"command"`
	RestartDelay config.Duration `toml:"restart_delay"`

	parserConfig     *parsers.Config
	parser           parsers.Parser
	serializerConfig *serializers.Config
	serializer       serializers.Serializer
	acc              telegraf.Accumulator
	process          *process.Process
}

func New() *Execd {
	return &Execd{
		RestartDelay: config.Duration(10 * time.Second),
		parserConfig: &parsers.Config{
			DataFormat: "influx",
		},
		serializerConfig: &serializers.Config{
			DataFormat: "influx",
		},
	}
}

func (e *Execd) SampleConfig() string {
	return sampleConfig
}

func (e *Execd) Description() string {
	return "Run executable as long-running processor plugin"
}

func (e *Execd) Start(acc telegraf.Accumulator) error {
	var err error
	e.parser, err = parsers.NewParser(e.parserConfig)
	if err != nil {
		return fmt.Errorf("error creating parser: %w", err)
	}
	e.serializer, err = serializers.NewSerializer(e.serializerConfig)
	if err != nil {
		return fmt.Errorf("error creating serializer: %w", err)
	}
	e.acc = acc

	if len(e.Command) == 0 {
		return fmt.Errorf("no command specified")
	}

	e.process, err = process.New(e.Command)
	if err != nil {
		return fmt.Errorf("error creating new process: %w", err)
	}

	e.process.RestartDelay = time.Duration(e.RestartDelay)
	e.process.ReadStdoutFn = e.cmdReadOut
	e.process.ReadStderrFn = e.cmdReadErr

	if err = e.process.Start(); err != nil {
		return fmt.Errorf("failed to start process %s: %w", e.Command, err)
	}

	return nil
}

func (e *Execd) Add(m telegraf.Metric, acc telegraf.Accumulator) {
	b, err := e.serializer.Serialize(m)
	if err != nil {
		acc.AddError(fmt.Errorf("metric serializing error: %w", err))
		return
	}

	_, err = e.process.Stdin.Write(b)
	if err != nil {
		acc.AddError(fmt.Errorf("error writing to process stdin: %w", err))
		return
	}

	// We cannot maintain tracking metrics at the moment because input/output
	// is done asynchronously and we don't have any metric metadata to tie the
	// output metric back to the original input metric.
	m.Drop()
}

func (e *Execd) Stop() error {
	e.process.Stop()
	return nil
}

func (e *Execd) cmdReadOut(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		metrics, err := e.parser.Parse(scanner.Bytes())
		if err != nil {
			log.Println(fmt.Errorf("Parse error: %s", err))
		}

		for _, metric := range metrics {
			e.acc.AddMetric(metric)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(fmt.Errorf("Error reading stdout: %s", err))
	}
}

func (e *Execd) cmdReadErr(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		log.Printf("stderr: %q", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println(fmt.Errorf("Error reading stderr: %s", err))
	}
}

func init() {
	processors.AddStreaming("execd", func() telegraf.StreamingProcessor {
		return New()
	})
}
