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

	ctx := context.Background()
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
