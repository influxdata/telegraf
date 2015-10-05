package packets

import (
	"fmt"
	"github.com/pborman/uuid"
	"io"
)

//PubrelPacket is an internal representation of the fields of the
//Pubrel MQTT packet
type PubrelPacket struct {
	FixedHeader
	MessageID uint16
	uuid      uuid.UUID
}

func (pr *PubrelPacket) String() string {
	str := fmt.Sprintf("%s\n", pr.FixedHeader)
	str += fmt.Sprintf("MessageID: %d", pr.MessageID)
	return str
}

func (pr *PubrelPacket) Write(w io.Writer) error {
	var err error
	pr.FixedHeader.RemainingLength = 2
	packet := pr.FixedHeader.pack()
	packet.Write(encodeUint16(pr.MessageID))
	_, err = packet.WriteTo(w)

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (pr *PubrelPacket) Unpack(b io.Reader) {
	pr.MessageID = decodeUint16(b)
}

//Details returns a Details struct containing the Qos and
//MessageID of this ControlPacket
func (pr *PubrelPacket) Details() Details {
	return Details{Qos: pr.Qos, MessageID: pr.MessageID}
}

//UUID returns the unique ID assigned to the ControlPacket when
//it was originally received. Note: this is not related to the
//MessageID field for MQTT packets
func (pr *PubrelPacket) UUID() uuid.UUID {
	return pr.uuid
}
