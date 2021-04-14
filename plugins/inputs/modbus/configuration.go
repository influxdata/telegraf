package modbus

import (
	"fmt"
	"sort"
)

func (m *Modbus) initRegisters() error {
	err := m.initRegister(m.DiscreteInputs, cDiscreteInputs)
	if err != nil {
		return err
	}

	err = m.initRegister(m.Coils, cCoils)
	if err != nil {
		return err
	}

	err = m.initRegister(m.HoldingRegisters, cHoldingRegisters)
	if err != nil {
		return err
	}

	return m.initRegister(m.InputRegisters, cInputRegisters)
}

func (m *Modbus) initRegister(fields []fieldContainer, name string) error {
	if len(fields) == 0 {
		return nil
	}

	err := validateFieldContainers(fields, name)
	if err != nil {
		return err
	}

	addrs := []uint16{}
	for _, field := range fields {
		addrs = append(addrs, field.Address...)
	}

	addrs = removeDuplicates(addrs)
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

	ii := 0
	maxQuantity := 1
	var registersRange []registerRange
	if name == cDiscreteInputs || name == cCoils {
		maxQuantity = 2000
	} else if name == cInputRegisters || name == cHoldingRegisters {
		maxQuantity = 125
	}

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

	m.registers = append(m.registers, register{name, registersRange, fields})

	return nil
}

func validateFieldContainers(t []fieldContainer, n string) error {
	nameEncountered := map[string]bool{}
	for _, item := range t {
		//check empty name
		if item.Name == "" {
			return fmt.Errorf("empty name in '%s'", n)
		}

		//search name duplicate
		canonicalName := item.Measurement + "." + item.Name
		if nameEncountered[canonicalName] {
			return fmt.Errorf("name '%s' is duplicated in measurement '%s' '%s' - '%s'", item.Name, item.Measurement, n, item.Name)
		}
		nameEncountered[canonicalName] = true

		if n == cInputRegisters || n == cHoldingRegisters {
			// search byte order
			switch item.ByteOrder {
			case "AB", "BA", "ABCD", "CDAB", "BADC", "DCBA", "ABCDEFGH", "HGFEDCBA", "BADCFEHG", "GHEFCDAB":
			default:
				return fmt.Errorf("invalid byte order '%s' in '%s' - '%s'", item.ByteOrder, n, item.Name)
			}

			// search data type
			switch item.DataType {
			case "UINT16", "INT16", "UINT32", "INT32", "UINT64", "INT64", "FLOAT32-IEEE", "FLOAT64-IEEE", "FLOAT32", "FIXED", "UFIXED":
			default:
				return fmt.Errorf("invalid data type '%s' in '%s' - '%s'", item.DataType, n, item.Name)
			}

			// check scale
			if item.Scale == 0.0 {
				return fmt.Errorf("invalid scale '%f' in '%s' - '%s'", item.Scale, n, item.Name)
			}
		}

		// check address
		if len(item.Address) != 1 && len(item.Address) != 2 && len(item.Address) != 4 {
			return fmt.Errorf("invalid address '%v' length '%v' in '%s' - '%s'", item.Address, len(item.Address), n, item.Name)
		}

		if n == cInputRegisters || n == cHoldingRegisters {
			if 2*len(item.Address) != len(item.ByteOrder) {
				return fmt.Errorf("invalid byte order '%s' and address '%v'  in '%s' - '%s'", item.ByteOrder, item.Address, n, item.Name)
			}

			// search duplicated
			if len(item.Address) > len(removeDuplicates(item.Address)) {
				return fmt.Errorf("duplicate address '%v'  in '%s' - '%s'", item.Address, n, item.Name)
			}
		} else if len(item.Address) != 1 {
			return fmt.Errorf("invalid address'%v' length'%v' in '%s' - '%s'", item.Address, len(item.Address), n, item.Name)
		}
	}
	return nil
}

func removeDuplicates(elements []uint16) []uint16 {
	encountered := map[uint16]bool{}
	result := []uint16{}

	for _, addr := range elements {
		if ! encountered[addr] {
			encountered[addr] = true
			result = append(result, addr)
		}
	}

	return result
}
