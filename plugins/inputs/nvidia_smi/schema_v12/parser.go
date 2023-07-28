package schema_v12

import (
	"encoding/xml"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/common"
)

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

	for i, gpu := range s.Gpu {
		tags := map[string]string{
			"index": strconv.Itoa(i),
		}
		fields := map[string]interface{}{}

		common.SetTagIfUsed(tags, "pstate", gpu.PerformanceState)
		common.SetTagIfUsed(tags, "name", gpu.ProductName)
		common.SetTagIfUsed(tags, "arch", gpu.ProductArchitecture)
		common.SetTagIfUsed(tags, "uuid", gpu.UUID)
		common.SetTagIfUsed(tags, "compute_mode", gpu.ComputeMode)

		common.SetIfUsed("str", fields, "driver_version", s.DriverVersion)
		common.SetIfUsed("str", fields, "cuda_version", s.CudaVersion)
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
		common.SetIfUsed("int", fields, "utilization_gpu", gpu.Utilization.GpuUtil)
		common.SetIfUsed("int", fields, "utilization_memory", gpu.Utilization.MemoryUtil)
		common.SetIfUsed("int", fields, "utilization_encoder", gpu.Utilization.EncoderUtil)
		common.SetIfUsed("int", fields, "utilization_decoder", gpu.Utilization.DecoderUtil)
		common.SetIfUsed("int", fields, "pcie_link_gen_current", gpu.Pci.PciGpuLinkInfo.PcieGen.CurrentLinkGen)
		common.SetIfUsed("int", fields, "pcie_link_width_current", gpu.Pci.PciGpuLinkInfo.LinkWidths.CurrentLinkWidth)
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
		common.SetIfUsed("float", fields, "power_draw", gpu.GpuPowerReadings.PowerDraw)
		common.SetIfUsed("float", fields, "module_power_draw", gpu.ModulePowerReadings.PowerDraw)
		acc.AddFields("nvidia_smi", fields, tags, timestamp)
	}

	return nil
}
