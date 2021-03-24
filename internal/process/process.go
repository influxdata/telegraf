package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
)

// Process is a long-running process manager that will restart processes if they stop.
type Process struct {
	Cmd          *exec.Cmd
	Stdin        io.WriteCloser
	Stdout       io.ReadCloser
	Stderr       io.ReadCloser
	ReadStdoutFn func(io.Reader)
	ReadStderrFn func(io.Reader)
	RestartDelay time.Duration
	Log          telegraf.Logger

	name       string
	args       []string
	pid        int32
	cancel     context.CancelFunc
	mainLoopWg sync.WaitGroup
}

// New creates a new process wrapper
func New(command []string) (*Process, error) {
	if len(command) == 0 {
		return nil, errors.New("no command")
	}

	p := &Process{
		RestartDelay: 5 * time.Second,
		name:         command[0],
		args:         []string{},
	}

	if len(command) > 1 {
		p.args = command[1:]
	}

	return p, nil
}

// Start the process. A &Process can only be started once. It will restart itself
// as necessary.
func (p *Process) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	if err := p.cmdStart(); err != nil {
		return err
	}

	p.mainLoopWg.Add(1)
	go func() {
		if err := p.cmdLoop(ctx); err != nil {
			p.Log.Errorf("Process quit with message: %v", err)
		}
		p.mainLoopWg.Done()
	}()

	return nil
}

// Stop is called when the process isn't needed anymore
func (p *Process) Stop() {
	if p.cancel != nil {
		// signal our intent to shutdown and not restart the process
		p.cancel()
	}
	// close stdin so the app can shut down gracefully.
	p.Stdin.Close()
	p.mainLoopWg.Wait()
}

func (p *Process) cmdStart() error {
	p.Cmd = exec.Command(p.name, p.args...)

	var err error
	p.Stdin, err = p.Cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error opening stdin pipe: %w", err)
	}

	p.Stdout, err = p.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error opening stdout pipe: %w", err)
	}

	p.Stderr, err = p.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error opening stderr pipe: %w", err)
	}

	p.Log.Infof("Starting process: %s %s", p.name, p.args)

	if err := p.Cmd.Start(); err != nil {
		return fmt.Errorf("error starting process: %s", err)
	}
	atomic.StoreInt32(&p.pid, int32(p.Cmd.Process.Pid))
	return nil
}

func (p *Process) Pid() int {
	pid := atomic.LoadInt32(&p.pid)
	return int(pid)
}

// cmdLoop watches an already running process, restarting it when appropriate.
func (p *Process) cmdLoop(ctx context.Context) error {
	for {
		err := p.cmdWait(ctx)
		if isQuitting(ctx) {
			p.Log.Infof("Process %s shut down", p.Cmd.Path)
			return nil
		}

		p.Log.Errorf("Process %s exited: %v", p.Cmd.Path, err)
		p.Log.Infof("Restarting in %s...", p.RestartDelay)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(p.RestartDelay):
			// Continue the loop and restart the process
			if err := p.cmdStart(); err != nil {
				return err
			}
		}
	}
}

// cmdWait waits for the process to finish.
func (p *Process) cmdWait(ctx context.Context) error {
	var wg sync.WaitGroup

	if p.ReadStdoutFn == nil {
		p.ReadStdoutFn = defaultReadPipe
	}
	if p.ReadStderrFn == nil {
		p.ReadStderrFn = defaultReadPipe
	}

	processCtx, processCancel := context.WithCancel(context.Background())
	defer processCancel()

	wg.Add(1)
	go func() {
		p.ReadStdoutFn(p.Stdout)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		p.ReadStderrFn(p.Stderr)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		select {
		case <-ctx.Done():
			gracefulStop(processCtx, p.Cmd, 5*time.Second)
		case <-processCtx.Done():
		}
		wg.Done()
	}()

	err := p.Cmd.Wait()
	processCancel()
	wg.Wait()
	return err
}

func isQuitting(ctx context.Context) bool {
	return ctx.Err() != nil
}

func defaultReadPipe(r io.Reader) {
	io.Copy(ioutil.Discard, r)
}
