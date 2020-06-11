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
	ecsMetadataPath, _  = url.Parse("/v2/metadata")
	ecsMetaStatsPath, _ = url.Parse("/v2/stats")
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
func NewClient(timeout time.Duration) (*EcsClient, error) {
	c := &http.Client{
		Timeout: timeout,
	}

	return &EcsClient{
		client: c,
	}, nil
}

// EcsClient contains ECS connection config
type EcsClient struct {
	client   httpClient
	BaseURL  *url.URL
	taskURL  string
	statsURL string
}

// Task calls the ECS metadata endpoint and returns a populated Task
func (c *EcsClient) Task() (*Task, error) {
	if c.taskURL == "" {
		c.taskURL = c.BaseURL.ResolveReference(ecsMetadataPath).String()
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

// ContainerStats calls the ECS stats endpoint and returns a populated container stats map
func (c *EcsClient) ContainerStats() (map[string]types.StatsJSON, error) {
	if c.statsURL == "" {
		c.statsURL = c.BaseURL.ResolveReference(ecsMetaStatsPath).String()
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
