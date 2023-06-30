package modbus

import (
	_ "embed"
	"errors"
	"fmt"
	"hash/maphash"

	"github.com/influxdata/telegraf"
)

//go:embed sample_metric.conf
var sampleConfigPartPerMetric string

type metricFieldDefinition struct {
	RegisterType string  `toml:"register"`
	Address      uint16  `toml:"address"`
	Name         string  `toml:"name"`
	InputType    string  `toml:"type"`
	Scale        float64 `toml:"scale"`
	OutputType   string  `toml:"output"`
}

type metricDefinition struct {
	SlaveID     byte                    `toml:"slave_id"`
	ByteOrder   string                  `toml:"byte_order"`
	Measurement string                  `toml:"measurement"`
	Fields      []metricFieldDefinition `toml:"fields"`
	Tags        map[string]string       `toml:"tags"`
}

type ConfigurationPerMetric struct {
	Optimization      string             `toml:"optimization"`
	MaxExtraRegisters uint16             `toml:"optimization_max_register_fill"`
	Metrics           []metricDefinition `toml:"metric"`
	workarounds       ModbusWorkarounds
	logger            telegraf.Logger
}

func (c *ConfigurationPerMetric) SampleConfigPart() string {
	return sampleConfigPartPerMetric
}

func (c *ConfigurationPerMetric) Check() error {
	seed := maphash.MakeSeed()
	seenFields := make(map[uint64]bool)

	// Check optimization algorithm
	switch c.Optimization {
	case "", "none":
		c.Optimization = "none"
	case "max_insert":
		if c.MaxExtraRegisters == 0 {
			c.MaxExtraRegisters = 50
		}
	default:
		return fmt.Errorf("unknown optimization %q", c.Optimization)
	}

	for defidx, def := range c.Metrics {
		// Check byte order of the data
		switch def.ByteOrder {
		case "":
			def.ByteOrder = "ABCD"
		case "ABCD", "DCBA", "BADC", "CDAB", "MSW-BE", "MSW-LE", "LSW-LE", "LSW-BE":
		default:
			return fmt.Errorf("unknown byte-order %q", def.ByteOrder)
		}

		// Set the default for measurement if required
		if def.Measurement == "" {
			def.Measurement = "modbus"
		}

		// Reject any configuration without fields as it
		// makes no sense to not define anything but a request.
		if len(def.Fields) == 0 {
			return errors.New("found request section without fields")
		}

		// Check the fields
		for fidx, f := range def.Fields {
			// Name is mandatory
			if f.Name == "" {
				return fmt.Errorf("empty field name in request for slave %d", def.SlaveID)
			}

			// Check register type
			switch f.RegisterType {
			case "":
				f.RegisterType = "holding"
			case "coil", "discrete", "holding", "input":
			default:
				return fmt.Errorf("unknown register-type %q for field %q", f.RegisterType, f.Name)
			}

			// Check the input and output type for all fields as we later need
			// it to determine the number of registers to query.
			switch f.RegisterType {
			case "holding", "input":
				// Check the input type
				switch f.InputType {
				case "":
				case "INT8L", "INT8H", "INT16", "INT32", "INT64":
				case "UINT8L", "UINT8H", "UINT16", "UINT32", "UINT64":
				case "FLOAT16", "FLOAT32", "FLOAT64":
				default:
					return fmt.Errorf("unknown register data-type %q for field %q", f.InputType, f.Name)
				}

				// Check output type
				switch f.OutputType {
				case "", "INT64", "UINT64", "FLOAT64":
				default:
					return fmt.Errorf("unknown output data-type %q for field %q", f.OutputType, f.Name)
				}
			case "coil", "discrete":
				// Bit register types can only be UINT64 or BOOL
				switch f.OutputType {
				case "", "UINT16", "BOOL":
				default:
					return fmt.Errorf("unknown output data-type %q for field %q", f.OutputType, f.Name)
				}
			}
			def.Fields[fidx] = f

			// Check for duplicate field definitions
			id, err := c.fieldID(seed, def, f)
			if err != nil {
				return fmt.Errorf("cannot determine field id for %q: %w", f.Name, err)
			}
			if seenFields[id] {
				return fmt.Errorf("field %q duplicated in measurement %q (slave %d)", f.Name, def.Measurement, def.SlaveID)
			}
			seenFields[id] = true
		}
		c.Metrics[defidx] = def
	}

	return nil
}

