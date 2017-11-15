package nsca

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	STATE_OK = iota
	STATE_WARNING
	STATE_CRITICAL
	STATE_UNKNOWN
)

const (
	ENCRYPT_NONE        = iota /* no encryption */
	ENCRYPT_XOR                /* not really encrypted, just obfuscated */
	ENCRYPT_DES                /* DES */
	ENCRYPT_3DES               /* 3DES or Triple DES */
	ENCRYPT_CAST128            /* CAST-128 */            /* UNUSED */
	ENCRYPT_CAST256            /* CAST-256 */            /* UNUSED */
	ENCRYPT_XTEA               /* xTEA */                /* UNUSED */
	ENCRYPT_3WAY               /* 3-WAY */               /* UNUSED */
	ENCRYPT_BLOWFISH           /* SKIPJACK */            /* UNUSED */
	ENCRYPT_TWOFISH            /* TWOFISH */             /* UNUSED */
	ENCRYPT_LOKI97             /* LOKI97 */              /* UNUSED */
	ENCRYPT_RC2                /* RC2 */                 /* UNUSED */
	ENCRYPT_ARCFOUR            /* RC4 */                 /* UNUSED */
	ENCRYPT_RC6                /* RC6 */                 /* UNUSED */
	ENCRYPT_RIJNDAEL128        /* RIJNDAEL-128 */        /* AES-128 */
	ENCRYPT_RIJNDAEL192        /* RIJNDAEL-192 */        /* AES-192 */
	ENCRYPT_RIJNDAEL256        /* RIJNDAEL-256 */        /* AES-256 */
	ENCRYPT_MARS               /* MARS */                /* UNUSED */
	ENCRYPT_PANAMA             /* PANAMA */              /* UNUSED */
	ENCRYPT_WAKE               /* WAKE */                /* UNUSED */
	ENCRYPT_SERPENT            /* SERPENT */             /* UNUSED */
	ENCRYPT_IDEA               /* IDEA */                /* UNUSED */
	ENCRYPT_ENIGMA             /* ENIGMA (Unix crypt) */ /* UNUSED */
	ENCRYPT_GOST               /* GOST */                /* UNUSED */
	ENCRYPT_SAFER64            /* SAFER-sk64 */          /* UNUSED */
	ENCRYPT_SAFER128           /* SAFER-sk128 */         /* UNUSED */
	ENCRYPT_SAFERPLUS          /* SAFER+ */              /* UNUSED */
)

type dataPacket struct {
	packetVersion      int16
	crc32              uint32
	timestamp          uint32
	returnCode         int16
	hostName           string // 64 char max
	serviceDescription string // 128 char max
	pluginOutput       string // 512 char max
}

type initializationPacket struct {
	iv        []byte // 128 bytes
	timestamp uint32
}

type encryption struct {
	method   int
	iv       []byte
	password []byte
}

