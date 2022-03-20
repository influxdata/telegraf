package modbus

import (
	"fmt"
	"hash/maphash"
)

const sampleConfigPartPerRequest = `
  ## Per request definition
  ##

  ## Define a request sent to the device
  ## Multiple of those requests can be defined. Data will be collated into metrics at the end of data collection.
  # [[inputs.modbus.request]]
    ## ID of the modbus slave device to query.
    ## If you need to query multiple slave-devices, create several "request" definitions.
    # slave_id = 0

    ## Byte order of the data.
    ##  |---ABCD or MSW-BE -- Big Endian (Motorola)
    ##  |---DCBA or LSW-LE -- Little Endian (Intel)
    ##  |---BADC or MSW-LE -- Big Endian with byte swap
    ##  |---CDAB or LSW-BE -- Little Endian with byte swap
    # byte_order = "ABCD"

    ## Type of the register for the request
    ## Can be "coil", "discrete", "holding" or "input"
    # register = "holding"

    ## Name of the measurement.
    ## Can be overriden by the individual field definitions. Defaults to "modbus"
    # measurement = "modbus"

    ## Field definitions
    ## Analog Variables, Input Registers and Holding Registers
    ## address        - address of the register to query. For coil and discrete inputs this is the bit address.
    ## name *1        - field name
    ## type *1,2      - type of the modbus field, can be INT16, UINT16, INT32, UINT32, INT64, UINT64 and
    ##                  FLOAT32, FLOAT64 (IEEE 754 binary representation)
    ## scale *1,2     - (optional) factor to scale the variable with
    ## output *1,2    - (optional) type of resulting field, can be INT64, UINT64 or FLOAT64. Defaults to FLOAT64 if
    ##                  "scale" is provided and to the input "type" class otherwise (i.e. INT* -> INT64, etc).
    ## measurement *1 - (optional) measurement name, defaults to the setting of the request
    ## omit           - (optional) omit this field. Useful to leave out single values when querying many registers
    ##                  with a single request. Defaults to "false".
    ##
    ## *1: Those fields are ignored if field is omitted ("omit"=true)
    ##
    ## *2: Thise fields are ignored for both "coil" and "discrete"-input type of registers. For those register types
    ##     the fields are output as zero or one in UINT64 format by default.

    ## Coil / discrete input example
    # fields = [
    #   { address=0, name="motor1_run"},
    #   { address=1, name="jog", measurement="motor"},
    #   { address=2, name="motor1_stop", omit=true},
    #   { address=3, name="motor1_overheating"},
    # ]

    ## Per-request tags
    ## These tags take precedence over predefined tags.
    # [[inputs.modbus.request.tags]]
    #	  name = "value"

    ## Holding / input example
    ## All of those examples will result in FLOAT64 field outputs
    # fields = [
    #   { address=0, name="voltage",      type="INT16",   scale=0.1   },
    #   { address=1, name="current",      type="INT32",   scale=0.001 },
    #   { address=3, name="power",        type="UINT32",  omit=true   },
    #   { address=5, name="energy",       type="FLOAT32", scale=0.001, measurement="W" },
    #   { address=7, name="frequency",    type="UINT32",  scale=0.1   },
    #   { address=8, name="power_factor", type="INT64",   scale=0.01  },
    # ]

    ## Holding / input example with type conversions
    # fields = [
    #   { address=0, name="rpm",         type="INT16"                   },  # will result in INT64 field
    #   { address=1, name="temperature", type="INT16", scale=0.1        },  # will result in FLOAT64 field
    #   { address=2, name="force",       type="INT32", output="FLOAT64" },  # will result in FLOAT64 field
    #   { address=4, name="hours",       type="UINT32"                  },  # will result in UIN64 field
    # ]

    ## Per-request tags
		## These tags take precedence over predefined tags.
    # [[inputs.modbus.request.tags]]
    #	  name = "value"
`

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
	SlaveID      byte                     `toml:"slave_id"`
	ByteOrder    string                   `toml:"byte_order"`
	RegisterType string                   `toml:"register"`
	Measurement  string                   `toml:"measurement"`
	Fields       []requestFieldDefinition `toml:"fields"`
	Tags         map[string]string        `toml:"tags"`
}

