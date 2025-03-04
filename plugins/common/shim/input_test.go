package shim

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
)

func TestInputShimTimer(t *testing.T) {
	stdoutReader, stdoutWriter := io.Pipe()

	stdin, _ := io.Pipe() // hold the stdin pipe open

	metricProcessed, _ := runInputPlugin(t, 10*time.Millisecond, stdin, stdoutWriter, nil)

	<-metricProcessed
	r := bufio.NewReader(stdoutReader)
	out, err := r.ReadString('\n')
	require.NoError(t, err)
	require.Contains(t, out, "\n")
	metricLine := strings.Split(out, "\n")[0]
	require.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
}

func TestInputShimStdinSignalingWorks(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	metricProcessed, exited := runInputPlugin(t, 40*time.Second, stdinReader, stdoutWriter, nil)

	_, err := stdinWriter.Write([]byte("\n"))
	require.NoError(t, err)

	<-metricProcessed

	r := bufio.NewReader(stdoutReader)
	out, err := r.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "measurement,tag=tag field=1i 1234000005678\n", out)

	err = stdinWriter.Close()
	require.NoError(t, err)
	go func() {
		if _, err = io.ReadAll(r); err != nil {
			t.Error(err)
		}
	}()
	// check that it exits cleanly
	<-exited
}

func runInputPlugin(t *testing.T, interval time.Duration, stdin io.Reader, stdout, stderr io.Writer) (processed, exited chan bool) {
	processed = make(chan bool, 1)
	exited = make(chan bool, 1)
	inp := &testInput{
		metricProcessed: processed,
	}

	shim := New()
	if stdin != nil {
		shim.stdin = stdin
	}
	if stdout != nil {
		shim.stdout = stdout
	}
	if stderr != nil {
		shim.stderr = stderr
	}
	err := shim.AddInput(inp)
	require.NoError(t, err)
	go func(e chan bool) {
		if err := shim.Run(interval); err != nil {
			t.Error(err)
		}
		e <- true
	}(exited)
	return processed, exited
}

type testInput struct {
	metricProcessed chan bool
}

func (*testInput) SampleConfig() string {
	return ""
}

func (i *testInput) Gather(acc telegraf.Accumulator) error {
	acc.AddFields("measurement",
		map[string]interface{}{
			"field": 1,
		},
		map[string]string{
			"tag": "tag",
		}, time.Unix(1234, 5678))
	i.metricProcessed <- true
	return nil
}

func (*testInput) Start(telegraf.Accumulator) error {
	return nil
}

func (*testInput) Stop() {
}

type serviceInput struct {
	ServiceName string `toml:"service_name"`
	SecretToken string `toml:"secret_token"`
	SecretValue string `toml:"secret_value"`
}

func (*serviceInput) SampleConfig() string {
	return ""
}

func (*serviceInput) Gather(acc telegraf.Accumulator) error {
	acc.AddFields("measurement",
		map[string]interface{}{
			"field": 1,
		},
		map[string]string{
			"tag": "tag",
		}, time.Unix(1234, 5678))

	return nil
}

func (*serviceInput) Start(telegraf.Accumulator) error {
	return nil
}

func (*serviceInput) Stop() {
}
