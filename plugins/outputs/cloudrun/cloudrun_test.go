package cloudrun

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func defaultCloudrun() *CloudRun {
	return &CloudRun{
		DisableConvertPaths: true,
		CredentialsFile:     "./testdata/test_key_file.json",
	}
}

func (cr *CloudRun) setUpDefaultTestClient() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cr.Timeout))
	defer cancel()

	// TODO: Evaluate whether to request token here.
	// err := cr.getAccessToken(ctx)
	// if err != nil {
	// 	return err
	// }

	client, err := cr.HTTPClientConfig.CreateClient(ctx, cr.Log)
	if err != nil {
		return err
	}

	cr.client = client
	return nil
}

// func TestCloudRun_SetSerializer(t *testing.T) {
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
// 	type args struct {
// 		serializer serializers.Serializer
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
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
// 			cr.SetSerializer(tt.args.serializer)
// 		})
// 	}
// }

// func TestCloudRun_Connect(t *testing.T) {
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
// 		name    string
// 		fields  fields
// 		wantErr bool
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
// 			if err := cr.Connect(); (err != nil) != tt.wantErr {
// 				t.Errorf("CloudRun.Connect() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestCloudRun_setUpDefaultClient(t *testing.T) {
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
// 		name    string
// 		fields  fields
// 		wantErr bool
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
// 			if err := cr.setUpDefaultClient(); (err != nil) != tt.wantErr {
// 				t.Errorf("CloudRun.setUpDefaultClient() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestCloudRun_getAccessToken(t *testing.T) {
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
// 	type args struct {
// 		ctx context.Context
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
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
// 			if err := cr.getAccessToken(tt.args.ctx); (err != nil) != tt.wantErr {
// 				t.Errorf("CloudRun.getAccessToken() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestCloudRun_Close(t *testing.T) {
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
// 		name    string
// 		fields  fields
// 		wantErr bool
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
// 			if err := cr.Close(); (err != nil) != tt.wantErr {
// 				t.Errorf("CloudRun.Close() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
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
	cr := defaultCloudrun()
	cr.serializer = influx.NewSerializer()
	cr.setUpDefaultTestClient()

	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			// This receives metrics
			// TODO: Inspect metrics payload? How to test...
			fmt.Println("Hello root")
			w.WriteHeader(http.StatusAccepted)

		case "/token":
			fmt.Println("Hello token")
			// fmt.Println(r.Header)
			w.WriteHeader(http.StatusAccepted)
			// TODO: assess responding with the same token every time.
			// fmt.Println("Get your tokens here")
			_, _ = w.Write([]byte(`{"id_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg"}`))
			// _, err := w.Write([]byte([]byte(fmt.Sprintf(`{"id_token":"%s"`, cr.accessToken))))
			// require.NoError(t, err)
		default:
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
}

// func TestCloudRun_send(t *testing.T) {
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
// 	type args struct {
// 		reqBody []byte
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
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
// 			if err := cr.send(tt.args.reqBody); (err != nil) != tt.wantErr {
// 				t.Errorf("CloudRun.send() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
