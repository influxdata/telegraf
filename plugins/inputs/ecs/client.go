package ecs

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/docker/api/types/container"
)

var (
	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html
	ecsMetadataPathV2  = "/v2/metadata"
	ecsMetaStatsPathV2 = "/v2/stats"

	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html
	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4.html
	ecsMetadataPath  = "/task"
	ecsMetaStatsPath = "/task/stats"
)

// client is the ECS client contract
type client interface {
	task() (*ecsTask, error)
	containerStats() (map[string]*container.StatsResponse, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// newClient constructs an ECS client with the passed configuration params
func newClient(timeout time.Duration, endpoint string, version int) (*ecsClient, error) {
	if version < 2 || version > 4 {
		const msg = "expected metadata version 2, 3 or 4, got %d"
		return nil, fmt.Errorf(msg, version)
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Timeout: timeout,
	}

	return &ecsClient{
		client:   c,
		baseURL:  baseURL,
		taskURL:  resolveTaskURL(baseURL, version),
		statsURL: resolveStatsURL(baseURL, version),
		version:  version,
	}, nil
}

func resolveTaskURL(base *url.URL, version int) string {
	var path string
	switch version {
	case 2:
		path = ecsMetadataPathV2
	case 3:
		path = ecsMetadataPath
	case 4:
		path = ecsMetadataPath
	default:
		const msg = "resolveTaskURL: unexpected version %d"
		panic(fmt.Errorf(msg, version))
	}
	return resolveURL(base, path)
}

func resolveStatsURL(base *url.URL, version int) string {
	var path string
	switch version {
	case 2:
		path = ecsMetaStatsPathV2
	case 3:
		path = ecsMetaStatsPath
	case 4:
		path = ecsMetaStatsPath
	default:
		// Should never happen.
		const msg = "resolveStatsURL: unexpected version %d"
		panic(fmt.Errorf(msg, version))
	}
	return resolveURL(base, path)
}

// resolveURL returns a URL string by concatenating the string representation of base
// and path. This is consistent with AWS metadata documentation:
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html#task-metadata-endpoint-v3-paths
func resolveURL(base *url.URL, path string) string {
	return base.String() + path
}

// ecsClient contains ECS connection config
type ecsClient struct {
	client   httpClient
	version  int
	baseURL  *url.URL
	taskURL  string
	statsURL string
}

// task calls the ECS metadata endpoint and returns a populated task
func (c *ecsClient) task() (*ecsTask, error) {
	req, err := http.NewRequest("GET", c.taskURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//nolint:errcheck // LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("%s returned HTTP status %s: %q", c.taskURL, resp.Status, body)
	}

	task, err := unmarshalTask(resp.Body)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// containerStats calls the ECS stats endpoint and returns a populated container stats map
func (c *ecsClient) containerStats() (map[string]*container.StatsResponse, error) {
	req, err := http.NewRequest("GET", c.statsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//nolint:errcheck // LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("%s returned HTTP status %s: %q", c.statsURL, resp.Status, body)
	}

	return unmarshalStats(resp.Body)
}

// pollSync executes task and containerStats in parallel.
// If both succeed, both structs are returned.
// If either errors, a single error is returned.
func pollSync(c client) (*ecsTask, map[string]*container.StatsResponse, error) {
	var task *ecsTask
	var stats map[string]*container.StatsResponse
	var err error

	if stats, err = c.containerStats(); err != nil {
		return nil, nil, err
	}

	if task, err = c.task(); err != nil {
		return nil, nil, err
	}

	return task, stats, nil
}
