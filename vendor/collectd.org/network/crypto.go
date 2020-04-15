package network // import "collectd.org/network"

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// PasswordLookup is used when parsing signed and encrypted network traffic to
// look up the password associated with a given username.
type PasswordLookup interface {
	Password(user string) (string, error)
}

// AuthFile implements the PasswordLookup interface in the same way the
// collectd network plugin implements it, i.e. by stat'ing and reading a file.
//
// The file has a very simple syntax with one username / password mapping per
// line, separated by a colon. For example:
//
//   alice: w0nderl4nd
//   bob:   bu1|der
type AuthFile struct {
	name string
	last time.Time
	data map[string]string
	lock *sync.Mutex
}

// NewAuthFile initializes and returns a new AuthFile.
func NewAuthFile(name string) *AuthFile {
	return &AuthFile{
		name: name,
		lock: &sync.Mutex{},
	}
}

// Password looks up a user in the file and returns the associated password.
func (a *AuthFile) Password(user string) (string, error) {
	if a == nil {
		return "", fmt.Errorf("no AuthFile")
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	if err := a.update(); err != nil {
		return "", err
	}

	pwd, ok := a.data[user]
	if !ok {
		return "", fmt.Errorf("no such user: %q", user)
	}

	return pwd, nil
}

func (a *AuthFile) update() error {
	fi, err := os.Stat(a.name)
	if err != nil {
		return err
	}

	if !fi.ModTime().After(a.last) {
		// up to date
		return nil
	}

	file, err := os.Open(a.name)
	if err != nil {
		return err
	}
	defer file.Close()

	newData := make(map[string]string)

	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			break
		}

		line = strings.Trim(line, " \r\n\t\v")
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}

		user := strings.TrimSpace(fields[0])
		pass := strings.TrimSpace(fields[1])
		if strings.HasPrefix(user, "#") {
			continue
		}

		newData[user] = pass
	}

	a.data = newData
	a.last = fi.ModTime()
	return nil
}

func signSHA256(payload []byte, username, password string) []byte {
	mac := hmac.New(sha256.New, bytes.NewBufferString(password).Bytes())

	usernameBuffer := bytes.NewBufferString(username)

	size := uint16(36 + usernameBuffer.Len())

	mac.Write(usernameBuffer.Bytes())
	mac.Write(payload)

	out := new(bytes.Buffer)
	binary.Write(out, binary.BigEndian, uint16(typeSignSHA256))
	binary.Write(out, binary.BigEndian, size)
	out.Write(mac.Sum(nil))
	out.Write(usernameBuffer.Bytes())
	out.Write(payload)

	return out.Bytes()
}

func verifySHA256(part, payload []byte, lookup PasswordLookup) (bool, error) {
	if lookup == nil {
		return false, errors.New("no PasswordLookup available")
	}

	if len(part) <= 32 {
		return false, fmt.Errorf("part too small (%d bytes)", len(part))
	}

	hash := part[:32]
	user := bytes.NewBuffer(part[32:]).String()

	password, err := lookup.Password(user)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, bytes.NewBufferString(password).Bytes())

	mac.Write(part[32:])
	mac.Write(payload)

	return bytes.Equal(hash, mac.Sum(nil)), nil
}

func createCipher(password string, iv []byte) (cipher.Stream, error) {
	passwordHash := sha256.Sum256(bytes.NewBufferString(password).Bytes())

	blockCipher, err := aes.NewCipher(passwordHash[:])
	if err != nil {
		return nil, err
	}

	streamCipher := cipher.NewOFB(blockCipher, iv)
	return streamCipher, nil
}

func encryptAES256(plaintext []byte, username, password string) ([]byte, error) {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		log.Printf("rand.Read: %v", err)
		return nil, err
	}

	streamCipher, err := createCipher(password, iv)
	if err != nil {
		return nil, err
	}

	usernameBuffer := bytes.NewBufferString(username)

	size := uint16(42 + usernameBuffer.Len() + len(plaintext))

	checksum := sha1.Sum(plaintext)

	out := new(bytes.Buffer)
	binary.Write(out, binary.BigEndian, uint16(typeEncryptAES256))
	binary.Write(out, binary.BigEndian, size)
	binary.Write(out, binary.BigEndian, uint16(usernameBuffer.Len()))
	out.Write(usernameBuffer.Bytes())
	out.Write(iv)

	w := &cipher.StreamWriter{S: streamCipher, W: out}
	w.Write(checksum[:])
	w.Write(plaintext)

	return out.Bytes(), nil
}

func decryptAES256(ciphertext []byte, lookup PasswordLookup) ([]byte, error) {
	if lookup == nil {
		return nil, errors.New("no PasswordLookup available")
	}
	if len(ciphertext) < 2 {
		return nil, errors.New("buffer too short")
	}

	buf := bytes.NewBuffer(ciphertext)
	userLen := int(binary.BigEndian.Uint16(buf.Next(2)))
	if 42+userLen >= buf.Len() {
		return nil, fmt.Errorf("invalid username length %d", userLen)
	}
	user := bytes.NewBuffer(buf.Next(userLen)).String()

	password, err := lookup.Password(user)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, 16)
	if n, err := buf.Read(iv); n != 16 || err != nil {
		return nil, fmt.Errorf("reading IV failed: %v", err)
	}

	streamCipher, err := createCipher(password, iv)
	if err != nil {
		return nil, err
	}

	r := &cipher.StreamReader{S: streamCipher, R: buf}

	plaintext := make([]byte, buf.Len())
	if n, err := r.Read(plaintext); n != len(plaintext) || err != nil {
		return nil, fmt.Errorf("decryption failure: got (%d, %v), want (%d, nil)", n, err, len(plaintext))
	}

	checksumWant := plaintext[:20]
	plaintext = plaintext[20:]
	checksumGot := sha1.Sum(plaintext)

	if !bytes.Equal(checksumGot[:], checksumWant[:]) {
		return nil, errors.New("checksum mismatch")
	}

	return plaintext, nil
}
