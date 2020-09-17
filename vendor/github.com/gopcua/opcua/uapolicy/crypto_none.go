package uapolicy

const (
	NoneBlockSize  = 1
	NoneMinPadding = 0
)

type None struct{}

func (c *None) Decrypt(src []byte) ([]byte, error) {
	var b []byte
	return append(b, src...), nil
}

func (c *None) Encrypt(src []byte) ([]byte, error) {
	var b []byte
	return append(b, src...), nil
}

func (s *None) Signature(msg []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

func (s *None) Verify(msg, signature []byte) error {
	return nil
}
