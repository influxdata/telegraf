package encryption

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed sample_decryption.conf
var sampleConfigDecryption string

type Decrypter interface {
	Decrypt(data []byte) ([]byte, error)
}

type DecryptionConfig struct {
	Cipher string       `toml:"cipher"`
	Aes    AesEncryptor `toml:"aes"`
}

func (*DecryptionConfig) SampleConfig(prefix string) string {
	tmpl := template.Must(template.New("cfg").Parse(sampleConfigDecryption))
	params := struct {
		Prefix string
	}{
		Prefix: prefix,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		panic(fmt.Errorf("creating sample config for prefix %q failed: %w", prefix, err))
	}
	return buf.String()
}

func (c *DecryptionConfig) CreateDecrypter() (Decrypter, error) {
	// For ciphers that allowing variants (e.g. AES256/CBC/PKCS#5Padding)
	// can specify the variant using <algorithm>[/param 1>[/<param 2>]...]
	// where all parameters will be passed on to the decrypter.
	parts := strings.Split(c.Cipher, "/")
	switch strings.ToLower(parts[0]) {
	case "", "none":
		return nil, nil
	case "aes", "aes128", "aes192", "aes256":
		c.Aes.Variant = parts
		if err := c.Aes.Init(); err != nil {
			return nil, fmt.Errorf("init of AES decrypter failed: %w", err)
		}
		return &c.Aes, nil
	}
	return nil, fmt.Errorf("unknown cipher %q", c.Cipher)
}
