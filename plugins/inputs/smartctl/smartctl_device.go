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
		// Error running the command and unable to parse the JSON, then bail
		if jsonErr := json.Unmarshal(out, &device); jsonErr != nil {
			return fmt.Errorf("error running smartctl with %s: %w", args, err)
		}

		// If we were able to parse the result, then only exit if we get an error
		// as sometimes we can get warnings, that still produce data.
		if len(device.Smartctl.Messages) > 0 &&
			device.Smartctl.Messages[0].Severity == "error" &&
			device.Smartctl.Messages[0].String != "" {
			return fmt.Errorf("error running smartctl with %s got smartctl error message: %s", args, device.Smartctl.Messages[0].String)
		}
	}

	if err := json.Unmarshal(out, &device); err != nil {
		return fmt.Errorf("error unable to unmarshall response %s: %w", args, err)
	}

	t := time.Now()

	tags := map[string]string{
		"name":   device.Device.Name,
		"type":   device.Device.Type,
		"serial": device.SerialNumber,
	}

	if device.ModelName != "" {
		tags["model"] = device.ModelName
	}
	if device.Vendor != "" {
		tags["vendor"] = device.Vendor
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

	if device.SCSIVendor != "" {
		fields["scsi_vendor"] = device.SCSIVendor
	}
	if device.SCSIModelName != "" {
		fields["scsi_model"] = device.SCSIModelName
	}
	if device.SCSIRevision != "" {
		fields["scsi_revision"] = device.SCSIRevision
	}
	if device.SCSIVersion != "" {
		fields["scsi_version"] = device.SCSIVersion
	}
	if device.SCSITransportProtocol.Name != "" {
		fields["scsi_transport_protocol"] = device.SCSITransportProtocol.Name
	}
	if device.SCSIProtectionType != 0 {
		fields["scsi_protection_type"] = device.SCSIProtectionType
	}
	if device.SCSIProtectionIntervalBytesPerLB != 0 {
		fields["scsi_protection_interval_bytes_per_lb"] = device.SCSIProtectionIntervalBytesPerLB
	}
	if device.SCSIGrownDefectList != 0 {
		fields["scsi_grown_defect_list"] = device.SCSIGrownDefectList
	}
	if device.LogicalBlockSize != 0 {
		fields["logical_block_size"] = device.LogicalBlockSize
	}
	if device.RotationRate != 0 {
		fields["rotation_rate"] = device.RotationRate
	}
	if device.SCSIStartStopCycleCounter.SpecifiedCycleCountOverDeviceLifetime != 0 {
		fields["specified_cycle_count_over_device_lifetime"] = device.SCSIStartStopCycleCounter.SpecifiedCycleCountOverDeviceLifetime
	}
	if device.SCSIStartStopCycleCounter.AccumulatedStartStopCycles != 0 {
		fields["accumulated_start_stop_cycles"] = device.SCSIStartStopCycleCounter.AccumulatedStartStopCycles
	}
	if device.PowerOnTime.Hours != 0 {
		fields["power_on_hours"] = device.PowerOnTime.Hours
	}
	if device.PowerOnTime.Minutes != 0 {
		fields["power_on_minutes"] = device.PowerOnTime.Minutes
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

	// Check for SCSI error counter entries
	if device.Device.Type == "scsi" {
		counterTags := make(map[string]string, len(tags)+1)
		for k, v := range tags {
			counterTags[k] = v
		}

		counterTags["page"] = "read"
		fields := map[string]interface{}{
			"errors_corrected_by_eccfast":          device.ScsiErrorCounterLog.Read.ErrorsCorrectedByEccfast,
			"errors_corrected_by_eccdelayed":       device.ScsiErrorCounterLog.Read.ErrorsCorrectedByEccdelayed,
			"errors_corrected_by_rereads_rewrites": device.ScsiErrorCounterLog.Read.ErrorsCorrectedByRereadsRewrites,
			"total_errors_corrected":               device.ScsiErrorCounterLog.Read.TotalErrorsCorrected,
			"correction_algorithm_invocations":     device.ScsiErrorCounterLog.Read.CorrectionAlgorithmInvocations,
			"gigabytes_processed":                  device.ScsiErrorCounterLog.Read.GigabytesProcessed,
			"total_uncorrected_errors":             device.ScsiErrorCounterLog.Read.TotalUncorrectedErrors,
		}
		acc.AddFields("smartctl_scsi_error_counter_log", fields, counterTags, t)

		counterTags["page"] = "write"
		fields = map[string]interface{}{
			"errors_corrected_by_eccfast":          device.ScsiErrorCounterLog.Write.ErrorsCorrectedByEccfast,
			"errors_corrected_by_eccdelayed":       device.ScsiErrorCounterLog.Write.ErrorsCorrectedByEccdelayed,
			"errors_corrected_by_rereads_rewrites": device.ScsiErrorCounterLog.Write.ErrorsCorrectedByRereadsRewrites,
			"total_errors_corrected":               device.ScsiErrorCounterLog.Write.TotalErrorsCorrected,
			"correction_algorithm_invocations":     device.ScsiErrorCounterLog.Write.CorrectionAlgorithmInvocations,
			"gigabytes_processed":                  device.ScsiErrorCounterLog.Write.GigabytesProcessed,
			"total_uncorrected_errors":             device.ScsiErrorCounterLog.Write.TotalUncorrectedErrors,
		}
		acc.AddFields("smartctl_scsi_error_counter_log", fields, counterTags, t)

		counterTags["page"] = "verify"
		fields = map[string]interface{}{
			"errors_corrected_by_eccfast":          device.ScsiErrorCounterLog.Verify.ErrorsCorrectedByEccfast,
			"errors_corrected_by_eccdelayed":       device.ScsiErrorCounterLog.Verify.ErrorsCorrectedByEccdelayed,
			"errors_corrected_by_rereads_rewrites": device.ScsiErrorCounterLog.Verify.ErrorsCorrectedByRereadsRewrites,
			"total_errors_corrected":               device.ScsiErrorCounterLog.Verify.TotalErrorsCorrected,
			"correction_algorithm_invocations":     device.ScsiErrorCounterLog.Verify.CorrectionAlgorithmInvocations,
			"gigabytes_processed":                  device.ScsiErrorCounterLog.Verify.GigabytesProcessed,
			"total_uncorrected_errors":             device.ScsiErrorCounterLog.Verify.TotalUncorrectedErrors,
		}
		acc.AddFields("smartctl_scsi_error_counter_log", fields, counterTags, t)
	}

	return nil
}
