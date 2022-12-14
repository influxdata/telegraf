package encryption

type Decryptor interface {
	Init() error
	Decrypt(data []byte) ([]byte, error)
}
