package packets

import (
	"bytes"
	"fmt"
	"github.com/pborman/uuid"
	"io"
)

//PublishPacket is an internal representation of the fields of the
//Publish MQTT packet
type PublishPacket struct {
	FixedHeader
	TopicName string
	MessageID uint16
	Payload   []byte
	uuid      uuid.UUID
}

func (p *PublishPacket) String() string {
	str := fmt.Sprintf("%s\n", p.FixedHeader)
	str += fmt.Sprintf("topicName: %s MessageID: %d\n", p.TopicName, p.MessageID)
	str += fmt.Sprintf("payload: %s\n", string(p.Payload))
	return str
}

func (p *PublishPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeString(p.TopicName))
	if p.Qos > 0 {
		body.Write(encodeUint16(p.MessageID))
	}
	p.FixedHeader.RemainingLength = body.Len() + len(p.Payload)
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	packet.Write(p.Payload)
	_, err = w.Write(packet.Bytes())

	return err
}

//Unpack decodes the details of a ControlPacket after the fixed
//header has been read
func (p *PublishPacket) Unpack(b io.Reader) {
	var payloadLength = p.FixedHeader.RemainingLength
	p.TopicName = decodeString(b)
	if p.Qos > 0 {
		p.MessageID = decodeUint16(b)
		payloadLength -= len(p.TopicName) + 4
	} else {
		payloadLength -= len(p.TopicName) + 2
	}
	p.Payload = make([]byte, payloadLength)
	b.Read(p.Payload)
}

//Copy creates a new PublishPacket with the same topic and payload
//but an empty fixed header, useful for when you want to deliver
//a message with different properties such as Qos but the same
//content
func (p *PublishPacket) Copy() *PublishPacket {
	newP := NewControlPacket(Publish).(*PublishPacket)
	newP.TopicName = p.TopicName
	newP.Payload = p.Payload

	return newP
}

//Details returns a Details struct containing the Qos and
//MessageID of this ControlPacket
func (p *PublishPacket) Details() Details {
	return Details{Qos: p.Qos, MessageID: p.MessageID}
}

//UUID returns the unique ID assigned to the ControlPacket when
//it was originally received. Note: this is not related to the
//MessageID field for MQTT packets
func (p *PublishPacket) UUID() uuid.UUID {
	return p.uuid
}
