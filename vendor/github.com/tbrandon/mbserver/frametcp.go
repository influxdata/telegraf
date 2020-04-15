package mbserver

import (
	"encoding/binary"
	"fmt"
)

// TCPFrame is the Modbus TCP frame.
type TCPFrame struct {
	TransactionIdentifier uint16
	ProtocolIdentifier    uint16
	Length                uint16
	Device                uint8
	Function              uint8
	Data                  []byte
}

// NewTCPFrame converts a packet to a Modbus TCP frame.
func NewTCPFrame(packet []byte) (*TCPFrame, error) {
	// Check if the packet is too short.
	if len(packet) < 9 {
		return nil, fmt.Errorf("TCP Frame error: packet less than 9 bytes")
	}

	frame := &TCPFrame{
		TransactionIdentifier: binary.BigEndian.Uint16(packet[0:2]),
		ProtocolIdentifier:    binary.BigEndian.Uint16(packet[2:4]),
		Length:                binary.BigEndian.Uint16(packet[4:6]),
		Device:                uint8(packet[6]),
		Function:              uint8(packet[7]),
		Data:                  packet[8:],
	}

	// Check expected vs actual packet length.
	if int(frame.Length) != len(frame.Data)+2 {
		return nil, fmt.Errorf("specified packet length does not match actual packet length")
	}

	return frame, nil
}

// Copy the TCPFrame.
func (frame *TCPFrame) Copy() Framer {
	copy := *frame
	return &copy
}

// Bytes returns the Modbus byte stream based on the TCPFrame fields
func (frame *TCPFrame) Bytes() []byte {
	bytes := make([]byte, 8)

	binary.BigEndian.PutUint16(bytes[0:2], frame.TransactionIdentifier)
	binary.BigEndian.PutUint16(bytes[2:4], frame.ProtocolIdentifier)
	binary.BigEndian.PutUint16(bytes[4:6], uint16(2+len(frame.Data)))
	bytes[6] = frame.Device
	bytes[7] = frame.Function
	bytes = append(bytes, frame.Data...)

	return bytes
}

// GetFunction returns the Modbus function code.
func (frame *TCPFrame) GetFunction() uint8 {
	return frame.Function
}

// GetData returns the TCPFrame Data byte field.
func (frame *TCPFrame) GetData() []byte {
	return frame.Data
}

// SetData sets the TCPFrame Data byte field and updates the frame length
// accordingly.
func (frame *TCPFrame) SetData(data []byte) {
	frame.Data = data
	frame.setLength()
}

// SetException sets the Modbus exception code in the frame.
func (frame *TCPFrame) SetException(exception *Exception) {
	frame.Function = frame.Function | 0x80
	frame.Data = []byte{byte(*exception)}
	frame.setLength()
}

func (frame *TCPFrame) setLength() {
	frame.Length = uint16(len(frame.Data) + 2)
}
