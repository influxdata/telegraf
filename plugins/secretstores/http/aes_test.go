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
	tests := []struct {
		cipher    string
		encrypted string
		key       string
		iv        string
	}{
		{
			cipher:    "AES128/CBC/PKCS#5",
			encrypted: "9E36B490B0B1D6CE28550DF9DE65FC0013FF9F0939E24DA4A24324BDB5EABA04",
			key:       keySource[:32],
			iv:        "0123456789abcdef",
		},
		{
			cipher:    "AES192/CBC/PKCS#5",
			encrypted: "D3A5A0004B6783351F89B00C1D4154EDF2321EDAD3111B5551C18836B9FCFD62",
			key:       keySource[:48],
			iv:        "0123456789abcdef",
		},
		{
			cipher:    "AES256/CBC/PKCS#5",
			encrypted: "9751D7FB4B1497DEBC8A95C5D88097ECB1B8E63979E2D41E7ECD304D6B39B808",
			key:       keySource[:64],
			iv:        "0123456789abcdef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.cipher, func(t *testing.T) {
			decrypter := AesEncryptor{
				Variant: strings.Split(tt.cipher, "/"),
				Key:     config.NewSecret([]byte(tt.key)),
				Vec:     config.NewSecret([]byte(tt.iv)),
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
	cipher := AesEncryptor{
		Variant: []string{"aes128"},
	}
	require.ErrorContains(t, cipher.Init(), "please specify cipher mode")

	cipher = AesEncryptor{
		Variant: []string{"AES128", "CBC", "PKCS#5", "superfluous"},
	}
	require.ErrorContains(t, cipher.Init(), "too many variant elements")

	cipher = AesEncryptor{
		Variant: []string{"rsa", "cbc"},
	}
	require.ErrorContains(t, cipher.Init(), `requested AES but specified "rsa"`)

	cipher = AesEncryptor{}
	require.ErrorContains(t, cipher.Init(), "please specify cipher")

	cipher = AesEncryptor{
		Variant: []string{"aes64", "cbc"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported AES cipher")

	cipher = AesEncryptor{
		Variant: []string{"aes128", "foo"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported cipher mode")

	cipher = AesEncryptor{
		Variant: []string{"aes128", "cbc", "bar"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported padding")

	cipher = AesEncryptor{
		Variant: []string{"aes128", "cbc", "none"},
	}
	require.ErrorContains(t, cipher.Init(), "either key or password has to be specified")

	cipher = AesEncryptor{
		Variant: []string{"aes128", "cbc", "none"},
		KDFConfig: KDFConfig{
			Passwd: config.NewSecret([]byte("secret")),
		},
	}
	require.ErrorContains(t, cipher.Init(), "salt and iterations required for password-based-keys")

	cipher = AesEncryptor{
		Variant: []string{"aes256", "cbc"},
		Key:     config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0")),
	}
	require.ErrorContains(t, cipher.Init(), "key length (128 bit) does not match cipher (256 bit)")

	cipher = AesEncryptor{
		Variant: []string{"aes128", "cbc"},
		Key:     config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0")),
	}
	require.ErrorContains(t, cipher.Init(), "'init_vector' has to be specified or derived from password")
}
