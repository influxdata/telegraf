package ecs

import (
	"fmt"
	"io"
	"io/ioutil"
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
func NewClient(timeout time.Duration, version int) (*EcsClient, error) {
	if version != 2 && version != 3 {
		const msg = "expected metadata version 2 or 3, got %d"
		return nil, fmt.Errorf(msg, version)
	}

	c := &http.Client{
		Timeout: timeout,
	}

	return &EcsClient{
		client:  c,
		version: version,
	}, nil
}

// EcsClient contains ECS connection config
type EcsClient struct {
	client   httpClient
	version  int
	BaseURL  *url.URL
	taskURL  string
	statsURL string
}

// Task calls the ECS metadata endpoint and returns a populated Task
func (c *EcsClient) Task() (*Task, error) {
	if c.taskURL == "" {
		path := getMetadataPath(c.version)
		c.taskURL = c.BaseURL.String() + path
	}

	req, _ := http.NewRequest("GET", c.taskURL, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// ignore the err here; LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("%s returned HTTP status %s: %q", c.taskURL, resp.Status, body)
	}

	task, err := unmarshalTask(resp.Body)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func getMetadataPath(version int) string {
	if version == 3 {
		return ecsMetadataPathV3
	}
	return ecsMetadataPath
}

// ContainerStats calls the ECS stats endpoint and returns a populated container stats map
func (c *EcsClient) ContainerStats() (map[string]types.StatsJSON, error) {
	if c.statsURL == "" {
		path := getMetaStatsPath(c.version)
		c.statsURL = c.BaseURL.String() + path
	}

	req, _ := http.NewRequest("GET", c.statsURL, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return map[string]types.StatsJSON{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// ignore the err here; LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("%s returned HTTP status %s: %q", c.statsURL, resp.Status, body)
	}

	statsMap, err := unmarshalStats(resp.Body)
	if err != nil {
		return map[string]types.StatsJSON{}, err
	}

	return statsMap, nil
}

func getMetaStatsPath(version int) string {
	if version == 3 {
		return ecsMetaStatsPathV3
	}
	return ecsMetaStatsPath
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
