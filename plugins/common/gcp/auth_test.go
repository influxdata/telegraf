package gcp

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAccessToken(t *testing.T) {
	fakeServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.WriteHeader(http.StatusAccepted)
			_, err := w.Write([]byte(`{"id_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg"}`))
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
			t.Fatalf("unexpected path: " + r.URL.Path)
		}
	}))
	// Matches the token_uri field in ./testdata/test_key_file.json
	l, _ := net.Listen("tcp", "localhost:58025")
	fakeServer.Listener = l
	fakeServer.Start()
	defer fakeServer.Close()

	type args struct {
		secret string
		url    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Successfully request access token",
			args: args{
				secret: "./testdata/test_key_file.json",
				url:    fakeServer.URL,
			},
			want:    "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijg2NzUzMDliMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwOi8vMTI3LjAuMC4xOjU4MDI1LyIsImF6cCI6InRlc3Qtc2VydmljZS1hY2NvdW50LWVtYWlsQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ0ZXN0LXNlcnZpY2UtYWNjb3VudC1lbWFpbEBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjk0NjY4NDgwMCwiaWF0Ijo5NDY2ODEyMDAsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMudGVzdC5jb20iLCJzdWIiOiIxMTAzMDAwMDk4MTM3Mzg2NzUzMDkifQ.qi2LsXP2o6nl-rbYKUlHAgTBY0QoU7Nhty5NGR4GMdc8OoGEPW-vlD0WBSaKSr11vyFcIO4ftFDWXElo9Ut-AIQPKVxinsjHIU2-LoIATgI1kyifFLyU_pBecwcI4CIXEcDK5wEkfonWFSkyDZHBeZFKbJXlQXtxj0OHvQ-DEEepXLuKY6v3s4U6GyD9_ppYUy6gzDZPYUbfPfgxCj_Jbv6qkLU0DiZ7F5-do6X6n-qkpgCRLTGHcY__rn8oe8_pSimsyJEeY49ZQ5lj4mXkVCwgL9bvL1_eW1p6sgbHaBnPKVPbM7S1_cBmzgSonm__qWyZUxfDgNdigtNsvzBQTg",
			wantErr: false,
		},
		{
			name: "Nonexistent JSON secret file",
			args: args{
				secret: "./testdata/nonexistent.json",
				url:    fakeServer.URL,
			},
			wantErr: true,
		},
		{
			name: "Mismatched token URI",
			args: args{
				secret: "./testdata/nonexistent.json",
				url:    "http://localhost:12345/token",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAccessToken(tt.args.secret, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
