package modbus_server

import (
	"github.com/influxdata/telegraf"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

// Handler object, passed to the NewServer() constructor above.
type Handler struct {
	// this lock is used to avoid concurrency issues between goroutines, as
	// Handler methods are called from different goroutines
	// (1 goroutine per client)
	lock sync.RWMutex

	// these are here to hold client-provided (written) values, for both coils and
	// holding registers

	coils            []bool
	coilOffset       uint16
	holdingRegisters []uint16
	registerOffset   uint16
	LastEdit         chan time.Time
	logger           telegraf.Logger
}

func NewRequestHandler(coilsLen, coilOffset, registersLen, registerOffset uint16, logger telegraf.Logger) (*Handler, error) {
	if coilsLen == 0 && registersLen == 0 {
		return nil, modbus.ErrConfigurationError
	}

	return &Handler{
		coils:            make([]bool, coilsLen),
		coilOffset:       coilOffset,
		holdingRegisters: make([]uint16, registersLen),
		registerOffset:   registerOffset,
		LastEdit:         make(chan time.Time, 1),
		logger:           logger,
	}, nil

}

func (h *Handler) updateLastEdit() {
	// Check if the channel is empty. If empty write the current time to the channel, otherwise update the time.
	select {
	case _ = <-h.LastEdit:
		h.LastEdit <- time.Now()
	default:
		h.LastEdit <- time.Now()
	}
}

func (h *Handler) ReadCoilsAndRegisters() ([]bool, []uint16, uint16, uint16) {
	h.lock.Lock()
	defer h.lock.Unlock()

	coils := make([]bool, len(h.coils))
	registers := make([]uint16, len(h.holdingRegisters))

	copy(coils, h.coils)
	copy(registers, h.holdingRegisters)

	return coils, registers, h.coilOffset, h.registerOffset
}

func (h *Handler) getRegisters(address uint16, quantity uint16) ([]uint16, error) {
	if address < h.registerOffset || address+quantity > h.registerOffset+uint16(len(h.holdingRegisters)) {
		h.logger.Errorf("Reading address out of range: %v, %v, %v", address, quantity, h.registerOffset)
		return nil, modbus.ErrIllegalDataAddress
	}

	res := make([]uint16, quantity)
	copy(res, h.holdingRegisters[address-h.registerOffset:address-h.registerOffset+quantity])

	return res, nil
}

func (h *Handler) setRegisters(address uint16, values []uint16) ([]uint16, error) {
	res := []uint16{}
	for i, value := range values {
		//check if the address is within the range of the holding registers, if not skip the value
		if address+uint16(i) >= h.registerOffset+uint16(len(h.holdingRegisters)) || address+uint16(i) < h.registerOffset {
			continue
		} else {
			h.holdingRegisters[address-h.registerOffset+uint16(i)] = value
			res = append(res, value)
		}
	}
	return res, nil
}

func (h *Handler) WriteBitToHoldingRegister(address uint16, bitValue bool, bitIndex uint8) (register uint16, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	registers, err := h.getRegisters(address, 1)
	if err != nil {
		return 0, err
	}

	currentValue := registers[0]
	if bitValue {
		// Set the bit (use OR to ensure the bit is 1)
		currentValue |= 1 << bitIndex
	} else {
		// Clear the bit (use AND with NOT to ensure the bit is 0)
		currentValue &^= 1 << bitIndex
	}

	registers, err = h.setRegisters(address, []uint16{currentValue})
	if err != nil || len(registers) == 0 {
		return 0, err
	}

	h.updateLastEdit()
	return registers[0], nil
}

func (h *Handler) WriteCoils(address uint16, values []bool) ([]bool, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	res := []bool{}
	for i, value := range values {
		//check if the address is within the range of the coils, if not skip the value
		if address+uint16(i) >= h.coilOffset+uint16(len(h.coils)) || address+uint16(i) < h.coilOffset {
			continue
		} else {
			h.coils[address-h.coilOffset+uint16(i)] = value
			res = append(res, value)
		}
	}

	h.updateLastEdit()
	return res, nil
}

func (h *Handler) ReadCoils(address, quantity uint16) ([]bool, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	//check if the address is within the range of the coils
	if address < h.coilOffset || address+quantity > h.coilOffset+uint16(len(h.coils)) {
		h.logger.Errorf("Reading address out of range: %v, %v, %v", address, quantity, h.coilOffset)
		return nil, modbus.ErrIllegalDataAddress
	}

	res := make([]bool, quantity)
	copy(res, h.coils[address-h.coilOffset:address-h.coilOffset+quantity])
	return res, nil
}

func (h *Handler) WriteHoldingRegisters(address uint16, values []uint16) ([]uint16, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	res, err := h.setRegisters(address, values)

	if err != nil {
		return res, err
	}

	h.updateLastEdit()
	return res, nil
}

func (h *Handler) ReadHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if address < h.registerOffset || address+quantity > h.registerOffset+uint16(len(h.holdingRegisters)) {
		h.logger.Errorf("Reading address out of range: %v, %v, %v", address, quantity, h.registerOffset)
		return nil, modbus.ErrIllegalDataAddress
	}

	res := make([]uint16, quantity)
	copy(res, h.holdingRegisters[address-h.registerOffset:address-h.registerOffset+quantity])
	return res, nil
}

// Coil handler method.
// This method gets called whenever a valid modbus request asking for a coil operation is
// received by the server.
func (h *Handler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	h.logger.Debugf("Handling coils request: %+v", req)
	if req.IsWrite {
		h.logger.Debugf("Writing coils: %+v, args: %+v", req.Addr, req.Args)
		res, err = h.WriteCoils(req.Addr, req.Args)
		h.logger.Debugf("Write coils: %+v", res)
		// Check if the channel is empty. If empty write the current time to the channel, otherwise update the time.
	} else {
		h.logger.Debugf("Reading coils: %+v, quantity %+v", req.Addr, req.Quantity)
		res, err = h.ReadCoils(req.Addr, req.Quantity)
		h.logger.Debugf("Read coils: %+v", res)
	}
	return
}

// Discrete input handler method.
// Note that we're returning ErrIllegalFunction unconditionally.
// This will cause the client to receive "illegal function", which is the modbus way of
// reporting that this server does not support/implement the discrete input type.
func (h *Handler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	// this is the equivalent of saying
	// "discrete inputs are not supported by this device"
	// (try it with modbus-cli --target tcp://localhost:5502 rdi:1)
	h.logger.Error("Discrete inputs are not supported by this device")
	err = modbus.ErrIllegalFunction

	return
}

// Holding register handler method.
// This method gets called whenever a valid modbus request asking for a holding register
// operation (either read or write) received by the server.
func (h *Handler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	h.logger.Debugf("Handling register reguest: %+v", req)
	if req.IsWrite {
		h.logger.Debugf("Writing registers: %+v, args: %+v", req.Addr, req.Args)
		res, err = h.WriteHoldingRegisters(req.Addr, req.Args)
		h.logger.Debugf("Write registers: %+v", res)
	} else {
		h.logger.Debugf("Reading registers: %+v, quantity: %+v ", req.Addr, req.Quantity)
		res, err = h.ReadHoldingRegisters(req.Addr, req.Quantity)
		h.logger.Debugf("Read registers: %+v", res)
	}

	return
}

// Input register handler method.
// This method gets called whenever a valid modbus request asking for an input register
// operation is received by the server.
// Note that input registers are always read-only as per the modbus spec.
func (h *Handler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	h.logger.Error("Input registers are not supported by this device")
	err = modbus.ErrIllegalFunction
	return
}
