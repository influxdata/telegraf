package hc3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
)

func Parse(acc telegraf.Accumulator, sectionBytes []byte, roomBytes []byte, deviecsBytes []byte) error {
	var tmpSections []Sections
	if err := json.Unmarshal(sectionBytes, &tmpSections); err != nil {
		return err
	}
	sections := map[uint16]string{}
	for _, v := range tmpSections {
		sections[v.ID] = v.Name
	}

	var tmpRooms []Rooms
	if err := json.Unmarshal(roomBytes, &tmpRooms); err != nil {
		return err
	}
	rooms := map[uint16]LinkRoomsSections{}
	for _, v := range tmpRooms {
		rooms[v.ID] = LinkRoomsSections{Name: v.Name, SectionID: v.SectionID}
	}

	var devices []Devices
	if err := json.Unmarshal(deviecsBytes, &devices); err != nil {
		return err
	}

	for _, device := range devices {
		// skip device in some cases
		if device.RoomID == 0 ||
			!device.Enabled ||
			device.Properties.Dead ||
			device.Type == "com.fibaro.zwaveDevice" {
			continue
		}

		tags := map[string]string{
			"deviceId": strconv.FormatUint(uint64(device.ID), 10),
			"section":  sections[rooms[device.RoomID].SectionID],
			"room":     rooms[device.RoomID].Name,
			"name":     device.Name,
			"type":     device.Type,
		}
		fields := make(map[string]interface{})

		if device.Properties.BatteryLevel != nil {
			fields["batteryLevel"] = *device.Properties.BatteryLevel
		}

		if device.Properties.Energy != nil {
			fields["energy"] = *device.Properties.Energy
		}

		if device.Properties.Power != nil {
			fields["power"] = *device.Properties.Power
		}

		// Value can be a JSON bool, string, or numeric value
		if device.Properties.Value != nil {
			switch v := device.Properties.Value.(type) {
			case string:
				if fValue, err := strconv.ParseFloat(v, 64); err == nil {
					fields["value"] = fValue
				}
			case bool:
				switch v {
				case true:
					fields["value"] = 1.0
				case false:
					fields["value"] = 0.0
				}
			case float64:
				fields["value"] = v
			case int:
				fields["value"] = float64(v)
			default:
				acc.AddError(fmt.Errorf("unknown value type %T: %s", v, v))
			}
		}

		acc.AddFields("fibaro", fields, tags)
	}

	return nil
}