type ConfigurationPerRequest struct {
	Requests []requestDefinition `toml:"request"`
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

		// Set the default for measurement if required
		if def.Measurement == "" {
			def.Measurement = "modbus"
		}

		// Check the fields
		for fidx, f := range def.Fields {
			// Check the input type for all fields except the bit-field ones.
			// We later need the type (even for omitted fields) to determine the length.
			if def.RegisterType == cHoldingRegisters || def.RegisterType == cInputRegisters {
				switch f.InputType {
				case "INT16", "UINT16", "INT32", "UINT32", "INT64", "UINT64", "FLOAT32", "FLOAT64":
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

			// Check fields only relevant for non-bit register types
			if def.RegisterType == cHoldingRegisters || def.RegisterType == cInputRegisters {
				// Check output type
				switch f.OutputType {
				case "", "INT64", "UINT64", "FLOAT64":
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
			id, err := c.fieldID(seed, def.SlaveID, def.RegisterType, def.Measurement, f.Name)
			if err != nil {
				return fmt.Errorf("cannot determine field id for %q: %v", f.Name, err)
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

		switch def.RegisterType {
		case "coil":
			requests := groupFieldsToRequests(fields, def.Tags, maxQuantityCoils)
			set.coil = append(set.coil, requests...)
		case "discrete":
			requests := groupFieldsToRequests(fields, def.Tags, maxQuantityDiscreteInput)
			set.discrete = append(set.discrete, requests...)
		case "holding":
			requests := groupFieldsToRequests(fields, def.Tags, maxQuantityHoldingRegisters)
			set.holding = append(set.holding, requests...)
		case "input":
			requests := groupFieldsToRequests(fields, def.Tags, maxQuantityInputRegisters)
			set.input = append(set.input, requests...)
		default:
			return nil, fmt.Errorf("unknown register type %q", def.RegisterType)
		}
		result[def.SlaveID] = set
	}

	return result, nil
}

func (c *ConfigurationPerRequest) initFields(fieldDefs []requestFieldDefinition, typed bool, byteOrder string) ([]field, error) {
	// Construct the fields from the field definitions
	fields := make([]field, 0, len(fieldDefs))
	for _, def := range fieldDefs {
		f, err := c.newFieldFromDefinition(def, typed, byteOrder)
		if err != nil {
			return nil, fmt.Errorf("initializing field %q failed: %v", def.Name, err)
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

func (c *ConfigurationPerRequest) fieldID(seed maphash.Seed, slave byte, register, measurement, name string) (uint64, error) {
	var mh maphash.Hash
	mh.SetSeed(seed)

	if err := mh.WriteByte(slave); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(register); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(measurement); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}
	if _, err := mh.WriteString(name); err != nil {
		return 0, err
	}
	if err := mh.WriteByte(0); err != nil {
		return 0, err
	}

	return mh.Sum64(), nil
}

func (c *ConfigurationPerRequest) determineOutputDatatype(input string) (string, error) {
	// Handle our special types
	switch input {
	case "INT16", "INT32", "INT64":
		return "INT64", nil
	case "UINT16", "UINT32", "UINT64":
		return "UINT64", nil
	case "FLOAT32", "FLOAT64":
		return "FLOAT64", nil
	}
	return "unknown", fmt.Errorf("invalid input datatype %q for determining output", input)
}

func (c *ConfigurationPerRequest) determineFieldLength(input string) (uint16, error) {
	// Handle our special types
	switch input {
	case "INT16", "UINT16":
		return 1, nil
	case "INT32", "UINT32", "FLOAT32":
		return 2, nil
	case "INT64", "UINT64", "FLOAT64":
		return 4, nil
	}
	return 0, fmt.Errorf("invalid input datatype %q for determining field length", input)
}
