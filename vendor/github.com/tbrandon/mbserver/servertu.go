package mbserver

import (
	"io"
	"log"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Fatalf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequests(port)
	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
	for {
		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return
		}

		if bytesRead != 0 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				return
			}

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}
