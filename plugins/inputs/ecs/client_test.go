package ecs

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

type pollMock struct {
	getTask  func() (*ecsTask, error)
	getStats func() (map[string]*container.StatsResponse, error)
}

func (p *pollMock) task() (*ecsTask, error) {
	return p.getTask()
}

func (p *pollMock) containerStats() (map[string]*container.StatsResponse, error) {
	return p.getStats()
}

func TestEcsClient_PollSync(t *testing.T) {
	tests := []struct {
		name    string
		mock    *pollMock
		want    *ecsTask
		want1   map[string]*container.StatsResponse
		wantErr bool
	}{
		{
			name: "success",
			mock: &pollMock{
				getTask: func() (*ecsTask, error) {
					return &validMeta, nil
				},
				getStats: func() (map[string]*container.StatsResponse, error) {
					return validStats, nil
				},
			},
			want:  &validMeta,
			want1: validStats,
		},
		{
			name: "task err",
			mock: &pollMock{
				getTask: func() (*ecsTask, error) {
					return nil, errors.New("err")
				},
				getStats: func() (map[string]*container.StatsResponse, error) {
					return validStats, nil
				},
			},
			wantErr: true,
		},
		{
			name: "stats err",
			mock: &pollMock{
				getTask: func() (*ecsTask, error) {
					return &validMeta, nil
				},
				getStats: func() (map[string]*container.StatsResponse, error) {
					return nil, errors.New("err")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := pollSync(tt.mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("ecsClient.pollSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got, "ecsClient.pollSync() got = %v, want %v", got, tt.want)
			require.Equal(t, tt.want1, got1, "ecsClient.pollSync() got1 = %v, want %v", got1, tt.want1)
		})
	}
}

type mockDo struct {
	do func() (*http.Response, error)
}

func (m mockDo) Do(*http.Request) (*http.Response, error) {
	return m.do()
}

func TestEcsClient_Task(t *testing.T) {
	tests := []struct {
		name    string
		client  httpClient
		want    *ecsTask
		wantErr bool
	}{
		{
			name: "happy",
			client: mockDo{
				do: func() (*http.Response, error) {
					rc, err := os.Open("testdata/metadata.golden")
					if err != nil {
						return nil, err
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(rc),
					}, nil
				},
			},
			want: &validMeta,
		},
		{
			name: "do err",
			client: mockDo{
				do: func() (*http.Response, error) {
					return nil, errors.New("err")
				},
			},
			wantErr: true,
		},
		{
			name: "malformed 500 resp",
			client: mockDo{
				do: func() (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "malformed 200 resp",
			client: mockDo{
				do: func() (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ecsClient{
				client:  tt.client,
				taskURL: "abc",
			}
			got, err := c.task()
			if (err != nil) != tt.wantErr {
				t.Errorf("ecsClient.task() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got, "ecsClient.task() = %v, want %v", got, tt.want)
		})
	}
}

func TestEcsClient_ContainerStats(t *testing.T) {
	tests := []struct {
		name    string
		client  httpClient
		want    map[string]*container.StatsResponse
		wantErr bool
	}{
		{
			name: "happy",
			client: mockDo{
				do: func() (*http.Response, error) {
					rc, err := os.Open("testdata/stats.golden")
					if err != nil {
						return nil, err
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(rc),
					}, nil
				},
			},
			want: validStats,
		},
		{
			name: "do err",
			client: mockDo{
				do: func() (*http.Response, error) {
					return nil, errors.New("err")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "malformed 200 resp",
			client: mockDo{
				do: func() (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "malformed 500 resp",
			client: mockDo{
				do: func() (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ecsClient{
				client:   tt.client,
				statsURL: "abc",
			}
			got, err := c.containerStats()
			if (err != nil) != tt.wantErr {
				t.Errorf("ecsClient.containerStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got, "ecsClient.containerStats() = %v, want %v", got, tt.want)
		})
	}
}

func TestResolveTaskURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		ver  int
		exp  string
	}{
		{
			name: "default v2 endpoint",
			base: v2Endpoint,
			ver:  2,
			exp:  "http://169.254.170.2/v2/metadata",
		},
		{
			name: "custom v2 endpoint",
			base: "http://192.168.0.1",
			ver:  2,
			exp:  "http://192.168.0.1/v2/metadata",
		},
		{
			name: "theoretical v3 endpoint",
			base: "http://169.254.170.2/v3/metadata",
			ver:  3,
			exp:  "http://169.254.170.2/v3/metadata/task",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse(tt.base)
			require.NoError(t, err)

			act := resolveTaskURL(baseURL, tt.ver)
			require.Equal(t, tt.exp, act)
		})
	}
}

func TestResolveStatsURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		ver  int
		exp  string
	}{
		{
			name: "default v2 endpoint",
			base: v2Endpoint,
			ver:  2,
			exp:  "http://169.254.170.2/v2/stats",
		},
		{
			name: "custom v2 endpoint",
			base: "http://192.168.0.1",
			ver:  2,
			exp:  "http://192.168.0.1/v2/stats",
		},
		{
			name: "theoretical v3 endpoint",
			base: "http://169.254.170.2/v3/metadata",
			ver:  3,
			exp:  "http://169.254.170.2/v3/metadata/task/stats",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse(tt.base)
			require.NoError(t, err)

			act := resolveStatsURL(baseURL, tt.ver)
			require.Equal(t, tt.exp, act)
		})
	}
}
