package modbus

import (
	"fmt"
	"sort"
)

type fieldDefinition struct {
	Measurement string   `toml:"measurement"`
	Name        string   `toml:"name"`
	ByteOrder   string   `toml:"byte_order"`
	DataType    string   `toml:"data_type"`
	Scale       float64  `toml:"scale"`
	Address     []uint16 `toml:"address"`
	value       interface{}
}

type ConfigurationOriginal struct {
	SlaveID          int               `toml:"slave_id"`
	DiscreteInputs   []fieldDefinition `toml:"discrete_inputs"`
	Coils            []fieldDefinition `toml:"coils"`
	HoldingRegisters []fieldDefinition `toml:"holding_registers"`
	InputRegisters   []fieldDefinition `toml:"input_registers"`
}

func (c *ConfigurationOriginal) Process(m *Modbus) error {
	r, err := m.initRequests(c.DiscreteInputs, cDiscreteInputs, maxQuantityDiscreteInput)
	if err != nil {
		return err
	}
	m.requests = append(m.requests, r...)

	r, err = c.initRequests(c.Coils, cCoils, maxQuantityCoils)
	if err != nil {
		return err
	}
	m.requests = append(m.requests, r...)

	r, err = m.initRequests(c.HoldingRegisters, cHoldingRegisters, maxQuantityHoldingRegisters)
	if err != nil {
		return err
	}
	m.requests = append(m.requests, r...)

	r, err = c.initRequests(m.InputRegisters, cInputRegisters, maxQuantityInputRegisters)
	if err != nil {
		return err
	}
	m.requests = append(m.requests, r...)

	return nil
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

	if err := c.validateFieldDefinitions(c.InputRegisters, cInputRegisters); err != nil {
		return err
	}

	return nil
}

func (c *ConfigurationOriginal) initRequests(fieldDefs []fieldDefinition, registerType string, maxQuantity uint16) ([]request, error) {
	return c.initRequestsPerSlaveAndType(fieldDefs, c.SlaveID, registerType, maxQuantity)
}

func (c *ConfigurationOriginal) initRequestsPerSlaveAndType(fieldDefs []fieldDefinition, slaveID int, registerType string, maxQuantity uint16) ([]request, error) {
	if len(fieldDefs) < 1 {
		return nil, nil
	}

	// Construct the fields from the field definitions
	fields := make([]field, 0, len(fieldDefs))
	for _, def := range fieldDefs {
		f, err := c.initField(def)
		if err != nil {
			return nil, fmt.Errorf("initializing field %q failed: %v", def.Name, err)
		}
		fields = append(fields, f)
	}

	// Sort the fields by address (ascending) and length
	sort.Slice(fields, func(i, j int) bool {
		addrI := fields[i].address
		addrJ := fields[j].address
		return addrI < addrJ || (addrI == addrJ && fields[i].length > fields[j].length)
	})

	// Construct the consecutive register chunks for the addresses and construct Modbus requests.
	// For field addresses like [1, 2, 3, 5, 6, 10, 11, 12, 14] we should construct the following
	// requests (1, 3) , (5, 2) , (10, 3), (14 , 1). Furthermore, we should respect field boundaries
	// and the given maximum chunk sizes.
	var requests []request

	current := request{
		slaveID:      c.SlaveID,
		registerType: registerType,
		address:      fields[0].address,
		length:       fields[0].length,
		fields:       []field{fields[0]},
	}

	for _, f := range fields[1:] {
		// Check if we need to interrupt the current chunk and require a new one
		needInterrupt := f.address != current.address+current.length           // not consecutive
		needInterrupt = needInterrupt || f.length+current.length > maxQuantity // too large

		if !needInterrupt {
			// Still save to add the field to the current request
			current.length += f.length
			current.fields = append(current.fields, f) // TODO: omit the field with a future flag
			continue
		}

		// Finish the current request, add it to the list and construct a new one
		requests = append(requests, current)
		current = request{
			slaveID:      c.SlaveID,
			registerType: registerType,
			address:      f.address,
			length:       f.length,
			fields:       []field{f},
		}
	}
	requests = append(requests, current)

	return requests, nil
}

func (c *ConfigurationOriginal) initField(def fieldDefinition) (field, error) {
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
	case "FLOAT32-IEEE", "FLOAT64-IEEE":
		return "FLOAT64", nil
	}
	return normalizeOutputDatatype(dataType)
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
