package elasticsearch

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

// roundTripper holds client settings and applies per-request authentication, headers, and compression through the configured HTTP client.
type roundTripper struct {
	urls               []string
	basicAuth          bool
	userinfo           *url.Userinfo
	username           config.Secret
	password           config.Secret
	bearerToken        config.Secret
	headers            http.Header
	enableGzip         bool
	enableSniffer      bool
	discoveryInterval  time.Duration
	httpClient         *http.Client
	log                telegraf.Logger
	bypassProductCheck bool
}

func (t *roundTripper) configureAuth(username, password, bearerToken config.Secret) error {
	if !username.Empty() && !password.Empty() {
		t.basicAuth = true
		t.username, t.password = username, password
	}
	if !t.basicAuth {
		for _, address := range t.urls {
			u, err := url.Parse(address)
			if err != nil {
				return fmt.Errorf("parsing URL %q failed: %w", address, err)
			}
			if u.User != nil {
				t.userinfo = u.User
				break
			}
		}
	}
	t.bearerToken = bearerToken
	return nil
}

func (t roundTripper) probeVersion(ctx context.Context) (string, uint64, error) {
	probe, err := elasticsearch5.NewClient(elasticsearch5.Config{
		Addresses: t.urls,
		Transport: t,
	})
	if err != nil {
		return "", 0, fmt.Errorf("creating ES client failed: %w", err)
	}
	res, err := probe.Info(probe.Info.WithContext(ctx))
	if err != nil {
		return "", 0, fmt.Errorf("checking ES version failed: %w", err)
	}
	defer res.Body.Close()
	if err := checkForError(res.StatusCode, res.Body); err != nil {
		return "", 0, fmt.Errorf("checking ES version failed: %w", err)
	}
	var info serverInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", 0, fmt.Errorf("checking ES version failed: %w", err)
	}
	version, err := semver.NewVersion(info.Version.Number)
	if err != nil {
		return "", 0, fmt.Errorf("parsing server version %q failed: %w", info.Version.Number, err)
	}
	return info.Version.Number, version.Major(), nil
}

func encodeBulk(documents []document, docType string) (*bytes.Buffer, error) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	for _, doc := range documents {
		metadata := bulkMetadata{Index: doc.index, ID: doc.id, Type: docType, Pipeline: doc.pipeline}
		if err := encoder.Encode(map[string]bulkMetadata{doc.opType: metadata}); err != nil {
			return nil, fmt.Errorf("encoding bulk metadata failed: %w", err)
		}
		if err := encoder.Encode(doc.body); err != nil {
			return nil, fmt.Errorf("encoding document body failed: %w", err)
		}
	}
	return &body, nil
}

func checkBulkResponse(status int, body io.Reader, log telegraf.Logger) error {
	if err := checkForError(status, body); err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	var response bulkResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return fmt.Errorf("decoding message body failed: %w", err)
	}
	if !response.Errors {
		return nil
	}
	var failed int
	for _, item := range response.Items {
		for _, result := range item {
			if result.Status >= 200 && result.Status < 300 {
				continue
			}
			if failed == 0 && result.Error != nil {
				// Preserve the legacy failure-list index: only the first failure was logged, at index 0.
				log.Errorf(
					"Elasticsearch indexing failure, id: %d, status: %d, error: %s, caused by: %s, %s",
					0, result.Status, result.Error.Reason, result.Error.CausedBy["reason"], result.Error.CausedBy["type"],
				)
			}
			failed++
		}
	}
	return fmt.Errorf("elasticsearch failed to index %d metrics", failed)
}

func checkForError(status int, body io.Reader) error {
	if status >= 200 && status < 300 {
		return nil
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("received error %d (%s)", status, http.StatusText(status))
	var response apiErrorResponse
	if err := json.Unmarshal(data, &response); err == nil && response.Error.Reason != "" {
		message += ": " + response.Error.Reason
		if response.Error.Type != "" {
			message += " [type=" + response.Error.Type + "]"
		}
	} else if text := strings.TrimSpace(string(data)); text != "" {
		message += ": " + text
	}
	return errors.New(message)
}

func (t roundTripper) RoundTrip(original *http.Request) (*http.Response, error) {
	req := original.Clone(original.Context())
	if err := t.setAuth(req); err != nil {
		if req.Body != nil {
			req.Body.Close()
		}
		return nil, err
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if strings.HasSuffix(req.URL.Path, "/_bulk") {
		// The v5/v6/v7 APIs default to JSON even for newline-delimited bulk bodies.
		req.Header.Set("Content-Type", "application/x-ndjson")
	} else if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	// Preserve the existing bearer-token override of configured headers.
	if t.bearerToken.Empty() {
		for name, values := range t.headers {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}
	}
	if t.enableGzip && req.Body != nil && req.Body != http.NoBody {
		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		_, err := io.Copy(writer, req.Body)
		req.Body.Close()
		if err != nil {
			writer.Close()
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		data := compressed.Bytes()
		req.Body = io.NopCloser(bytes.NewReader(data))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
		req.ContentLength = int64(compressed.Len())
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Add("Vary", "Accept-Encoding")
	}
	res, err := t.httpClient.Do(req)
	if err == nil && t.bypassProductCheck {
		// Preserve support for OpenSearch in compatibility mode. The official client's
		// product check validates identity, not API compatibility.
		res.Header.Set("X-Elastic-Product", "Elasticsearch")
	}
	return res, err
}

func (t roundTripper) setAuth(req *http.Request) error {
	if t.basicAuth {
		username, err := t.username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := t.password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		req.SetBasicAuth(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	} else if t.userinfo != nil {
		password, _ := t.userinfo.Password()
		req.SetBasicAuth(t.userinfo.Username(), password)
	}
	if !t.bearerToken.Empty() {
		token, err := t.bearerToken.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		req.Header.Add("Authorization", "Bearer "+token.String())
		token.Destroy()
	}
	return nil
}

// preserveConnectionPool keeps the previous v6/v7 SDK pool when discovery finds no eligible nodes.
func preserveConnectionPool[Connection, Selector, Pool any](newPool func([]Connection, Selector) (Pool, error)) func([]Connection, Selector) Pool {
	var pool Pool
	return func(conns []Connection, selector Selector) Pool {
		if len(conns) > 0 {
			//nolint:errcheck // The v6/v7 SDK pool constructors always return a nil error.
			pool, _ = newPool(conns, selector)
		}
		return pool
	}
}

// startDiscovery runs node discovery immediately and then at the configured interval.
func startDiscovery(ctx context.Context, interval time.Duration, discover func() error, log telegraf.Logger) {
	if err := discover(); err != nil && ctx.Err() == nil {
		log.Errorf("Discovering Elasticsearch nodes failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := discover(); err != nil && ctx.Err() == nil {
				log.Errorf("Discovering Elasticsearch nodes failed: %v", err)
			}
		}
	}
}
