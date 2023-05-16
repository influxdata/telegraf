package http

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestAES(t *testing.T) {
	keySource := hex.EncodeToString([]byte("0123456789abcdefghijklmnopqrstuvwxyz"))
	expected := "my $ecret-Passw0rd"
	iv := hex.EncodeToString([]byte("0123456789abcdef"))
	tests := []struct {
		cipher    string
		encrypted string
		key       string
	}{
		{
			cipher:    "AES128/CBC/PKCS#5",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			key:       keySource[:32],
		},
		{
			cipher:    "AES192/CBC/PKCS#5",
			encrypted: "D3A5A0004B6783351F89B00C1D4154EDF2321EDAD3111B5551C18836B9FCFD62",
			key:       keySource[:48],
		},
		{
			cipher:    "AES256/CBC/PKCS#5",
			encrypted: "9751D7FB4B1497DEBC8A95C5D88097ECB1B8E63979E2D41E7ECD304D6B39B808",
			key:       keySource[:64],
		},
	}
	for _, tt := range tests {
		t.Run(tt.cipher, func(t *testing.T) {
			decrypter := AesEncryptor{
				Variant: strings.Split(tt.cipher, "/"),
				Key:     config.NewSecret([]byte(tt.key)),
				Vec:     config.NewSecret([]byte(iv)),
			}
			require.NoError(t, decrypter.Init())
			enc, err := hex.DecodeString(tt.encrypted)
			require.NoError(t, err)
			dec, err := decrypter.Decrypt(enc)
			require.NoError(t, err)
			require.Equal(t, expected, string(dec))
		})
	}
}

func TestAESNoPadding(t *testing.T) {
	keySource := hex.EncodeToString([]byte("0123456789abcdefghijklmnopqrstuvwxyz"))
	expected := "my $ecret-Passw0rd"
	iv := hex.EncodeToString([]byte("0123456789abcdef"))
	tests := []struct {
		cipher    string
		encrypted string
		key       string
	}{
		{
			cipher:    "AES128/CBC",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			key:       keySource[:32],
		},
		{
			cipher:    "AES192/CBC",
			encrypted: "D3A5A0004B6783351F89B00C1D4154EDF2321EDAD3111B5551C18836B9FCFD62",
			key:       keySource[:48],
		},
		{
			cipher:    "AES256/CBC",
			encrypted: "9751D7FB4B1497DEBC8A95C5D88097ECB1B8E63979E2D41E7ECD304D6B39B808",
			key:       keySource[:64],
		},
	}
	for _, tt := range tests {
		t.Run(tt.cipher, func(t *testing.T) {
			decrypter := AesEncryptor{
				Variant: strings.Split(tt.cipher, "/"),
				Key:     config.NewSecret([]byte(tt.key)),
				Vec:     config.NewSecret([]byte(iv)),
			}
			require.NoError(t, decrypter.Init())
			enc, err := hex.DecodeString(tt.encrypted)
			require.NoError(t, err)
			dec, err := decrypter.Decrypt(enc)
			require.NoError(t, err)
			require.Len(t, string(dec), 32)
			require.Contains(t, string(dec), expected)
		})
	}
}

func TestAESKDF(t *testing.T) {
	expected := "my $ecret-Passw0rd"
	iv := hex.EncodeToString([]byte("asupersecretiv42"))
	tests := []struct {
		cipher     string
		password   string
		salt       string
		iterations int
		encrypted  string
	}{
		{
			cipher:     "AES256/CBC/PKCS#5",
			password:   "a secret password",
			salt:       "somerandombytes",
			iterations: 2000,
			encrypted:  "224b169206ce918f167ae0da18f4de45bede0d2c853d45e55f1422d1446037bf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.cipher, func(t *testing.T) {
			decrypter := AesEncryptor{
				Variant: strings.Split(tt.cipher, "/"),
				KDFConfig: KDFConfig{
					Algorithm:  "PBKDF2-HMAC-SHA256",
					Passwd:     config.NewSecret([]byte(tt.password)),
					Salt:       config.NewSecret([]byte(tt.salt)),
					Iterations: tt.iterations,
				},
				Vec: config.NewSecret([]byte(iv)),
			}
			require.NoError(t, decrypter.Init())
			enc, err := hex.DecodeString(tt.encrypted)
			require.NoError(t, err)
			dec, err := decrypter.Decrypt(enc)
			require.NoError(t, err)
			require.Equal(t, expected, string(dec))
		})
	}
}

