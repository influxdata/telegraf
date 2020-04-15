package ldap

import (
	"errors"

	ber "gopkg.in/asn1-ber.v1"
)

var (
	errRespChanClosed = errors.New("ldap: response channel closed")
	errCouldNotRetMsg = errors.New("ldap: could not retrieve message")
)

type request interface {
	appendTo(*ber.Packet) error
}

type requestFunc func(*ber.Packet) error

func (f requestFunc) appendTo(p *ber.Packet) error {
	return f(p)
}

func (l *Conn) doRequest(req request) (*messageContext, error) {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, l.nextMessageID(), "MessageID"))
	if err := req.appendTo(packet); err != nil {
		return nil, err
	}

	if l.Debug {
		ber.PrintPacket(packet)
	}

	msgCtx, err := l.sendMessage(packet)
	if err != nil {
		return nil, err
	}
	l.Debug.Printf("%d: returning", msgCtx.id)
	return msgCtx, nil
}

func (l *Conn) readPacket(msgCtx *messageContext) (*ber.Packet, error) {
	l.Debug.Printf("%d: waiting for response", msgCtx.id)
	packetResponse, ok := <-msgCtx.responses
	if !ok {
		return nil, NewError(ErrorNetwork, errRespChanClosed)
	}
	packet, err := packetResponse.ReadPacket()
	l.Debug.Printf("%d: got response %p", msgCtx.id, packet)
	if err != nil {
		return nil, err
	}

	if packet == nil {
		return nil, NewError(ErrorNetwork, errCouldNotRetMsg)
	}

	if l.Debug {
		if err = addLDAPDescriptions(packet); err != nil {
			return nil, err
		}
		ber.PrintPacket(packet)
	}
	return packet, nil
}
