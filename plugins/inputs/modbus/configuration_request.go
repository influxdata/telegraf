package modbus

import (
	_ "embed"
	"errors"
	"fmt"
	"hash/maphash"

	"github.com/influxdata/telegraf"
)

//go:embed sample_request.conf
var sampleConfigPartPerRequest string

type requestFieldDefinition struct {
	Address     uint16  `toml:"address"`
	Name        string  `toml:"name"`
	InputType   string  `toml:"type"`
	Scale       float64 `toml:"scale"`
	OutputType  string  `toml:"output"`
	Measurement string  `toml:"measurement"`
	Omit        bool    `toml:"omit"`
}

type requestDefinition struct {
	SlaveID           byte                     `toml:"slave_id"`
	ByteOrder         string                   `toml:"byte_order"`
	RegisterType      string                   `toml:"register"`
	Measurement       string                   `toml:"measurement"`
	Optimization      string                   `toml:"optimization"`
	MaxExtraRegisters uint16                   `toml:"optimization_max_register_fill"`
	Fields            []requestFieldDefinition `toml:"fields"`
	Tags              map[string]string        `toml:"tags"`
}

type ConfigurationPerRequest struct {
	Requests    []requestDefinition `toml:"request"`
	workarounds ModbusWorkarounds
	logger      telegraf.Logger
}

func (c *ConfigurationPerRequest) SampleConfigPart() string {
	return sampleConfigPartPerRequest
}

func (c *ConfigurationPerRequest) Check() error {
	seed := maphash.MakeSeed()
	seenFields := make(map[uint64]bool)

	for _, def := range c.Requests {
		// Check byte order of the data
		switch def.ByteOrder {
		case "":
			def.ByteOrder = "ABCD"
		case "ABCD", "DCBA", "BADC", "CDAB", "MSW-BE", "MSW-LE", "LSW-LE", "LSW-BE":
		default:
			return fmt.Errorf("unknown byte-order %q", def.ByteOrder)
		}

		// Check register type
		switch def.RegisterType {
		case "":
			def.RegisterType = "holding"
		case "coil", "discrete", "holding", "input":
		default:
			return fmt.Errorf("unknown register-type %q", def.RegisterType)
		}
		// Check for valid optimization
		switch def.Optimization {
		case "", "none", "shrink", "rearrange", "aggressive":
		case "max_insert":
			switch def.RegisterType {
			case "coil":
				if def.MaxExtraRegisters <= 0 || def.MaxExtraRegisters > maxQuantityCoils {
					return fmt.Errorf("optimization_max_register_fill has to be between 1 and %d", maxQuantityCoils)
				}
			case "discrete":
				if def.MaxExtraRegisters <= 0 || def.MaxExtraRegisters > maxQuantityDiscreteInput {
					return fmt.Errorf("optimization_max_register_fill has to be between 1 and %d", maxQuantityDiscreteInput)
				}
			case "holding":
				if def.MaxExtraRegisters <= 0 || def.MaxExtraRegisters > maxQuantityHoldingRegisters {
					return fmt.Errorf("optimization_max_register_fill has to be between 1 and %d", maxQuantityHoldingRegisters)
				}
			case "input":
				if def.MaxExtraRegisters <= 0 || def.MaxExtraRegisters > maxQuantityInputRegisters {
					return fmt.Errorf("optimization_max_register_fill has to be between 1 and %d", maxQuantityInputRegisters)
				}
			}
		default:
			return fmt.Errorf("unknown optimization %q", def.Optimization)
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
			// Check the input type for all fields except the bit-field ones.
			// We later need the type (even for omitted fields) to determine the length.
			if def.RegisterType == "holding" || def.RegisterType == "input" {
				switch f.InputType {
				case "":
				case "INT8L", "INT8H", "INT16", "INT32", "INT64":
				case "UINT8L", "UINT8H", "UINT16", "UINT32", "UINT64":
				case "FLOAT16", "FLOAT32", "FLOAT64":
				default:
					return fmt.Errorf("unknown register data-type %q for field %q", f.InputType, f.Name)
				}
			}

			// Other properties don't need to be checked for omitted fields
			if f.Omit {
				continue
			}

			// Name is mandatory
			if f.Name == "" {
				return fmt.Errorf("empty field name in request for slave %d", def.SlaveID)
			}

			// Check output type
			if def.RegisterType == "holding" || def.RegisterType == "input" {
				switch f.OutputType {
				case "", "INT64", "UINT64", "FLOAT64":
				default:
					return fmt.Errorf("unknown output data-type %q for field %q", f.OutputType, f.Name)
				}
			} else {
				// Bit register types can only be UINT64 or BOOL
				switch f.OutputType {
				case "", "UINT16", "BOOL":
				default:
					return fmt.Errorf("unknown output data-type %q for field %q", f.OutputType, f.Name)
				}
			}

			// Handle the default for measurement
			if f.Measurement == "" {
				f.Measurement = def.Measurement
			}
			def.Fields[fidx] = f

			// Check for duplicate field definitions
			id, err := c.fieldID(seed, def, f)
			if err != nil {
				return fmt.Errorf("cannot determine field id for %q: %w", f.Name, err)
			}
			if seenFields[id] {
				return fmt.Errorf("field %q duplicated in measurement %q (slave %d/%q)", f.Name, f.Measurement, def.SlaveID, def.RegisterType)
			}
			seenFields[id] = true
		}
	}

	return nil
}

