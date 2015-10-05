package packets

import (
	"bytes"
	"fmt"
	"github.com/pborman/uuid"
	"io"
)

//SubackPacket is an internal representation of the fields of the
//Suback MQTT packet
type SubackPacket struct {
	FixedHeader
	MessageID   uint16
	GrantedQoss []byte
	uuid        uuid.UUID
}

func (sa *SubackPacket) String() string {
	str := fmt.Sprintf("%s\n", sa.FixedHeader)
	str += fmt.Sprintf("MessageID: %d", sa.MessageID)
	return str
}

func (sa *SubackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(encodeUint16(sa.MessageID))
	body.Write(sa.GrantedQoss)
	sa.FixedHeader.RemainingLength = body.Len()
	packet := sa.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (sa *SubackPacket) Unpack(b io.Reader) {
	var qosBuffer bytes.Buffer
	sa.MessageID = decodeUint16(b)
	qosBuffer.ReadFrom(b)
	sa.GrantedQoss = qosBuffer.Bytes()
}

//Details returns a Details struct containing the Qos and
//MessageID of this ControlPacket
func (sa *SubackPacket) Details() Details {
	return Details{Qos: 0, MessageID: sa.MessageID}
}

//UUID returns the unique ID assigned to the ControlPacket when
//it was originally received. Note: this is not related to the
//MessageID field for MQTT packets
func (sa *SubackPacket) UUID() uuid.UUID {
	return sa.uuid
}
