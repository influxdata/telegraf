package modbus

import (
	"fmt"
)

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
}

func (c *ConfigurationOriginal) Process() (map[byte]requestSet, error) {
	coil, err := c.initRequests(c.Coils, cCoils, maxQuantityCoils)
	if err != nil {
		return nil, err
	}

	discrete, err := c.initRequests(c.DiscreteInputs, cDiscreteInputs, maxQuantityDiscreteInput)
	if err != nil {
		return nil, err
	}

	holding, err := c.initRequests(c.HoldingRegisters, cHoldingRegisters, maxQuantityHoldingRegisters)
	if err != nil {
		return nil, err
	}

	input, err := c.initRequests(c.InputRegisters, cInputRegisters, maxQuantityInputRegisters)
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

func (c *ConfigurationOriginal) initRequests(fieldDefs []fieldDefinition, registerType string, maxQuantity uint16) ([]request, error) {
	fields, err := c.initFields(fieldDefs)
	if err != nil {
		return nil, err
	}
	return newRequestsFromFields(fields, c.SlaveID, registerType, maxQuantity), nil
}

func (c *ConfigurationOriginal) initFields(fieldDefs []fieldDefinition) ([]field, error) {
	// Construct the fields from the field definitions
	fields := make([]field, 0, len(fieldDefs))
	for _, def := range fieldDefs {
		f, err := c.newFieldFromDefinition(def)
		if err != nil {
			return nil, fmt.Errorf("initializing field %q failed: %v", def.Name, err)
		}
		fields = append(fields, f)
	}

	return fields, nil
}

func (c *ConfigurationOriginal) newFieldFromDefinition(def fieldDefinition) (field, error) {
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
		scale:       def.Scale,
		address:     def.Address[0],
		length:      uint16(len(def.Address)),
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
			return fmt.Errorf("empty name in '%s'", registerType)
		}

		//search name duplicate
		canonicalName := item.Measurement + "." + item.Name
		if nameEncountered[canonicalName] {
			return fmt.Errorf("name '%s' is duplicated in measurement '%s' '%s' - '%s'", item.Name, item.Measurement, registerType, item.Name)
		}
		nameEncountered[canonicalName] = true

		if registerType == cInputRegisters || registerType == cHoldingRegisters {
			// search byte order
			switch item.ByteOrder {
			case "AB", "BA", "ABCD", "CDAB", "BADC", "DCBA", "ABCDEFGH", "HGFEDCBA", "BADCFEHG", "GHEFCDAB":
			default:
				return fmt.Errorf("invalid byte order '%s' in '%s' - '%s'", item.ByteOrder, registerType, item.Name)
			}

			// search data type
			switch item.DataType {
			case "UINT16", "INT16", "UINT32", "INT32", "UINT64", "INT64", "FLOAT32-IEEE", "FLOAT64-IEEE", "FLOAT32", "FIXED", "UFIXED":
			default:
				return fmt.Errorf("invalid data type '%s' in '%s' - '%s'", item.DataType, registerType, item.Name)
			}

			// check scale
			if item.Scale == 0.0 {
				return fmt.Errorf("invalid scale '%f' in '%s' - '%s'", item.Scale, registerType, item.Name)
			}
		}

		// check address
		if len(item.Address) != 1 && len(item.Address) != 2 && len(item.Address) != 4 {
			return fmt.Errorf("invalid address '%v' length '%v' in '%s' - '%s'", item.Address, len(item.Address), registerType, item.Name)
		}

		if registerType == cInputRegisters || registerType == cHoldingRegisters {
			if 2*len(item.Address) != len(item.ByteOrder) {
				return fmt.Errorf("invalid byte order '%s' and address '%v'  in '%s' - '%s'", item.ByteOrder, item.Address, registerType, item.Name)
			}

			// search duplicated
			if len(item.Address) > len(removeDuplicates(item.Address)) {
				return fmt.Errorf("duplicate address '%v'  in '%s' - '%s'", item.Address, registerType, item.Name)
			}
		} else if len(item.Address) != 1 {
			return fmt.Errorf("invalid address'%v' length'%v' in '%s' - '%s'", item.Address, len(item.Address), registerType, item.Name)
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
