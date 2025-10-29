package http

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"net"
	"syscall"
	"runtime"

	"github.com/influxdata/telegraf"
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

  ## Specify a source IP Address
  # source_ip = "10.1.1.1"

  ## Specify a source interface name
  # source_interface = "eth0"

  ## For status codes not in non_retryable_statuscodes list, number of times to retry
  ## before metric buffer is discarded.
  ## max_retries = 0, NEVER retry and discard the metric buffer for (<200 or >300); negates 
  ## retryable_statuscodes list.
  ## When not set or max_retries = -1, maintain default behaviour which is to continually retry.
  # max_retries = 3

  ## Optional list of status codes (<200 or >300) upon which requests should not be retried
  # non_retryable_statuscodes = [400, 500]
  
  ## Specific set of status codes (<200 or >300) to retry.  All other status codes will cause
  ## metric buffer to be discarded.
  # retryable_statuscodes = [402, 503, 504]
`

const (
	defaultContentType    = "text/plain; charset=utf-8"
	defaultMethod         = http.MethodPost
	defaultUseBatchFormat = true
	defaultMaxRetries     = -1
)

type HTTP struct {
	URL             string            `toml:"url"`
	Method          string            `toml:"method"`
	Username        string            `toml:"username"`
	Password        string            `toml:"password"`
	Headers         map[string]string `toml:"headers"`
	ContentEncoding string            `toml:"content_encoding"`
	UseBatchFormat  bool              `toml:"use_batch_format"`
	NonRetryableStatusCodes []int     `toml:"non_retryable_statuscodes"` // Port 1.22.0
	RetryableStatusCodes []int        `toml:"retryable_statuscodes"`     // EXTR Specific
	MaxRetries         int32          `toml:"max_retries"`               // EXTR Specific
	SourceIP           string         `toml:"source_ip"`                 // EXTR Specific
	Interface          string         `toml:"source_interface"`          // EXTR Specific
	httpconfig.HTTPClientConfig
	Log telegraf.Logger               `toml:"-"`

	FailCount   int32
	client     *http.Client
	serializer serializers.Serializer
}

func (h *HTTP) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

func (h *HTTP) Connect() error {
		
	dialer := net.Dialer{
		Timeout:   time.Duration(h.Timeout),
		KeepAlive: 30 * time.Second,
	}

	if h.SourceIP != "" {
	 	localAddr, err := net.ResolveIPAddr("ip", h.SourceIP)
	 	if err != nil {
	 		return fmt.Errorf("unable to resolve source_ip %s: %v", h.SourceIP, err)
	 	}
	 	dialer.LocalAddr = &net.TCPAddr{
	 		IP: localAddr.IP,
	 	}
	}

	if h.Interface != "" {
		if runtime.GOOS == "linux" {
			dialer.Control = func(network, address string, c syscall.RawConn) error {
				var ctrlErr error
				err := c.Control(func(fd uintptr) {
					ctrlErr = syscall.SetsockoptString(
						int(fd),
						syscall.SOL_SOCKET,
						syscall.SO_BINDTODEVICE,
						h.Interface,
					)
				})
				if err != nil {
					return err
				}
				return ctrlErr
			}
		} else {
			h.Log.Warnf("source_interface option is not supported on %s", runtime.GOOS)
		}
	}

	if h.Method == "" {
		h.Method = http.MethodPost
	}
	h.Method = strings.ToUpper(h.Method)
	if h.Method != http.MethodPost && h.Method != http.MethodPut {
		return fmt.Errorf("invalid HTTP method %s for URL %s", h.Method, h.URL)
	}

	ctx := context.Background()
	client, err := h.HTTPClientConfig.CreateClient(ctx, h.Log)
	if err != nil {
		return err
	}

	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.DialContext = dialer.DialContext
	} else {
		h.Log.Errorf("Unable to set custom dialer: client transport is not *http.Transport (type: %T). Source IP and interface binding will be ignored.", client.Transport)
	}

	h.client = client

	return nil
}

func (h *HTTP) Close() error {
	return nil
}

func (h *HTTP) Description() string {
	return "A plugin that can transmit metrics over HTTP"
}

func (h *HTTP) SampleConfig() string {
	return sampleConfig
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

func (h *HTTP) checkRetriesFailed() bool {

	if h.MaxRetries == 0 {
		// Drop, never retry
		return true
	} else if (h.MaxRetries < 0) {
		// Always retry	
		return false
	} else {
		// Retry up to maxRetries
		atomic.AddInt32(&h.FailCount, 1)

		if atomic.LoadInt32(&h.FailCount) > h.MaxRetries {
			h.Log.Errorf("%s FailCount %d > MaxRetries %d. Metrics are dropped.",
					h.URL, h.FailCount, h.MaxRetries )
			atomic.StoreInt32(&h.FailCount, 0)

			// Drop, max retry hit.
			return true
		} else {
			// Retry
			return false
		}
	}
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
	
	req, err := http.NewRequest(h.Method, h.URL, reqBodyBuffer)
	if err != nil {
		h.Log.Errorf("Failed http.NewRequest() return err:%v", err)
		return err
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

		// This will fail if:
		//   - endpoint not reachable or not listening on port
		//   - server not responding withing timeout
		//   - URL schema incorrect

		if h.checkRetriesFailed() {
			h.Log.Errorf("%s", err)
			return nil
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		if h.MaxRetries == 0 {
			// Drop, never retry
			h.Log.Errorf("%s HTTP Response:%v. MaxRetries:%v. Metrics are dropped.", h.URL, resp.StatusCode, h.MaxRetries)
			return nil
		}

		// Port from v1.22.0
		for _, nonRetryableStatusCode := range h.NonRetryableStatusCodes {
			if resp.StatusCode == nonRetryableStatusCode {
				h.Log.Errorf("%s HTTP Response:%v. non-retryable status. Metrics are dropped.", h.URL, resp.StatusCode)
				if h.MaxRetries > 0 {
					atomic.StoreInt32(&h.FailCount, 0)
				}
				return nil
			}
		}

		// EXTR - Only retry if specific retry list is provided HTTP response code is present in list
		if len(h.RetryableStatusCodes) > 0 {
			
			var retryableCodeFound bool = false

			for _, retryableStatusCode := range h.RetryableStatusCodes {
				if resp.StatusCode == retryableStatusCode {
					retryableCodeFound = true
					break
				}
			}

			if retryableCodeFound == false {
				h.Log.Errorf("%s HTTP Response:%v not found in retryable code list. Metrics are dropped", h.URL, resp.StatusCode)
				if h.MaxRetries > 0 {
					atomic.StoreInt32(&h.FailCount, 0)
				}
				return nil
			}
		}

		errorLine := ""
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxErrMsgLen))
		if scanner.Scan() {
			errorLine = scanner.Text()
		}
		
		if h.checkRetriesFailed() {
			return nil
		}

		return fmt.Errorf("when writing to [%s] received status code: %d. body: %s", h.URL, resp.StatusCode, errorLine)
	}

	// Don't really care about response. Just want read and drain the body so HTTP connection is reused
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("when writing to [%s] received error: %v", h.URL, err)
	}

	if h.MaxRetries > 0 {
		atomic.StoreInt32(&h.FailCount, 0)
	}
	return nil
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &HTTP{
			Method:         defaultMethod,
			URL:            defaultURL,
			UseBatchFormat: defaultUseBatchFormat,
			MaxRetries:     defaultMaxRetries,
		}
	})
}
