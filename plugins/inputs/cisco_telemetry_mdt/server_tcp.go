package cisco_telemetry_mdt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

func (c *CiscoTelemetryMDT) acceptTCPClients() {
	// Keep track of all active connections, so we can close them if necessary
	var mutex sync.Mutex
	clients := make(map[net.Conn]bool)

	for {
		conn, err := c.listener.Accept()
		var neterr *net.OpError
		if errors.As(err, &neterr) && (neterr.Timeout() || neterr.Temporary()) {
			continue
		} else if err != nil {
			break // Stop() will close the connection so Accept() will fail here
		}

		mutex.Lock()
		clients[conn] = true
		mutex.Unlock()

		// Individual client connection routine
		c.wg.Add(1)
		go func() {
			c.Log.Debugf("Accepted Cisco MDT TCP dialout connection from %s", conn.RemoteAddr())
			if err := c.handleTCPClient(conn); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					c.acc.AddError(err)
				}
			}
			c.Log.Debugf("Closed Cisco MDT TCP dialout connection from %s", conn.RemoteAddr())

			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()

			if err := conn.Close(); err != nil {
				c.Log.Warnf("closing connection failed: %v", err)
			}
			c.wg.Done()
		}()
	}

	// Close all remaining client connections
	mutex.Lock()
	for client := range clients {
		if err := client.Close(); err != nil {
			c.Log.Errorf("Failed to close TCP dialout client: %v", err)
		}
	}
	mutex.Unlock()
}

// Handle a TCP telemetry client
func (c *CiscoTelemetryMDT) handleTCPClient(conn net.Conn) error {
	var payload bytes.Buffer
	var hdr header

	for {
		// Read and validate dialout telemetry header
		if err := binary.Read(conn, binary.BigEndian, &hdr); err != nil {
			return fmt.Errorf("reading header failed: %w", err)
		}

		// Maximum telemetry payload size (in bytes) to accept for GRPC dialout transport
		maxMsgSize := uint32(1024 * 1024)
		if c.MaxMsgSize > 0 {
			maxMsgSize = uint32(c.MaxMsgSize)
		}

		if hdr.MsgLen > maxMsgSize {
			return fmt.Errorf("dialout packet too long: %v", hdr.MsgLen)
		} else if hdr.MsgFlags != 0 {
			return fmt.Errorf("invalid dialout flags: %v", hdr.MsgFlags)
		}

		// Read and handle telemetry packet
		payload.Reset()
		if size, err := payload.ReadFrom(io.LimitReader(conn, int64(hdr.MsgLen))); size != int64(hdr.MsgLen) {
			if err != nil {
				return fmt.Errorf("reading payload failed: %w", err)
			}
			return errors.New("premature EOF during TCP dialout")
		}

		c.handleTelemetry(payload.Bytes())
	}
}
