//go:generate ../../../tools/readme_config_includer/generator
package fireboard

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Fireboard struct {
	AuthToken   string          `toml:"auth_token"`
	URL         string          `toml:"url"`
	HTTPTimeout config.Duration `toml:"http_timeout"`

	client *http.Client
}

type rtt struct {
	Temp       float64 `json:"temp"`
	Channel    int64   `json:"channel"`
	DegreeType int     `json:"degreetype"`
	Created    string  `json:"created"`
}

type channel struct {
	Channel      int64   `json:"channel"`
	ChannelLabel string  `json:"channel_label"`
	DegreeType   int     `json:"degreetype"`
	Alerts       []alert `json:"alerts"`
}

type alert struct {
	Channel int64    `json:"channel"`
	TempMax *float64 `json:"temp_max"`
	TempMin *float64 `json:"temp_min"`
	Enabled bool     `json:"enabled"`
}

type lastDriveLog struct {
	DrivePer    float64 `json:"driveper"`
	Setpoint    float64 `json:"setpoint"`
	CurrentTemp float64 `json:"currenttemp"`
	VBatt       float64 `json:"vbatt"`
	PowerMode   int     `json:"powermode"`
	ModeType    int     `json:"modetype"`
	DriveType   int     `json:"drivetype"`
	TiedChannel int64   `json:"tiedchannel"`
	JSONRaw     string  `json:"jsonraw"`
}

type driveJSON struct {
	PelState       string  `json:"pelState"`
	FanSpeedTarget float64 `json:"fanSpeedTarget"`
	LidPaused      bool    `json:"lidPaused"`
	PelProfile     string  `json:"pelProfile"`
	POut           float64 `json:"pOut"`
	PErrSum        float64 `json:"pErrSum"`
	PKdOut         float64 `json:"pKdOut"`
	PKpOut         float64 `json:"pKpOut"`
	PKiOut         float64 `json:"pKiOut"`
	PLastErr       float64 `json:"pLastErr"`
	PKp            float64 `json:"pKp"`
	PKi            float64 `json:"pKi"`
	PKd            float64 `json:"pKd"`
}

type deviceLog struct {
	OnboardTemp float64 `json:"onboardTemp"`
	SignalLevel int     `json:"signallevel"`
	VBatt       float64 `json:"vBatt"`
	VBattPer    float64 `json:"vBattPer"`
}

type fireboardStats struct {
	Title        string        `json:"title"`
	UUID         string        `json:"uuid"`
	Model        string        `json:"model"`
	LatestTemps  []rtt         `json:"latest_temps"`
	Channels     []channel     `json:"channels"`
	LastDriveLog *lastDriveLog `json:"last_drivelog"`
	DeviceLog    *deviceLog    `json:"device_log"`
}

func (*Fireboard) SampleConfig() string {
	return sampleConfig
}

func (r *Fireboard) Init() error {
	if len(r.AuthToken) == 0 {
		return errors.New("you must specify an authToken")
	}
	if len(r.URL) == 0 {
		r.URL = "https://fireboard.io/api/v1/devices.json"
	}
	// Have a default timeout of 4s
	if r.HTTPTimeout == 0 {
		r.HTTPTimeout = config.Duration(time.Second * 4)
	}

	r.client.Timeout = time.Duration(r.HTTPTimeout)

	return nil
}

func (r *Fireboard) Gather(acc telegraf.Accumulator) error {
	// Perform the GET request to the fireboard servers
	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+r.AuthToken)
	req.Header.Set("User-Agent", "Telegraf")
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Successful responses will always return status code 200
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("fireboard server responded with %d [Forbidden], verify your authToken", resp.StatusCode)
		}
		return fmt.Errorf("fireboard responded with unexpected status code %d", resp.StatusCode)
	}
	// Decode the response JSON into a new stats struct
	var stats []fireboardStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return fmt.Errorf("unable to decode fireboard response: %w", err)
	}
	// Range over all devices, gathering stats. Returns early in case of any error.
	for _, s := range stats {
		gatherTemps(s, acc)
		gatherDrive(s, acc)
		gatherAlerts(s, acc)
		gatherDevice(s, acc)
	}
	return nil
}

// Return text description of degree type (scale)
func scale(n int) string {
	switch n {
	case 1:
		return "Celcius"
	case 2:
		return "Fahrenheit"
	default:
		return ""
	}
}

