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
	Init(ctx context.Context, cfg *Config, agent AgentController) error
	Close() error
}

// StoragePlugin is the interface to implement if you're building a plugin that implements state storage
type StoragePlugin interface {
	GetName() string
	Init() error

	Load(namespace string) map[string]interface{}
	Save(namespace string, values map[string]interface{})

	LoadKey(namespace, key string) interface{}
	SaveKey(namespace, key string, value interface{})

	Close() error
}

type AgentController interface {
	RunningInputs() []*models.RunningInput
	RunningProcessors() []models.ProcessorRunner
	RunningOutputs() []*models.RunningOutput

	StartInput(input *models.RunningInput) error
	RunInput(input *models.RunningInput, startTime time.Time)
	StopInput(i *models.RunningInput)

	StartProcessor(processor models.ProcessorRunner) error
	RunProcessor(p models.ProcessorRunner)
	StopProcessor(p models.ProcessorRunner)

	StartOutput(output *models.RunningOutput) error
	RunOutput(ctx context.Context, output *models.RunningOutput)
	StopOutput(p *models.RunningOutput)
}
