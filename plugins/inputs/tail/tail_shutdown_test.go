package tail

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/testutil"
)

// trackingMaker is a minimal agent.MetricMaker so the test can drive the real
// tracking accumulator, whose delivery channel is sized to the in-flight
// budget and is what panics with "channel is full" in issue #19073.
type trackingMaker struct{}

func (trackingMaker) LogName() string                              { return "tail" }
func (trackingMaker) MakeMetric(m telegraf.Metric) telegraf.Metric { return m }
func (trackingMaker) Log() telegraf.Logger                         { return logger.New("inputs", "tail", "") }

// TestTailNoSemaphoreLeakOnClose reproduces issue #19073. With multiple
// tailers sharing a single in-flight budget, closing a tailer that is blocked
// adding a line must not free a budget slot without a matching delivery.
// Releasing a slot lets another tailer add beyond the budget and overflow the
// accumulator delivery channel.
func TestTailNoSemaphoreLeakOnClose(t *testing.T) {
	const maxUndelivered = 1

	dir := t.TempDir()
	files := []string{
		filepath.Join(dir, "a.log"),
		filepath.Join(dir, "b.log"),
		filepath.Join(dir, "c.log"),
	}
	// Many lines per file so that, once one line takes the single in-flight
	// slot, every receiver stays blocked trying to add a further line.
	content := bytes.Repeat([]byte("m value=1\n"), 50)
	for _, f := range files {
		require.NoError(t, os.WriteFile(f, content, 0600))
	}

	plugin := newTestTail()
	plugin.Log = testutil.Logger{}
	plugin.InitialReadOffset = "beginning"
	plugin.MaxUndeliveredLines = maxUndelivered
	plugin.Files = files
	plugin.SetParserFunc(newInfluxParser)
	require.NoError(t, plugin.Init())

	// Real tracking accumulator with a delivery channel sized to the budget.
	// Metrics are collected but never accepted, i.e. deliveries are withheld,
	// so the budget can only be freed by an erroneous release. The number of
	// collected metrics therefore equals the in-flight count.
	src := make(chan telegraf.Metric, 1000)
	acc := agent.NewAccumulator(trackingMaker{}, src).WithTracking(maxUndelivered)

	require.NoError(t, plugin.Start(acc))

	// Wait until the budget is taken, after which every receiver is blocked
	// trying to add its next line.
	require.Eventually(t, func() bool {
		return len(src) == maxUndelivered
	}, 5*time.Second, 10*time.Millisecond)

	// Closing the tailers must not free a budget slot without a delivery.
	plugin.Stop()

	// Start() and Stop() join all receiver goroutines, so the in-flight count
	// is now final. It must never exceed the configured budget.
	require.LessOrEqual(t, len(src), maxUndelivered,
		"in-flight metrics exceeded max_undelivered_lines after closing tailers")
}
