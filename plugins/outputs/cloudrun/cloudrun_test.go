package cloudrun

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// Default config used by Tests
func defaultCloudrun() *CloudRun {
	return &CloudRun{
		DisableConvertPaths: true,
		CredentialsFile:     "./testdata/test_key_file.json",
	}
}

// Function to generate fake access token
func (cr *CloudRun) getFakeAccessToken(ctx context.Context) error {
	data, err := ioutil.ReadFile(cr.CredentialsFile)
	if err != nil {
		return err
	}

	conf, err := google.JWTConfigFromJSON(data, cr.URL)
	if err != nil {
		return err
	}

	jwtConfig := &jwt.Config{
		Email:         conf.Email,
		TokenURL:      conf.TokenURL,
		PrivateKey:    conf.PrivateKey,
		PrivateClaims: map[string]interface{}{"target_audience": cr.URL},
	}

	token, err := jwtConfig.TokenSource(ctx).Token()
	if err != nil {
		fmt.Println("To err is human", err)
		return err
	}

	cr.accessToken = token.Extra("id_token").(string)

	return nil
}

func TestCloudRun_Write(t *testing.T) {
	cr := defaultCloudrun()
	cr.serializer = influx.NewSerializer()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cr.Timeout))
	defer cancel()

	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("r.URL.Path", r.URL.Path)
		switch r.URL.Path {
		case "/":
			fmt.Println("Hello root")
		case "/token":
			w.WriteHeader(http.StatusAccepted)
			_, err := w.Write([]byte(`{"id_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg"}`))
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
			// t.Fatalf("unexpected path: " + r.URL.Path)
		}
	}))
	// Matches the token_uri field in ./testdata/test_key_file.json
	l, _ := net.Listen("tcp", "localhost:58025")
	fakeServer.Listener = l
	fakeServer.Start()
	defer fakeServer.Close()

	cr.URL = fakeServer.URL

	cr.getFakeAccessToken(ctx)

	err := cr.Connect()
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
			fmt.Println("Writing metrics")
			if err := cr.Write(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	err = cr.Close()
	require.NoError(t, err)
}
