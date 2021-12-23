package cloudrun

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtGo "github.com/golang-jwt/jwt/v4"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"

	// TODO: package jws deprecated
	"golang.org/x/oauth2/jws"
	"golang.org/x/oauth2/jwt"
)

// Default config used by Tests
func defaultCloudrun() *CloudRun {
	return &CloudRun{
		DisableConvertPaths: true,
		CredentialsFile:     "./testdata/test_key_file.json",
	}
}

func TestCloudRun_Write(t *testing.T) {
	cr := defaultCloudrun()
	cr.serializer = influx.NewSerializer()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cr.Timeout))
	defer cancel()

	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			// fmt.Println("Hello root")
			// fmt.Println(r.Header)
		case "/token":
			// fmt.Println("Hello token")
			// fmt.Println(r.Header)
			w.WriteHeader(http.StatusAccepted)
			_, err := w.Write([]byte(`{"id_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg"}`))
			require.NoError(t, err)
		default:
			// fmt.Println("Hello rootier root")
			// fmt.Println(r.Header)
			w.WriteHeader(http.StatusNotFound)
			// require.NoError(t, err)
			// t.Fatalf("unexpected path: " + r.URL.Path)
		}
	}))

	l, _ := net.Listen("tcp", "localhost:58025")
	fakeServer.Listener = l
	fakeServer.Start()
	defer fakeServer.Close()

	cr.URL = fakeServer.URL
	err := cr.getFakeAccessToken(ctx)
	require.NoError(t, err)

	claims := jwtGo.RegisteredClaims{}
	jwtGo.ParseWithClaims(cr.accessToken, &claims, func(token *jwtGo.Token) (interface{}, error) {
		return nil, nil
	})

	now := time.Now().Unix()
	iat := now
	exp := now + 3600
	updatedClaims := &jws.ClaimSet{
		Aud: claims.Audience[0],
		Exp: exp,
		Iat: iat,
		Iss: claims.Issuer,
		Sub: claims.Subject,
	}

	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
		KeyID:     "8675309",
	}

	data, err := ioutil.ReadFile(cr.CredentialsFile)
	require.NoError(t, err)

	conf, err := google.JWTConfigFromJSON(data)
	require.NoError(t, err)

	block, _ := pem.Decode(conf.PrivateKey)
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	require.NoError(t, err)

	rsaKey, _ := parsedKey.(*rsa.PrivateKey)

	cr.accessToken, err = jws.Encode(jwsHeader, updatedClaims, rsaKey)
	require.NoError(t, err)

	err = cr.Connect()
	require.NoError(t, err)

	fmt.Println("🤞")
	tests := []struct {
		name    string
		metrics []telegraf.Metric
		wantErr bool
	}{
		{
			name:    "write success",
			metrics: testutil.MockMetrics(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// fmt.Println("tt.metrics", tt.metrics)
			if err := cr.Write(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	err = cr.Close()
	require.NoError(t, err)
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

// { // Testing access token request
// 	name: "Successfully request access token",
// 	args: args{
// 		secret: "./testdata/test_key_file.json",
// 		url:    fakeServer.URL,
// 	},
// 	want:    "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg",
// 	wantErr: false,
// },
// {
// 	name: "Nonexistent JSON secret file",
// 	args: args{
// 		secret: "./testdata/nonexistent.json",
// 		url:    fakeServer.URL,
// 	},
// 	wantErr: true,
// },
// {
// 	name: "Mismatched token URI",
// 	args: args{
// 		secret: "./testdata/nonexistent.json",
// 		url:    "http://localhost:12345/token",
// 	},
// 	wantErr: true,
// },
