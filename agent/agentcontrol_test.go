package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2021, 4, 9, 0, 0, 0, 0, time.UTC)

func TestAgentPluginControllerLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.NewConfig()
	a := NewAgent(ctx, cfg)
	// cfg.SetAgent(a)
	inp := &testInputPlugin{}
	_ = inp.Init()
	ri := models.NewRunningInput(inp, &models.InputConfig{Name: "in"})
	a.AddInput(ri)
	go a.RunInput(ri, time.Now())

	m, _ := metric.New("testing",
		map[string]string{
			"country": "canada",
		},
		map[string]interface{}{
			"population": 37_590_000,
		},
		now)

	rp := models.NewRunningProcessor(&testProcessorPlugin{}, &models.ProcessorConfig{Name: "proc"})
	a.AddProcessor(rp)
	go a.RunProcessor(rp)

	outputCtx, outputCancel := context.WithCancel(context.Background())
	o := &testOutputPlugin{}
	_ = o.Init()
	ro := models.NewRunningOutput(o, &models.OutputConfig{Name: "out"}, 100, 100)
	a.AddOutput(ro)
	go a.RunOutput(outputCtx, ro)

	go a.RunWithAPI(outputCancel)

	inp.injectMetric(m)

	waitForStatus(t, ri, "running", 1*time.Second)
	waitForStatus(t, rp, "running", 1*time.Second)
	waitForStatus(t, ro, "running", 1*time.Second)

	cancel()

	waitForStatus(t, ri, "dead", 1*time.Second)
	waitForStatus(t, rp, "dead", 1*time.Second)
	waitForStatus(t, ro, "dead", 1*time.Second)

	require.Len(t, o.receivedMetrics, 1)
	expected := testutil.MustMetric("testing",
		map[string]string{
			"country": "canada",
		},
		map[string]interface{}{
			"population": 37_590_000,
			"capital":    "Ottawa",
		},
		now)
	testutil.RequireMetricEqual(t, expected, o.receivedMetrics[0])

}

func TestAgentPluginConnectionsAfterAddAndRemoveProcessor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.NewConfig()
	a := NewAgent(ctx, cfg)
	// cfg.SetAgent(a)

	// start an input
	inp := &testInputPlugin{}
	_ = inp.Init()
	ri := models.NewRunningInput(inp, &models.InputConfig{Name: "in"})
	a.AddInput(ri)
	go a.RunInput(ri, time.Now())

	// start output
	outputCtx, outputCancel := context.WithCancel(context.Background())
	o := &testOutputPlugin{}
	_ = o.Init()
	ro := models.NewRunningOutput(o, &models.OutputConfig{Name: "out"}, 100, 100)
	a.AddOutput(ro)
	go a.RunOutput(outputCtx, ro)

	// Run agent
	go a.RunWithAPI(outputCancel)

	// wait for plugins to start
	waitForStatus(t, ri, "running", 1*time.Second)
	waitForStatus(t, ro, "running", 1*time.Second)

	// inject a metric into the input plugin as if it collected it
	m, _ := metric.New("mojo", nil, map[string]interface{}{"jenkins": "leroy"}, now)
	inp.injectMetric(m)

	// wait for the output to get it
	o.wait(1)
	testutil.RequireMetricEqual(t, m, o.receivedMetrics[0])
	o.clear()

	// spin up new processor
	rp := models.NewRunningProcessor(&testProcessorPlugin{}, &models.ProcessorConfig{Name: "proc"})
	a.AddProcessor(rp)
	go a.RunProcessor(rp)

	// wait for the processor to start
	waitForStatus(t, rp, "running", 5*time.Second)

	// inject a metric into the input
	inp.injectMetric(m)
	// wait for it to arrive
	o.wait(1)

	// create the expected output for comparison
	expected := m.Copy()
	expected.AddField("capital", "Ottawa")

	testutil.RequireMetricEqual(t, expected, o.receivedMetrics[0])

	o.clear()

	// stop processor and wait for it to stop
	a.StopProcessor(rp)
	waitForStatus(t, rp, "dead", 5*time.Second)

	// inject a new metric
	inp.injectMetric(m)

	// wait for the output to get it
	o.wait(1)
	testutil.RequireMetricEqual(t, m, o.receivedMetrics[0])
	o.clear()

	// cancel the app's context
	cancel()

	// wait for plugins to stop
	waitForStatus(t, ri, "dead", 5*time.Second)
	waitForStatus(t, ro, "dead", 5*time.Second)
}

type hasState interface {
	GetState() models.PluginState
}

func waitForStatus(t *testing.T, stateable hasState, waitStatus string, timeout time.Duration) {
	timeoutAt := time.Now().Add(timeout)
	for timeoutAt.After(time.Now()) {
		if stateable.GetState().String() == waitStatus {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.FailNow(t, "timed out waiting for status "+waitStatus)
}

type testInputPlugin struct {
	sync.Mutex
	*sync.Cond
	started bool
	acc     telegraf.Accumulator
}

func (p *testInputPlugin) Init() error {
	p.Cond = sync.NewCond(&p.Mutex)
	return nil
}
func (p *testInputPlugin) SampleConfig() string                { return "" }
func (p *testInputPlugin) Description() string                 { return "testInputPlugin" }
func (p *testInputPlugin) Gather(a telegraf.Accumulator) error { return nil }
func (p *testInputPlugin) Start(a telegraf.Accumulator) error {
	println("locking")
	p.Lock()
	defer p.Unlock()
	println("started")
	p.acc = a
	p.started = true
	p.Cond.Broadcast()
	return nil
}
func (p *testInputPlugin) Stop() {
	println("stopping input (waiting for lock)")
	p.Lock()
	defer p.Unlock()
	println("stopped input")
}
func (p *testInputPlugin) injectMetric(m telegraf.Metric) {
	p.Lock()
	defer p.Unlock()
	for !p.started {
		p.Cond.Wait()
	}
	p.acc.AddMetric(m)
}

type testProcessorPlugin struct {
}

func (p *testProcessorPlugin) Init() error                          { return nil }
func (p *testProcessorPlugin) SampleConfig() string                 { return "" }
func (p *testProcessorPlugin) Description() string                  { return "testProcessorPlugin" }
func (p *testProcessorPlugin) Start(acc telegraf.Accumulator) error { return nil }
func (p *testProcessorPlugin) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	metric.AddField("capital", "Ottawa")
	acc.AddMetric(metric)
	return nil
}
func (p *testProcessorPlugin) Stop() error { return nil }

type testOutputPlugin struct {
	sync.Mutex
	*sync.Cond
	receivedMetrics []telegraf.Metric
}

func (p *testOutputPlugin) Init() error {
	p.Cond = sync.NewCond(&p.Mutex)
	return nil
}
func (p *testOutputPlugin) SampleConfig() string { return "" }
func (p *testOutputPlugin) Description() string  { return "testOutputPlugin" }
func (p *testOutputPlugin) Connect() error       { return nil }
func (p *testOutputPlugin) Close() error         { return nil }
func (p *testOutputPlugin) Write(metrics []telegraf.Metric) error {
	p.Lock()
	defer p.Unlock()
	p.receivedMetrics = append(p.receivedMetrics, metrics...)
	p.Broadcast()
	return nil
}

// Wait for the given number of metrics to arrive
func (p *testOutputPlugin) wait(n int) {
	p.Lock()
	defer p.Unlock()
	for len(p.receivedMetrics) < n {
		p.Cond.Wait()
	}
}

func (p *testOutputPlugin) clear() {
	p.Lock()
	defer p.Unlock()
	p.receivedMetrics = nil
}
