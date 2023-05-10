package http

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestAESInitErrors(t *testing.T) {
	cipher := AesEncryptor{
		Variant: []string{"AES128", "CBC", "PKCS#5", "superfluous"},
	}
	require.ErrorContains(t, cipher.Init(), "too many variant elements")

	cipher = AesEncryptor{
		Variant: []string{"rsa"},
	}
	require.ErrorContains(t, cipher.Init(), `requested AES but specified "rsa"`)

	cipher = AesEncryptor{}
	require.ErrorContains(t, cipher.Init(), "please specify cipher")

	cipher = AesEncryptor{
		Variant: []string{"aes64"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported AES cipher")

	cipher = AesEncryptor{
		Variant: []string{"aes128", "foo"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported block mode")

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
		Variant: []string{"aes256"},
		Key:     config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0")),
	}
	require.ErrorContains(t, cipher.Init(), "key length (128 bit) does not match cipher (256 bit)")

	cipher = AesEncryptor{
		Variant: []string{"aes128"},
		Key:     config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0")),
	}
	require.ErrorContains(t, cipher.Init(), "'init_vector' has to be specified or derived from password")
}
