package podman

import "errors"

var (
	errNoStats = errors.New("no container stats retrieved")
)