func (c *ConfigurationPerRequest) Process() (map[byte]requestSet, error) {
	result := map[byte]requestSet{}

	for _, def := range c.Requests {
		// Set default
		if def.RegisterType == "" {
			def.RegisterType = "holding"
		}

		// Construct the fields
		isTyped := def.RegisterType == "holding" || def.RegisterType == "input"
		fields, err := c.initFields(def.Fields, isTyped, def.ByteOrder)
		if err != nil {
			return nil, err
		}

		// Make sure we have a set to work with
		set, found := result[def.SlaveID]
		if !found {
			set = requestSet{
				coil:     []request{},
				discrete: []request{},
				holding:  []request{},
				input:    []request{},
			}
		}

		params := groupingParams{
			MaxExtraRegisters: def.MaxExtraRegisters,
			Optimization:      def.Optimization,
			Tags:              def.Tags,
			Log:               c.logger,
		}
		switch def.RegisterType {
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
			return nil, fmt.Errorf("unknown register type %q", def.RegisterType)
		}
		if !set.Empty() {
			result[def.SlaveID] = set
		}
	}

	return result, nil
}

func (c *ConfigurationPerRequest) initFields(fieldDefs []requestFieldDefinition, typed bool, byteOrder string) ([]field, error) {
	// Construct the fields from the field definitions
	fields := make([]field, 0, len(fieldDefs))
	for _, def := range fieldDefs {
		f, err := c.newFieldFromDefinition(def, typed, byteOrder)
		if err != nil {
			return nil, fmt.Errorf("initializing field %q failed: %w", def.Name, err)
		}
		fields = append(fields, f)
	}

	return fields, nil
}

func (c *ConfigurationPerRequest) newFieldFromDefinition(def requestFieldDefinition, typed bool, byteOrder string) (field, error) {
	var err error

	fieldLength := uint16(1)
	if typed {
		if fieldLength, err = c.determineFieldLength(def.InputType); err != nil {
			return field{}, err
		}
	}

	// Initialize the field
	f := field{
		measurement: def.Measurement,
		name:        def.Name,
		address:     def.Address,
		length:      fieldLength,
		omit:        def.Omit,
	}

	// Handle type conversions for coil and discrete registers
	if !typed {
		f.converter, err = determineUntypedConverter(def.OutputType)
		if err != nil {
			return field{}, err
		}
	}

	// No more processing for un-typed (coil and discrete registers) or omitted fields
	if !typed || def.Omit {
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

func (c *ConfigurationPerRequest) fieldID(seed maphash.Seed, def requestDefinition, field requestFieldDefinition) (uint64, error) {
	var mh maphash.Hash
	mh.SetSeed(seed)

	if err := mh.WriteByte(def.SlaveID); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(def.RegisterType); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(field.Measurement); err != nil {
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

func (c *ConfigurationPerRequest) determineOutputDatatype(input string) (string, error) {
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

func (c *ConfigurationPerRequest) determineFieldLength(input string) (uint16, error) {
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
