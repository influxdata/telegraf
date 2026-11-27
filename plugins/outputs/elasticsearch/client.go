package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	elasticsearch5 "github.com/elastic/go-elasticsearch/v5"
	esapi5 "github.com/elastic/go-elasticsearch/v5/esapi"

	"github.com/influxdata/telegraf"
)

// roundTripper applies per-request authentication and headers through the configured HTTP client.
type roundTripper struct {
	setAuth            func(*http.Request) error
	headers            http.Header
	httpClient         *http.Client
	bypassProductCheck bool
}

func (a *Elasticsearch) newRoundTripper() roundTripper {
	return roundTripper{
		setAuth:    a.setAuth,
		headers:    a.authorizationHeaders(),
		httpClient: a.httpClient,
	}
}

func (t roundTripper) RoundTrip(original *http.Request) (*http.Response, error) {
	req := original.Clone(original.Context())
	// Per-request secret authentication must remain in the transport because
	// v6/v7 discovery requests bypass the client methods.
	if err := t.setAuth(req); err != nil {
		if req.Body != nil {
			req.Body.Close()
		}
		return nil, err
	}
	for name, values := range t.headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	res, err := t.httpClient.Do(req)
	if err == nil && t.bypassProductCheck {
		// Preserve support for OpenSearch in compatibility mode. The official client's
		// product check validates identity, not API compatibility.
		res.Header.Set("X-Elastic-Product", "Elasticsearch")
	}
	return res, err
}

func (a *Elasticsearch) setAuth(req *http.Request) error {
	if err := a.setAuthBasic(req); err != nil {
		return err
	}
	return a.setAuthBearer(req)
}

func (a *Elasticsearch) setAuthBasic(req *http.Request) error {
	if !a.Username.Empty() && !a.Password.Empty() {
		username, err := a.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := a.Password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		req.SetBasicAuth(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	}
	return nil
}

func (a *Elasticsearch) setAuthBearer(req *http.Request) error {
	if !a.AuthBearerToken.Empty() {
		token, err := a.AuthBearerToken.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		req.Header.Add("Authorization", "Bearer "+token.String())
		token.Destroy()
	}
	return nil
}

// clientHeaders returns configured headers without Authorization for the SDK clients.
func (a *Elasticsearch) clientHeaders() http.Header {
	// TODO: Revisit the existing bearer-token override of configured headers.
	if !a.AuthBearerToken.Empty() {
		return nil
	}
	headers := a.httpHeaders.Clone()
	headers.Del("Authorization")
	return headers
}

// authorizationHeaders returns configured Authorization values unless bearer authentication
// is set. They are kept out of the SDK client headers because SetBasicAuth in RoundTrip
// would replace them; RoundTrip adds them after Basic authentication instead.
func (a *Elasticsearch) authorizationHeaders() http.Header {
	// See the bearer-token override TODO in clientHeaders.
	if !a.AuthBearerToken.Empty() {
		return nil
	}
	return http.Header{"Authorization": a.httpHeaders.Values("Authorization")}
}

func (a *Elasticsearch) probeVersion(ctx context.Context) (string, uint64, error) {
	transport := a.newRoundTripper()
	probe, err := elasticsearch5.NewClient(elasticsearch5.Config{
		Addresses: a.URLs,
		Transport: transport,
	})
	if err != nil {
		return "", 0, fmt.Errorf("creating ES client failed: %w", err)
	}
	request := esapi5.InfoRequest{Header: a.clientHeaders()}
	res, err := request.Do(ctx, probe)
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

// startDiscovery runs periodic discovery for v6 and v7 until canceled.
func startDiscovery(ctx context.Context, discover func() error, log telegraf.Logger) {
	ticker := time.NewTicker(15 * time.Minute)
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
