package api_test

import (
	"net"
	"testing"

	"github.com/kentik/libkflow/api"
	"github.com/kentik/libkflow/api/test"
	"github.com/stretchr/testify/assert"
)

func TestGetDeviceByID(t *testing.T) {
	client, _, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	device2, err := client.GetDeviceByID(device.ID)

	assert.NoError(err)
	assert.EqualValues(device, device2)
}

func TestGetDeviceByName(t *testing.T) {
	client, _, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	device2, err := client.GetDeviceByName(device.Name)

	assert.NoError(err)
	assert.EqualValues(device, device2)
}

func TestGetDeviceByIP(t *testing.T) {
	client, _, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	device2, err := client.GetDeviceByIP(device.IP)

	assert.NoError(err)
	assert.EqualValues(device, device2)
}

func TestGetDeviceByIF(t *testing.T) {
	client, _, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	ifs, err := net.Interfaces()
	assert.NoError(err)

	device2, err := client.GetDeviceByIF(ifs[0].Name)

	assert.NoError(err)
	assert.EqualValues(device, device2)
}

func TestGetInvalidDevice(t *testing.T) {
	client, _, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	_, err = client.GetDeviceByName(device.Name + "-invalid")
	assert.Error(err)
	assert.True(api.IsErrorWithStatusCode(err, 404))

	_, err = client.GetDeviceByIF("invalid")
	assert.Error(err)

	_, err = client.GetDeviceByIP(net.ParseIP("0.0.0.0"))
	assert.Error(err)
	assert.True(api.IsErrorWithStatusCode(err, 404))
}
