//go:generate ../../../tools/readme_config_includer/generator
package http

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/blues/jsonata-go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

const defaultIdleConnTimeoutMinutes = 5

type HTTP struct {
	URL                string            `toml:"url"`
	Headers            map[string]string `toml:"headers"`
	Username           config.Secret     `toml:"username"`
	Password           config.Secret     `toml:"password"`
	Token              config.Secret     `toml:"token"`
	SuccessStatusCodes []int             `toml:"success_status_codes"`
	Transformation     string            `toml:"transformation"`
	Log                telegraf.Logger   `toml:"-"`
	common_http.HTTPClientConfig
	DecryptionConfig

	client      *http.Client
	transformer *jsonata.Expr
	cache       map[string]string
	decrypter   Decrypter
}

func (*HTTP) SampleConfig() string {
	return sampleConfig
}

func (h *HTTP) Init() error {
	ctx := context.Background()

	// Prevent idle connections from hanging around forever on telegraf reload
	if h.HTTPClientConfig.IdleConnTimeout == 0 {
		h.HTTPClientConfig.IdleConnTimeout = config.Duration(defaultIdleConnTimeoutMinutes * time.Minute)
	}

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
	h.decrypter, err = h.DecryptionConfig.CreateDecrypter()
	if err != nil {
		return fmt.Errorf("creating decryptor failed: %w", err)
	}

	return nil
}

// Get searches for the given key and return the secret
func (h *HTTP) Get(key string) ([]byte, error) {
	v, found := h.cache[key]
	if !found {
		return nil, errors.New("not found")
	}

	if h.decrypter != nil {
		// We got binary data delivered in a string, so try to
		// decode it assuming base64-encoding.
		buf, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("base64 decoding failed: %w", err)
		}
		return h.decrypter.Decrypt(buf)
	}

	return []byte(v), nil
}

// Set sets the given secret for the given key
func (*HTTP) Set(_, _ string) error {
	return errors.New("setting secrets not supported")
}

// List lists all known secret keys
func (h *HTTP) List() ([]string, error) {
	keys := make([]string, 0, len(h.cache))
	for k := range h.cache {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetResolver returns a function to resolve the given key.
func (h *HTTP) GetResolver(key string) (telegraf.ResolveFunc, error) {
	// Download and parse the credentials
	if err := h.download(); err != nil {
		return nil, err
	}

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
		return fmt.Errorf("reading body failed: %w", err)
	}

	// Transform the data to the expected form if given
	if h.transformer != nil {
		out, err := h.transformer.EvalBytes(data)
		if err != nil {
			return fmt.Errorf("transforming data failed: %w", err)
		}
		data = out
	}

	// Extract the data from the resulting data
	if err := json.Unmarshal(data, &h.cache); err != nil {
		var terr *json.UnmarshalTypeError
		if errors.As(err, &terr) {
			return fmt.Errorf("%w; maybe missing or wrong data transformation", err)
		}
		return err
	}

	return nil
}

func (h *HTTP) query() ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, h.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	for k, v := range h.Headers {
		if strings.EqualFold(k, "host") {
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

	// Try to wipe the bearer token if any
	request.SetBasicAuth("---", "---")
	request.Header.Set("Authorization", "---")

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
	if !h.Username.Empty() && !h.Password.Empty() {
		username, err := h.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		defer username.Destroy()
		password, err := h.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		defer password.Destroy()
		request.SetBasicAuth(username.String(), password.String())
	}

	if !h.Token.Empty() {
		token, err := h.Token.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		defer token.Destroy()
		bearer := "Bearer " + strings.TrimSpace(token.String())
		request.Header.Set("Authorization", bearer)
	}

	return nil
}

// Register the secret-store on load.
func init() {
	secretstores.Add("http", func(string) telegraf.SecretStore {
		return &HTTP{}
	})
}
