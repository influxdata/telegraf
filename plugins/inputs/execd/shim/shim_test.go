package shim

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/selfstat"
)

func TestShimWorks(t *testing.T) {
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes

	wait := runInputPlugin(10 * time.Millisecond)
	<-wait
	for stdoutBytes.Len() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	out := string(stdoutBytes.Bytes())
	if assert.Contains(t, out, "\n") {
		metricLine := strings.Split(out, "\n")[0]
		assert.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
	}
}

func TestShimStdinSignalingWorks(t *testing.T) {
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes
	stdinBytes := bytes.NewBufferString("")
	stdin = stdinBytes

	wait := runInputPlugin(40 * time.Second)

	stdinBytes.WriteString("\n")
	<-wait
	for stdoutBytes.Len() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	out := string(stdoutBytes.Bytes())
	if assert.Contains(t, out, "\n") {
		metricLine := strings.Split(out, "\n")[0]
		assert.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
	}
}

func TestShimUSR1SignalingWorks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
		return
	}
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes

	wait := runInputPlugin(40 * time.Second)

	// sleep a bit to avoid a race condition where the input hasn't loaded yet.
	time.Sleep(10 * time.Millisecond)

	// signal USR1 to yourself.
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	assert.NoError(t, err)
	err = process.Signal(syscall.SIGUSR1)
	assert.NoError(t, err)

	<-wait
	for stdoutBytes.Len() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	out := string(stdoutBytes.Bytes())
	if assert.Contains(t, out, "\n") {
		metricLine := strings.Split(out, "\n")[0]
		assert.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
	}
}

func runInputPlugin(timeout time.Duration) chan bool {
	wait := make(chan bool)
	inp := &models.RunningInput{
		Input: &testInput{
			wait: wait,
		},
		Config:          &models.InputConfig{},
		GatherTime:      selfstat.Register("", "", map[string]string{}),
		MetricsGathered: selfstat.Register("", "", map[string]string{}),
	}

	cfg := &config.Config{
		Inputs: []*models.RunningInput{
			inp,
		},
	}
	go RunPlugins(cfg, timeout) // we aren't using the timer here
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
