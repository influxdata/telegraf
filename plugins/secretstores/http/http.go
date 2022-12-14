//go:generate ../../../tools/readme_config_includer/generator
package http

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/blues/jsonata-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/encryption"
	chttp "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type store struct {
	PlainText map[string]string `json:"plain"`
	Encrypted map[string][]byte `json:"encrypted"`
}

type HTTP struct {
	URL                string                  `toml:"url"`
	Headers            map[string]string       `toml:"headers"`
	Username           config.Secret           `toml:"username"`
	Password           config.Secret           `toml:"password"`
	BearerToken        string                  `toml:"bearer_token"`
	SuccessStatusCodes []int                   `toml:"success_status_codes"`
	Transformation     string                  `toml:"transformation"`
	Cipher             string                  `toml:"cipher"`
	CipherAes          encryption.AesEncryptor `toml:"aes"`
	Log                telegraf.Logger         `toml:"-"`
	chttp.HTTPClientConfig

	client      *http.Client
	transformer *jsonata.Expr
	cache       store
	decryptor   encryption.Decryptor
}

func (*HTTP) SampleConfig() string {
	return sampleConfig
}

func (h *HTTP) Init() error {
	ctx := context.Background()
	client, err := h.HTTPClientConfig.CreateClient(ctx, h.Log)
	if err != nil {
		return err
	}
	h.client = client

	// Set default as [200]
	if len(h.SuccessStatusCodes) == 0 {
		h.SuccessStatusCodes = []int{200}
	}

	// Setup the data transformer if any
	if h.Transformation != "" {
		e, err := jsonata.Compile(h.Transformation)
		if err != nil {
			return fmt.Errorf("setting up data transformation failed: %w", err)
		}
		h.transformer = e
	}

	// Setup the decryption infrastructure
	parts := strings.Split(h.Cipher, "/")
	switch strings.ToLower(parts[0]) {
	case "":
	case "aes":
		h.CipherAes.Variant = parts
		if err := h.CipherAes.Init(); err != nil {
			return fmt.Errorf("AES init failed: %w", err)
		}
		h.decryptor = &h.CipherAes
	default:
		return fmt.Errorf("unknown cipher %q", h.Cipher)
	}

	// Download and parse the credentials
	if err := h.download(); err != nil {
		return err
	}

	// Check if we do have a decryptor for encrypted data
	if len(h.cache.Encrypted) > 0 && h.decryptor == nil {
		return errors.New("got encrypted secrets but no cipher specified")
	}

	return nil
}

// Get searches for the given key and return the secret
func (h *HTTP) Get(key string) ([]byte, error) {
	if v, found := h.cache.PlainText[key]; found {
		return []byte(v), nil
	}

	v, found := h.cache.Encrypted[key]
	if !found {
		return nil, errors.New("not found")
	}
	return h.decryptor.Decrypt(v)
}

// Set sets the given secret for the given key
func (h *HTTP) Set(key, value string) error {
	return errors.New("setting secrets not supported")
}

// List lists all known secret keys
func (h *HTTP) List() ([]string, error) {
	keys := make([]string, 0, len(h.cache.PlainText)+len(h.cache.Encrypted))
	for k := range h.cache.PlainText {
		keys = append(keys, k)
	}
	for k := range h.cache.Encrypted {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetResolver returns a function to resolve the given key.
func (h *HTTP) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := h.Get(key)
		return s, false, err
	}
	return resolver, nil
}

func (h *HTTP) download() error {
	// Get the raw data form the URL
	data, err := h.query()
	if err != nil {
		return fmt.Errorf("reading body failed: %v", err)
	}

	// Transform the data to the expected form if given
	if h.transformer != nil {
		out, err := h.transformer.EvalBytes(data)
		if err != nil {
			return fmt.Errorf("transforming data failed: %v", err)
		}
		data = out
	}

	// Extract the data from the resulting data
	return json.Unmarshal(data, &h.cache)
}

func (h *HTTP) query() ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, h.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	if h.BearerToken != "" {
		token, err := os.ReadFile(h.BearerToken)
		if err != nil {
			return nil, fmt.Errorf("reading bearer file failed: %w", err)
		}
		bearer := "Bearer " + strings.Trim(string(token), "\n")
		request.Header.Set("Authorization", bearer)
	}

	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			request.Host = v
		} else {
			request.Header.Add(k, v)
		}
	}

	if err := h.setRequestAuth(request); err != nil {
		return nil, err
	}

	resp, err := h.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	responseHasSuccessCode := false
	for _, statusCode := range h.SuccessStatusCodes {
		if resp.StatusCode == statusCode {
			responseHasSuccessCode = true
			break
		}
	}

	if !responseHasSuccessCode {
		msg := "received status code %d (%s), expected any value out of %v"
		return nil, fmt.Errorf(msg, resp.StatusCode, http.StatusText(resp.StatusCode), h.SuccessStatusCodes)
	}

	return io.ReadAll(resp.Body)
}

func (h *HTTP) setRequestAuth(request *http.Request) error {
	username, err := h.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %v", err)
	}
	defer config.ReleaseSecret(username)
	password, err := h.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %v", err)
	}
	defer config.ReleaseSecret(password)
	if len(username) != 0 || len(password) != 0 {
		request.SetBasicAuth(string(username), string(password))
	}
	return nil
}

// Register the secret-store on load.
func init() {
	secretstores.Add("http", func(id string) telegraf.SecretStore {
		return &HTTP{}
	})
}
