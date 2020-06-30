package process

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// test that a restarting process resets pipes properly
func TestRestartingRebindsPipes(t *testing.T) {
	exe, err := os.Executable()
	require.NoError(t, err)

	p, err := New([]string{exe, "-external"})
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

	syscall.Kill(p.Pid(), syscall.SIGKILL)

	for atomic.LoadInt64(&linesRead) < 2 {
		time.Sleep(1 * time.Millisecond)
	}

	p.Stop()
}

var external = flag.Bool("external", false,
	"if true, run externalProcess instead of tests")

func TestMain(m *testing.M) {
	flag.Parse()
	if *external {
		externalProcess()
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

// externalProcess is an external "misbehaving" process that won't exit
// cleanly, and crashes on SIGUSR1.
func externalProcess() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	fmt.Fprintln(os.Stdout, "started")
	<-sigs
	os.Exit(2)
}
