package modbus

import (
	_ "embed"
	"fmt"
)

//go:embed sample_register.conf
var sampleConfigPartPerRegister string

type fieldDefinition struct {
	Measurement string   `toml:"measurement"`
	Name        string   `toml:"name"`
	ByteOrder   string   `toml:"byte_order"`
	DataType    string   `toml:"data_type"`
	Scale       float64  `toml:"scale"`
	Address     []uint16 `toml:"address"`
}

type ConfigurationOriginal struct {
	SlaveID          byte              `toml:"slave_id"`
	DiscreteInputs   []fieldDefinition `toml:"discrete_inputs"`
	Coils            []fieldDefinition `toml:"coils"`
	HoldingRegisters []fieldDefinition `toml:"holding_registers"`
	InputRegisters   []fieldDefinition `toml:"input_registers"`
	workarounds      ModbusWorkarounds
}

func (c *ConfigurationOriginal) SampleConfigPart() string {
	return sampleConfigPartPerRegister
}

func (c *ConfigurationOriginal) Check() error {
	if err := c.validateFieldDefinitions(c.DiscreteInputs, cDiscreteInputs); err != nil {
		return err
	}

	if err := c.validateFieldDefinitions(c.Coils, cCoils); err != nil {
		return err
	}

	if err := c.validateFieldDefinitions(c.HoldingRegisters, cHoldingRegisters); err != nil {
		return err
	}

	return c.validateFieldDefinitions(c.InputRegisters, cInputRegisters)
}

func (c *ConfigurationOriginal) Process() (map[byte]requestSet, error) {
	maxQuantity := uint16(1)
	if !c.workarounds.OnRequestPerField {
		maxQuantity = maxQuantityCoils
	}
	coil, err := c.initRequests(c.Coils, maxQuantity, false)
	if err != nil {
		return nil, err
	}

	if !c.workarounds.OnRequestPerField {
		maxQuantity = maxQuantityDiscreteInput
	}
	discrete, err := c.initRequests(c.DiscreteInputs, maxQuantity, false)
	if err != nil {
		return nil, err
	}

	if !c.workarounds.OnRequestPerField {
		maxQuantity = maxQuantityHoldingRegisters
	}
	holding, err := c.initRequests(c.HoldingRegisters, maxQuantity, true)
	if err != nil {
		return nil, err
	}

	if !c.workarounds.OnRequestPerField {
		maxQuantity = maxQuantityInputRegisters
	}
	input, err := c.initRequests(c.InputRegisters, maxQuantity, true)
	if err != nil {
		return nil, err
	}

	return map[byte]requestSet{
		c.SlaveID: {
			coil:     coil,
			discrete: discrete,
			holding:  holding,
			input:    input,
		},
	}, nil
}

func (c *ConfigurationOriginal) initRequests(fieldDefs []fieldDefinition, maxQuantity uint16, typed bool) ([]request, error) {
	fields, err := c.initFields(fieldDefs, typed)
	if err != nil {
		return nil, err
	}
	params := groupingParams{
		MaxBatchSize:    maxQuantity,
		Optimization:    "none",
		EnforceFromZero: c.workarounds.ReadCoilsStartingAtZero,
	}

	return groupFieldsToRequests(fields, params), nil
}

func (c *ConfigurationOriginal) initFields(fieldDefs []fieldDefinition, typed bool) ([]field, error) {
	// Construct the fields from the field definitions
	fields := make([]field, 0, len(fieldDefs))
	for _, def := range fieldDefs {
		f, err := c.newFieldFromDefinition(def, typed)
		if err != nil {
			return nil, fmt.Errorf("initializing field %q failed: %w", def.Name, err)
		}
		fields = append(fields, f)
	}

	return fields, nil
}

func (c *ConfigurationOriginal) newFieldFromDefinition(def fieldDefinition, typed bool) (field, error) {
	// Check if the addresses are consecutive
	expected := def.Address[0]
	for _, current := range def.Address[1:] {
		expected++
		if current != expected {
			return field{}, fmt.Errorf("addresses of field %q are not consecutive", def.Name)
		}
	}

	// Initialize the field
	f := field{
		measurement: def.Measurement,
		name:        def.Name,
		address:     def.Address[0],
		length:      uint16(len(def.Address)),
	}

	// Handle coil and discrete registers which do have a limited datatype set
	if !typed {
		var err error
		f.converter, err = determineUntypedConverter(def.DataType)
		if err != nil {
			return field{}, err
		}
		return f, nil
	}

	if def.DataType != "" {
		inType, err := c.normalizeInputDatatype(def.DataType, len(def.Address))
		if err != nil {
			return f, err
		}
		outType, err := c.normalizeOutputDatatype(def.DataType)
		if err != nil {
			return f, err
		}
		byteOrder, err := c.normalizeByteOrder(def.ByteOrder)
		if err != nil {
			return f, err
		}

		f.converter, err = determineConverter(inType, byteOrder, outType, def.Scale)
		if err != nil {
			return f, err
		}
	}

	return f, nil
}

