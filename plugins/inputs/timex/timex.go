//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package timex

import (
	_ "embed"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	// https://man7.org/linux/man-pages/man2/adjtimex.2.html#NOTES.
	// There the frequency is represented as a fixed-point number with a scaling factor of 2^16 (65536).
	ppm16 = float64(65536)
)

// Timex gathers system time metrics using the Linux kernel adjtimex syscall.
type Timex struct{}

func (*Timex) SampleConfig() string {
	return sampleConfig
}

func (*Timex) Gather(acc telegraf.Accumulator) error {
	var timex unix.Timex
	status, err := unix.Adjtimex(&timex)
	if err != nil {
		return fmt.Errorf("failed to get time adjtimex stats: %w", err)
	}

	// Check the return status for clock state
	// https://github.com/torvalds/linux/blob/master/include/uapi/linux/timex.h
	synced := status != unix.TIME_ERROR

	// https://man7.org/linux/man-pages/man2/adjtimex.2.html
	// validate the status to determine if the time is in nanoseconds or microseconds
	// STA_NANO (0x2000): time is in nanoseconds
	// STA_MICRO (0x4000): time is in microseconds
	multiplier := int64(1000)
	if (timex.Status & unix.STA_NANO) != 0 {
		multiplier = int64(1)
	}

	var statusOutput string
	switch status {
	case unix.TIME_OK:
		statusOutput = "ok"
	case unix.TIME_INS:
		statusOutput = "insert"
	case unix.TIME_DEL:
		statusOutput = "delete"
	case unix.TIME_OOP:
		statusOutput = "progress"
	case unix.TIME_WAIT:
		statusOutput = "wait"
	case unix.TIME_ERROR:
		statusOutput = "error"
	default:
		statusOutput = fmt.Sprintf("unknown-%d", status)
	}

	tags := map[string]string{
		"status": statusOutput,
	}

	fields := map[string]interface{}{
		"offset_ns":                    int64(timex.Offset) * multiplier, //nolint:unconvert // Conversion needed for some architectures
		"frequency_offset_ppm":         float64(timex.Freq) / ppm16,
		"maxerror_ns":                  timex.Maxerror * 1000,
		"estimated_error_ns":           timex.Esterror * 1000,
		"status":                       timex.Status,
		"loop_time_constant":           timex.Constant,
		"tick_ns":                      timex.Tick * 1000,
		"pps_frequency_ppm":            float64(timex.Ppsfreq) / ppm16,
		"pps_jitter_ns":                int64(timex.Jitter) * multiplier, //nolint:unconvert // Conversion needed for some architectures
		"pps_shift_sec":                timex.Shift,
		"pps_stability_ppm":            float64(timex.Stabil) / ppm16,
		"pps_jitter_total":             timex.Jitcnt,
		"pps_calibration_total":        timex.Calcnt,
		"pps_error_total":              timex.Errcnt,
		"pps_stability_exceeded_total": timex.Stbcnt,
		"tai_offset_sec":               timex.Tai,
		"synchronized":                 synced,
	}

	acc.AddGauge("timex", fields, tags)

	return nil
}

func init() {
	inputs.Add("timex", func() telegraf.Input {
		return &Timex{}
	})
}
