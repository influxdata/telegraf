package network // import "collectd.org/network"

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

type mockPasswordLookup map[string]string

func (l mockPasswordLookup) Password(user string) (string, error) {
	pass, ok := l[user]
	if !ok {
		return "", errors.New("not found")
	}

	return pass, nil
}

func TestSign(t *testing.T) {
	want := []byte{
		2, 0, 0, 41,
		0xcd, 0xa5, 0x9a, 0x37, 0xb0, 0x81, 0xc2, 0x31,
		0x24, 0x2a, 0x6d, 0xbd, 0xfb, 0x44, 0xdb, 0xd7,
		0x41, 0x2a, 0xf4, 0x29, 0x83, 0xde, 0xa5, 0x11,
		0x96, 0xd2, 0xe9, 0x30, 0x21, 0xae, 0xc5, 0x45,
		'a', 'd', 'm', 'i', 'n',
		'c', 'o', 'l', 'l', 'e', 'c', 't', 'd',
	}
	got := signSHA256([]byte{'c', 'o', 'l', 'l', 'e', 'c', 't', 'd'}, "admin", "admin")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	passwords := mockPasswordLookup{
		"admin": "admin",
	}
	ok, err := verifySHA256(want[4:41], want[41:], passwords)
	if !ok || err != nil {
		t.Errorf("got (%v, %v), want (true, nil)", ok, err)
	}

	want[41], want[42] = want[42], want[41] // corrupt data
	ok, err = verifySHA256(want[4:41], want[41:], passwords)
	if ok || err != nil {
		t.Errorf("got (%v, %v), want (false, nil)", ok, err)
	}

	want[41], want[42] = want[42], want[41] // fix data
	passwords["admin"] = "test123"          // different password
	ok, err = verifySHA256(want[4:41], want[41:], passwords)
	if ok || err != nil {
		t.Errorf("got (%v, %v), want (false, nil)", ok, err)
	}
}

func TestEncrypt(t *testing.T) {
	plaintext := []byte{'c', 'o', 'l', 'l', 'e', 'c', 't', 'd'}
	// actual ciphertext depends on IV -- only check the first part
	want := []byte{
		0x02, 0x10, // part type
		0x00, 0x37, // part length
		0x00, 0x05, // username length
		0x61, 0x64, 0x6d, 0x69, 0x6e, // username
		// IV
		// SHA1
		// encrypted data
	}

	ciphertext, err := encryptAES256(plaintext, "admin", "admin")
	if !bytes.Equal(want, ciphertext[:11]) || err != nil {
		t.Errorf("got (%v, %v), want (%v, nil)", ciphertext[:11], err, want)
	}

	passwords := mockPasswordLookup{
		"admin": "admin",
	}
	if got, err := decryptAES256(ciphertext[4:], passwords); !bytes.Equal(got, plaintext) || err != nil {
		t.Errorf("got (%v, %v), want (%v, nil)", got, err, plaintext)
	}

	ciphertext[47], ciphertext[48] = ciphertext[48], ciphertext[47] // corrupt data
	if got, err := decryptAES256(ciphertext[4:], passwords); got != nil || err == nil {
		t.Errorf("got (%v, %v), want (nil, \"checksum mismatch\")", got, err)
	}

	ciphertext[47], ciphertext[48] = ciphertext[48], ciphertext[47] // fix data
	passwords["admin"] = "test123"                                  // different password
	if got, err := decryptAES256(ciphertext[4:], passwords); got != nil || err == nil {
		t.Errorf("got (%v, %v), want (nil, \"no such user\")", got, err)
	}
}
