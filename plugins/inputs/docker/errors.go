package docker

import "errors"

var (
	errInfoTimeout    = errors.New("timeout retrieving docker engine info")
	errStatsTimeout   = errors.New("timeout retrieving container stats")
	errInspectTimeout = errors.New("timeout retrieving container environment")
	errListTimeout    = errors.New("timeout retrieving container list")
	errServiceTimeout = errors.New("timeout retrieving swarm service list")
)
