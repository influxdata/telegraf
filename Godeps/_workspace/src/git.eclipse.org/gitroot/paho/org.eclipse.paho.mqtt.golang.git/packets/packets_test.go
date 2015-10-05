package packets

import (
	"bytes"
	"testing"
)

func TestPacketNames(t *testing.T) {
	if PacketNames[1] != "CONNECT" {
		t.Errorf("PacketNames[1] is %s, should be %s", PacketNames[1], "CONNECT")
	}
	if PacketNames[2] != "CONNACK" {
		t.Errorf("PacketNames[2] is %s, should be %s", PacketNames[2], "CONNACK")
	}
	if PacketNames[3] != "PUBLISH" {
		t.Errorf("PacketNames[3] is %s, should be %s", PacketNames[3], "PUBLISH")
	}
	if PacketNames[4] != "PUBACK" {
		t.Errorf("PacketNames[4] is %s, should be %s", PacketNames[4], "PUBACK")
	}
	if PacketNames[5] != "PUBREC" {
		t.Errorf("PacketNames[5] is %s, should be %s", PacketNames[5], "PUBREC")
	}
	if PacketNames[6] != "PUBREL" {
		t.Errorf("PacketNames[6] is %s, should be %s", PacketNames[6], "PUBREL")
	}
	if PacketNames[7] != "PUBCOMP" {
		t.Errorf("PacketNames[7] is %s, should be %s", PacketNames[7], "PUBCOMP")
	}
	if PacketNames[8] != "SUBSCRIBE" {
		t.Errorf("PacketNames[8] is %s, should be %s", PacketNames[8], "SUBSCRIBE")
	}
	if PacketNames[9] != "SUBACK" {
		t.Errorf("PacketNames[9] is %s, should be %s", PacketNames[9], "SUBACK")
	}
	if PacketNames[10] != "UNSUBSCRIBE" {
		t.Errorf("PacketNames[10] is %s, should be %s", PacketNames[10], "UNSUBSCRIBE")
	}
	if PacketNames[11] != "UNSUBACK" {
		t.Errorf("PacketNames[11] is %s, should be %s", PacketNames[11], "UNSUBACK")
	}
	if PacketNames[12] != "PINGREQ" {
		t.Errorf("PacketNames[12] is %s, should be %s", PacketNames[12], "PINGREQ")
	}
	if PacketNames[13] != "PINGRESP" {
		t.Errorf("PacketNames[13] is %s, should be %s", PacketNames[13], "PINGRESP")
	}
	if PacketNames[14] != "DISCONNECT" {
		t.Errorf("PacketNames[14] is %s, should be %s", PacketNames[14], "DISCONNECT")
	}
}

func TestPacketConsts(t *testing.T) {
	if Connect != 1 {
		t.Errorf("Const for Connect is %d, should be %d", Connect, 1)
	}
	if Connack != 2 {
		t.Errorf("Const for Connack is %d, should be %d", Connack, 2)
	}
	if Publish != 3 {
		t.Errorf("Const for Publish is %d, should be %d", Publish, 3)
	}
	if Puback != 4 {
		t.Errorf("Const for Puback is %d, should be %d", Puback, 4)
	}
	if Pubrec != 5 {
		t.Errorf("Const for Pubrec is %d, should be %d", Pubrec, 5)
	}
	if Pubrel != 6 {
		t.Errorf("Const for Pubrel is %d, should be %d", Pubrel, 6)
	}
	if Pubcomp != 7 {
		t.Errorf("Const for Pubcomp is %d, should be %d", Pubcomp, 7)
	}
	if Subscribe != 8 {
		t.Errorf("Const for Subscribe is %d, should be %d", Subscribe, 8)
	}
	if Suback != 9 {
		t.Errorf("Const for Suback is %d, should be %d", Suback, 9)
	}
	if Unsubscribe != 10 {
		t.Errorf("Const for Unsubscribe is %d, should be %d", Unsubscribe, 10)
	}
	if Unsuback != 11 {
		t.Errorf("Const for Unsuback is %d, should be %d", Unsuback, 11)
	}
	if Pingreq != 12 {
		t.Errorf("Const for Pingreq is %d, should be %d", Pingreq, 12)
	}
	if Pingresp != 13 {
		t.Errorf("Const for Pingresp is %d, should be %d", Pingresp, 13)
	}
	if Disconnect != 14 {
		t.Errorf("Const for Disconnect is %d, should be %d", Disconnect, 14)
	}
}

