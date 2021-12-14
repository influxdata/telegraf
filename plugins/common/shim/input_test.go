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
		_, _ = io.ReadAll(r)
	}()
	// check that it exits cleanly
	<-exited
}

func runInputPlugin(t *testing.T, interval time.Duration, stdin io.Reader, stdout, stderr io.Writer) (metricProcessed chan bool, exited chan bool) {
	metricProcessed = make(chan bool, 1)
	exited = make(chan bool, 1)
	inp := &testInput{
		metricProcessed: metricProcessed,
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
	go func() {
		err := shim.Run(interval)
		require.NoError(t, err)
		exited <- true
	}()
	return metricProcessed, exited
}

type testInput struct {
	metricProcessed chan bool
}

func (i *testInput) SampleConfig() string {
	return ""
}

func (i *testInput) Description() string {
	return "test"
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

func (i *testInput) Start(_ telegraf.Accumulator) error {
	return nil
}

func (i *testInput) Stop() {
}

type serviceInput struct {
	ServiceName string `toml:"service_name"`
	SecretToken string `toml:"secret_token"`
	SecretValue string `toml:"secret_value"`
}

func (i *serviceInput) SampleConfig() string {
	return ""
}

func (i *serviceInput) Description() string {
	return ""
}

func (i *serviceInput) Gather(acc telegraf.Accumulator) error {
	acc.AddFields("measurement",
		map[string]interface{}{
			"field": 1,
		},
		map[string]string{
			"tag": "tag",
		}, time.Unix(1234, 5678))

	return nil
}

func (i *serviceInput) Start(_ telegraf.Accumulator) error {
	return nil
}

func (i *serviceInput) Stop() {
}
