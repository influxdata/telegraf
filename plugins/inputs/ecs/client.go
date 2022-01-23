package ecs

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/docker/api/types"
)

var (
	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html
	ecsMetadataPath  = "/v2/metadata"
	ecsMetaStatsPath = "/v2/stats"

	// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html
	ecsMetadataPathV3  = "/task"
	ecsMetaStatsPathV3 = "/task/stats"
)

// Client is the ECS client contract
type Client interface {
	Task() (*Task, error)
	ContainerStats() (map[string]types.StatsJSON, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient constructs an ECS client with the passed configuration params
func NewClient(timeout time.Duration, endpoint string, version int) (*EcsClient, error) {
	if version != 2 && version != 3 {
		const msg = "expected metadata version 2 or 3, got %d"
		return nil, fmt.Errorf(msg, version)
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Timeout: timeout,
	}

	return &EcsClient{
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
		path = ecsMetadataPath
	case 3:
		path = ecsMetadataPathV3
	default:
		// Should never happen.
		const msg = "resolveTaskURL: unexpected version %d"
		panic(fmt.Errorf(msg, version))
	}
	return resolveURL(base, path)
}

func resolveStatsURL(base *url.URL, version int) string {
	var path string
	switch version {
	case 2:
		path = ecsMetaStatsPath
	case 3:
		path = ecsMetaStatsPathV3
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

// EcsClient contains ECS connection config
type EcsClient struct {
	client   httpClient
	version  int
	baseURL  *url.URL
	taskURL  string
	statsURL string
}

// Task calls the ECS metadata endpoint and returns a populated Task
func (c *EcsClient) Task() (*Task, error) {
	req, _ := http.NewRequest("GET", c.taskURL, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// ignore the err here; LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("%s returned HTTP status %s: %q", c.taskURL, resp.Status, body)
	}

	task, err := unmarshalTask(resp.Body)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// ContainerStats calls the ECS stats endpoint and returns a populated container stats map
func (c *EcsClient) ContainerStats() (map[string]types.StatsJSON, error) {
	req, _ := http.NewRequest("GET", c.statsURL, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return map[string]types.StatsJSON{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// ignore the err here; LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("%s returned HTTP status %s: %q", c.statsURL, resp.Status, body)
	}

	statsMap, err := unmarshalStats(resp.Body)
	if err != nil {
		return map[string]types.StatsJSON{}, err
	}

	return statsMap, nil
}

// PollSync executes Task and ContainerStats in parallel. If both succeed, both structs are returned.
// If either errors, a single error is returned.
func PollSync(c Client) (*Task, map[string]types.StatsJSON, error) {
	var task *Task
	var stats map[string]types.StatsJSON
	var err error

	if stats, err = c.ContainerStats(); err != nil {
		return nil, nil, err
	}

	if task, err = c.Task(); err != nil {
		return nil, nil, err
	}

	return task, stats, nil
}