func TestConnackConsts(t *testing.T) {
	if Accepted != 0x00 {
		t.Errorf("Const for Accepted is %d, should be %d", Accepted, 0)
	}
	if ErrRefusedBadProtocolVersion != 0x01 {
		t.Errorf("Const for RefusedBadProtocolVersion is %d, should be %d", ErrRefusedBadProtocolVersion, 1)
	}
	if ErrRefusedIDRejected != 0x02 {
		t.Errorf("Const for RefusedIDRejected is %d, should be %d", ErrRefusedIDRejected, 2)
	}
	if ErrRefusedServerUnavailable != 0x03 {
		t.Errorf("Const for RefusedServerUnavailable is %d, should be %d", ErrRefusedServerUnavailable, 3)
	}
	if ErrRefusedBadUsernameOrPassword != 0x04 {
		t.Errorf("Const for RefusedBadUsernameOrPassword is %d, should be %d", ErrRefusedBadUsernameOrPassword, 4)
	}
	if ErrRefusedNotAuthorised != 0x05 {
		t.Errorf("Const for RefusedNotAuthorised is %d, should be %d", ErrRefusedNotAuthorised, 5)
	}
}

func TestConnectPacket(t *testing.T) {
	connectPacketBytes := bytes.NewBuffer([]byte{16, 52, 0, 4, 77, 81, 84, 84, 4, 204, 0, 0, 0, 0, 0, 4, 116, 101, 115, 116, 0, 12, 84, 101, 115, 116, 32, 80, 97, 121, 108, 111, 97, 100, 0, 8, 116, 101, 115, 116, 117, 115, 101, 114, 0, 8, 116, 101, 115, 116, 112, 97, 115, 115})
	packet, err := ReadPacket(connectPacketBytes)
	if err != nil {
		t.Fatalf("Error reading packet: %s", err.Error())
	}
	cp := packet.(*ConnectPacket)
	if cp.ProtocolName != "MQTT" {
		t.Errorf("Connect Packet ProtocolName is %s, should be %s", cp.ProtocolName, "MQTT")
	}
	if cp.ProtocolVersion != 4 {
		t.Errorf("Connect Packet ProtocolVersion is %d, should be %d", cp.ProtocolVersion, 4)
	}
	if cp.UsernameFlag != true {
		t.Errorf("Connect Packet UsernameFlag is %t, should be %t", cp.UsernameFlag, true)
	}
	if cp.Username != "testuser" {
		t.Errorf("Connect Packet Username is %s, should be %s", cp.Username, "testuser")
	}
	if cp.PasswordFlag != true {
		t.Errorf("Connect Packet PasswordFlag is %t, should be %t", cp.PasswordFlag, true)
	}
	if string(cp.Password) != "testpass" {
		t.Errorf("Connect Packet Password is %s, should be %s", string(cp.Password), "testpass")
	}
	if cp.WillFlag != true {
		t.Errorf("Connect Packet WillFlag is %t, should be %t", cp.WillFlag, true)
	}
	if cp.WillTopic != "test" {
		t.Errorf("Connect Packet WillTopic is %s, should be %s", cp.WillTopic, "test")
	}
	if cp.WillQos != 1 {
		t.Errorf("Connect Packet WillQos is %d, should be %d", cp.WillQos, 1)
	}
	if cp.WillRetain != false {
		t.Errorf("Connect Packet WillRetain is %t, should be %t", cp.WillRetain, false)
	}
	if string(cp.WillMessage) != "Test Payload" {
		t.Errorf("Connect Packet WillMessage is %s, should be %s", string(cp.WillMessage), "Test Payload")
	}
}
