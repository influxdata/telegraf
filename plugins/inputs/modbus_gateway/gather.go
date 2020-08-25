package modbus_gateway

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/prometheus/common/log"
	"math"
	"time"
)

func (m *ModbusGateway) Gather(acc telegraf.Accumulator) error {
	if !m.isConnected {
		err := connect(m)
		if err != nil {
			m.isConnected = false
			return err
		}
	}

	grouper := metric.NewSeriesGrouper()

	for _, req := range m.Requests {
		now := time.Now()
		m.tcpHandler.SlaveId = req.Unit
		var resp []byte
		var err error

		if req.RequestType == "holding" {
			resp, err = m.client.ReadHoldingRegisters(req.Address, req.Count)
		} else if req.RequestType == "input" {
			resp, err = m.client.ReadInputRegisters(req.Address, req.Count)
		} else {
			return fmt.Errorf("Don't know how to poll register type \"%s\"", req.RequestType)
		}

		if err == nil {
			reader := bytes.NewReader(resp)

			for _, f := range req.Fields {
				/*
				 * Look up the byte ordering for this request, with higher level overrides
				 */
				var orderSpec string
				if f.Order != "" {
					orderSpec = f.Order
				} else if req.Order != "" {
					orderSpec = req.Order
				} else if m.Order != "" {
					orderSpec = m.Order
				} else {
					orderSpec = "ABCD"
				}
				var byteOrder *CustomByteOrder = getOrCreateByteOrder(orderSpec)

				switch f.InputType {
				case "UINT16":
					var value uint16
					binary.Read(reader, byteOrder, &value)
					if f.Omit == false {
						grouper.Add(req.MeasurementName, nil, now, f.Name, scale(&f, value))
					}
					break
				case "INT16":
					var value int16
					binary.Read(reader, byteOrder, &value)
					if f.Omit == false {
						grouper.Add(req.MeasurementName, nil, now, f.Name, scale(&f, value))
					}
					break
				case "UINT32":
					var value uint32
					binary.Read(reader, byteOrder, &value)
					if f.Omit == false {
						grouper.Add(req.MeasurementName, nil, now, f.Name, scale(&f, value))
					}
					break
				case "INT32":
					var value int32
					binary.Read(reader, byteOrder, &value)
					if f.Omit == false {
						grouper.Add(req.MeasurementName, nil, now, f.Name, scale(&f, value))
					}
					break
				default:
					m.Log.Warnf("Invalid conversion type %s", f.InputType)
				}
			}
		} else {
			m.Log.Info("Modbus Error: ", err)
		}
	}

	for _, metric := range grouper.Metrics() {
		m.Log.Infof("write %v", metric)
		acc.AddMetric(metric)
	}

	return nil
}

func scale(f *FieldDef, value interface{}) interface{} {
	switch f.OutputFormat {
	case "FLOAT", "FLOAT64":
		switch v := value.(type) {
		case int:
			return float64((float64(v) * f.Scale) + f.Offset)
		case int16:
			return float64((float64(v) * f.Scale) + f.Offset)
		case uint16:
			return float64((float64(v) * f.Scale) + f.Offset)
		case int32:
			return float64((float64(v) * f.Scale) + f.Offset)
		case uint32:
			return float64((float64(v) * f.Scale) + f.Offset)
		default:
			return nil
		}

	case "FLOAT32":
		switch v := value.(type) {
		case int:
			return float32((float64(v) * f.Scale) + f.Offset)
		case int16:
			return float32((float64(v) * f.Scale) + f.Offset)
		case uint16:
			return float32((float64(v) * f.Scale) + f.Offset)
		case int32:
			return float32((float64(v) * f.Scale) + f.Offset)
		case uint32:
			return float32((float64(v) * f.Scale) + f.Offset)
		default:
			return nil
		}

	case "INT", "INT64":
		switch v := value.(type) {
		case int:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int16:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint16:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int32:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint32:
			return int64(math.Round((float64(v) * f.Scale) + f.Offset))
		default:
			return nil
		}

	case "UINT", "UINT64":
		switch v := value.(type) {
		case int:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int16:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint16:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case int32:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		case uint32:
			return uint64(math.Round((float64(v) * f.Scale) + f.Offset))
		default:
			return nil
		}

	default:
		log.Warn("Invalid output format")
		return nil
	}
}
