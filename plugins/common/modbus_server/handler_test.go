package modbus_server

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/simonvetter/modbus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewRequestHandler(t *testing.T) {
	logger := testutil.Logger{}
	handler, err := NewRequestHandler(10, 0, 10, 0, logger)
	require.NoError(t, err)
	require.NotNil(t, handler)
	require.Equal(t, 10, len(handler.coils))
	require.Equal(t, 10, len(handler.holdingRegisters))
}

func TestWriteCoils(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(10, 0, 0, 0, logger)
	values := []bool{true, false, true}
	res, err := handler.WriteCoils(0, values)
	require.NoError(t, err)
	require.Equal(t, values, res)

	// writing outside the server memory
	res, err = handler.WriteCoils(20, values)
	require.NoError(t, err)
	require.Equal(t, []bool{}, res)

	// writing partly outside the server memory
	res, err = handler.WriteCoils(8, values)
	require.NoError(t, err)
	require.Equal(t, []bool{true, false}, res)
	res, err = handler.ReadCoils(0, 10)
	require.Equal(t, []bool{true, false, true, false, false, false, false, false, true, false}, res)
}

func TestReadCoils(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(10, 0, 0, 0, logger)
	values := []bool{true, false, true}
	handler.WriteCoils(0, values)
	res, err := handler.ReadCoils(0, 3)
	require.NoError(t, err)
	require.Equal(t, values, res)
}

func TestWriteHoldingRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(0, 0, 10, 0, logger)
	values := []uint16{123, 456, 789}
	res, err := handler.WriteHoldingRegisters(0, values)
	require.NoError(t, err)
	require.Equal(t, values, res)

	// writing outside the server memory
	res, err = handler.WriteHoldingRegisters(20, values)
	require.NoError(t, err)
	require.Equal(t, []uint16{}, res)

	// writing partly outside the server memory
	res, err = handler.WriteHoldingRegisters(8, values)
	require.NoError(t, err)
	require.Equal(t, []uint16{123, 456}, res)
	res, err = handler.ReadHoldingRegisters(0, 10)
	require.Equal(t, []uint16{123, 456, 789, 0, 0, 0, 0, 0, 123, 456}, res)

}

func TestReadHoldingRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(0, 0, 10, 0, logger)
	values := []uint16{123, 456, 789}
	handler.WriteHoldingRegisters(0, values)
	res, err := handler.ReadHoldingRegisters(0, 3)
	require.NoError(t, err)
	require.Equal(t, values, res)
}

func TestHandleCoils(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(10, 0, 0, 0, logger)
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
	handler, _ := NewRequestHandler(10, 0, 10, 0, logger)
	req := &modbus.DiscreteInputsRequest{}
	_, err := handler.HandleDiscreteInputs(req)
	require.Error(t, err)
	require.Equal(t, modbus.ErrIllegalFunction, err)
}

func TestHandleHoldingRegisters(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(0, 0, 10, 0, logger)
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
	handler, _ := NewRequestHandler(0, 0, 10, 0, logger)
	req := &modbus.InputRegistersRequest{}
	_, err := handler.HandleInputRegisters(req)
	require.Error(t, err)
	require.Equal(t, modbus.ErrIllegalFunction, err)
}

func TestWriteBitToHoldingRegister(t *testing.T) {
	logger := testutil.Logger{}
	handler, _ := NewRequestHandler(0, 0, 10, 0, logger)

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
