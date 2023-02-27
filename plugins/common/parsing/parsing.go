package parsing

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"strconv"
	"strings"
)

var (
	ErrEmptyDelimiter = errors.New("parsing: delimiter must not be empty")
	ErrEmptyWildcard  = errors.New("parsing: wildcard must not be empty")
)

type Config struct {
	config    []ConfigEntry
	baseName  string
	delimiter string
	wildcard  string
	log       telegraf.Logger
}

// NewConfig creates a new parser to extract measurements, tags and fields from structures like
// mqtt topic or nats subjects.
// each entry in the config describes how this structure should be interpreted.
// baseName is just the individual name of the structure (e.g: 'topic' for mqtt).
// delimiter is the character or string that divides the topic in different parts.
// wildcard is the character or string that tells that a part of the topic can be any string.
func NewConfig(entries []ConfigEntry,
	baseName string,
	delimiter string,
	wildcard string,
	log telegraf.Logger) *Config {
	return &Config{
		config:    entries,
		baseName:  baseName,
		delimiter: delimiter,
		wildcard:  wildcard,
		log:       log,
	}
}

// ConfigEntry describes how a specific structure can be parsed into a telegraf.Metric.
// And incoming structure is first matched against the Base string. This string can include wildcards.
// If the incoming structure matches, measurement name, tags and fields can be extracted from it
// and are added to the metric.
type ConfigEntry struct {
	Base        string            `toml:"topic"`
	Measurement string            `toml:"measurement"`
	Tags        string            `toml:"tags"`
	Fields      string            `toml:"fields"`
	FieldTypes  map[string]string `toml:"types"`

	measurementIndex int
	splitTags        []string
	splitFields      []string
	splitSubject     []string
}

// Init validates and configures the parser and has to be called before using it.
func (c *Config) Init() error {
	if c.wildcard == "" {
		return ErrEmptyWildcard
	}

	if c.baseName == "" {
		c.baseName = "topic"
	}

	if c.delimiter == "" {
		return ErrEmptyDelimiter
	}

	for i := range c.config {
		splitMeasurement := strings.Split(c.config[i].Measurement, c.delimiter)
		for j := range splitMeasurement {
			if splitMeasurement[j] != "_" && splitMeasurement[j] != "" {
				c.config[i].measurementIndex = j
				break
			}
		}
		c.config[i].splitTags = strings.Split(c.config[i].Tags, c.delimiter)
		c.config[i].splitFields = strings.Split(c.config[i].Fields, c.delimiter)
		c.config[i].splitSubject = strings.Split(c.config[i].Base, c.delimiter)

		if len(splitMeasurement) != len(c.config[i].splitSubject) && len(splitMeasurement) != 1 {
			err := fmt.Errorf("config error %s parsing: measurement length %d does not equal %s length %d",
				c.baseName, len(splitMeasurement), c.baseName, len(c.config[i].splitSubject))
			c.log.Error(err.Error())
			return err
		}

		if len(c.config[i].splitFields) != len(c.config[i].splitSubject) && c.config[i].Fields != "" {
			err := fmt.Errorf("config error %s parsing: fields length does not equal %s length",
				c.baseName, c.baseName)
			c.log.Error(err.Error())
			return err
		}

		if len(c.config[i].splitTags) != len(c.config[i].splitSubject) && c.config[i].Tags != "" {
			err := fmt.Errorf("config error %s parsing: tags length does not equal %s length",
				c.baseName, c.baseName)
			c.log.Error(err.Error())
			return err
		}
	}
	return nil
}

// Parse evaluates the incoming structure (e.g. topic for mqtt) and modifies the metric, if the config matches
func (c *Config) Parse(str string, metric telegraf.Metric) error {
	values := strings.Split(str, c.delimiter)
	for _, p := range c.config {
		if !c.compare(p.splitSubject, values) {
			continue
		}

		if p.Measurement != "" {
			metric.SetName(values[p.measurementIndex])
		}

		if p.Tags != "" {
			err := parseMetric(p.splitTags, values, p.FieldTypes, true, metric)
			if err != nil {
				return err
			}
		}

		if p.Fields != "" {
			err := parseMetric(p.splitFields, values, p.FieldTypes, false, metric)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// compare is used to support wild cards like `+` for mqtt or `*` for nats, which allows any value
func (c *Config) compare(expected []string, incoming []string) bool {
	if len(expected) != len(incoming) {
		return false
	}

	for i, expected := range expected {
		if incoming[i] != expected && expected != c.wildcard {
			return false
		}
	}

	return true
}

// parseMetric gets multiple fields based on the user configuration (ConfigEntry.Fields)
func parseMetric(keys []string, values []string, types map[string]string, isTag bool, metric telegraf.Metric) error {
	for i, k := range keys {
		if k == "_" || k == "" {
			continue
		}
		if isTag {
			metric.AddTag(k, values[i])
		} else {
			newType, err := typeConvert(types, values[i], k)
			if err != nil {
				return err
			}
			metric.AddField(k, newType)
		}
	}
	return nil
}

func typeConvert(types map[string]string, subjectValue string, key string) (interface{}, error) {
	var newType interface{}
	var err error
	// If the user configured inputs.mqtt_consumer.subject.types, check for the desired type
	if desiredType, ok := types[key]; ok {
		switch desiredType {
		case "uint":
			newType, err = strconv.ParseUint(subjectValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert field '%s' to type uint: %w", subjectValue, err)
			}
		case "int":
			newType, err = strconv.ParseInt(subjectValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert field '%s' to type int: %w", subjectValue, err)
			}
		case "float":
			newType, err = strconv.ParseFloat(subjectValue, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert field '%s' to type float: %w", subjectValue, err)
			}
		default:
			return nil, fmt.Errorf("converting to the type %s is not supported: use int, uint, or float", desiredType)
		}
	} else {
		newType = subjectValue
	}

	return newType, nil
}