func TestAESInitErrors(t *testing.T) {
	tests := []struct {
		name     string
		variant  []string
		key      string
		iv       string
		kdfcfg   *KDFConfig
		expected string
	}{
		{
			name:     "no mode",
			variant:  []string{"AES128"},
			expected: "please specify cipher mode",
		},
		{
			name:     "too many elements",
			variant:  []string{"AES128", "CBC", "PKCS#5", "superfluous"},
			expected: "too many variant elements",
		},
		{
			name:     "no AES",
			variant:  []string{"rsa", "cbc"},
			expected: `requested AES but specified "rsa"`,
		},
		{
			name:     "no cipher",
			expected: "please specify cipher",
		},
		{
			name:     "unsupported cipher",
			variant:  []string{"aes64", "cbc"},
			expected: "unsupported AES cipher",
		},
		{
			name:     "unsupported mode",
			variant:  []string{"aes128", "foo"},
			expected: "unsupported cipher mode",
		},
		{
			name:     "unsupported padding",
			variant:  []string{"aes128", "cbc", "bar"},
			expected: "unsupported padding",
		},
		{
			name:     "missing key",
			variant:  []string{"aes128", "cbc", "none"},
			expected: "either key or password has to be specified",
		},
		{
			name:     "wrong key length",
			variant:  []string{"aes256", "cbc"},
			key:      "63238c069e3c5d6aaa20048c43ce4ed0",
			expected: "key length (128 bit) does not match cipher (256 bit)",
		},
		{
			name:     "invalid key",
			variant:  []string{"aes256", "cbc"},
			key:      "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			expected: "decoding key failed: encoding/hex: invalid byte: U+0078 'x'",
		},
		{
			name:     "missing IV",
			variant:  []string{"aes128", "cbc"},
			key:      "63238c069e3c5d6aaa20048c43ce4ed0",
			expected: "'init_vector' has to be specified or derived from password",
		},
		{
			name:     "invalid IV",
			variant:  []string{"aes128", "cbc"},
			key:      "63238c069e3c5d6aaa20048c43ce4ed0",
			iv:       "abcd",
			expected: "init vector size must match block size",
		},
		{
			name:    "missing salt and iterations",
			variant: []string{"aes128", "cbc", "none"},
			kdfcfg: &KDFConfig{
				Passwd: config.NewSecret([]byte("secret")),
			},
			expected: "salt and iterations required for password-based-keys",
		},
		{
			name:    "wrong keygen algorithm",
			variant: []string{"aes128", "cbc", "none"},
			kdfcfg: &KDFConfig{
				Algorithm:  "foo",
				Passwd:     config.NewSecret([]byte("secret")),
				Salt:       config.NewSecret([]byte("salt")),
				Iterations: 2000,
			},
			expected: "unknown key-derivation function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.expected)

			decrypter := AesEncryptor{
				Variant: tt.variant,
			}
			if tt.key != "" {
				decrypter.Key = config.NewSecret([]byte(tt.key))
			}
			if tt.iv != "" {
				decrypter.Vec = config.NewSecret([]byte(tt.iv))
			}
			if tt.kdfcfg != nil {
				decrypter.KDFConfig = *tt.kdfcfg
			}
			require.ErrorContains(t, decrypter.Init(), tt.expected)
		})
	}
}

func TestAESDecryptError(t *testing.T) {
	tests := []struct {
		name      string
		encrypted string
		messMode  string
		messKey   string
		messIV    string
		expected  string
	}{
		{
			name:      "wrong data length",
			encrypted: "abcd",
			expected:  "invalid data size",
		},
		{
			name:      "mode tampered",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			messMode:  "tampered",
			expected:  `unsupported cipher mode "tampered"`,
		},
		{
			name:      "invalid key",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			messKey:   "tampered",
			expected:  "decoding key failed: encoding/hex: invalid byte: U+0074 't'",
		},
		{
			name:      "wrong key length",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			messKey:   "01234567",
			expected:  "creating AES cipher failed: crypto/aes: invalid key size",
		},
		{
			name:      "invalid key",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			messIV:    "tampered",
			expected:  "decoding init vector failed: encoding/hex: invalid byte: U+0074 't'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.expected)

			decrypter := AesEncryptor{
				Variant: []string{"AES128", "CBC", "PKCS#5"},
				Key:     config.NewSecret([]byte(hex.EncodeToString([]byte("0123456789abcdef")))),
				Vec:     config.NewSecret([]byte(hex.EncodeToString([]byte("0123456789abcdef")))),
			}
			require.NoError(t, decrypter.Init())
			enc, err := hex.DecodeString(tt.encrypted)
			require.NoError(t, err)

			// Mess with the internal values for testing
			if tt.messMode != "" {
				decrypter.mode = tt.messMode
			}
			if tt.messKey != "" {
				decrypter.Key = config.NewSecret([]byte(tt.messKey))
			}
			if tt.messIV != "" {
				decrypter.Vec = config.NewSecret([]byte(tt.messIV))
			}
			_, err = decrypter.Decrypt(enc)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}
