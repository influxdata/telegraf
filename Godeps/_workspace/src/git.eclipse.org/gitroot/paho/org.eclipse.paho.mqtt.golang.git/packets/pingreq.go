package packets

import (
	"fmt"
	"github.com/pborman/uuid"
	"io"
)

//PingreqPacket is an internal representation of the fields of the
//Pingreq MQTT packet
type PingreqPacket struct {
	FixedHeader
	uuid uuid.UUID
}

func (pr *PingreqPacket) String() string {
	str := fmt.Sprintf("%s", pr.FixedHeader)
	return str
}

func (pr *PingreqPacket) Write(w io.Writer) error {
	packet := pr.FixedHeader.pack()
	_, err := packet.WriteTo(w)

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (pr *PingreqPacket) Unpack(b io.Reader) {
}

//Details returns a Details struct containing the Qos and
//MessageID of this ControlPacket
func (pr *PingreqPacket) Details() Details {
	return Details{Qos: 0, MessageID: 0}
}

//UUID returns the unique ID assigned to the ControlPacket when
//it was originally received. Note: this is not related to the
//MessageID field for MQTT packets
func (pr *PingreqPacket) UUID() uuid.UUID {
	return pr.uuid
}
