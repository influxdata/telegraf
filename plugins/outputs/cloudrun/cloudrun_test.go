package cloudrun

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
)

// Default config used by Tests
func defaultCloudrun() *CloudRun {
	return &CloudRun{
		Timeout:      config.Duration(defaultClientTimeout),
		Method:       defaultMethod,
		ConvertPaths: true,
		JSONSecret:   "../../common/gcp/testdata/test_key_file.json",
	}
}

// Function to generate fake access token
func generateFakeAccessToken(saKeyfile string) (string, error) {
	now := time.Now().Unix()

	claims := &jws.ClaimSet{
		Aud: "https://endpoint.app/",
		Exp: now + 3600,
		Iat: now,
		Iss: "https://accounts.google.com",
		Sub: "8675309",
	}

	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
		KeyID:     "8675309",
	}

	sa, err := ioutil.ReadFile(saKeyfile)
	if err != nil {
		return "", fmt.Errorf("could not read service account file: %v", err)
	}

	conf, err := google.JWTConfigFromJSON(sa)
	if err != nil {
		return "", fmt.Errorf("could not parse service account JSON: %v", err)
	}

	block, _ := pem.Decode(conf.PrivateKey)

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("private key parse error: %v", err)
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	// Sign the JWT with the service account's private key.
	if !ok {
		return "", errors.New("private key failed rsa.PrivateKey type assertion")
	}

	return jws.Encode(jwsHeader, claims, rsaKey)
}

// TODO: This is may only be useful as a reference
func TestCloudRun_Close(t *testing.T) {
	cr := defaultCloudrun()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "close success", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cr.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: This is may only be useful as a reference
func TestCloudRun_Connect(t *testing.T) {
	cr := defaultCloudrun()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "connect success", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cr.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloudRun_Write(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		// TODO: For stronger testing, check the mock metrics received here to ensure accuracy
	}))
	defer fakeServer.Close()

	cr := defaultCloudrun()
	cr.serializer = influx.NewSerializer()
	cr.URL = fakeServer.URL

	fakeAccessToken, err := generateFakeAccessToken(cr.JSONSecret)
	require.NoError(t, err)
	cr.accessToken = fakeAccessToken

	err = cr.Connect()
	require.NoError(t, err)

	tests := []struct {
		name    string
		metrics []telegraf.Metric
		wantErr bool
	}{
		// TODO: Write failure test cases
		{
			name:    "write success",
			metrics: testutil.MockMetrics(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cr.Write(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