// gatherDrive emits a fireboard_drive measurement for pellet/fan controller devices.
// It is a no-op when LastDriveLog is nil.
func gatherDrive(s fireboardStats, acc telegraf.Accumulator) {
	if s.LastDriveLog == nil {
		return
	}
	d := s.LastDriveLog

	tags := map[string]string{
		"title": s.Title,
		"uuid":  s.UUID,
		"model": s.Model,
	}
	fields := map[string]interface{}{
		"drive_percent": d.DrivePer,
		"setpoint":      d.Setpoint,
		"current_temp":  d.CurrentTemp,
		"vbatt":         d.VBatt,
		"power_mode":    d.PowerMode,
		"mode_type":     d.ModeType,
		"drive_type":    d.DriveType,
		"tied_channel":  d.TiedChannel,
	}

	// Parse the embedded jsonraw string for pellet/PID fields.
	if d.JSONRaw != "" {
		var dj driveJSON
		if err := json.Unmarshal([]byte(d.JSONRaw), &dj); err == nil {
			tags["pel_state"] = dj.PelState
			tags["pel_profile"] = dj.PelProfile
			fields["fan_speed_target"] = dj.FanSpeedTarget
			fields["lid_paused"] = dj.LidPaused
			fields["pid_output"] = dj.POut
			fields["pid_error_sum"] = dj.PErrSum
			fields["pid_kd_output"] = dj.PKdOut
			fields["pid_kp_output"] = dj.PKpOut
			fields["pid_ki_output"] = dj.PKiOut
			fields["pid_last_error"] = dj.PLastErr
			fields["pid_kp"] = dj.PKp
			fields["pid_ki"] = dj.PKi
			fields["pid_kd"] = dj.PKd
		} else {
			acc.AddError(fmt.Errorf("fireboard: failed to parse jsonraw for device %q: %w", s.Title, err))
		}
	}

	acc.AddFields("fireboard_drive", fields, tags)
}

// gatherAlerts emits a fireboard_alert measurement for each alert entry across all channels.
// Only channels with active temperature data in latest_temps are emitted; this prevents
// alert metrics from being produced when the device is off or probes are disconnected.
func gatherAlerts(s fireboardStats, acc telegraf.Accumulator) {
	activeChannels := make(map[int64]bool, len(s.LatestTemps))
	for _, t := range s.LatestTemps {
		activeChannels[t.Channel] = true
	}

	for _, ch := range s.Channels {
		if !activeChannels[ch.Channel] {
			continue
		}
		for _, a := range ch.Alerts {
			tags := map[string]string{
				"title":        s.Title,
				"uuid":         s.UUID,
				"channel":      strconv.FormatInt(ch.Channel, 10),
				"channel_name": ch.ChannelLabel,
				"scale":        scale(ch.DegreeType),
			}
			fields := map[string]interface{}{
				"enabled": a.Enabled,
			}
			if a.TempMax != nil {
				fields["temp_max"] = *a.TempMax
			}
			if a.TempMin != nil {
				fields["temp_min"] = *a.TempMin
			}
			acc.AddFields("fireboard_alert", fields, tags)
		}
	}
}

// gatherDevice emits a fireboard_device measurement with device health data.
// It is a no-op when DeviceLog is nil.
func gatherDevice(s fireboardStats, acc telegraf.Accumulator) {
	if s.DeviceLog == nil {
		return
	}
	dl := s.DeviceLog
	tags := map[string]string{
		"title": s.Title,
		"uuid":  s.UUID,
		"model": s.Model,
	}
	fields := map[string]interface{}{
		"onboard_temp":  dl.OnboardTemp,
		"signal_level":  dl.SignalLevel,
		"vbatt":         dl.VBatt,
		"vbatt_percent": dl.VBattPer,
	}
	acc.AddFields("fireboard_device", fields, tags)
}

// Gathers stats from a single device, adding them to the accumulator
func gatherTemps(s fireboardStats, acc telegraf.Accumulator) {
	// Build channel number → label lookup
	channelNames := make(map[int64]string, len(s.Channels))
	for _, ch := range s.Channels {
		channelNames[ch.Channel] = ch.ChannelLabel
	}

	for _, t := range s.LatestTemps {
		tags := map[string]string{
			"title":        s.Title,
			"uuid":         s.UUID,
			"channel":      strconv.FormatInt(t.Channel, 10),
			"scale":        scale(t.DegreeType),
			"channel_name": channelNames[t.Channel],
		}
		fields := map[string]interface{}{
			"temperature": t.Temp,
		}
		acc.AddFields("fireboard", fields, tags)
	}
}

func newFireboard() *Fireboard {
	tr := &http.Transport{ResponseHeaderTimeout: 3 * time.Second}
	client := &http.Client{
		Transport: tr,
		Timeout:   4 * time.Second,
	}
	return &Fireboard{client: client}
}

func init() {
	inputs.Add("fireboard", func() telegraf.Input {
		return newFireboard()
	})
}
