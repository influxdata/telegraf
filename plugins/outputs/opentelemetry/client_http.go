package opentelemetry

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"

	"github.com/influxdata/telegraf/internal"
)

type httpClient struct {
	httpClient   *http.Client
	url          string
	encodingType string
	compress     string
	headers      map[string]string
}

func (h *httpClient) Connect(cfg *clientConfig) error {
	h.httpClient = &http.Client{}
	h.url = cfg.ServiceAddress
	h.encodingType = cfg.Encoding
	h.compress = cfg.Compression
	h.headers = cfg.Headers

	tlsConfig, err := cfg.TLSConfig.TLSConfig()
	if err != nil {
		return err
	}

	// For coralogix, we enforce HTTP connection with TLS
	if tlsConfig == nil && cfg.CoralogixConfig != nil {
		tlsConfig = &tls.Config{}
	}

	if tlsConfig != nil {
		h.httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	return nil
}

func (h *httpClient) Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	var err error
	var requestBytes []byte
	var encoding string

	switch h.encodingType {
	case "protobuf":
		requestBytes, err = request.MarshalProto()
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
		encoding = "application/x-protobuf"
	case "json":
		requestBytes, err = request.MarshalJSON()
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
		encoding = "application/json"
	default:
		return pmetricotlp.ExportResponse{}, fmt.Errorf("unsupported content type '%s'", h.encodingType)
	}
	var reader io.Reader
	reader = bytes.NewReader(requestBytes)

	if h.compress != "" && h.compress != "none" {
		reader = internal.CompressWithGzip(reader)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, reader)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}
	for key, value := range h.headers {
		httpRequest.Header.Set(key, value)
	}

	httpRequest.Header.Set("Content-Type", encoding)
	httpRequest.Header.Set("User-Agent", userAgent)
	if h.compress != "" && h.compress != "none" {
		httpRequest.Header.Set("Content-Encoding", "gzip")
	}

	httpResponse, err := h.httpClient.Do(httpRequest)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return pmetricotlp.ExportResponse{}, fmt.Errorf("received unexpected status: %s (%d)",
			http.StatusText(httpResponse.StatusCode), httpResponse.StatusCode)
	}

	r := io.LimitReader(httpResponse.Body, 64*1024)
	responseBytes, err := io.ReadAll(r)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}

	exportResponse := pmetricotlp.NewExportResponse()
	switch httpResponse.Header.Get("Content-Type") {
	case "application/x-protobuf":
		err = exportResponse.UnmarshalProto(responseBytes)
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
	case "application/json":
		err = exportResponse.UnmarshalJSON(responseBytes)
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
	}

	return exportResponse, nil
}

func (h *httpClient) Close() error {
	h.httpClient.CloseIdleConnections()
	return nil
}
