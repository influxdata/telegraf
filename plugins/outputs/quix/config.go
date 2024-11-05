// config.go
package quix

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BrokerConfig holds the broker configuration fields from the Quix API response
type BrokerConfig struct {
	BootstrapServers string `json:"bootstrap.servers"`
	SaslMechanism    string `json:"sasl.mechanism"`
	SaslUsername     string `json:"sasl.username"`
	SaslPassword     string `json:"sasl.password"`
	SecurityProtocol string `json:"security.protocol"`
	SSLCertBase64    string `json:"ssl.ca.cert"`
}

// fetchBrokerConfig retrieves broker configuration from the Quix API
func (q *Quix) fetchBrokerConfig() (*BrokerConfig, []byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/workspaces/%s/broker/librdkafka", q.APIURL, q.Workspace), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.AuthToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var config BrokerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, nil, err
	}

	decodedCert, err := base64.StdEncoding.DecodeString(config.SSLCertBase64)
	if err != nil {
		return nil, nil, err
	}

	q.Log.Infof("Fetched broker configuration from Quix API.")
	return &config, decodedCert, nil
}

// parseTimestampUnits parses the timestamp units for metrics serialization
func parseTimestampUnits(units string) (time.Duration, error) {
	return time.ParseDuration(units)
}
