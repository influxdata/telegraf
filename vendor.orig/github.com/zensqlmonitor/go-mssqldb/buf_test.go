package mssql

import (
	"bytes"
	"testing"
)

type closableBuffer struct {
	*bytes.Buffer
}

func (closableBuffer) Close() error {
	return nil
}

func makeBuf(bufSize uint16, testData []byte) *tdsBuffer {
	buffer := closableBuffer{bytes.NewBuffer(testData)}
	return newTdsBuffer(bufSize, &buffer)
}

func TestStreamShorterThanHeader(t *testing.T) {
	//buffer := closableBuffer{*bytes.NewBuffer([]byte{0xFF, 0xFF})}
	//buffer := closableBuffer{*bytes.NewBuffer([]byte{0x6F, 0x96, 0x19, 0xFF, 0x8B, 0x86, 0xD0, 0x11, 0xB4, 0x2D, 0x00, 0xC0, 0x4F, 0xC9, 0x64, 0xFF})}
	//tdsBuffer := newTdsBuffer(100, &buffer)
	buffer := makeBuf(100, []byte{0xFF, 0xFF})
	_, err := buffer.BeginRead()
	if err == nil {
		t.Fatal("BeginRead was expected to return error but it didn't")
	} else {
		t.Log("BeginRead failed as expected with error:", err.Error())
	}
}

func TestInvalidLengthInHeaderTooLong(t *testing.T) {
	buffer := makeBuf(8, []byte{0xFF, 0xFF, 0x0, 0x9, 0xff, 0xff, 0xff, 0xff})
	_, err := buffer.BeginRead()
	if err == nil {
		t.Fatal("BeginRead was expected to return error but it didn't")
	} else {
		t.Log("BeginRead failed as expected with error:", err.Error())
	}
}

func TestInvalidLengthInHeaderTooShort(t *testing.T) {
	buffer := makeBuf(100, []byte{0xFF, 0xFF, 0x0, 0x1, 0xff, 0xff, 0xff, 0xff})
	_, err := buffer.BeginRead()
	if err == nil {
		t.Fatal("BeginRead was expected to return error but it didn't")
	} else {
		t.Log("BeginRead failed as expected with error:", err.Error())
	}
}

func TestInvalidLengthInHeaderLongerThanIncomingBuffer(t *testing.T) {
	buffer := makeBuf(9, []byte{0xFF, 0xFF, 0x0, 0x9, 0xff, 0xff, 0xff, 0xff})
	_, err := buffer.BeginRead()
	if err == nil {
		t.Fatal("BeginRead was expected to return error but it didn't")
	} else {
		t.Log("BeginRead failed as expected with error:", err.Error())
	}
}

func TestBeginReadSucceeds(t *testing.T) {
	buffer := makeBuf(9, []byte{0x01 /*id*/, 0xFF /*status*/, 0x0, 0x9 /*size*/, 0xff, 0xff, 0xff, 0xff, 0x02 /*test byte*/})

	id, err := buffer.BeginRead()
	if err != nil {
		t.Fatal("BeginRead failed:", err.Error())
	}
	if id != 1 {
		t.Fatalf("Expected id to be 1 but it is %d", id)
	}

	b, err := buffer.ReadByte()
	if err != nil {
		t.Fatal("ReadByte failed:", err.Error())
	}
	if b != 2 {
		t.Fatalf("Expected read byte to be 2 but it is %d", b)
	}

	// should fail because no more bytes left
	_, err = buffer.ReadByte()
	if err == nil {
		t.Fatal("ReadByte was expected to return error but it didn't")
	} else {
		t.Log("ReadByte failed as expected with error:", err.Error())
	}

	testBuf := []byte{0, 1, 2}
	// should fail because no more bytes left
	_, err = buffer.Read(testBuf)
	if err == nil {
		t.Fatal("Read was expected to return error but it didn't")
	} else {
		t.Log("Read failed as expected with error:", err.Error())
	}
}

func TestReadByteFailsOnSecondPacket(t *testing.T) {
	buffer := makeBuf(9, []byte{
		0x01 /*id*/, 0x0 /*not final*/, 0x0, 0x9 /*size*/, 0xff, 0xff, 0xff, 0xff, 0x02, /*test byte*/
		0x01 /*next id, this packet is invalid, it is too short*/})

	_, err := buffer.BeginRead()
	if err != nil {
		t.Fatal("BeginRead failed:", err.Error())
	}

	_, err = buffer.ReadByte()
	if err != nil {
		t.Fatal("ReadByte failed:", err.Error())
	}

	_, err = buffer.ReadByte()
	if err == nil {
		t.Fatal("ReadByte was expected to return error but it didn't")
	} else {
		t.Log("ReadByte failed as expected with error:", err.Error())
	}
}

func TestReadFailsOnSecondPacket(t *testing.T) {
	buffer := makeBuf(9, []byte{
		0x01 /*id*/, 0x0 /*not final*/, 0x0, 0x9 /*size*/, 0xff, 0xff, 0xff, 0xff, 0x02, /*test byte*/
		0x01 /*next id, this packet is invalid, it is too short*/})

	_, err := buffer.BeginRead()
	if err != nil {
		t.Fatal("BeginRead failed:", err.Error())
	}

	testBuf := []byte{0}
	_, err = buffer.Read(testBuf)
	if err != nil {
		t.Fatal("Read failed:", err.Error())
	}
	if testBuf[0] != 2 {
		t.Fatal("Read returned invalid value")
	}

	_, err = buffer.Read(testBuf)
	if err == nil {
		t.Fatal("ReadByte was expected to return error but it didn't")
	} else {
		t.Log("ReadByte failed as expected with error:", err.Error())
	}
}

func TestWrite(t *testing.T) {
	memBuf := bytes.NewBuffer([]byte{})
	buf := newTdsBuffer(11, closableBuffer{memBuf})
	buf.BeginPacket(1)
	err := buf.WriteByte(2)
	if err != nil {
		t.Fatal("WriteByte failed:", err.Error())
	}
	wrote, err := buf.Write([]byte{3, 4})
	if err != nil {
		t.Fatal("Write failed:", err.Error())
	}
	if wrote != 2 {
		t.Fatalf("Write returned invalid value of written bytes %d", wrote)
	}

	err = buf.FinishPacket()
	if err != nil {
		t.Fatal("FinishPacket failed:", err.Error())
	}
	if bytes.Compare(memBuf.Bytes(), []byte{1, 1, 0, 11, 0, 0, 1, 0, 2, 3, 4}) != 0 {
		t.Fatalf("Written buffer has invalid content: %v", memBuf.Bytes())
	}

	buf.BeginPacket(2)
	wrote, err = buf.Write([]byte{3, 4, 5, 6})
	if err != nil {
		t.Fatal("Write failed:", err.Error())
	}
	if wrote != 4 {
		t.Fatalf("Write returned invalid value of written bytes %d", wrote)
	}
	err = buf.FinishPacket()
	if err != nil {
		t.Fatal("FinishPacket failed:", err.Error())
	}
	expectedBuf := []byte{
		1, 1, 0, 11, 0, 0, 1, 0, 2, 3, 4, // packet 1
		2, 0, 0, 11, 0, 0, 1, 0, 3, 4, 5, // packet 2
		2, 1, 0, 9, 0, 0, 2, 0, 6, // packet 3
	}
	if bytes.Compare(memBuf.Bytes(), expectedBuf) != 0 {
		t.Fatalf("Written buffer has invalid content: %v", memBuf.Bytes())
	}
}
