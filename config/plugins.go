package config

import (
	"context"
	"time"

	"github.com/influxdata/telegraf/models"
)

type ConfigCreator func() ConfigPlugin
type StorageCreator func() StoragePlugin

var (
	ConfigPlugins  = map[string]ConfigCreator{}
	StoragePlugins = map[string]StorageCreator{}
)

// ConfigPlugin is the interface for implemnting plugins that change how Telegraf works
type ConfigPlugin interface {
	GetName() string
	Init(ctx context.Context, outputCtx context.Context, cfg *Config, agent AgentController) error
	Close() error
}

// StoragePlugin is the interface to implement if you're building a plugin that implements state storage
type StoragePlugin interface {
	Init() error

	Load(namespace, key string, obj interface{}) error
	Save(namespace, key string, obj interface{}) error

	Close() error
}

// AgentController represents the plugin management that the agent is currently doing
type AgentController interface {
	RunningInputs() []*models.RunningInput
	RunningProcessors() []models.ProcessorRunner
	RunningOutputs() []*models.RunningOutput

	AddInput(input *models.RunningInput)
	AddProcessor(processor models.ProcessorRunner)
	AddOutput(output *models.RunningOutput)

	RunInput(input *models.RunningInput, startTime time.Time)
	RunProcessor(p models.ProcessorRunner)
	RunOutput(ctx context.Context, output *models.RunningOutput)
	RunConfigPlugin(ctx context.Context, plugin ConfigPlugin)

	StopInput(i *models.RunningInput)
	StopProcessor(p models.ProcessorRunner)
	StopOutput(p *models.RunningOutput)
}
