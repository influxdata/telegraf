package schema_v12

import (
	"encoding/xml"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/common"
)

// Parse parses the XML-encoded data from nvidia-smi and adds measurements.
func Parse(acc telegraf.Accumulator, buf []byte) error {
	var s smi
	if err := xml.Unmarshal(buf, &s); err != nil {
		return err
	}

	timestamp := time.Now()
	if s.Timestamp != "" {
		if t, err := time.ParseInLocation(time.ANSIC, s.Timestamp, time.Local); err == nil {
			timestamp = t
		}
	}

	for i := range s.Gpu {
		gpu := &s.Gpu[i]

		tags := map[string]string{
			"index": strconv.Itoa(i),
		}
		fields := make(map[string]interface{}, 67)

		common.SetTagIfUsed(tags, "pstate", gpu.PerformanceState)
		common.SetTagIfUsed(tags, "name", gpu.ProductName)
		common.SetTagIfUsed(tags, "arch", gpu.ProductArchitecture)
		common.SetTagIfUsed(tags, "uuid", gpu.UUID)
		common.SetTagIfUsed(tags, "compute_mode", gpu.ComputeMode)

		common.SetIfUsed("str", fields, "driver_version", s.DriverVersion)
		common.SetIfUsed("str", fields, "cuda_version", s.CudaVersion)
		common.SetIfUsed("str", fields, "serial", gpu.Serial)
		common.SetIfUsed("str", fields, "vbios_version", gpu.VbiosVersion)
		common.SetIfUsed("str", fields, "display_active", gpu.DisplayActive)
		common.SetIfUsed("str", fields, "display_mode", gpu.DisplayMode)
		common.SetIfUsed("str", fields, "current_ecc", gpu.EccMode.CurrentEcc)
		common.SetIfUsed("int", fields, "fan_speed", gpu.FanSpeed)
		common.SetIfUsed("int", fields, "memory_total", gpu.FbMemoryUsage.Total)
		common.SetIfUsed("int", fields, "memory_used", gpu.FbMemoryUsage.Used)
		common.SetIfUsed("int", fields, "memory_free", gpu.FbMemoryUsage.Free)
		common.SetIfUsed("int", fields, "memory_reserved", gpu.FbMemoryUsage.Reserved)
		common.SetIfUsed("int", fields, "retired_pages_multiple_single_bit", gpu.RetiredPages.MultipleSingleBitRetirement.RetiredCount)
		common.SetIfUsed("int", fields, "retired_pages_double_bit", gpu.RetiredPages.DoubleBitRetirement.RetiredCount)
		common.SetIfUsed("str", fields, "retired_pages_blacklist", gpu.RetiredPages.PendingBlacklist)
		common.SetIfUsed("str", fields, "retired_pages_pending", gpu.RetiredPages.PendingRetirement)
		common.SetIfUsed("int", fields, "remapped_rows_correctable", gpu.RemappedRows.Correctable)
		common.SetIfUsed("int", fields, "remapped_rows_uncorrectable", gpu.RemappedRows.Uncorrectable)
		common.SetIfUsed("str", fields, "remapped_rows_pending", gpu.RemappedRows.Pending)
		common.SetIfUsed("str", fields, "remapped_rows_failure", gpu.RemappedRows.Failure)
		common.SetIfUsed("int", fields, "temperature_gpu", gpu.Temperature.GpuTemp)
		common.SetIfUsed("int", fields, "temperature_gpu_tlimit", gpu.Temperature.GpuTempTlimit)
		common.SetIfUsed("int", fields, "temperature_memory", gpu.Temperature.MemoryTemp)
		common.SetIfUsed("int", fields, "temperature_gpu_target", gpu.Temperature.GpuTargetTemperature)
		common.SetIfUsed("int", fields, "temperature_gpu_target_min", gpu.SupportedGpuTargetTemp.GpuTargetTempMin)
		common.SetIfUsed("int", fields, "temperature_gpu_target_max", gpu.SupportedGpuTargetTemp.GpuTargetTempMax)
		common.SetIfUsed("int", fields, "temperature_max_threshold", gpu.Temperature.GpuTempMaxThreshold)
		common.SetIfUsed("int", fields, "temperature_slow_threshold", gpu.Temperature.GpuTempSlowThreshold)
		common.SetIfUsed("int", fields, "temperature_max_gpu_threshold", gpu.Temperature.GpuTempMaxGpuThreshold)
		common.SetIfUsed("int", fields, "temperature_max_mem_threshold", gpu.Temperature.GpuTempMaxMemThreshold)
		common.SetIfUsed("int", fields, "utilization_gpu", gpu.Utilization.GpuUtil)
		common.SetIfUsed("int", fields, "utilization_memory", gpu.Utilization.MemoryUtil)
		common.SetIfUsed("int", fields, "utilization_encoder", gpu.Utilization.EncoderUtil)
		common.SetIfUsed("int", fields, "utilization_decoder", gpu.Utilization.DecoderUtil)
		common.SetIfUsed("int", fields, "utilization_jpeg", gpu.Utilization.JpegUtil)
		common.SetIfUsed("int", fields, "utilization_ofa", gpu.Utilization.OfaUtil)
		common.SetIfUsed("int", fields, "pcie_link_gen_current", gpu.Pci.PciGpuLinkInfo.PcieGen.CurrentLinkGen)
		common.SetIfUsed("int", fields, "pcie_link_width_current", gpu.Pci.PciGpuLinkInfo.LinkWidths.CurrentLinkWidth)
		common.SetIfUsed("int", fields, "pcie_rx_util", gpu.Pci.RxUtil)
		common.SetIfUsed("int", fields, "pcie_tx_util", gpu.Pci.TxUtil)
		common.SetIfUsed("int", fields, "encoder_stats_session_count", gpu.EncoderStats.SessionCount)
		common.SetIfUsed("int", fields, "encoder_stats_average_fps", gpu.EncoderStats.AverageFps)
		common.SetIfUsed("int", fields, "encoder_stats_average_latency", gpu.EncoderStats.AverageLatency)
		common.SetIfUsed("int", fields, "fbc_stats_session_count", gpu.FbcStats.SessionCount)
		common.SetIfUsed("int", fields, "fbc_stats_average_fps", gpu.FbcStats.AverageFps)
		common.SetIfUsed("int", fields, "fbc_stats_average_latency", gpu.FbcStats.AverageLatency)
		common.SetIfUsed("int", fields, "clocks_current_graphics", gpu.Clocks.GraphicsClock)
		common.SetIfUsed("int", fields, "clocks_current_sm", gpu.Clocks.SmClock)
		common.SetIfUsed("int", fields, "clocks_current_memory", gpu.Clocks.MemClock)
		common.SetIfUsed("int", fields, "clocks_current_video", gpu.Clocks.VideoClock)
		throttle := gpu.ClocksThrottleReasons
		reasons := gpu.ClocksEventReasons
		common.SetIfUsed("str", fields, "clocks_event_reason_sw_power_cap", throttle.ClocksThrottleReasonSwPowerCap)
		common.SetIfUsed("str", fields, "clocks_event_reason_sw_power_cap", reasons.ClocksEventReasonSwPowerCap)
		common.SetIfUsed("str", fields, "clocks_event_reason_sw_thermal_slowdown", throttle.ClocksThrottleReasonSwThermalSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_sw_thermal_slowdown", reasons.ClocksEventReasonSwThermalSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_hw_thermal_slowdown", throttle.ClocksThrottleReasonHwThermalSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_hw_thermal_slowdown", reasons.ClocksEventReasonHwThermalSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_hw_power_brake_slowdown", throttle.ClocksThrottleReasonHwPowerBrakeSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_hw_power_brake_slowdown", reasons.ClocksEventReasonHwPowerBrakeSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_hw_slowdown", throttle.ClocksThrottleReasonHwSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_hw_slowdown", reasons.ClocksEventReasonHwSlowdown)
		common.SetIfUsed("str", fields, "clocks_event_reason_sync_boost", throttle.ClocksThrottleReasonSyncBoost)
		common.SetIfUsed("str", fields, "clocks_event_reason_sync_boost", reasons.ClocksEventReasonSyncBoost)
		common.SetIfUsed("str", fields, "clocks_event_reason_gpu_idle", throttle.ClocksThrottleReasonGpuIdle)
		common.SetIfUsed("str", fields, "clocks_event_reason_gpu_idle", reasons.ClocksEventReasonGpuIdle)
		common.SetIfUsed("str", fields, "clocks_event_reason_applications_clocks_setting", throttle.ClocksThrottleReasonApplicationsClocksSetting)
		common.SetIfUsed("str", fields, "clocks_event_reason_applications_clocks_setting", reasons.ClocksEventReasonApplicationsClocksSetting)
		common.SetIfUsed("str", fields, "clocks_event_reason_display_clocks_setting", throttle.ClocksThrottleReasonDisplayClocksSetting)
		common.SetIfUsed("str", fields, "clocks_event_reason_display_clocks_setting", reasons.ClocksEventReasonDisplayClocksSetting)
		common.SetIfUsed("float", fields, "power_draw", gpu.PowerReadings.PowerDraw)
		common.SetIfUsed("float", fields, "power_draw", gpu.PowerReadings.InstantPowerDraw)
		common.SetIfUsed("float", fields, "power_limit", gpu.PowerReadings.PowerLimit)
		common.SetIfUsed("float", fields, "power_draw_average", gpu.PowerReadings.AveragePowerDraw)
		common.SetIfUsed("float", fields, "power_limit_default", gpu.PowerReadings.DefaultPowerLimit)
		common.SetIfUsed("float", fields, "power_limit_min", gpu.PowerReadings.MinPowerLimit)
		common.SetIfUsed("float", fields, "power_limit_max", gpu.PowerReadings.MaxPowerLimit)
		common.SetIfUsed("float", fields, "power_draw", gpu.GpuPowerReadings.PowerDraw)
		common.SetIfUsed("float", fields, "power_draw", gpu.GpuPowerReadings.InstantPowerDraw)
		common.SetIfUsed("float", fields, "power_limit", gpu.GpuPowerReadings.CurrentPowerLimit)
		common.SetIfUsed("float", fields, "power_limit", gpu.GpuPowerReadings.PowerLimit)
		common.SetIfUsed("float", fields, "power_draw_average", gpu.GpuPowerReadings.AveragePowerDraw)
		common.SetIfUsed("float", fields, "power_limit_default", gpu.GpuPowerReadings.DefaultPowerLimit)
		common.SetIfUsed("float", fields, "power_limit_min", gpu.GpuPowerReadings.MinPowerLimit)
		common.SetIfUsed("float", fields, "power_limit_max", gpu.GpuPowerReadings.MaxPowerLimit)
		common.SetIfUsed("float", fields, "power_limit_requested", gpu.GpuPowerReadings.RequestedPowerLimit)
		common.SetIfUsed("float", fields, "module_power_draw", gpu.ModulePowerReadings.PowerDraw)
		common.SetIfUsed("float", fields, "module_power_draw", gpu.ModulePowerReadings.InstantPowerDraw)
		acc.AddFields("nvidia_smi", fields, tags, timestamp)

		for _, device := range gpu.MigDevices.MigDevice {
			tags := make(map[string]string, 8)
			common.SetTagIfUsed(tags, "index", device.Index)
			common.SetTagIfUsed(tags, "gpu_index", device.GpuInstanceID)
			common.SetTagIfUsed(tags, "compute_index", device.ComputeInstanceID)
			common.SetTagIfUsed(tags, "pstate", gpu.PerformanceState)
			common.SetTagIfUsed(tags, "name", gpu.ProductName)
			common.SetTagIfUsed(tags, "arch", gpu.ProductArchitecture)
			common.SetTagIfUsed(tags, "uuid", gpu.UUID)
			common.SetTagIfUsed(tags, "compute_mode", gpu.ComputeMode)

			fields := make(map[string]interface{}, 8)
			common.SetIfUsed("int", fields, "sram_uncorrectable", device.EccErrorCount.VolatileCount.SramUncorrectable)
			common.SetIfUsed("int", fields, "memory_fb_total", device.FbMemoryUsage.Total)
			common.SetIfUsed("int", fields, "memory_fb_reserved", device.FbMemoryUsage.Reserved)
			common.SetIfUsed("int", fields, "memory_fb_used", device.FbMemoryUsage.Used)
			common.SetIfUsed("int", fields, "memory_fb_free", device.FbMemoryUsage.Free)
			common.SetIfUsed("int", fields, "memory_bar1_total", device.Bar1MemoryUsage.Total)
			common.SetIfUsed("int", fields, "memory_bar1_used", device.Bar1MemoryUsage.Used)
			common.SetIfUsed("int", fields, "memory_bar1_free", device.Bar1MemoryUsage.Free)

			acc.AddFields("nvidia_smi_mig", fields, tags, timestamp)
		}

		for _, process := range gpu.Processes.ProcessInfo {
			tags := make(map[string]string, 2)
			common.SetTagIfUsed(tags, "name", process.ProcessName)
			common.SetTagIfUsed(tags, "type", process.Type)

			fields := make(map[string]interface{}, 2)
			common.SetIfUsed("int", fields, "pid", process.Pid)
			common.SetIfUsed("int", fields, "used_memory", process.UsedMemory)

			acc.AddFields("nvidia_smi_process", fields, tags, timestamp)
		}
	}

	return nil
}
