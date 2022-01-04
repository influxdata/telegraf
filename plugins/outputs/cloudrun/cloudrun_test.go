package cloudrun

import (
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudRun_Write(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusAccepted)
		case "/token":
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	l, _ := net.Listen("tcp", "localhost:58025")
	fakeServer.Listener = l
	fakeServer.Start()
	defer fakeServer.Close()

	tests := []struct {
		name    string
		metrics []telegraf.Metric
		plugin  *CloudRun
		wantErr bool
	}{
		{
			name:    "write success",
			metrics: testutil.MockMetrics(),
			plugin: &CloudRun{
				CredentialsFile: "./testdata/test_key_file.json",
			},
			wantErr: false,
		},
		{
			name:    "no credentials file",
			metrics: testutil.MockMetrics(),
			plugin: &CloudRun{
				CredentialsFile: "./testdata/missing.json",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.SetSerializer(influx.NewSerializer())
			if err := tt.plugin.Connect(); tt.wantErr {
				if runtime.GOOS == "windows" {
					require.Equal(t, "open ./testdata/missing.json: The system cannot find the file specified.", err.Error())
				} else {
					require.Equal(t, "open ./testdata/missing.json: no such file or directory", err.Error())
				}
			} else {
				require.NoError(t, err)
			}

			tt.plugin.URL = fakeServer.URL

			if err := tt.plugin.Write(tt.metrics); tt.wantErr {
				if runtime.GOOS == "windows" {
					require.Equal(t, "open ./testdata/missing.json: The system cannot find the file specified.", err.Error())
				} else {
					require.Equal(t, "open ./testdata/missing.json: no such file or directory", err.Error())
				}
			} else {
				require.NoError(t, err)
			}

			err := tt.plugin.Close()
			require.NoError(t, err)
		})
	}
}
