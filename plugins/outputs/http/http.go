package http

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	awsV2 "github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/internal"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const (
	maxErrMsgLen = 1024
	defaultURL   = "http://127.0.0.1:8080/telegraf"
)

var sampleConfig = `
  ## URL is the address to send metrics to
  url = "http://127.0.0.1:8080/telegraf"

  ## Timeout for HTTP message
  # timeout = "5s"

  ## HTTP method, one of: "POST" or "PUT"
  # method = "POST"

  ## HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## OAuth2 Client Credentials Grant
  # client_id = "clientid"
  # client_secret = "secret"
  # token_url = "https://indentityprovider/oauth2/v1/token"
  # scopes = ["urn:opc:idm:__myscopes__"]

  # OAuth2 Authorization Code Grant
  # credentials_file = "/etc/telegraf/keyfile.json"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional Cookie authentication
  # cookie_auth_url = "https://localhost/authMe"
  # cookie_auth_method = "POST"
  # cookie_auth_username = "username"
  # cookie_auth_password = "pa$$word"
  # cookie_auth_headers = '{"Content-Type": "application/json", "X-MY-HEADER":"hello"}'
  # cookie_auth_body = '{"username": "user", "password": "pa$$word", "authenticate": "me"}'
  ## cookie_auth_renewal not set or set to "0" will auth once and never renew the cookie
  # cookie_auth_renewal = "5m"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"

  ## Use batch serialization format (default) instead of line based format.
  ## Batch format is more efficient and should be used unless line based
  ## format is really needed.
  # use_batch_format = true

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## Additional HTTP headers
  # [outputs.http.headers]
  #   # Should be set manually to "application/json" for json data_format
  #   Content-Type = "text/plain; charset=utf-8"

  ## Idle (keep-alive) connection timeout.
  ## Maximum amount of time before idle connection is closed.
  ## Zero means no limit.
  # idle_conn_timeout = 0

  ## Amazon Region
  #region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #web_identity_token_file = ""
  #role_session_name = ""
  #profile = ""
  #shared_credential_file = ""
`

const (
	defaultContentType    = "text/plain; charset=utf-8"
	defaultMethod         = http.MethodPost
	defaultUseBatchFormat = true
)

type HTTP struct {
	URL                     string            `toml:"url"`
	Method                  string            `toml:"method"`
	Username                string            `toml:"username"`
	Password                string            `toml:"password"`
	Headers                 map[string]string `toml:"headers"`
	ContentEncoding         string            `toml:"content_encoding"`
	UseBatchFormat          bool              `toml:"use_batch_format"`
	AwsService              string            `toml:"aws_service"`
	NonRetryableStatusCodes []int             `toml:"non_retryable_statuscodes"`
	httpconfig.HTTPClientConfig
	Log telegraf.Logger `toml:"-"`

	client     *http.Client
	serializer serializers.Serializer

	awsCfg *awsV2.Config
	internalaws.CredentialConfig
}

func (h *HTTP) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

func (h *HTTP) Connect() error {
	if h.AwsService != "" {
		cfg, err := h.CredentialConfig.Credentials()
		if err == nil {
			h.awsCfg = &cfg
		}
	}

	if h.Method == "" {
		h.Method = http.MethodPost
	}
	h.Method = strings.ToUpper(h.Method)
	if h.Method != http.MethodPost && h.Method != http.MethodPut {
		return fmt.Errorf("invalid method [%s] %s", h.URL, h.Method)
	}

	// ctx := context.Background()
	// ctx := context.WithValue(context.Background(), oauth2.HTTPClient, h.HTTPClientConfig)
	// TODO: use h.TokenURL, or h.URL? passing it in to ctx...
	ctx := context.WithValue(context.Background(), "url", h.TokenURL)

	client, err := h.HTTPClientConfig.CreateClient(ctx, h.Log)
	if err != nil {
		return err
	}

	h.client = client

	return nil
}

func (h *HTTP) Close() error {
	return nil
}

func (h *HTTP) Write(metrics []telegraf.Metric) error {
	var reqBody []byte

	if h.UseBatchFormat {
		var err error
		reqBody, err = h.serializer.SerializeBatch(metrics)
		if err != nil {
			return err
		}

		return h.writeMetric(reqBody)
	}

	for _, metric := range metrics {
		var err error
		reqBody, err = h.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		if err := h.writeMetric(reqBody); err != nil {
			return err
		}
	}
	return nil
}

func (h *HTTP) writeMetric(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)

	var err error
	if h.ContentEncoding == "gzip" {
		rc, err := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

	var payloadHash *string
	if h.awsCfg != nil {
		// We need a local copy of the full buffer, the signature scheme requires a sha256 of the request body.
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, reqBodyBuffer)
		if err != nil {
			return err
		}

		sum := sha256.Sum256(buf.Bytes())
		reqBodyBuffer = buf

		// sha256 is hex encoded
		hash := fmt.Sprintf("%x", sum)
		payloadHash = &hash
	}

	req, err := http.NewRequest(h.Method, h.URL, reqBodyBuffer)
	if err != nil {
		return err
	}

	if h.awsCfg != nil {
		signer := v4.NewSigner()
		ctx := context.Background()

		credentials, err := h.awsCfg.Credentials.Retrieve(ctx)
		if err != nil {
			return err
		}

		err = signer.SignHTTP(ctx, credentials, req, *payloadHash, h.AwsService, h.Region, time.Now().UTC())
		if err != nil {
			return err
		}
	}

	if h.Username != "" || h.Password != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}

	// Authorization Code Grant
	if h.CredentialsFile != "" {
		claims := jwtGo.RegisteredClaims{}
		_, err := jwtGo.ParseWithClaims(h.HTTPClientConfig.AccessToken, &claims, func(token *jwtGo.Token) (interface{}, error) {
			return nil, nil
		})
		if err != nil {
			return err
		}

		// Request new token if expired
		if !claims.VerifyExpiresAt(time.Now(), true) {
			// token is expired
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Duration(h.Timeout))
			defer cancel()

			err = h.OAuth2Config.GetAccessToken(ctx, h.URL)
			if err != nil {
				return err
			}
		}

		bearerToken := "Bearer " + h.HTTPClientConfig.AccessToken
		req.Header.Set("Authorization", bearerToken)
		req.Header.Set("User-Agent", internal.ProductToken())
		req.Header.Set("Accept", "application/json")
	}

	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Set("Content-Type", defaultContentType)
	if h.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}
	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		for _, nonRetryableStatusCode := range h.NonRetryableStatusCodes {
			if resp.StatusCode == nonRetryableStatusCode {
				h.Log.Errorf("Received non-retryable status %v. Metrics are lost.", resp.StatusCode)
				return nil
			}
		}

		errorLine := ""
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxErrMsgLen))
		if scanner.Scan() {
			errorLine = scanner.Text()
		}

		return fmt.Errorf("when writing to [%s] received status code: %d. body: %s", h.URL, resp.StatusCode, errorLine)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("when writing to [%s] received error: %v", h.URL, err)
	}

	return nil
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &HTTP{
			Method:         defaultMethod,
			URL:            defaultURL,
			UseBatchFormat: defaultUseBatchFormat,
		}
	})
}
