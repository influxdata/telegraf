//go:build !windows

package process

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

// test that a restarting process resets pipes properly
func TestRestartingRebindsPipes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long running test in short mode")
	}

	exe, err := os.Executable()
	require.NoError(t, err)

	p, err := New([]string{exe, "-external"}, []string{"INTERNAL_PROCESS_MODE=application"})
	p.RestartDelay = 100 * time.Nanosecond
	p.Log = testutil.Logger{}
	require.NoError(t, err)

	linesRead := int64(0)
	p.ReadStdoutFn = func(r io.Reader) {
		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			atomic.AddInt64(&linesRead, 1)
		}
	}

	require.NoError(t, p.Start())

	for atomic.LoadInt64(&linesRead) < 1 {
		time.Sleep(1 * time.Millisecond)
	}

	require.NoError(t, syscall.Kill(p.Pid(), syscall.SIGKILL))

	for atomic.LoadInt64(&linesRead) < 2 {
		time.Sleep(1 * time.Millisecond)
	}

	// the mainLoopWg.Wait() call p.Stop() makes takes multiple seconds to complete
	p.Stop()
}

// test that output written right before the process exits is not lost
func TestReadsOutputAfterExit(t *testing.T) {
	exe, err := os.Executable()
	require.NoError(t, err)

	p, err := New([]string{exe, "-echo"}, []string{"INTERNAL_PROCESS_MODE=application"})
	require.NoError(t, err)
	p.RestartDelay = time.Hour
	p.Log = testutil.Logger{}

	output := make(chan string, 1)
	p.ReadStdoutFn = func(r io.Reader) {
		// Only start reading once the process has exited
		for running := true; running; {
			_, running = p.State()
		}
		buf, err := io.ReadAll(r)
		if err != nil {
			output <- err.Error()
			return
		}
		output <- string(buf)
	}

	require.NoError(t, p.Start())
	defer p.Stop()

	require.Equal(t, "hello\n", <-output)
}

var external = flag.Bool("external", false,
	"if true, run externalProcess instead of tests")
var echo = flag.Bool("echo", false,
	"if true, print a line and exit instead of running tests")

func TestMain(m *testing.M) {
	flag.Parse()
	runMode := os.Getenv("INTERNAL_PROCESS_MODE")
	if *echo && runMode == "application" {
		fmt.Fprintln(os.Stdout, "hello")
		os.Exit(0)
	}
	if *external && runMode == "application" {
		externalProcess()
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

// externalProcess is an external "misbehaving" process that won't exit
// cleanly.
func externalProcess() {
	wait := make(chan int)
	fmt.Fprintln(os.Stdout, "started")
	<-wait
	os.Exit(2) //nolint:revive // os.Exit called intentionally
}
