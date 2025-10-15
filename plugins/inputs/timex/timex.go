//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package timex

import (
	_ "embed"
	"fmt"
	"time"

	"golang.org/x/sys/unix"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Timex struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Timex) SampleConfig() string {
	return sampleConfig
}

func (*Timex) Init() error {
	return nil
}

func (*Timex) Gather(acc telegraf.Accumulator) error {
	var timex = new(unix.Timex)

	status, err := unix.Adjtimex(timex)
	if err != nil {
		return fmt.Errorf("failed to get time adjtimex stats: %w", err)
	}

	// Check the return status for clock state
	// https://github.com/torvalds/linux/blob/master/include/uapi/linux/timex.h
	syncStatus := 1
	if status == unix.TIME_ERROR {
		syncStatus = 0
	}

	microseconds := float64(1 * time.Second.Microseconds())
	nanoseconds := float64(1 * time.Second.Nanoseconds())
	divisor := microseconds

	// https://man7.org/linux/man-pages/man2/adjtimex.2.html
	// Notes for frequency adjustment ppm
	ppm16 := float64(microseconds * 65536)

	// https://man7.org/linux/man-pages/man2/adjtimex.2.html
	// validate the status to determine if the time is in nanoseconds or microseconds
	// STA_NANO (0x2000): time is in nanoseconds
	// STA_MICRO (0x4000): time is in microseconds
	if (timex.Status & unix.STA_NANO) != 0 {
		divisor = nanoseconds
	}

	fields := map[string]interface{}{
		"offset_seconds":               float64(timex.Offset) / divisor,
		"frequency_adjustment_ratio":   1 + float64(timex.Freq)/ppm16,
		"maxerror_seconds":             float64(timex.Maxerror) / microseconds,
		"estimated_error_seconds":      float64(timex.Esterror) / microseconds,
		"status":                       float64(timex.Status),
		"loop_time_constant":           float64(timex.Constant),
		"tick_seconds":                 float64(timex.Tick) / microseconds,
		"pps_frequency_hertz":          float64(timex.Ppsfreq) / ppm16,
		"pps_jitter_seconds":           float64(timex.Jitter) / divisor,
		"pps_shift_seconds":            float64(timex.Shift),
		"pps_stability_hertz":          float64(timex.Stabil) / ppm16,
		"pps_jitter_total":             float64(timex.Jitcnt),
		"pps_calibration_total":        float64(timex.Calcnt),
		"pps_error_total":              float64(timex.Errcnt),
		"pps_stability_exceeded_total": float64(timex.Stbcnt),
		"tai_offset_seconds":           float64(timex.Tai),
		"sync_status":                  syncStatus,
	}

	now := time.Now()

	acc.AddGauge("timex", fields, make(map[string]string, 0), now)

	return nil
}

func init() {
	inputs.Add("timex", func() telegraf.Input {
		return &Timex{}
	})
}
