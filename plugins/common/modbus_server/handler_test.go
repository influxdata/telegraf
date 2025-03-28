package modbus_server

import (
	"testing"

	"github.com/simonvetter/modbus"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestNewRequestHandler(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(10, 0, 10, 0, logger)
	require.NoError(t, err)
	require.NotNil(t, handler)
	require.Len(t, handler.coils, 10)
	require.Len(t, handler.holdingRegisters, 10)
}

func TestWriteCoils(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(10, 0, 0, 0, logger)
	require.NoError(t, err)

	values := []bool{true, false, true}
	res, err := handler.WriteCoils(0, values)
	require.NoError(t, err)
	require.Equal(t, values, res)

	// writing outside the server memory
	res, err = handler.WriteCoils(20, values)
	require.NoError(t, err)

	require.Empty(t, res)

	// writing partly outside the server memory
	res, err = handler.WriteCoils(8, values)
	require.NoError(t, err)
	require.Equal(t, []bool{true, false}, res)
	res, err = handler.ReadCoils(0, 10)
	require.Equal(t, []bool{true, false, true, false, false, false, false, false, true, false}, res)
	require.NoError(t, err)
}

func TestReadCoils(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(10, 0, 0, 0, logger)
	require.NoError(t, err)
	values := []bool{true, false, true}
	_, err = handler.WriteCoils(0, values)
	require.NoError(t, err)
	res, err := handler.ReadCoils(0, 3)
	require.NoError(t, err)
	require.Equal(t, values, res)
}

func TestWriteHoldingRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(0, 0, 10, 0, logger)
	require.NoError(t, err)
	values := []uint16{123, 456, 789}
	res, err := handler.WriteHoldingRegisters(0, values)
	require.NoError(t, err)
	require.Equal(t, values, res)

	// writing outside the server memory
	res, err = handler.WriteHoldingRegisters(20, values)
	require.NoError(t, err)
	require.Empty(t, res)

	// writing partly outside the server memory
	res, err = handler.WriteHoldingRegisters(8, values)
	require.NoError(t, err)
	require.Equal(t, []uint16{123, 456}, res)
	res, err = handler.ReadHoldingRegisters(0, 10)
	require.Equal(t, []uint16{123, 456, 789, 0, 0, 0, 0, 0, 123, 456}, res)
	require.NoError(t, err)
}

func TestReadHoldingRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(0, 0, 10, 0, logger)
	require.NoError(t, err)
	values := []uint16{123, 456, 789}
	_, err = handler.WriteHoldingRegisters(0, values)
	require.NoError(t, err)
	res, err := handler.ReadHoldingRegisters(0, 3)
	require.NoError(t, err)
	require.Equal(t, values, res)
}

func TestHandleCoils(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(10, 0, 0, 0, logger)
	require.NoError(t, err)

	req := &modbus.CoilsRequest{
		IsWrite:  true,
		Addr:     0,
		Args:     []bool{true, false, true},
		Quantity: 3,
	}
	res, err := handler.HandleCoils(req)
	require.NoError(t, err)
	require.Equal(t, req.Args, res)

	req = &modbus.CoilsRequest{
		IsWrite:  false,
		Addr:     0,
		Quantity: 3,
	}
	res, err = handler.HandleCoils(req)
	require.NoError(t, err)
	require.Equal(t, []bool{true, false, true}, res)
}

func TestHandleDiscreteInputs(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(10, 0, 10, 0, logger)
	require.NoError(t, err)

	req := &modbus.DiscreteInputsRequest{}
	_, err = handler.HandleDiscreteInputs(req)
	require.Error(t, err)
	require.Equal(t, modbus.ErrIllegalFunction, err)
}

func TestHandleHoldingRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(0, 0, 10, 0, logger)
	require.NoError(t, err)

	req := &modbus.HoldingRegistersRequest{
		IsWrite:  true,
		Addr:     0,
		Args:     []uint16{123, 456, 789},
		Quantity: 3,
	}
	res, err := handler.HandleHoldingRegisters(req)
	require.NoError(t, err)
	require.Equal(t, req.Args, res)

	req = &modbus.HoldingRegistersRequest{
		IsWrite:  false,
		Addr:     0,
		Quantity: 3,
	}
	res, err = handler.HandleHoldingRegisters(req)
	require.NoError(t, err)
	require.Equal(t, []uint16{123, 456, 789}, res)
}

func TestHandleInputRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(0, 0, 10, 0, logger)
	require.NoError(t, err)

	req := &modbus.InputRegistersRequest{}
	_, err = handler.HandleInputRegisters(req)
	require.Error(t, err)
	require.Equal(t, modbus.ErrIllegalFunction, err)
}

func TestWriteBitToHoldingRegister(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(0, 0, 10, 0, logger)
	require.NoError(t, err)

	// Test setting a bit
	register, err := handler.WriteBitToHoldingRegister(0, true, 0)
	require.NoError(t, err)
	require.Equal(t, uint16(1), register)

	// Test clearing a bit
	register, err = handler.WriteBitToHoldingRegister(0, false, 0)
	require.NoError(t, err)
	require.Equal(t, uint16(0), register)

	// Test setting a different bit
	register, err = handler.WriteBitToHoldingRegister(0, true, 1)
	require.NoError(t, err)
	require.Equal(t, uint16(2), register)

	// Test setting a bit out of range
	_, err = handler.WriteBitToHoldingRegister(20, true, 0)
	require.Error(t, err)
	require.Equal(t, modbus.ErrIllegalDataAddress, err)
}
