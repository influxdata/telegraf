package modbus_gateway

import (
	"fmt"
	"strings"
)

func (m *ModbusGateway) Init() error {
	for i := range m.Requests {
		request := &m.Requests[i]

		/*
		 * If no register type was specified, default to "holding
		 */
		if request.Type == "" {
			request.Type = "holding"
		} else {
			/*
			 * User specified the register type - make sure they made a valid selection
			 */
			request.Type = strings.ToLower(request.Type)
			if request.Type != "holding" && request.Type != "input" {
				return fmt.Errorf("Request type must be \"holding\" or \"input\"")
			}
		}

		/*
		 * Check field mappings
		 */
		for j := range m.Requests[i].Fields {
			field := &m.Requests[i].Fields[j]

			if field.Scale == 0.0 {
				field.Scale = 1.0
			}

			field.Type = strings.ToUpper(field.Type)
			if field.Type == "" {
				field.Type = "UINT16"
			}
		}
	}
	return nil
}
