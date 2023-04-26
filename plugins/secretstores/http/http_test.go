package http

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	secretstores.Add("http", func(id string) telegraf.SecretStore {
		return &HTTP{Log: testutil.Logger{}}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		t.Run(fname, func(t *testing.T) {
			testdataPath := filepath.Join("testcases", fname)
			configFilename := filepath.Join(testdataPath, "telegraf.conf")
			inputFilename := filepath.Join(testdataPath, "secrets.json")
			expectedFilename := filepath.Join(testdataPath, "expected.json")

			// Read the input data
			input, err := os.ReadFile(inputFilename)
			require.NoError(t, err)

			// Read the expected output data
			buf, err := os.ReadFile(expectedFilename)
			require.NoError(t, err)
			var expected map[string]string
			require.NoError(t, json.Unmarshal(buf, &expected))

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.NotEmpty(t, cfg.SecretStores)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/secrets" {
					_, _ = w.Write(input)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			us, err := url.Parse(server.URL)
			require.NoError(t, err)

			var id string
			var plugin telegraf.SecretStore
			actual := make(map[string]string, len(expected))
			for id, plugin = range cfg.SecretStores {
				// Setup dummy server and redirect the plugin's URL to that dummy
				httpPlugin, ok := plugin.(*HTTP)
				require.True(t, ok)

				u, err := url.Parse(httpPlugin.URL)
				require.NoError(t, err)
				u.Host = us.Host
				httpPlugin.URL = u.String()
				require.NoError(t, httpPlugin.download())

				// Retrieve the secrets from the plugin
				keys, err := plugin.List()
				require.NoError(t, err)

				for _, k := range keys {
					v, err := plugin.Get(k)
					require.NoError(t, err)
					actual[id+"."+k] = string(v)
				}
			}
			require.EqualValues(t, expected, actual)
		})
	}
}

func TestInitErrors(t *testing.T) {
	plugin := &HTTP{Transformation: "{some: malformed"}
	require.ErrorContains(t, plugin.Init(), "setting up data transformation failed")

	plugin = &HTTP{DecryptionConfig: DecryptionConfig{Cipher: "non-existing/CBC/lala"}}
	require.ErrorContains(t, plugin.Init(), "creating decryptor failed: unknown cipher")
}

func TestSetNotSupported(t *testing.T) {
	plugin := &HTTP{}
	require.NoError(t, plugin.Init())

	require.ErrorContains(t, plugin.Set("key", "value"), "setting secrets not supported")
}

func TestGetErrors(t *testing.T) {
	plugin := &HTTP{
		DecryptionConfig: DecryptionConfig{
			Cipher: "AES256/CBC/PKCS#5",
			Aes: AesEncryptor{
				Key: config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0a910eef95f22f55bacdddacafa06b656")),
				Vec: config.NewSecret([]byte("asupersecretiv42")),
			},
		},
	}
	require.NoError(t, plugin.Init())

	_, err := plugin.Get("OMG")
	require.ErrorContains(t, err, "not found")

	plugin.cache = map[string]string{"test": "aedMZXaLR246OHHjVtJKXQ=X"}
	_, err = plugin.Get("test")
	require.ErrorContains(t, err, "base64 decoding failed")
}

func TestResolver(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"test": "aedMZXaLR246OHHjVtJKXQ=="}`))
	}))
	defer server.Close()

	plugin := &HTTP{
		URL: server.URL,
		DecryptionConfig: DecryptionConfig{
			Cipher: "AES256/CBC/PKCS#5",
			Aes: AesEncryptor{
				Key: config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0a910eef95f22f55bacdddacafa06b656")),
				Vec: config.NewSecret([]byte("asupersecretiv42")),
			},
		},
	}
	plugin.Timeout = config.Duration(200 * time.Millisecond)
	require.NoError(t, plugin.Init())

	resolver, err := plugin.GetResolver("test")
	require.NoError(t, err)

	s, _, err := resolver()
	require.NoError(t, err)
	require.Equal(t, "password-B", string(s))
}

func TestGetResolverErrors(t *testing.T) {
	dummy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer dummy.Close()

	plugin := &HTTP{
		URL: "http://" + dummy.Addr().String(),
	}
	plugin.Timeout = config.Duration(200 * time.Millisecond)
	require.NoError(t, plugin.Init())

	_, err = plugin.GetResolver("test")
	require.ErrorContains(t, err, "context deadline exceeded")
	dummy.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"test": "aedMZXaLR246OHHjVtJKXQ=="}]`))
	}))
	defer server.Close()

	plugin = &HTTP{
		URL: server.URL,
		DecryptionConfig: DecryptionConfig{
			Cipher: "AES256/CBC/PKCS#5",
			Aes: AesEncryptor{
				Key: config.NewSecret([]byte("63238c069e3c5d6aaa20048c43ce4ed0a910eef95f22f55bacdddacafa06b656")),
				Vec: config.NewSecret([]byte("asupersecretiv42")),
			},
		},
	}
	plugin.Timeout = config.Duration(200 * time.Millisecond)
	require.NoError(t, plugin.Init())

	_, err = plugin.GetResolver("test")
	require.ErrorContains(t, err, "maybe missing or wrong data transformation")

	plugin.Transformation = "{awe:skds}"
	require.NoError(t, plugin.Init())

	_, err = plugin.GetResolver("test")
	require.ErrorContains(t, err, "transforming data failed")
}

func TestInitAESErrors(t *testing.T) {
	cipher := AesEncryptor{
		Variant: []string{"AES", "CBC", "PKCS#5", "superfluous"},
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
		Variant: []string{"aes", "foo"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported block mode")

	cipher = AesEncryptor{
		Variant: []string{"aes", "cbc", "bar"},
	}
	require.ErrorContains(t, cipher.Init(), "unsupported padding")

	cipher = AesEncryptor{
		Variant: []string{"aes", "cbc", "none"},
	}
	require.ErrorContains(t, cipher.Init(), "either key or password has to be specified")

	cipher = AesEncryptor{
		Variant: []string{"aes", "cbc", "none"},
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
