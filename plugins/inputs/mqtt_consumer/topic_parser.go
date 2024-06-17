package mqtt_consumer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrNoMatch = errors.New("not matching")

type TopicParsingConfig struct {
	Topic       string            `toml:"topic"`
	Measurement string            `toml:"measurement"`
	Tags        string            `toml:"tags"`
	Fields      string            `toml:"fields"`
	FieldTypes  map[string]string `toml:"types"`
}

type TopicParser struct {
	splitTopic []string

	extractMeasurement bool
	measurementIndex   int
	splitTags          []string
	splitFields        []string
	fieldTypes         map[string]string
}

func (cfg *TopicParsingConfig) NewParser() (*TopicParser, error) {
	p := &TopicParser{
		extractMeasurement: cfg.Measurement != "",
		fieldTypes:         cfg.FieldTypes,
	}

	splitMeasurement := strings.Split(cfg.Measurement, "/")
	for j := range splitMeasurement {
		if splitMeasurement[j] != "_" && splitMeasurement[j] != "" {
			p.measurementIndex = j
			break
		}
	}
	p.splitTags = strings.Split(cfg.Tags, "/")
	p.splitFields = strings.Split(cfg.Fields, "/")
	p.splitTopic = strings.Split(cfg.Topic, "/")

	if len(splitMeasurement) != len(p.splitTopic) && len(splitMeasurement) != 1 {
		return nil, errors.New("measurement length does not equal topic length")
	}

	if len(p.splitFields) != len(p.splitTopic) && cfg.Fields != "" {
		return nil, errors.New("fields length does not equal topic length")
	}

	if len(p.splitTags) != len(p.splitTopic) && cfg.Tags != "" {
		return nil, errors.New("tags length does not equal topic length")
	}

	return p, nil
}

func (p *TopicParser) Parse(topic string) (string, map[string]string, map[string]interface{}, error) {
	// Split the actual topic into its elements
	values := strings.Split(topic, "/")
	if !compareTopics(p.splitTopic, values) {
		return "", nil, nil, ErrNoMatch
	}

	// Extract the measurement name
	var measurement string
	if p.extractMeasurement {
		measurement = values[p.measurementIndex]
	}

	// Extract the tags
	tags := make(map[string]string, len(p.splitTags))
	for i, k := range p.splitTags {
		if k == "_" || k == "" {
			continue
		}
		tags[k] = values[i]
	}

	// Extract the fields
	fields := make(map[string]interface{}, len(p.splitFields))
	for i, k := range p.splitFields {
		if k == "_" || k == "" {
			continue
		}
		v, err := p.convertToFieldType(values[i], k)
		if err != nil {
			return "", nil, nil, err
		}
		fields[k] = v
	}

	return measurement, tags, fields, nil
}

func (p *TopicParser) convertToFieldType(value string, key string) (interface{}, error) {
	// If the user configured inputs.mqtt_consumer.topic.types, check for the desired type
	desiredType, ok := p.fieldTypes[key]
	if !ok {
		return value, nil
	}

	var v interface{}
	var err error
	switch desiredType {
	case "uint":
		if v, err = strconv.ParseUint(value, 10, 64); err != nil {
			return nil, fmt.Errorf("unable to convert field %q to type uint: %w", value, err)
		}
	case "int":
		if v, err = strconv.ParseInt(value, 10, 64); err != nil {
			return nil, fmt.Errorf("unable to convert field %q to type int: %w", value, err)
		}
	case "float":
		if v, err = strconv.ParseFloat(value, 64); err != nil {
			return nil, fmt.Errorf("unable to convert field %q to type float: %w", value, err)
		}
	default:
		return nil, fmt.Errorf("converting to the type %s is not supported: use int, uint, or float", desiredType)
	}

	return v, nil
}
