package ecs

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

// Task is the ECS task representation
type Task struct {
	Cluster       string
	TaskARN       string
	Family        string
	Revision      string
	DesiredStatus string
	KnownStatus   string
	Containers    []Container
	Limits        map[string]float64
	PullStartedAt time.Time
	PullStoppedAt time.Time
}

// Container is the ECS metadata container representation
type Container struct {
	ID            string `json:"DockerId"`
	Name          string
	DockerName    string
	Image         string
	ImageID       string
	Labels        map[string]string
	DesiredStatus string
	KnownStatus   string
	Limits        map[string]float64
	CreatedAt     time.Time
	StartedAt     time.Time
	Stats         types.StatsJSON
	Type          string
	Networks      []Network
}

// Network is a docker network configuration
type Network struct {
	NetworkMode   string
	IPv4Addresses []string
}

func unmarshalTask(r io.Reader) (*Task, error) {
	task := &Task{}
	err := json.NewDecoder(r).Decode(task)
	return task, err
}

// docker parsers
func unmarshalStats(r io.Reader) (map[string]types.StatsJSON, error) {
	var statsMap map[string]types.StatsJSON
	err := json.NewDecoder(r).Decode(&statsMap)
	return statsMap, err
}

// interleaves Stats in to the Container objects in the Task
func mergeTaskStats(task *Task, stats map[string]types.StatsJSON) {
	for i, c := range task.Containers {
		if strings.Trim(c.ID, " ") == "" {
			continue
		}
		stat, ok := stats[c.ID]
		if !ok {
			continue
		}
		task.Containers[i].Stats = stat
	}
}
