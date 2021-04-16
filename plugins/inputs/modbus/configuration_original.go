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
	if len(c.DiscreteInputs) > 0 {
		r, err := m.initRegister(c.DiscreteInputs, maxQuantityDiscreteInput)
		if err != nil {
			return err
		}
		r.Type = cDiscreteInputs
		m.requests = append(m.requests, r)
	}

	if len(c.Coils) > 0 {
		r, err := c.initRegister(c.Coils, maxQuantityCoils)
		if err != nil {
			return err
		}
		r.Type = cCoils
		m.requests = append(m.requests, r)
	}

	if len(c.HoldingRegisters) > 0 {
		r, err := m.initRegister(c.HoldingRegisters, maxQuantityHoldingRegisters)
		if err != nil {
			return err
		}
		r.Type = cHoldingRegisters
		m.requests = append(m.requests, r)
	}

	if len(c.InputRegisters) > 0 {
		r, err := c.initRegister(m.InputRegisters, maxQuantityInputRegisters)
		if err != nil {
			return err
		}
		r.Type = cInputRegisters
		m.requests = append(m.requests, r)
	}

	return nil
}

func (c *ConfigurationOriginal) Check() error {
	if len(c.DiscreteInputs) > 0 {
		if err := c.validateFieldDefinitions(c.DiscreteInputs, cDiscreteInputs); err != nil {
			return err
		}
	}

	if len(c.Coils) > 0 {
		if err := c.validateFieldDefinitions(c.Coils, cCoils); err != nil {
			return err
		}
	}

	if len(c.HoldingRegisters) > 0 {
		if err := c.validateFieldDefinitions(c.HoldingRegisters, cHoldingRegisters); err != nil {
			return err
		}
	}

	if len(c.InputRegisters) > 0 {
		if err := c.validateFieldDefinitions(c.InputRegisters, cInputRegisters); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigurationOriginal) initRegister(fieldDefs []fieldDefinition, maxQuantity int) (request, error) {
	addrs := []uint16{}
	for _, def := range fieldDefs {
		addrs = append(addrs, def.Address...)
	}

	fields := make([]field, 0, len(fieldDefs))
	for _, def := range fieldDefs {
		f, err := c.initField(def)
		if err != nil {
			return request{}, fmt.Errorf("initializing field %q failed: %v", def.Name, err)
		}
		fields = append(fields, f)
	}

	addrs = removeDuplicates(addrs)
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

	ii := 0

	var registersRange []registerRange

	// Get range of consecutive integers
	// [1, 2, 3, 5, 6, 10, 11, 12, 14]
	// (1, 3) , (5, 2) , (10, 3), (14 , 1)
	for range addrs {
		if ii >= len(addrs) {
			break
		}
		quantity := 1
		start := addrs[ii]
		end := start

		for ii < len(addrs)-1 && addrs[ii+1]-addrs[ii] == 1 && quantity < maxQuantity {
			end = addrs[ii+1]
			ii++
			quantity++
		}
		ii++

		registersRange = append(registersRange, registerRange{start, end - start + 1})
	}

	return request{
		SlaveID:        c.SlaveID,
		RegistersRange: registersRange,
		Fields:         fields,
	}, nil
}

func (c *ConfigurationOriginal) initField(def fieldDefinition) (field, error) {
	f := field{
		Measurement: def.Measurement,
		Name:        def.Name,
		Scale:       def.Scale,
		Address:     def.Address,
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
