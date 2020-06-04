package execd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
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
	Command      []string
	RestartDelay config.Duration

	parserConfig     *parsers.Config
	parser           parsers.Parser
	serializerConfig *serializers.Config
	serializer       serializers.Serializer
	acc              telegraf.Accumulator
	inCh             chan telegraf.Metric
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	stdout           io.ReadCloser
	stderr           io.ReadCloser
	cancel           context.CancelFunc
	wg               sync.WaitGroup
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
	e.inCh = make(chan telegraf.Metric)

	if len(e.Command) == 0 {
		return fmt.Errorf("FATAL no command specified")
	}

	e.wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	if err := e.cmdStart(); err != nil {
		return err
	}

	go func() {
		e.cmdLoop(ctx)
		e.wg.Done()
	}()

	return nil
}

func (e *Execd) Add(metric telegraf.Metric) {
	e.inCh <- metric
}

func (e *Execd) Stop() error {
	close(e.inCh)
	e.cancel()
	e.wg.Wait()
	return nil
}

// cmdLoop watches an already running process, restarting it when appropriate.
func (e *Execd) cmdLoop(ctx context.Context) error {
	for {
		// Use a buffered channel to ensure goroutine below can exit
		// if `ctx.Done` is selected and nothing reads on `done` anymore
		done := make(chan error, 1)
		go func() {
			done <- e.cmdWait()
		}()

		select {
		case <-ctx.Done():
			if e.stdin != nil {
				e.stdin.Close()
				// Immediately exit process but with a graceful shutdown
				// period before killing
				internal.WaitTimeout(e.cmd, 5*time.Second)
			}
			return nil
		case err := <-done:
			log.Printf("Process %s terminated: %s", e.Command, err)
			if isQuitting(ctx) {
				return err
			}
		}

		log.Printf("Restarting in %s...", time.Duration(e.RestartDelay))

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(e.RestartDelay)):
			// Continue the loop and restart the process
			if err := e.cmdStart(); err != nil {
				return err
			}
		}
	}
}

func isQuitting(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (e *Execd) cmdStart() (err error) {
	if len(e.Command) > 1 {
		e.cmd = exec.Command(e.Command[0], e.Command[1:]...)
	} else {
		e.cmd = exec.Command(e.Command[0])
	}

	e.stdin, err = e.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Error opening stdin pipe: %s", err)
	}

	e.stdout, err = e.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Error opening stdout pipe: %s", err)
	}

	e.stderr, err = e.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Error opening stderr pipe: %s", err)
	}

	log.Printf("Starting process: %s", e.Command)

	err = e.cmd.Start()
	if err != nil {
		return fmt.Errorf("Error starting process: %s", err)
	}

	return nil
}

func (e *Execd) cmdWait() error {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		e.cmdReadOut(e.stdout)
		wg.Done()
	}()

	go func() {
		e.cmdWriteIn(e.stdin)
		wg.Done()
	}()

	go func() {
		e.cmdReadErr(e.stderr)
		wg.Done()
	}()

	wg.Wait()
	return e.cmd.Wait()
}

func (e *Execd) cmdWriteIn(in io.Writer) {
	for m := range e.inCh {
		b, err := e.serializer.Serialize(m)
		if err != nil {
			log.Println(fmt.Errorf("Metric serializing error: %s", err))
			continue
		}
		_, err = e.stdin.Write(b)
		if err != nil {
			log.Println(fmt.Errorf("Error writing to process stdin: %s", err))
			continue
		}
	}
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
