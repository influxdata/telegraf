package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBasicStartup(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Start(acc))
	require.NoError(t, p.Stop())

	require.Len(t, acc.GetTelegrafMetrics(), 0)
	require.Len(t, acc.Errors, 0)
}

func TestBasicStartupWithEC2Tags(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	p.EC2Tags = []string{"Name"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Start(acc))
	require.NoError(t, p.Stop())

	require.Len(t, acc.GetTelegrafMetrics(), 0)
	require.Len(t, acc.Errors, 0)
}

func TestBasicInitNoTagsReturnAnError(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{}
	err := p.Init()
	require.Error(t, err)
}

func TestBasicInitInvalidTagsReturnAnError(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"dummy", "qwerty"}
	err := p.Init()
	require.Error(t, err)
}

func TestLoadingConfig(t *testing.T) {
	// confFile := []byte("[[processors.aws_ec2]]" + "\n" + sampleConfig)
	c := config.NewConfig()
	c.SetAgent(&testAgentController{})
	err := c.LoadConfigData(context.Background(), context.Background(), []byte(
		`
	[[processors.aws_ec2]]
		imds_tags = ["availabilityZone"]
		ec2_tags = ["availabilityZone"]
		timeout = "30s"
		max_parallel_calls = 10
	`,
	))

	require.NoError(t, err)
	require.Len(t, c.Processors, 1)
}

type testAgentController struct {
	inputs     []*models.RunningInput
	processors []models.ProcessorRunner
	outputs    []*models.RunningOutput
	// configs    []*config.RunningConfigPlugin
}

func (a *testAgentController) reset() {
	a.inputs = nil
	a.processors = nil
	a.outputs = nil
	// a.configs = nil
}

func (a *testAgentController) RunningInputs() []*models.RunningInput {
	return a.inputs
}
func (a *testAgentController) RunningProcessors() []models.ProcessorRunner {
	return a.processors
}
func (a *testAgentController) RunningOutputs() []*models.RunningOutput {
	return a.outputs
}
func (a *testAgentController) AddInput(input *models.RunningInput) {
	a.inputs = append(a.inputs, input)
}
func (a *testAgentController) AddProcessor(processor models.ProcessorRunner) {
	a.processors = append(a.processors, processor)
}
func (a *testAgentController) AddOutput(output *models.RunningOutput) {
	a.outputs = append(a.outputs, output)
}
func (a *testAgentController) RunInput(input *models.RunningInput, startTime time.Time)        {}
func (a *testAgentController) RunProcessor(p models.ProcessorRunner)                           {}
func (a *testAgentController) RunOutput(ctx context.Context, output *models.RunningOutput)     {}
func (a *testAgentController) RunConfigPlugin(ctx context.Context, plugin config.ConfigPlugin) {}
func (a *testAgentController) StopInput(i *models.RunningInput)                                {}
func (a *testAgentController) StopProcessor(p models.ProcessorRunner)                          {}
func (a *testAgentController) StopOutput(p *models.RunningOutput)                              {}
