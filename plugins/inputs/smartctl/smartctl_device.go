package smartctl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

func (s *Smartctl) scanDevice(acc telegraf.Accumulator, deviceName string, deviceType string) error {
	args := []string{"--json", "--all", deviceName, "--device", deviceType, "--nocheck=" + s.NoCheck}
	cmd := execCommand(s.Path, args...)
	if s.UseSudo {
		cmd = execCommand("sudo", append([]string{"-n", s.Path}, args...)...)
	}

	var device smartctlDeviceJSON
	out, err := internal.CombinedOutputTimeout(cmd, time.Duration(s.Timeout))
	if err != nil {
		// Try to still unmarshal the output to see if a specific message can
		// be extracted from the output.
		if err := json.Unmarshal(out, &device); err == nil {
			if len(device.Smartctl.Messages) > 0 && device.Smartctl.Messages[0].String != "" {
				return fmt.Errorf("error running smartctl with %s: %s", args, device.Smartctl.Messages[0].String)
			}
		}
		return fmt.Errorf("error running smartctl with %s: %w", args, err)
	}
	t := time.Now()

	if err := json.Unmarshal(out, &device); err != nil {
		return fmt.Errorf("error unmarshalling smartctl output: %w", err)
	}

	tags := map[string]string{
		"name":   device.Device.Name,
		"type":   device.Device.Type,
		"model":  device.ModelName,
		"serial": device.SerialNumber,
	}

	// The JSON WWN is in decimal and needs to be converted to hex
	if device.Wwn.ID != 0 && device.Wwn.Naa != 0 && device.Wwn.Oui != 0 {
		tags["wwn"] = fmt.Sprintf("%01x%06x%09x", device.Wwn.Naa, device.Wwn.Oui, device.Wwn.ID)
	}

	fields := map[string]interface{}{
		"capacity":    device.UserCapacity.Bytes,
		"health_ok":   device.SmartStatus.Passed,
		"temperature": device.Temperature.Current,
		"firmware":    device.FirmwareVersion,
	}

	// Add NVMe specific fields
	if device.Device.Type == "nvme" {
		fields["critical_warning"] = device.NvmeSmartHealthInformationLog.CriticalWarning
		fields["temperature"] = device.NvmeSmartHealthInformationLog.Temperature
		fields["available_spare"] = device.NvmeSmartHealthInformationLog.AvailableSpare
		fields["available_spare_threshold"] = device.NvmeSmartHealthInformationLog.AvailableSpareThreshold
		fields["percentage_used"] = device.NvmeSmartHealthInformationLog.PercentageUsed
		fields["data_units_read"] = device.NvmeSmartHealthInformationLog.DataUnitsRead
		fields["data_units_written"] = device.NvmeSmartHealthInformationLog.DataUnitsWritten
		fields["host_reads"] = device.NvmeSmartHealthInformationLog.HostReads
		fields["host_writes"] = device.NvmeSmartHealthInformationLog.HostWrites
		fields["controller_busy_time"] = device.NvmeSmartHealthInformationLog.ControllerBusyTime
		fields["power_cycles"] = device.NvmeSmartHealthInformationLog.PowerCycles
		fields["power_on_hours"] = device.NvmeSmartHealthInformationLog.PowerOnHours
		fields["unsafe_shutdowns"] = device.NvmeSmartHealthInformationLog.UnsafeShutdowns
		fields["media_errors"] = device.NvmeSmartHealthInformationLog.MediaErrors
		fields["num_err_log_entries"] = device.NvmeSmartHealthInformationLog.NumErrLogEntries
		fields["warning_temp_time"] = device.NvmeSmartHealthInformationLog.WarningTempTime
		fields["critical_comp_time"] = device.NvmeSmartHealthInformationLog.CriticalCompTime
	}

	acc.AddFields("smartctl", fields, tags, t)

	// Check for ATA specific attribute fields
	for _, attribute := range device.AtaSmartAttributes.Table {
		attributeTags := make(map[string]string, len(tags)+1)
		for k, v := range tags {
			attributeTags[k] = v
		}
		attributeTags["name"] = attribute.Name

		fields := map[string]interface{}{
			"raw_value": attribute.Raw.Value,
			"worst":     attribute.Worst,
			"threshold": attribute.Thresh,
			"value":     attribute.Value,
		}

		acc.AddFields("smartctl_attributes", fields, attributeTags, t)
	}

	return nil
}