func (e *encryption) encrypt(b []byte) error {
	if e.method == ENCRYPT_NONE {
		return nil
	}
	if len(e.password) == 0 {
		return fmt.Errorf("Zero length password")
	}
	if e.method == ENCRYPT_XOR {
		for i := range b {
			b[i] = b[i] ^ e.iv[i%len(e.iv)] ^ e.password[i%len(e.password)]
		}
		return nil
	}
	var err error
	var block cipher.Block
	key := make([]byte, 128)
	copy(key, e.password)
	switch e.method {
	case ENCRYPT_DES:
		block, err = des.NewCipher(key[:des.BlockSize])
	case ENCRYPT_3DES:
		block, err = des.NewTripleDESCipher(key[:des.BlockSize*3])
	case ENCRYPT_RIJNDAEL128:
		block, err = aes.NewCipher(key[:16])
	case ENCRYPT_RIJNDAEL192:
		block, err = aes.NewCipher(key[:24])
	case ENCRYPT_RIJNDAEL256:
		block, err = aes.NewCipher(key[:32])
	case ENCRYPT_CAST128:
		fallthrough
	case ENCRYPT_CAST256:
		fallthrough
	case ENCRYPT_XTEA:
		fallthrough
	case ENCRYPT_3WAY:
		fallthrough
	case ENCRYPT_BLOWFISH:
		fallthrough
	case ENCRYPT_TWOFISH:
		fallthrough
	case ENCRYPT_LOKI97:
		fallthrough
	case ENCRYPT_RC2:
		fallthrough
	case ENCRYPT_ARCFOUR:
		fallthrough
	case ENCRYPT_RC6:
		fallthrough
	case ENCRYPT_MARS:
		fallthrough
	case ENCRYPT_PANAMA:
		fallthrough
	case ENCRYPT_WAKE:
		fallthrough
	case ENCRYPT_SERPENT:
		fallthrough
	case ENCRYPT_IDEA:
		fallthrough
	case ENCRYPT_ENIGMA:
		fallthrough
	case ENCRYPT_GOST:
		fallthrough
	case ENCRYPT_SAFER64:
		fallthrough
	case ENCRYPT_SAFER128:
		fallthrough
	case ENCRYPT_SAFERPLUS:
		err = fmt.Errorf("Unsupported encryption method")
	default:
		err = fmt.Errorf("Unrecognized encryption method")
	}
	if err != nil {
		return err
	}
	enc := cipher.NewCFBEncrypter(block, e.iv[:block.BlockSize()])
	enc.XORKeyStream(b, b)
	return nil
}

func newEncryption(method int, iv []byte, password string) *encryption {
	e := encryption{
		method:   method,
		iv:       make([]byte, len(iv)),
		password: make([]byte, len(password)),
	}
	copy(e.iv, iv)
	copy(e.password, password)
	return &e
}

func readInitializationPacket(reader io.Reader) (*initializationPacket, error) {
	p := initializationPacket{iv: make([]byte, 128)}
	err := binary.Read(reader, binary.BigEndian, p.iv)
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.BigEndian, &p.timestamp)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func makeBuffer(s string, length int) ([]byte, error) {
	if length == 0 {
		return make([]byte, 0), nil
	}
	b := make([]byte, length)
	n, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		return nil, fmt.Errorf("Unexpected result from rand.Read")
	}
	n = copy(b, s)
	if n == len(b) {
		b[len(b)-1] = 0
	} else {
		b[n] = 0
	}
	return b, nil
}

func newDataPacket(serverTimestamp uint32, returnCode int16, hostName, serviceDescription, pluginOutput string) *dataPacket {
	d := dataPacket{
		packetVersion:      3,
		timestamp:          serverTimestamp,
		returnCode:         returnCode,
		hostName:           hostName,
		serviceDescription: serviceDescription,
		pluginOutput:       pluginOutput,
	}
	return &d
}

func (p *dataPacket) write(w io.Writer, e *encryption) error {
	if p.packetVersion == 0 {
		p.packetVersion = 3
	}
	p.crc32 = 0
	hostName, err := makeBuffer(p.hostName, 64)
	if err != nil {
		return err
	}
	service, err := makeBuffer(p.serviceDescription, 128)
	if err != nil {
		return err
	}
	output, err := makeBuffer(p.pluginOutput, 512)
	if err != nil {
		return err
	}
	// 2 bytes for c struct padding
	padding, err := makeBuffer("", 2)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, p.packetVersion)
	binary.Write(buf, binary.BigEndian, padding)
	binary.Write(buf, binary.BigEndian, p.crc32)
	binary.Write(buf, binary.BigEndian, p.timestamp)
	binary.Write(buf, binary.BigEndian, p.returnCode)
	binary.Write(buf, binary.BigEndian, hostName)
	binary.Write(buf, binary.BigEndian, service)
	binary.Write(buf, binary.BigEndian, output)
	binary.Write(buf, binary.BigEndian, padding)

	b := buf.Bytes()
	p.crc32 = crc32.ChecksumIEEE(b)
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, p.crc32)
	copy(b[4:], crc)
	err = e.encrypt(b)
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("Wrong byte count returned from write: expected %d, got %d", len(b), n)
	}
	return nil
}
