//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && amd64

package turbostat

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/jpillora/backoff"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/turbostat/parser"
)

//go:embed sample.conf
var sampleConfig string

type Turbostat struct {
	Command []string        `toml:"command"`
	Log     telegraf.Logger `toml:"-"`

	cancel context.CancelFunc
}

func (*Turbostat) Gather(telegraf.Accumulator) error {
	return nil
}

func (*Turbostat) SampleConfig() string {
	return sampleConfig
}

// Start starts the plugin.
// Maintains a Turbostat process and parses metrics from stdout.
func (t *Turbostat) Start(a telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	go func() {
		retryDelay := &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    30 * time.Second,
			Factor: 2,
			Jitter: false,
		}
		for {
			t.Log.Info("Starting Turbostat")
			startedAt := time.Now()
			err := t.runTurbostat(ctx, a)
			t.Log.Errorf("Turbostat encountered an error: %s", err)
			// Exit the goroutine if the context was cancelled.
			select {
			case <-ctx.Done():
				return
			default:
			}
			// If Turbostat ran for more than 30s, there likely was
			// no major issue and we reset the backoff to the minimum.
			if time.Since(startedAt) > 30*time.Second {
				retryDelay.Reset()
			}
			delay := retryDelay.Duration()
			t.Log.Infof("Restarting Turbostat in %.1fs", delay.Seconds())
			// Wait until the delay elapses or the context is cancelled.
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}
	}()
	return nil
}

// Stop stops the plugin.
func (t *Turbostat) Stop() {
	t.cancel()
}

func init() {
	inputs.Add("turbostat", func() telegraf.Input {
		return &Turbostat{}
	})
}

// Reads error lines from Turbostat stderr and logs them.
func processStderr(r io.Reader, log telegraf.Logger) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		log.Info(line)
	}
	return s.Err()
}

// Starts a Turbostat process. One goroutine parses stdout for metrics and
// writes them to the accumulator. Another goroutine monitors stderr for
// messages and logs them. If an error occurs, the process and all goroutines
// terminate, and the function returns the first error encountered. Since
// Turbostat is not expected to terminate, this function never returns nil.
func (t *Turbostat) runTurbostat(c context.Context, a telegraf.Accumulator) error {
	if len(t.Command) < 1 {
		return errors.New("no command provided")
	}
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command[0], t.Command[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not start process: %w", err)
	}
	var once sync.Once
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err = parser.ProcessStream(stdout, a)
		if err != nil {
			once.Do(func() {
				errCh <- fmt.Errorf("error processing stdout: %w", err)
			})
			cancel()
		}
	}()
	go func() {
		defer wg.Done()
		err = processStderr(stderr, t.Log)
		if err != nil {
			once.Do(func() {
				errCh <- fmt.Errorf("error processing stderr: %w", err)
			})
			cancel()
		}
	}()
	wg.Wait()
	cmdErr := cmd.Wait()
	// Return the error on the channel if any.
	select {
	case err := <-errCh:
		return err
	default:
	}
	// Return the process error if any.
	if cmdErr != nil {
		return cmdErr
	}
	// No specific error occurred.
	return errors.New("process exited")
}
