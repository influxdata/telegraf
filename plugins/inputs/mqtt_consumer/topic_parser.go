package mqtt_consumer

import (
	"errors"
	"strings"
)

type TopicParsingConfig struct {
	Topic       string            `toml:"topic"`
	Measurement string            `toml:"measurement"`
	Tags        string            `toml:"tags"`
	Fields      string            `toml:"fields"`
	FieldTypes  map[string]string `toml:"types"`
	// cached split of user given information
	MeasurementIndex int
	SplitTags        []string
	SplitFields      []string
	SplitTopic       []string
}

func (cfg *TopicParsingConfig) Init() error {
	splitMeasurement := strings.Split(cfg.Measurement, "/")
	for j := range splitMeasurement {
		if splitMeasurement[j] != "_" && splitMeasurement[j] != "" {
			cfg.MeasurementIndex = j
			break
		}
	}
	cfg.SplitTags = strings.Split(cfg.Tags, "/")
	cfg.SplitFields = strings.Split(cfg.Fields, "/")
	cfg.SplitTopic = strings.Split(cfg.Topic, "/")

	if len(splitMeasurement) != len(cfg.SplitTopic) && len(splitMeasurement) != 1 {
		return errors.New("measurement length does not equal topic length")
	}

	if len(cfg.SplitFields) != len(cfg.SplitTopic) && cfg.Fields != "" {
		return errors.New("fields length does not equal topic length")
	}

	if len(cfg.SplitTags) != len(cfg.SplitTopic) && cfg.Tags != "" {
		return errors.New("tags length does not equal topic length")
	}

	return nil
}
