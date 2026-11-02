//go:build !windows && !darwin

package upower

import (
	"context"

	"github.com/godbus/dbus/v5"
)

const (
	busName         = "org.freedesktop.UPower"
	daemonPath      = "/org/freedesktop/UPower"
	deviceInterface = "org.freedesktop.UPower.Device"
)

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
