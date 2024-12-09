package config

import (
	"fmt"

	"github.com/awnumar/memguard"
)

type protectedSecretImpl struct{}

func (*protectedSecretImpl) Container(secret []byte) secretContainer {
	return &protectedSecretContainer{
		enclave: memguard.NewEnclave(secret),
	}
}

func (*protectedSecretImpl) EmptyBuffer() SecretBuffer {
	return &lockedBuffer{}
}

func (*protectedSecretImpl) Wipe(secret []byte) {
	memguard.WipeBytes(secret)
}

type lockedBuffer struct {
	buf *memguard.LockedBuffer
}

func (lb *lockedBuffer) Size() int {
	if lb.buf == nil {
		return 0
	}
	return lb.buf.Size()
}

func (lb *lockedBuffer) Grow(capacity int) {
	size := lb.Size()
	if capacity <= size {
		return
	}

	buf := memguard.NewBuffer(capacity)
	if lb.buf != nil {
		buf.Copy(lb.buf.Bytes())
	}
	lb.buf.Destroy()
	lb.buf = buf
}

func (lb *lockedBuffer) Bytes() []byte {
	if lb.buf == nil {
		return nil
	}
	return lb.buf.Bytes()
}

func (lb *lockedBuffer) TemporaryString() string {
	if lb.buf == nil {
		return ""
	}
	return lb.buf.String()
}

func (lb *lockedBuffer) String() string {
	if lb.buf == nil {
		return ""
	}
	return string(lb.buf.Bytes())
}

func (lb *lockedBuffer) Destroy() {
	if lb.buf == nil {
		return
	}
	lb.buf.Destroy()
	lb.buf = nil
}

type protectedSecretContainer struct {
	enclave *memguard.Enclave
}

func (c *protectedSecretContainer) Destroy() {
	if c.enclave == nil {
		return
	}

	// Wipe the secret from memory
	lockbuf, err := c.enclave.Open()
	if err == nil {
		lockbuf.Destroy()
	}
	c.enclave = nil
}

func (c *protectedSecretContainer) Equals(ref []byte) (bool, error) {
	if c.enclave == nil {
		return false, nil
	}

	// Get a locked-buffer of the secret to perform the comparison
	lockbuf, err := c.enclave.Open()
	if err != nil {
		return false, fmt.Errorf("opening enclave failed: %w", err)
	}
	defer lockbuf.Destroy()

	return lockbuf.EqualTo(ref), nil
}

func (c *protectedSecretContainer) Buffer() (SecretBuffer, error) {
	if c.enclave == nil {
		return &lockedBuffer{}, nil
	}

	// Get a locked-buffer of the secret to perform the comparison
	lockbuf, err := c.enclave.Open()
	if err != nil {
		return nil, fmt.Errorf("opening enclave failed: %w", err)
	}

	return &lockedBuffer{lockbuf}, nil
}

func (c *protectedSecretContainer) AsBuffer(secret []byte) SecretBuffer {
	return &lockedBuffer{memguard.NewBufferFromBytes(secret)}
}

func (c *protectedSecretContainer) Replace(secret []byte) {
	c.enclave = memguard.NewEnclave(secret)
}
