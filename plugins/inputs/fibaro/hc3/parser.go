package hc3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

func Parse(acc telegraf.Accumulator, sectionBytes []byte, roomBytes []byte, deviecsBytes []byte) error {
	var tmpSections []Sections
	if err := json.Unmarshal(sectionBytes, &tmpSections); err != nil {
		return err
	}
	sections := make(map[uint16]string, len(tmpSections))
	for _, v := range tmpSections {
		sections[v.ID] = v.Name
	}

	var tmpRooms []Rooms
	if err := json.Unmarshal(roomBytes, &tmpRooms); err != nil {
		return err
	}
	rooms := make(map[uint16]linkRoomsSections, len(tmpRooms))
	for _, v := range tmpRooms {
		rooms[v.ID] = linkRoomsSections{Name: v.Name, SectionID: v.SectionID}
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
			v, err := internal.ToFloat64(device.Properties.Value)
			if err != nil {
				acc.AddError(fmt.Errorf("unable to convert value: %w", err))
			} else {
				fields["value"] = v
			}
		}

		acc.AddFields("fibaro", fields, tags)
	}

	return nil
}
