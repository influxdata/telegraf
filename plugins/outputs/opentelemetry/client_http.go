package opentelemetry

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
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
}

func (h *httpClient) Connect(cfg *clientConfig) error {
	h.httpClient = &http.Client{}
	h.url = cfg.ServiceAddress
	h.encodingType = cfg.Encoding
	h.compress = cfg.Compression

	tlsConfig, err := cfg.TLSConfig.TLSConfig()
	if err != nil {
		return err
	}

	// For coralogix, we enforce HTTP connection with TLS
	if tlsConfig != nil && cfg.CoralogixConfig != nil {
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

	httpRequest.Header.Set("Content-Type", encoding)
	httpRequest.Header.Set("User-Agent", userAgent)
	if h.compress != "" && h.compress != "none" {
		httpRequest.Header.Set("Content-Encoding", "gzip")
	}

	httpResponse, err := h.httpClient.Do(httpRequest)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}
	defer func() {
		//nolint:errcheck // cannot fail with io.Discard
		// 64KB is a not specific limit. But, opentelemetry-collector also uses 64KB for safety.
		io.CopyN(io.Discard, httpResponse.Body, 64*1024)
		_ = httpResponse.Body.Close()
	}()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return pmetricotlp.ExportResponse{}, fmt.Errorf("received unexpected status: %s (%d)", http.StatusText(httpResponse.StatusCode), httpResponse.StatusCode)
	}

	responseBytes, err := readResponseBody(httpResponse)
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

// ref. https://github.com/open-telemetry/opentelemetry-collector/blob/7258150320ae4c3b489aa58bd2939ba358b23ae1/exporter/otlphttpexporter/otlp.go#L271
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp.ContentLength == 0 {
		return nil, nil
	}

	maxRead := resp.ContentLength

	// if maxRead == -1, the ContentLength header has not been sent, so read up to
	// the maximum permitted body size. If it is larger than the permitted body
	// size, still try to read from the body in case the value is an error. If the
	// body is larger than the maximum size, proto unmarshaling will likely fail.
	// 64KB is a not specific limit. But, opentelemetry-collector also uses 64KB for safety.
	if maxRead == -1 || maxRead > 64*1024 {
		maxRead = 64 * 1024
	}
	protoBytes := make([]byte, maxRead)
	n, err := io.ReadFull(resp.Body, protoBytes)

	// No bytes read and an EOF error indicates there is no body to read.
	if n == 0 && (err == nil || errors.Is(err, io.EOF)) {
		return nil, nil
	}

	// io.ReadFull will return io.ErrorUnexpectedEOF if the Content-Length header
	// wasn't set, since we will try to read past the length of the body. If this
	// is the case, the body will still have the full message in it, so we want to
	// ignore the error and parse the message.
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}

	return protoBytes[:n], nil
}
