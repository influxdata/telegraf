package ecs

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

type pollMock struct {
	task  func() (*Task, error)
	stats func() (map[string]types.StatsJSON, error)
}

func (p *pollMock) Task() (*Task, error) {
	return p.task()
}

func (p *pollMock) ContainerStats() (map[string]types.StatsJSON, error) {
	return p.stats()
}

func TestEcsClient_PollSync(t *testing.T) {

	tests := []struct {
		name    string
		mock    *pollMock
		want    *Task
		want1   map[string]types.StatsJSON
		wantErr bool
	}{
		{
			name: "success",
			mock: &pollMock{
				task: func() (*Task, error) {
					return &validMeta, nil
				},
				stats: func() (map[string]types.StatsJSON, error) {
					return validStats, nil
				},
			},
			want:  &validMeta,
			want1: validStats,
		},
		{
			name: "task err",
			mock: &pollMock{
				task: func() (*Task, error) {
					return nil, errors.New("err")
				},
				stats: func() (map[string]types.StatsJSON, error) {
					return validStats, nil
				},
			},
			wantErr: true,
		},
		{
			name: "stats err",
			mock: &pollMock{
				task: func() (*Task, error) {
					return &validMeta, nil
				},
				stats: func() (map[string]types.StatsJSON, error) {
					return nil, errors.New("err")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := PollSync(tt.mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("EcsClient.PollSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "EcsClient.PollSync() got = %v, want %v", got, tt.want)
			assert.Equal(t, tt.want1, got1, "EcsClient.PollSync() got1 = %v, want %v", got1, tt.want1)
		})
	}
}

type mockDo struct {
	do func(req *http.Request) (*http.Response, error)
}

func (m mockDo) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

func TestEcsClient_Task(t *testing.T) {
	rc, _ := os.Open("testdata/metadata.golden")
	tests := []struct {
		name    string
		client  httpClient
		want    *Task
		wantErr bool
	}{
		{
			name: "happy",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(rc),
					}, nil
				},
			},
			want: &validMeta,
		},
		{
			name: "do err",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("err")
				},
			},
			wantErr: true,
		},
		{
			name: "malformed 500 resp",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "malformed 200 resp",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &EcsClient{
				client:  tt.client,
				taskURL: "abc",
			}
			got, err := c.Task()
			if (err != nil) != tt.wantErr {
				t.Errorf("EcsClient.Task() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "EcsClient.Task() = %v, want %v", got, tt.want)
		})
	}
}

func TestEcsClient_ContainerStats(t *testing.T) {
	rc, _ := os.Open("testdata/stats.golden")
	tests := []struct {
		name    string
		client  httpClient
		want    map[string]types.StatsJSON
		wantErr bool
	}{
		{
			name: "happy",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(rc),
					}, nil
				},
			},
			want: validStats,
		},
		{
			name: "do err",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("err")
				},
			},
			want:    map[string]types.StatsJSON{},
			wantErr: true,
		},
		{
			name: "malformed 200 resp",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			want:    map[string]types.StatsJSON{},
			wantErr: true,
		},
		{
			name: "malformed 500 resp",
			client: mockDo{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("foo"))),
					}, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &EcsClient{
				client:   tt.client,
				statsURL: "abc",
			}
			got, err := c.ContainerStats()
			if (err != nil) != tt.wantErr {
				t.Errorf("EcsClient.ContainerStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "EcsClient.ContainerStats() = %v, want %v", got, tt.want)
		})
	}
}
