package hc2

import (
	"encoding/json"
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
			device.Properties.Dead == "true" ||
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
			if fValue, err := strconv.ParseFloat(*device.Properties.BatteryLevel, 64); err == nil {
				fields["batteryLevel"] = fValue
			}
		}

		if device.Properties.Energy != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.Energy, 64); err == nil {
				fields["energy"] = fValue
			}
		}

		if device.Properties.Power != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.Power, 64); err == nil {
				fields["power"] = fValue
			}
		}

		if device.Properties.Value != nil {
			value := device.Properties.Value
			switch value {
			case "true":
				value = "1"
			case "false":
				value = "0"
			}

			if fValue, err := strconv.ParseFloat(value.(string), 64); err == nil {
				fields["value"] = fValue
			}
		}

		if device.Properties.Value2 != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.Value2, 64); err == nil {
				fields["value2"] = fValue
			}
		}

		acc.AddFields("fibaro", fields, tags)
	}

	return nil
}
