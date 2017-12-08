package main

import (
	"testing"
	"unsafe"

	"github.com/kentik/libkflow/api/test"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	var n int
	cfg, assert := setupMainTest(t)

	// init with device ID
	n = int(kflowInit(cfg, nil, nil))
	assert.Equal(0, n)

	// init with device IP
	cfg.device_id = 0
	n = int(kflowInit(cfg, nil, nil))
	assert.Equal(0, n)
}

func TestInitInvalidConfig(t *testing.T) {
	var n int
	assert := assert.New(t)

	// NULL config
	n = int(kflowInit(nil, nil, nil))
	assert.Equal(EKFLOWCONFIG, n)

	// NULL API URL
	cfg := _Ctype_struct___3{}
	n = int(kflowInit(&cfg, nil, nil))
	assert.Equal(EKFLOWCONFIG, n)
}

func TestInitMissingProgram(t *testing.T) {
	cfg, assert := setupMainTest(t)
	cfg.program = nil
	n := int(kflowInit(cfg, nil, nil))
	assert.Equal(EKFLOWCONFIG, n)
}

func TestInitMissingVersion(t *testing.T) {
	cfg, assert := setupMainTest(t)
	cfg.version = nil
	n := int(kflowInit(cfg, nil, nil))
	assert.Equal(EKFLOWCONFIG, n)
}

func TestInitInvalidAuth(t *testing.T) {
	cfg, assert := setupMainTest(t)
	cfg.API.email = nil
	n := int(kflowInit(cfg, nil, nil))
	assert.Equal(EKFLOWAUTH, n)
}

func TestInitInvalidDevice(t *testing.T) {
	var n int
	cfg, assert := setupMainTest(t)

	// invalid device ID
	cfg.device_id = cfg.device_id + 1
	n = int(kflowInit(cfg, nil, nil))
	assert.Equal(EKFLOWNODEVICE, n)

	// invalid device IP
	cfg.device_id = 0
	cfg.device_ip = (*_Ctype_char)(unsafe.Pointer(&deviceip[1]))
	n = int(kflowInit(cfg, nil, nil))
	assert.Equal(EKFLOWNODEVICE, n)
}

func setupMainTest(t *testing.T) (*_Ctype_struct___3, *assert.Assertions) {
	client, server, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)

	apiurl = append([]byte(server.URL(test.API).String()), 0)
	email = append([]byte(client.Email))
	token = append([]byte(client.Token))
	deviceip = append([]byte(device.IP.String()), 0)
	program = append([]byte("test"), 0)
	version = append([]byte("0.0.1"), 0)

	cfg := _Ctype_struct___3{
		API: _Ctype_struct___4{
			email: (*_Ctype_char)(unsafe.Pointer(&email[0])),
			token: (*_Ctype_char)(unsafe.Pointer(&token[0])),
			URL:   (*_Ctype_char)(unsafe.Pointer(&apiurl[0])),
		},
		device_id: _Ctype_int(device.ID),
		device_ip: (*_Ctype_char)(unsafe.Pointer(&deviceip[0])),
		program:   (*_Ctype_char)(unsafe.Pointer(&program[0])),
		version:   (*_Ctype_char)(unsafe.Pointer(&version[0])),
	}

	return &cfg, assert
}

var (
	apiurl   []byte
	email    []byte
	token    []byte
	deviceip []byte
	program  []byte
	version  []byte
)