func (c *ConfigurationOriginal) validateFieldDefinitions(fieldDefs []fieldDefinition, registerType string) error {
	nameEncountered := map[string]bool{}
	for _, item := range fieldDefs {
		//check empty name
		if item.Name == "" {
			return fmt.Errorf("empty name in %q", registerType)
		}

		//search name duplicate
		canonicalName := item.Measurement + "." + item.Name
		if nameEncountered[canonicalName] {
			return fmt.Errorf("name %q is duplicated in measurement %q %q - %q", item.Name, item.Measurement, registerType, item.Name)
		}
		nameEncountered[canonicalName] = true

		if registerType == cInputRegisters || registerType == cHoldingRegisters {
			// search byte order
			switch item.ByteOrder {
			case "AB", "BA", "ABCD", "CDAB", "BADC", "DCBA", "ABCDEFGH", "HGFEDCBA", "BADCFEHG", "GHEFCDAB":
			default:
				return fmt.Errorf("invalid byte order %q in %q - %q", item.ByteOrder, registerType, item.Name)
			}

			// search data type
			switch item.DataType {
			case "INT8L", "INT8H", "UINT8L", "UINT8H",
				"UINT16", "INT16", "UINT32", "INT32", "UINT64", "INT64",
				"FLOAT16-IEEE", "FLOAT32-IEEE", "FLOAT64-IEEE", "FLOAT32", "FIXED", "UFIXED":
			default:
				return fmt.Errorf("invalid data type %q in %q - %q", item.DataType, registerType, item.Name)
			}

			// check scale
			if item.Scale == 0.0 {
				return fmt.Errorf("invalid scale '%f' in %q - %q", item.Scale, registerType, item.Name)
			}
		} else {
			// Bit-registers do have less data types
			switch item.DataType {
			case "", "UINT16", "BOOL":
			default:
				return fmt.Errorf("invalid data type %q in %q - %q", item.DataType, registerType, item.Name)
			}
		}

		// check address
		if len(item.Address) != 1 && len(item.Address) != 2 && len(item.Address) != 4 {
			return fmt.Errorf("invalid address '%v' length '%v' in %q - %q", item.Address, len(item.Address), registerType, item.Name)
		}

		if registerType == cInputRegisters || registerType == cHoldingRegisters {
			if 2*len(item.Address) != len(item.ByteOrder) {
				return fmt.Errorf("invalid byte order %q and address '%v'  in %q - %q", item.ByteOrder, item.Address, registerType, item.Name)
			}

			// search duplicated
			if len(item.Address) > len(removeDuplicates(item.Address)) {
				return fmt.Errorf("duplicate address '%v'  in %q - %q", item.Address, registerType, item.Name)
			}
		} else if len(item.Address) != 1 {
			return fmt.Errorf("invalid address'%v' length'%v' in %q - %q", item.Address, len(item.Address), registerType, item.Name)
		}
	}
	return nil
}

func (c *ConfigurationOriginal) normalizeInputDatatype(dataType string, words int) (string, error) {
	// Handle our special types
	switch dataType {
	case "FIXED":
		switch words {
		case 1:
			return "INT16", nil
		case 2:
			return "INT32", nil
		case 4:
			return "INT64", nil
		default:
			return "unknown", fmt.Errorf("invalid length %d for type %q", words, dataType)
		}
	case "FLOAT32", "UFIXED":
		switch words {
		case 1:
			return "UINT16", nil
		case 2:
			return "UINT32", nil
		case 4:
			return "UINT64", nil
		default:
			return "unknown", fmt.Errorf("invalid length %d for type %q", words, dataType)
		}
	case "FLOAT16-IEEE":
		return "FLOAT16", nil
	case "FLOAT32-IEEE":
		return "FLOAT32", nil
	case "FLOAT64-IEEE":
		return "FLOAT64", nil
	}
	return normalizeInputDatatype(dataType)
}

func (c *ConfigurationOriginal) normalizeOutputDatatype(dataType string) (string, error) {
	// Handle our special types
	switch dataType {
	case "FIXED", "FLOAT32", "UFIXED":
		return "FLOAT64", nil
	}
	return normalizeOutputDatatype("native")
}

func (c *ConfigurationOriginal) normalizeByteOrder(byteOrder string) (string, error) {
	// Handle our special types
	switch byteOrder {
	case "AB", "ABCDEFGH":
		return "ABCD", nil
	case "BADCFEHG":
		return "BADC", nil
	case "GHEFCDAB":
		return "CDAB", nil
	case "BA", "HGFEDCBA":
		return "DCBA", nil
	}
	return normalizeByteOrder(byteOrder)
}
