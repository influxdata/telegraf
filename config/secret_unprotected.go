package config

import (
	"bytes"
	"unsafe"
)

type unprotectedSecretImpl struct{}

func (*unprotectedSecretImpl) Container(secret []byte) secretContainer {
	return &unprotectedSecretContainer{buf: newUnlockedBuffer(secret)}
}

func (*unprotectedSecretImpl) EmptyBuffer() SecretBuffer {
	return &unlockedBuffer{}
}

func (*unprotectedSecretImpl) Wipe(secret []byte) {
	for i := range secret {
		secret[i] = 0
	}
}

type unlockedBuffer struct {
	content []byte
}

func newUnlockedBuffer(secret []byte) *unlockedBuffer {
	return &unlockedBuffer{bytes.Clone(secret)}
}

func (lb *unlockedBuffer) Size() int {
	return len(lb.content)
}

func (lb *unlockedBuffer) Grow(_ int) {
	// The underlying byte-buffer will grow dynamically
}

func (lb *unlockedBuffer) Bytes() []byte {
	return lb.content
}

func (lb *unlockedBuffer) TemporaryString() string {
	//nolint:gosec // G103: Valid use of unsafe call to cast underlying bytes to string
	return unsafe.String(&lb.content[0], len(lb.content))
}

func (lb *unlockedBuffer) String() string {
	return string(lb.content)
}

func (lb *unlockedBuffer) Destroy() {
	selectedImpl.Wipe(lb.content)
	lb.content = nil
}

type unprotectedSecretContainer struct {
	buf *unlockedBuffer
}

func (c *unprotectedSecretContainer) Destroy() {
	if c.buf == nil {
		return
	}

	// Wipe the secret from memory
	c.buf.Destroy()
	c.buf = nil
}

func (c *unprotectedSecretContainer) Equals(ref []byte) (bool, error) {
	if c.buf == nil {
		return false, nil
	}

	return bytes.Equal(c.buf.content, ref), nil
}

func (c *unprotectedSecretContainer) Buffer() (SecretBuffer, error) {
	if c.buf == nil {
		return &unlockedBuffer{}, nil
	}

	return newUnlockedBuffer(c.buf.content), nil
}

func (c *unprotectedSecretContainer) AsBuffer(secret []byte) SecretBuffer {
	return &unlockedBuffer{secret}
}

func (c *unprotectedSecretContainer) Replace(secret []byte) {
	c.buf = newUnlockedBuffer(secret)
}
