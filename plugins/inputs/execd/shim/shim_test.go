package shim

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
)

func TestShimWorks(t *testing.T) {
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes

	timeout := time.NewTimer(10 * time.Second)
	wait := runInputPlugin(t, 10*time.Millisecond)

	select {
	case <-wait:
	case <-timeout.C:
		require.Fail(t, "Timeout waiting for metric to arrive")
	}
	for stdoutBytes.Len() == 0 {
		select {
		case <-timeout.C:
			require.Fail(t, "Timeout waiting to read metric from stdout")
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	out := string(stdoutBytes.Bytes())
	require.Contains(t, out, "\n")
	metricLine := strings.Split(out, "\n")[0]
	require.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
}

func TestShimStdinSignalingWorks(t *testing.T) {
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes
	stdinBytes := bytes.NewBufferString("")
	stdin = stdinBytes

	timeout := time.NewTimer(10 * time.Second)
	wait := runInputPlugin(t, 40*time.Second)

	stdinBytes.WriteString("\n")

	select {
	case <-wait:
	case <-timeout.C:
		require.Fail(t, "Timeout waiting for metric to arrive")
	}

	for stdoutBytes.Len() == 0 {
		select {
		case <-timeout.C:
			require.Fail(t, "Timeout waiting to read metric from stdout")
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	out := string(stdoutBytes.Bytes())
	require.Contains(t, out, "\n")
	metricLine := strings.Split(out, "\n")[0]
	require.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
}

func runInputPlugin(t *testing.T, timeout time.Duration) chan bool {
	wait := make(chan bool)
	inp := &testInput{
		wait: wait,
	}

	shim := New()
	shim.AddInput(inp)
	go func() {
		err := shim.Run(timeout) // we aren't using the timer here
		require.NoError(t, err)
	}()
	return wait
}

type testInput struct {
	wait chan bool
}

func (i *testInput) SampleConfig() string {
	return ""
}

func (i *testInput) Description() string {
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
	i.wait <- true
	return nil
}

func (i *testInput) Start(acc telegraf.Accumulator) error {
	return nil
}

func (i *testInput) Stop() {
}
