//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows && !darwin

package upower

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	busName         = "org.freedesktop.UPower"
	daemonPath      = "/org/freedesktop/UPower"
	deviceInterface = "org.freedesktop.UPower.Device"
)

// Enumerations as defined in
// https://upower.freedesktop.org/docs/Device.html
var (
	deviceTypes = []string{
		"unknown", "line_power", "battery", "ups", "monitor", "mouse", "keyboard",
		"pda", "phone", "media_player", "tablet", "computer", "gaming_input", "pen",
		"touchpad", "modem", "network", "headset", "speakers", "headphones", "video",
		"other_audio", "remote_control", "printer", "scanner", "camera", "wearable",
		"toy", "bluetooth_generic",
	}
	deviceStates = []string{
		"unknown", "charging", "discharging", "empty", "fully_charged",
		"pending_charge", "pending_discharge",
	}
	technologies = []string{
		"unknown", "lithium_ion", "lithium_polymer", "lithium_iron_phosphate",
		"lead_acid", "nickel_cadmium", "nickel_metal_hydride",
	}
	warningLevels = []string{"unknown", "none", "discharging", "low", "critical", "action"}
)

type UPower struct {
	DeviceTypes []string        `toml:"device_types"`
	Timeout     config.Duration `toml:"timeout"`
	Log         telegraf.Logger `toml:"-"`

	client client
}

type client interface {
	connected() bool
	close() error
	enumerateDevices(ctx context.Context) ([]dbus.ObjectPath, error)
	deviceProperties(ctx context.Context, path dbus.ObjectPath) (map[string]dbus.Variant, error)
}

func (*UPower) SampleConfig() string {
	return sampleConfig
}

func (u *UPower) Init() error {
	if u.Timeout <= 0 {
		u.Timeout = config.Duration(5 * time.Second)
	}

	for _, t := range u.DeviceTypes {
		if !slices.Contains(deviceTypes, t) {
			return fmt.Errorf("invalid device type %q", t)
		}
	}

	return nil
}

func (u *UPower) Start(telegraf.Accumulator) error {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return fmt.Errorf("connecting to system bus failed: %w", err)
	}
	u.client = &dbusClient{conn: conn}

	return nil
}

func (u *UPower) Stop() {
	if u.client == nil {
		return
	}
	if err := u.client.close(); err != nil {
		u.Log.Errorf("Closing system bus connection failed: %v", err)
	}
	u.client = nil
}

func (u *UPower) Gather(acc telegraf.Accumulator) error {
	if u.client == nil || !u.client.connected() {
		u.Log.Debug("Connection to system bus lost, trying to reconnect...")
		u.Stop()
		if err := u.Start(acc); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.Timeout))
	defer cancel()

	paths, err := u.client.enumerateDevices(ctx)
	if err != nil {
		return fmt.Errorf("enumerating devices failed: %w", err)
	}

	for _, path := range paths {
		props, err := u.client.deviceProperties(ctx, path)
		if err != nil {
			acc.AddError(fmt.Errorf("getting properties of device %q failed: %w", path, err))
			continue
		}
		u.addDevice(acc, props)
	}

	return nil
}

func (u *UPower) addDevice(acc telegraf.Accumulator, props map[string]dbus.Variant) {
	deviceType := lookup(deviceTypes, property[uint32](props, "Type"))
	if len(u.DeviceTypes) > 0 && !slices.Contains(u.DeviceTypes, deviceType) {
		return
	}

	tags := map[string]string{
		"native_path": property[string](props, "NativePath"),
		"type":        deviceType,
	}
	for tag, name := range map[string]string{"vendor": "Vendor", "model": "Model", "serial": "Serial"} {
		if v := property[string](props, name); v != "" {
			tags[tag] = v
		}
	}

	// Line-power devices (e.g. AC adapters) only report their online state
	if deviceType == "line_power" {
		fields := map[string]any{
			"online": property[bool](props, "Online"),
		}
		acc.AddGauge("upower", fields, tags)
		return
	}

	tags["state"] = lookup(deviceStates, property[uint32](props, "State"))
	tags["technology"] = lookup(technologies, property[uint32](props, "Technology"))
	tags["warning_level"] = lookup(warningLevels, property[uint32](props, "WarningLevel"))

	fields := map[string]any{
		"percentage":            property[float64](props, "Percentage"),
		"energy_wh":             property[float64](props, "Energy"),
		"energy_empty_wh":       property[float64](props, "EnergyEmpty"),
		"energy_full_wh":        property[float64](props, "EnergyFull"),
		"energy_full_design_wh": property[float64](props, "EnergyFullDesign"),
		"energy_rate_w":         property[float64](props, "EnergyRate"),
		"voltage_v":             property[float64](props, "Voltage"),
		"temperature_c":         property[float64](props, "Temperature"),
		"capacity_percent":      property[float64](props, "Capacity"),
		"charge_cycles":         property[int32](props, "ChargeCycles"),
		"time_to_empty_sec":     property[int64](props, "TimeToEmpty"),
		"time_to_full_sec":      property[int64](props, "TimeToFull"),
		"is_present":            property[bool](props, "IsPresent"),
		"is_rechargeable":       property[bool](props, "IsRechargeable"),
		"power_supply":          property[bool](props, "PowerSupply"),
	}
	acc.AddGauge("upower", fields, tags)
}

func property[T any](props map[string]dbus.Variant, name string) T {
	var zero T
	v, ok := props[name]
	if !ok {
		return zero
	}
	value, ok := v.Value().(T)
	if !ok {
		return zero
	}
	return value
}

func lookup(names []string, code uint32) string {
	if int(code) < len(names) {
		return names[code]
	}
	return fmt.Sprintf("unknown-%d", code)
}

type dbusClient struct {
	conn *dbus.Conn
}

func (c *dbusClient) connected() bool {
	return c.conn.Connected()
}

func (c *dbusClient) close() error {
	return c.conn.Close()
}

func (c *dbusClient) enumerateDevices(ctx context.Context) ([]dbus.ObjectPath, error) {
	var paths []dbus.ObjectPath
	obj := c.conn.Object(busName, daemonPath)
	err := obj.CallWithContext(ctx, busName+".EnumerateDevices", 0).Store(&paths)
	return paths, err
}

func (c *dbusClient) deviceProperties(ctx context.Context, path dbus.ObjectPath) (map[string]dbus.Variant, error) {
	var props map[string]dbus.Variant
	obj := c.conn.Object(busName, path)
	err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, deviceInterface).Store(&props)
	return props, err
}

func init() {
	inputs.Add("upower", func() telegraf.Input {
		return &UPower{}
	})
}
