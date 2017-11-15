package nsca

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func TestMakeBuffer(t *testing.T) {
	x, err := makeBuffer("", 0)
	if err != nil || len(x) != 0 {
		t.Errorf("Bad return: x %q, len(x) %d, err %v", x, len(x), err)
	}
	x, err = makeBuffer("", 1)
	if err != nil || len(x) != 1 {
		t.Errorf("Bad return: x %q, len(x) %d, err %v", x, len(x), err)
	}
	if x[0] != 0 {
		t.Errorf("Bad null termination: x %q, len(x) %d, err %v", x, len(x), err)
	}
	x, err = makeBuffer("", 512)
	if err != nil || len(x) != 512 {
		t.Errorf("Bad return: x %q, len(x) %d, err %v", x, len(x), err)
	}
	if x[0] != 0 {
		t.Errorf("Bad null termination: x %q, len(x) %d, err %v", x, len(x), err)
	}
	x, err = makeBuffer("abcdef", 2)
	if err != nil || len(x) != 2 {
		t.Errorf("Bad return: x %q, len(x) %d, err %v", x, len(x), err)
	}
	if x[1] != 0 {
		t.Errorf("Bad null termination: x %q, len(x) %d, err %v", x, len(x), err)
	}
	x, err = makeBuffer("abcdef", 6)
	if err != nil || len(x) != 6 {
		t.Errorf("Bad return: x %q, len(x) %d, err %v", x, len(x), err)
	}
	x, err = makeBuffer("abcdef", 9)
	if err != nil || len(x) != 9 {
		t.Errorf("Bad return: x %q, len(x) %d, err %v", x, len(x), err)
	}
	if string(x[:6]) != "abcdef" {
		t.Errorf("Bad string value: x %q, len(x) %d, err %v", x, len(x), err)
	}
	if x[6] != 0 {
		t.Errorf("Bad null termination: x %q, len(x) %d, err %v", x, len(x), err)
	}
}

func testEncryptionMethod(method int, shouldFail bool, t *testing.T) {
	var e *encryption
	iv := make([]byte, 128)
	password := "abc"
	plain := []byte("hello")
	var err error
	e = newEncryption(method, iv, password)
	err = e.encrypt(plain)
	if !shouldFail && err != nil {
		t.Errorf("Encryption error on %d: %s", method, err)
	} else if shouldFail && err == nil {
		t.Errorf("Should have failed on %d, but no error", method)
	}
	// TODO: test some boundary conditions on iv, password and plain
	// TODO: implement a decrypt method so we can test round trip
}

func TestEncryption(t *testing.T) {
	testEncryptionMethod(ENCRYPT_NONE, false, t)
	testEncryptionMethod(ENCRYPT_XOR, false, t)
	testEncryptionMethod(ENCRYPT_DES, false, t)
	testEncryptionMethod(ENCRYPT_3DES, false, t)
	testEncryptionMethod(ENCRYPT_RIJNDAEL128, false, t)
	testEncryptionMethod(ENCRYPT_RIJNDAEL192, false, t)
	testEncryptionMethod(ENCRYPT_RIJNDAEL256, false, t)

	testEncryptionMethod(ENCRYPT_CAST128, true, t)
	testEncryptionMethod(ENCRYPT_CAST256, true, t)
	testEncryptionMethod(ENCRYPT_XTEA, true, t)
	testEncryptionMethod(ENCRYPT_3WAY, true, t)
	testEncryptionMethod(ENCRYPT_BLOWFISH, true, t)
	testEncryptionMethod(ENCRYPT_TWOFISH, true, t)
	testEncryptionMethod(ENCRYPT_LOKI97, true, t)
	testEncryptionMethod(ENCRYPT_RC2, true, t)
	testEncryptionMethod(ENCRYPT_ARCFOUR, true, t)
	testEncryptionMethod(ENCRYPT_RC6, true, t)
	testEncryptionMethod(ENCRYPT_MARS, true, t)
	testEncryptionMethod(ENCRYPT_PANAMA, true, t)
	testEncryptionMethod(ENCRYPT_WAKE, true, t)
	testEncryptionMethod(ENCRYPT_SERPENT, true, t)
	testEncryptionMethod(ENCRYPT_IDEA, true, t)
	testEncryptionMethod(ENCRYPT_ENIGMA, true, t)
	testEncryptionMethod(ENCRYPT_GOST, true, t)
	testEncryptionMethod(ENCRYPT_SAFER64, true, t)
	testEncryptionMethod(ENCRYPT_SAFER128, true, t)
	testEncryptionMethod(ENCRYPT_SAFERPLUS, true, t)
}

func TestSession(t *testing.T) {
	// read initialization from server
	packet := new(bytes.Buffer)
	iv := make([]byte, 128)
	rand.Read(iv)
	binary.Write(packet, binary.BigEndian, iv)
	timestamp := uint32(time.Now().Unix())
	binary.Write(packet, binary.BigEndian, timestamp)
	ip, err := readInitializationPacket(packet)
	if err != nil {
		t.Errorf("Error reading initialization packet: %s", err)
	} else {
		if ip.timestamp != timestamp {
			t.Errorf("Bad timestamp. Expected %d, got %d", timestamp, ip.timestamp)
		}
		if bytes.Compare(ip.iv, iv) != 0 {
			t.Errorf("Bad iv")
		}
	}
	// create Encryption
	enc := newEncryption(ENCRYPT_NONE, ip.iv, "testpassword")
	// create message
	msg := newDataPacket(ip.timestamp, STATE_OK, "testHost", "testService", "A plugin message")
	// write message
	writer := new(bytes.Buffer)
	err = msg.write(writer, enc)
	if err != nil {
		t.Errorf("Error writing message: %s", err)
	} else {
		// check message
	}
}

func TestServer(t *testing.T) {
	// TODO: disable the Skip if you have a real NSCA server to test against
	t.Skip("Skipping test that uses a real NSCA server")
	conn, err := net.Dial("tcp", ":5666")
	if err != nil {
		t.Fatalf("Could not connect to server: %s", err)
	}
	defer conn.Close()
	ip, err := readInitializationPacket(conn)
	if err != nil {
		t.Fatalf("Could not read initialization packet: %s", err)
	}
	// create Encryption
	enc := newEncryption(ENCRYPT_DES, ip.iv, "password")
	// create message
	msg := newDataPacket(ip.timestamp, STATE_OK, "testHost", "testService", "A plugin message")
	// write message
	for i := 0; i < 10; i++ {
		err = msg.write(conn, enc)
		if err != nil {
			t.Errorf("Error writing message: %s", err)
		}
	}
}