func (c *ConfigurationPerMetric) Process() (map[byte]requestSet, error) {
	collection := make(map[byte]map[string][]field)

	// Collect the requested registers across metrics and transform them into
	// requests. This will produce one request per slave and register-type
	for _, def := range c.Metrics {
		// Make sure we have a set to work with
		set, found := collection[def.SlaveID]
		if !found {
			set = make(map[string][]field)
		}

		for _, fdef := range def.Fields {
			// Construct the field from the field definition
			f, err := c.newField(fdef, def)
			if err != nil {
				return nil, fmt.Errorf("initializing field %q of measurement %q failed: %w", fdef.Name, def.Measurement, err)
			}

			// Attach the field to the correct set
			set[fdef.RegisterType] = append(set[fdef.RegisterType], f)
		}
		collection[def.SlaveID] = set
	}

	result := make(map[byte]requestSet)

	params := groupingParams{
		Optimization:      c.Optimization,
		MaxExtraRegisters: c.MaxExtraRegisters,
		Log:               c.logger,
	}
	for sid, scollection := range collection {
		var set requestSet
		for registerType, fields := range scollection {
			switch registerType {
			case "coil":
				params.MaxBatchSize = maxQuantityCoils
				if c.workarounds.OnRequestPerField {
					params.MaxBatchSize = 1
				}
				params.EnforceFromZero = c.workarounds.ReadCoilsStartingAtZero
				requests := groupFieldsToRequests(fields, params)
				set.coil = append(set.coil, requests...)
			case "discrete":
				params.MaxBatchSize = maxQuantityDiscreteInput
				if c.workarounds.OnRequestPerField {
					params.MaxBatchSize = 1
				}
				requests := groupFieldsToRequests(fields, params)
				set.discrete = append(set.discrete, requests...)
			case "holding":
				params.MaxBatchSize = maxQuantityHoldingRegisters
				if c.workarounds.OnRequestPerField {
					params.MaxBatchSize = 1
				}
				requests := groupFieldsToRequests(fields, params)
				set.holding = append(set.holding, requests...)
			case "input":
				params.MaxBatchSize = maxQuantityInputRegisters
				if c.workarounds.OnRequestPerField {
					params.MaxBatchSize = 1
				}
				requests := groupFieldsToRequests(fields, params)
				set.input = append(set.input, requests...)
			default:
				return nil, fmt.Errorf("unknown register type %q", registerType)
			}
		}
		if !set.Empty() {
			result[sid] = set
		}
	}

	return result, nil
}

func (c *ConfigurationPerMetric) newField(def metricFieldDefinition, mdef metricDefinition) (field, error) {
	typed := def.RegisterType == "holding" || def.RegisterType == "input"

	fieldLength := uint16(1)
	if typed {
		var err error
		if fieldLength, err = c.determineFieldLength(def.InputType); err != nil {
			return field{}, err
		}
	}

	// Initialize the field
	f := field{
		measurement: mdef.Measurement,
		name:        def.Name,
		address:     def.Address,
		length:      fieldLength,
		tags:        mdef.Tags,
	}

	// Handle type conversions for coil and discrete registers
	if !typed {
		var err error
		f.converter, err = determineUntypedConverter(def.OutputType)
		if err != nil {
			return field{}, err
		}
		// No more processing for un-typed (coil and discrete registers) fields
		return f, nil
	}

	// Automagically determine the output type...
	if def.OutputType == "" {
		if def.Scale == 0.0 {
			// For non-scaling cases we should choose the output corresponding to the input class
			// i.e. INT64 for INT*, UINT64 for UINT* etc.
			var err error
			if def.OutputType, err = c.determineOutputDatatype(def.InputType); err != nil {
				return field{}, err
			}
		} else {
			// For scaling cases we always want FLOAT64 by default
			def.OutputType = "FLOAT64"
		}
	}

	// Setting default byte-order
	byteOrder := mdef.ByteOrder
	if byteOrder == "" {
		byteOrder = "ABCD"
	}

	// Normalize the data relevant for determining the converter
	inType, err := normalizeInputDatatype(def.InputType)
	if err != nil {
		return field{}, err
	}
	outType, err := normalizeOutputDatatype(def.OutputType)
	if err != nil {
		return field{}, err
	}
	order, err := normalizeByteOrder(byteOrder)
	if err != nil {
		return field{}, err
	}

	f.converter, err = determineConverter(inType, order, outType, def.Scale)
	if err != nil {
		return field{}, err
	}

	return f, nil
}

func (c *ConfigurationPerMetric) fieldID(seed maphash.Seed, def metricDefinition, field metricFieldDefinition) (uint64, error) {
	var mh maphash.Hash
	mh.SetSeed(seed)

	if err := mh.WriteByte(def.SlaveID); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(field.RegisterType); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(def.Measurement); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(field.Name); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}

	// Tags
	for k, v := range def.Tags {
		if _, err := mh.WriteString(k); err != nil {
			return 0, err
		}
		if err := mh.WriteByte('='); err != nil {
			return 0, err
		}
		if _, err := mh.WriteString(v); err != nil {
			return 0, err
		}
		if err := mh.WriteByte(':'); err != nil {
			return 0, err
		}
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}

	return mh.Sum64(), nil
}

func (c *ConfigurationPerMetric) determineOutputDatatype(input string) (string, error) {
	// Handle our special types
	switch input {
	case "INT8L", "INT8H", "INT16", "INT32", "INT64":
		return "INT64", nil
	case "UINT8L", "UINT8H", "UINT16", "UINT32", "UINT64":
		return "UINT64", nil
	case "FLOAT16", "FLOAT32", "FLOAT64":
		return "FLOAT64", nil
	}
	return "unknown", fmt.Errorf("invalid input datatype %q for determining output", input)
}

func (c *ConfigurationPerMetric) determineFieldLength(input string) (uint16, error) {
	// Handle our special types
	switch input {
	case "INT8L", "INT8H", "UINT8L", "UINT8H":
		return 1, nil
	case "INT16", "UINT16", "FLOAT16":
		return 1, nil
	case "INT32", "UINT32", "FLOAT32":
		return 2, nil
	case "INT64", "UINT64", "FLOAT64":
		return 4, nil
	}
	return 0, fmt.Errorf("invalid input datatype %q for determining field length", input)
}
