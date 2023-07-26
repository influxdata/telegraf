//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package http

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type HTTP struct {
	URLs            []string `toml:"urls"`
	Method          string   `toml:"method"`
	Body            string   `toml:"body"`
	ContentEncoding string   `toml:"content_encoding"`

	// Basic authentication
	Username config.Secret `toml:"username"`
	Password config.Secret `toml:"password"`

	// Bearer authentication
	BearerToken string        `toml:"bearer_token" deprecated:"1.28.0;use 'token_file' instead"`
	Token       config.Secret `toml:"token"`
	TokenFile   string        `toml:"token_file"`

	Headers            map[string]string `toml:"headers"`
	SuccessStatusCodes []int             `toml:"success_status_codes"`
	Log                telegraf.Logger   `toml:"-"`

	httpconfig.HTTPClientConfig

	client     *http.Client
	parserFunc telegraf.ParserFunc
}

func (*HTTP) SampleConfig() string {
	return sampleConfig
}

func (h *HTTP) Init() error {
	// For backward compatibility
	if h.TokenFile != "" && h.BearerToken != "" && h.TokenFile != h.BearerToken {
		return fmt.Errorf("conflicting settings for 'bearer_token' and 'token_file'")
	} else if h.TokenFile == "" && h.BearerToken != "" {
		h.TokenFile = h.BearerToken
	}

	// We cannot use multiple sources for tokens
	if h.TokenFile != "" && !h.Token.Empty() {
		return fmt.Errorf("either use 'token_file' or 'token' not both")
	}

	// Create the client
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
	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (h *HTTP) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, u := range h.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := h.gatherURL(acc, url); err != nil {
				acc.AddError(fmt.Errorf("[url=%s]: %w", url, err))
			}
		}(u)
	}

	wg.Wait()

	return nil
}

// SetParserFunc takes the data_format from the config and finds the right parser for that format
func (h *HTTP) SetParserFunc(fn telegraf.ParserFunc) {
	h.parserFunc = fn
}

// Gathers data from a particular URL
// Parameters:
//
//	acc    : The telegraf Accumulator to use
//	url    : endpoint to send request to
//
// Returns:
//
//	error: Any error that may have occurred
func (h *HTTP) gatherURL(
	acc telegraf.Accumulator,
	url string,
) error {
	body := makeRequestBodyReader(h.ContentEncoding, h.Body)
	request, err := http.NewRequest(h.Method, url, body)
	if err != nil {
		return err
	}

	if !h.Token.Empty() {
		token, err := h.Token.Get()
		if err != nil {
			return err
		}
		bearer := "Bearer " + strings.TrimSpace(string(token))
		request.Header.Set("Authorization", bearer)
	} else if h.TokenFile != "" {
		token, err := os.ReadFile(h.BearerToken)
		if err != nil {
			return err
		}
		bearer := "Bearer " + strings.Trim(string(token), "\n")
		request.Header.Set("Authorization", bearer)
	}

	if h.ContentEncoding == "gzip" {
		request.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			request.Host = v
		} else {
			request.Header.Add(k, v)
		}
	}

	if err := h.setRequestAuth(request); err != nil {
		return err
	}

	resp, err := h.client.Do(request)
	if err != nil {
		return err
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
		return fmt.Errorf("received status code %d (%s), expected any value out of %v",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			h.SuccessStatusCodes)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body failed: %w", err)
	}

	// Instantiate a new parser for the new data to avoid trouble with stateful parsers
	parser, err := h.parserFunc()
	if err != nil {
		return fmt.Errorf("instantiating parser failed: %w", err)
	}
	metrics, err := parser.Parse(b)
	if err != nil {
		return fmt.Errorf("parsing metrics failed: %w", err)
	}

	for _, metric := range metrics {
		if !metric.HasTag("url") {
			metric.AddTag("url", url)
		}
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}

	return nil
}

func (h *HTTP) setRequestAuth(request *http.Request) error {
	if h.Username.Empty() && h.Password.Empty() {
		return nil
	}

	username, err := h.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	defer config.ReleaseSecret(username)

	password, err := h.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	defer config.ReleaseSecret(password)

	request.SetBasicAuth(string(username), string(password))

	return nil
}

func makeRequestBodyReader(contentEncoding, body string) io.Reader {
	if body == "" {
		return nil
	}

	var reader io.Reader = strings.NewReader(body)
	if contentEncoding == "gzip" {
		return internal.CompressWithGzip(reader)
	}

	return reader
}

func init() {
	inputs.Add("http", func() telegraf.Input {
		return &HTTP{
			Method: "GET",
		}
	})
}
