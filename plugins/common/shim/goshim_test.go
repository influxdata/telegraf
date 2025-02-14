package shim

import (
	"bufio"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/logger"
)

func TestShimSetsUpLogger(t *testing.T) {
	stderrReader, stderrWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()

	runErroringInputPlugin(t, 40*time.Second, stdinReader, nil, stderrWriter)

	_, err := stdinWriter.Write([]byte("\n"))
	require.NoError(t, err)

	r := bufio.NewReader(stderrReader)
	out, err := r.ReadString('\n')
	require.NoError(t, err)
	require.Contains(t, out, "Error in plugin: intentional")

	err = stdinWriter.Close()
	require.NoError(t, err)
}

func runErroringInputPlugin(t *testing.T, interval time.Duration, stdin io.Reader, stdout, stderr io.Writer) (processed, exited chan bool) {
	processed = make(chan bool, 1)
	exited = make(chan bool, 1)
	inp := &erroringInput{}

	shim := New()
	if stdin != nil {
		shim.stdin = stdin
	}
	if stdout != nil {
		shim.stdout = stdout
	}
	if stderr != nil {
		shim.stderr = stderr
		logger.RedirectLogging(stderr)
	}

	require.NoError(t, shim.AddInput(inp))
	go func(e chan bool) {
		if err := shim.Run(interval); err != nil {
			t.Error(err)
		}
		e <- true
	}(exited)
	return processed, exited
}

type erroringInput struct {
}

func (*erroringInput) SampleConfig() string {
	return ""
}

func (*erroringInput) Gather(acc telegraf.Accumulator) error {
	acc.AddError(errors.New("intentional"))
	return nil
}

func (*erroringInput) Start(telegraf.Accumulator) error {
	return nil
}

func (*erroringInput) Stop() {
}
