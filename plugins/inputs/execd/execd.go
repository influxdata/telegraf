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
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

const sampleConfig = `
  ## Program to run as daemon
  command = ["telegraf-smartctl", "-d", "/dev/sda"]

  ## Define how the process is signaled on each collection interval.
  ## Valid values are:
  ##   "none"   : Do not signal anything.
  ##              The process must output metrics by itself.
  ##   "STDIN"   : Send a newline on STDIN.
  ##   "SIGHUP"  : Send a HUP signal. Not available on Windows.
  ##   "SIGUSR1" : Send a USR1 signal. Not available on Windows.
  ##   "SIGUSR2" : Send a USR2 signal. Not available on Windows.
  signal = "none"

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

type Execd struct {
	Command      []string
	Signal       string
	RestartDelay config.Duration

	acc    telegraf.Accumulator
	cmd    *exec.Cmd
	parser parsers.Parser
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (e *Execd) SampleConfig() string {
	return sampleConfig
}

func (e *Execd) Description() string {
	return "Run executable as long-running input plugin"
}

func (e *Execd) SetParser(parser parsers.Parser) {
	e.parser = parser
}

func (e *Execd) Start(acc telegraf.Accumulator) error {
	e.acc = acc

	if len(e.Command) == 0 {
		return fmt.Errorf("FATAL no command specified")
	}

	e.wg.Add(1) // for the main loop

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

func (e *Execd) Stop() {
	e.cancel()
	e.wg.Wait()
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
				internal.WaitTimeout(e.cmd, 200*time.Millisecond)
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
	wg.Add(2)

	go func() {
		e.cmdReadOut(e.stdout)
		wg.Done()
	}()

	go func() {
		e.cmdReadErr(e.stderr)
		wg.Done()
	}()

	wg.Wait()
	return e.cmd.Wait()
}

func (e *Execd) cmdReadOut(out io.Reader) {
	if _, isInfluxParser := e.parser.(*influx.Parser); isInfluxParser {
		// work around the lack of built-in streaming parser. :(
		e.cmdReadOutStream(out)
		return
	}

	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		metrics, err := e.parser.Parse(scanner.Bytes())
		if err != nil {
			e.acc.AddError(fmt.Errorf("Parse error: %s", err))
		}

		for _, metric := range metrics {
			e.acc.AddMetric(metric)
		}
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("Error reading stdout: %s", err))
	}
}

func (e *Execd) cmdReadOutStream(out io.Reader) {
	parser := influx.NewStreamParser(out)

	for {
		metric, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				break // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				// parse error.
				e.acc.AddError(parseErr)
				continue
			}
			// some non-recoverable error?
			e.acc.AddError(err)
			return
		}

		e.acc.AddMetric(metric)
	}
}

func (e *Execd) cmdReadErr(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		log.Printf("stderr: %q", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("Error reading stderr: %s", err))
	}
}

func init() {
	inputs.Add("execd", func() telegraf.Input {
		return &Execd{
			Signal:       "none",
			RestartDelay: config.Duration(10 * time.Second),
		}
	})
}
