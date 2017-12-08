package libkflow

import (
	"errors"
	"fmt"
	"net"

	"github.com/kentik/libkflow/api"
)

var (
	ErrInvalidAuth   = errors.New("invalid API email/token")
	ErrInvalidConfig = errors.New("invalid config")
	ErrInvalidDevice = errors.New("invalid device")
)

// NewSenderWithDeviceID creates a new flow Sender given a device ID,
// error channel, and Config.
func NewSenderWithDeviceID(did int, errors chan<- error, cfg *Config) (*Sender, error) {
	client := cfg.client()

	d, err := lookupdev(client.GetDeviceByID(did))
	if err != nil {
		return nil, err
	}

	s, err := cfg.start(client, d, errors)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// NewSenderWithDeviceIF creates a new flow Sender given a device interface name,
// error channel, and Config.
func NewSenderWithDeviceIF(dif string, errors chan<- error, cfg *Config) (*Sender, error) {
	client := cfg.client()

	d, err := lookupdev(client.GetDeviceByIF(dif))
	if err != nil {
		return nil, err
	}

	s, err := cfg.start(client, d, errors)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// NewSenderWithDeviceIP creates a new flow Sender given a device IP address,
// error channel, and Config.
func NewSenderWithDeviceIP(dip net.IP, errors chan<- error, cfg *Config) (*Sender, error) {
	client := cfg.client()

	d, err := lookupdev(client.GetDeviceByIP(dip))
	if err != nil {
		return nil, err
	}

	s, err := cfg.start(client, d, errors)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func lookupdev(dev *api.Device, err error) (*api.Device, error) {
	if err != nil {
		switch {
		case api.IsErrorWithStatusCode(err, 401):
			return nil, ErrInvalidAuth
		case api.IsErrorWithStatusCode(err, 404):
			return nil, ErrInvalidDevice
		default:
			return nil, fmt.Errorf("device lookup error: %s", err)
		}
	}
	return dev, nil
}
