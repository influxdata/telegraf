package cisco_telemetry_mdt

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	mdtdialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	"google.golang.org/grpc/peer"
)

// MdtDialout RPC server method for grpc-dialout transport
func (c *CiscoTelemetryMDT) MdtDialout(stream mdtdialout.GRPCMdtDialout_MdtDialoutServer) error {
	peerInCtx, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		c.Log.Debugf("Accepted Cisco MDT GRPC dialout connection from %s", peerInCtx.Addr)
		defer c.Log.Debugf("Closed Cisco MDT GRPC dialout connection from %s", peerInCtx.Addr)
	}

	var chunkBuffer bytes.Buffer
	for {
		packet, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				c.acc.AddError(fmt.Errorf("receive error during GRPC dialout: %w", err))
			}
			return nil
		}

		if len(packet.Data) == 0 && len(packet.Errors) != 0 {
			c.acc.AddError(fmt.Errorf("error during GRPC dialout: %s", packet.Errors))
			break
		}

		// Reassemble chunked telemetry data received from NX-OS
		if packet.TotalSize == 0 {
			c.handleTelemetry(packet.Data)
		} else if int64(packet.TotalSize) <= int64(c.MaxMsgSize) {
			chunkBuffer.Write(packet.Data)
			if chunkBuffer.Len() >= int(packet.TotalSize) {
				c.handleTelemetry(chunkBuffer.Bytes())
				chunkBuffer.Reset()
			}
		} else {
			c.acc.AddError(fmt.Errorf("dropped too large packet: %dB > %dB", packet.TotalSize, c.MaxMsgSize))
		}
	}

	return nil
}
