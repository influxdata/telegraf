package cloudrun

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func defaultCloudrun() *CloudRun {
	return &CloudRun{
		DisableConvertPaths: true,
		CredentialsFile:     "./testdata/test_key_file.json",
	}
}

// This could instead be Connect()
// func (cr *CloudRun) setUpDefaultTestClient() error {
// 	ctx := context.Background()
// 	ctx, cancel := context.WithTimeout(ctx, time.Duration(cr.Timeout))
// 	defer cancel()

// 	// TODO: Evaluate whether to request token here.
// 	// err := cr.getAccessToken(ctx)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	client, err := cr.HTTPClientConfig.CreateClient(ctx, cr.Log)
// 	if err != nil {
// 		return err
// 	}

// 	cr.client = client
// 	return nil
// }

// func TestCloudRun_Description(t *testing.T) {
// 	type fields struct {
// 		URL                 string
// 		CredentialsFile     string
// 		DisableConvertPaths bool
// 		Log                 telegraf.Logger
// 		HTTPClientConfig    httpconfig.HTTPClientConfig
// 		client              *http.Client
// 		serializer          serializers.Serializer
// 		accessToken         string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cr := &CloudRun{
// 				URL:                 tt.fields.URL,
// 				CredentialsFile:     tt.fields.CredentialsFile,
// 				DisableConvertPaths: tt.fields.DisableConvertPaths,
// 				Log:                 tt.fields.Log,
// 				HTTPClientConfig:    tt.fields.HTTPClientConfig,
// 				client:              tt.fields.client,
// 				serializer:          tt.fields.serializer,
// 				accessToken:         tt.fields.accessToken,
// 			}
// 			if got := cr.Description(); got != tt.want {
// 				t.Errorf("CloudRun.Description() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestCloudRun_SampleConfig(t *testing.T) {
// 	type fields struct {
// 		URL                 string
// 		CredentialsFile     string
// 		DisableConvertPaths bool
// 		Log                 telegraf.Logger
// 		HTTPClientConfig    httpconfig.HTTPClientConfig
// 		client              *http.Client
// 		serializer          serializers.Serializer
// 		accessToken         string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cr := &CloudRun{
// 				URL:                 tt.fields.URL,
// 				CredentialsFile:     tt.fields.CredentialsFile,
// 				DisableConvertPaths: tt.fields.DisableConvertPaths,
// 				Log:                 tt.fields.Log,
// 				HTTPClientConfig:    tt.fields.HTTPClientConfig,
// 				client:              tt.fields.client,
// 				serializer:          tt.fields.serializer,
// 				accessToken:         tt.fields.accessToken,
// 			}
// 			if got := cr.SampleConfig(); got != tt.want {
// 				t.Errorf("CloudRun.SampleConfig() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestCloudRun_Write(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			// TODO: Inspect metrics payload? How to test...
			w.WriteHeader(http.StatusAccepted)
		case "/token":
			w.WriteHeader(http.StatusAccepted)
			// TODO: assess responding with the same token every time.
			_, _ = w.Write([]byte(`{"id_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	l, _ := net.Listen("tcp", "localhost:58025")
	fakeServer.Listener = l
	fakeServer.Start()
	defer fakeServer.Close()

	cr := defaultCloudrun()
	cr.SetSerializer(influx.NewSerializer())
	// cr.serializer = influx.NewSerializer()

	// cr.setUpDefaultTestClient()
	err := cr.Connect()
	require.NoError(t, err)

	cr.URL = fakeServer.URL

	tests := []struct {
		name    string
		metrics []telegraf.Metric
		wantErr bool
	}{
		// WIP: Adding test cases.
		{
			name:    "write success",
			metrics: testutil.MockMetrics(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cr.Write(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("CloudRun.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	err = cr.Close()
	require.NoError(t, err)
}
