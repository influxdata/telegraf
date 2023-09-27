package models

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/selfstat"
)

// SerializerConfig is the common config for all serializers.
type SerializerConfig struct {
	Parent      string
	Alias       string
	DataFormat  string
	DefaultTags map[string]string
}

type RunningSerializer struct {
	Serializer serializers.Serializer
	Config     *SerializerConfig
	log        telegraf.Logger

	MetricsSerialized selfstat.Stat
	BytesSerialized   selfstat.Stat
	SerializationTime selfstat.Stat
}

func NewRunningSerializer(serializer serializers.Serializer, config *SerializerConfig) *RunningSerializer {
	tags := map[string]string{"type": config.DataFormat}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	serializerErrorsRegister := selfstat.Register("serializer", "errors", tags)
	logger := NewLogger("serializers", config.DataFormat+"::"+config.Parent, config.Alias)
	logger.OnErr(func() {
		serializerErrorsRegister.Incr(1)
	})
	SetLoggerOnPlugin(serializer, logger)

	return &RunningSerializer{
		Serializer: serializer,
		Config:     config,
		MetricsSerialized: selfstat.Register(
			"serializer",
			"metrics_serialized",
			tags,
		),
		BytesSerialized: selfstat.Register(
			"serializer",
			"bytes_serialized",
			tags,
		),
		SerializationTime: selfstat.Register(
			"serializer",
			"serialization_time_ns",
			tags,
		),
		log: logger,
	}
}

func (r *RunningSerializer) LogName() string {
	return logName("parsers", r.Config.DataFormat+"::"+r.Config.Parent, r.Config.Alias)
}

func (r *RunningSerializer) Init() error {
	if p, ok := r.Serializer.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RunningSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	start := time.Now()
	buf, err := r.Serializer.Serialize(metric)
	elapsed := time.Since(start)
	r.SerializationTime.Incr(elapsed.Nanoseconds())
	r.MetricsSerialized.Incr(1)
	r.BytesSerialized.Incr(int64(len(buf)))

	return buf, err
}

func (r *RunningSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	start := time.Now()
	buf, err := r.Serializer.SerializeBatch(metrics)
	elapsed := time.Since(start)
	r.SerializationTime.Incr(elapsed.Nanoseconds())
	r.MetricsSerialized.Incr(int64(len(metrics)))
	r.BytesSerialized.Incr(int64(len(buf)))

	return buf, err
}

func (r *RunningSerializer) Log() telegraf.Logger {
	return r.log
}
