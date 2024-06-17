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
	topic []string

	extractMeasurement bool
	measurementIndex   int
	tagIndices         map[string]int
	fieldIndices       map[string]int
	fieldTypes         map[string]string
}

func (cfg *TopicParsingConfig) NewParser() (*TopicParser, error) {
	p := &TopicParser{
		extractMeasurement: cfg.Measurement != "",
		fieldTypes:         cfg.FieldTypes,
		topic:              strings.Split(cfg.Topic, "/"),
	}

	// Determine metric name selection
	measurementParts := strings.Split(cfg.Measurement, "/")
	for i, k := range measurementParts {
		if k != "_" && k != "" {
			p.measurementIndex = i
			break
		}
	}

	// Determine tag selections
	tagParts := strings.Split(cfg.Tags, "/")
	p.tagIndices = make(map[string]int, len(tagParts))
	for i, k := range tagParts {
		if k != "_" && k != "" {
			p.tagIndices[k] = i
		}
	}

	// Determine tag selections
	fieldParts := strings.Split(cfg.Fields, "/")
	p.fieldIndices = make(map[string]int, len(fieldParts))
	for i, k := range fieldParts {
		if k != "_" && k != "" {
			p.fieldIndices[k] = i
		}
	}

	if len(measurementParts) != len(p.topic) && len(measurementParts) != 1 {
		return nil, errors.New("measurement length does not equal topic length")
	}

	if len(fieldParts) != len(p.topic) && cfg.Fields != "" {
		return nil, errors.New("fields length does not equal topic length")
	}

	if len(tagParts) != len(p.topic) && cfg.Tags != "" {
		return nil, errors.New("tags length does not equal topic length")
	}

	return p, nil
}

func (p *TopicParser) Parse(topic string) (string, map[string]string, map[string]interface{}, error) {
	// Split the actual topic into its elements and check for a match
	topicParts := strings.Split(topic, "/")
	if len(p.topic) != len(topicParts) {
		return "", nil, nil, ErrNoMatch
	}
	for i, expected := range p.topic {
		if topicParts[i] != expected && expected != "+" {
			return "", nil, nil, ErrNoMatch
		}
	}

	// Extract the measurement name
	var measurement string
	if p.extractMeasurement {
		measurement = topicParts[p.measurementIndex]
	}

	// Extract the tags
	tags := make(map[string]string, len(p.tagIndices))
	for k, i := range p.tagIndices {
		tags[k] = topicParts[i]
	}

	// Extract the fields
	fields := make(map[string]interface{}, len(p.fieldIndices))
	for k, i := range p.fieldIndices {
		v, err := p.convertToFieldType(topicParts[i], k)
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
