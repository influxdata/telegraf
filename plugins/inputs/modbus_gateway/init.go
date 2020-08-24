package modbus_gateway

import (
	"fmt"
	"strings"
)

func (m *ModbusGateway) Init() error {
	for i := range m.Requests {
		request := &m.Requests[i]

		/*
		 * If no register type was specified, default to "holding"
		 */
		if request.RequestType == "" {
			request.RequestType = "holding"
		} else {
			/*
			 * User specified the register type - make sure they made a valid selection
			 */
			request.RequestType = strings.ToLower(request.RequestType)
			if request.RequestType != "holding" && request.RequestType != "input" {
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

			field.InputType = strings.ToUpper(field.InputType)
			if field.InputType == "" {
				field.InputType = "UINT16"
			}

			field.OutputFormat = strings.ToUpper(field.OutputFormat)
			if field.OutputFormat == "" {
				field.OutputFormat = "FLOAT64"
			} else {
				switch field.OutputFormat {
				case "INT", "UINT", "INT64", "UINT64", "FLOAT", "FLOAT32", "FLOAT64":
					break
				default:
					return fmt.Errorf("Invalid output format")

				}
			}

		}
	}

	/* Default order is ABCD */
	if m.Order == "" {
		m.Order = "ABCD"
	}
	return nil
}
