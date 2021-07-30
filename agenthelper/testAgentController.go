package agenthelper

// Do not import other Telegraf packages as it causes dependency loops
import (
	"context"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
)

// For use in unit tests where the test needs a config object but doesn't
// need to run its plugins. For example when testing parsing toml config.
type TestAgentController struct {
	inputs     []*models.RunningInput
	processors []models.ProcessorRunner
	outputs    []*models.RunningOutput
	// configs    []*config.RunningConfigPlugin
}

func (a *TestAgentController) Reset() {
	a.inputs = nil
	a.processors = nil
	a.outputs = nil
	// a.configs = nil
}

func (a *TestAgentController) RunningInputs() []*models.RunningInput {
	return a.inputs
}
func (a *TestAgentController) RunningProcessors() []models.ProcessorRunner {
	return a.processors
}
func (a *TestAgentController) RunningOutputs() []*models.RunningOutput {
	return a.outputs
}
func (a *TestAgentController) AddInput(input *models.RunningInput) {
	a.inputs = append(a.inputs, input)
}
func (a *TestAgentController) AddProcessor(processor models.ProcessorRunner) {
	a.processors = append(a.processors, processor)
}
func (a *TestAgentController) AddOutput(output *models.RunningOutput) {
	a.outputs = append(a.outputs, output)
}
func (a *TestAgentController) RunInput(input *models.RunningInput, startTime time.Time)        {}
func (a *TestAgentController) RunProcessor(p models.ProcessorRunner)                           {}
func (a *TestAgentController) RunOutput(ctx context.Context, output *models.RunningOutput)     {}
func (a *TestAgentController) RunConfigPlugin(ctx context.Context, plugin config.ConfigPlugin) {}
func (a *TestAgentController) StopInput(i *models.RunningInput)                                {}
func (a *TestAgentController) StopProcessor(p models.ProcessorRunner)                          {}
func (a *TestAgentController) StopOutput(p *models.RunningOutput)                              {}
