package ecs

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
)

const (
	ecsMetaScheme = "http"
)

var (
	ecsMetadataPath, _  = url.Parse("/v2/metadata")
	ecsMetaStatsPath, _ = url.Parse("/v2/stats")
)

// Client is the ECS client contract
type Client interface {
	Task() (Task, error)
	ContainerStats() (map[string]types.StatsJSON, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewEnvClient configures a new Client from the env
func NewEnvClient() (*EcsClient, error) {
	timeout := 5 * time.Second
	if t := os.Getenv("ECS_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			timeout = d
		}
	}

	return NewClient(
		timeout,
	)
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
func (c *EcsClient) Task() (Task, error) {
	if c.taskURL == "" {
		c.taskURL = c.BaseURL.ResolveReference(ecsMetadataPath).String()
	}

	req, _ := http.NewRequest("GET", c.taskURL, nil)
	resp, err := c.client.Do(req)

	if err != nil {
		log.Println("failed to GET metadata endpoint", err)
		return Task{}, err
	}

	task, err := unmarshalTask(resp.Body)
	if err != nil {
		log.Println("failed to decode response from metadata endpoint", err)
		return Task{}, err
	}

	return *task, nil
}

// ContainerStats calls the ECS stats endpoint and returns a populated container stats map
func (c *EcsClient) ContainerStats() (map[string]types.StatsJSON, error) {
	if c.statsURL == "" {
		c.statsURL = c.BaseURL.ResolveReference(ecsMetaStatsPath).String()
	}

	req, _ := http.NewRequest("GET", c.statsURL, nil)
	resp, err := c.client.Do(req)

	if err != nil {
		log.Println("failed to GET stats endpoint", err)
		return map[string]types.StatsJSON{}, err
	}

	statsMap, err := unmarshalStats(resp.Body)
	if err != nil {
		log.Println("failed to decode response from stats endpoint")
		return map[string]types.StatsJSON{}, err
	}

	return statsMap, nil
}

// PollSync executes Task and ContainerStats in parallel. If both succeed, both structs are returned.
// If either errors, a single error is returned.
func PollSync(c Client) (Task, map[string]types.StatsJSON, error) {

	var stats = map[string]types.StatsJSON{}
	var statsErr error
	var task Task
	var taskErr error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		stats, statsErr = c.ContainerStats()

		if statsErr != nil {
			log.Println("Failed to poll ECS endpoint:", statsErr)
			return
		}
	}()

	go func() {
		defer wg.Done()
		task, taskErr = c.Task()

		if taskErr != nil {
			log.Println("Failed to poll ECS endpoint:", taskErr)
			return
		}
	}()

	wg.Wait()

	if statsErr != nil || taskErr != nil {
		log.Printf("Stats or tasks polling failed. stats: %v, task: %v\n", statsErr, taskErr)
		return Task{}, nil, errors.New("polling failed")
	}

	return task, stats, nil
}
