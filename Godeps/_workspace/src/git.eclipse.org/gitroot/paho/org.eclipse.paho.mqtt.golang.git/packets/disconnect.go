package packets

import (
	"fmt"
	"github.com/pborman/uuid"
	"io"
)

//DisconnectPacket is an internal representation of the fields of the
//Disconnect MQTT packet
type DisconnectPacket struct {
	FixedHeader
	uuid uuid.UUID
}

func (d *DisconnectPacket) String() string {
	str := fmt.Sprintf("%s\n", d.FixedHeader)
	return str
}

func (d *DisconnectPacket) Write(w io.Writer) error {
	packet := d.FixedHeader.pack()
	_, err := packet.WriteTo(w)

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (d *DisconnectPacket) Unpack(b io.Reader) {
}

//Details returns a Details struct containing the Qos and
//MessageID of this ControlPacket
func (d *DisconnectPacket) Details() Details {
	return Details{Qos: 0, MessageID: 0}
}

//UUID returns the unique ID assigned to the ControlPacket when
//it was originally received. Note: this is not related to the
//MessageID field for MQTT packets
func (d *DisconnectPacket) UUID() uuid.UUID {
	return d.uuid
}
