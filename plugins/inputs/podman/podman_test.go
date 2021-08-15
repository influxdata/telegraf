package podman

import (
	"log"
	"os"
	"testing"

	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/influxdata/telegraf/testutil"
)

type MockClient struct {
	InfoF           func() (*define.Info, error)
	ContainerListF  func(filters map[string][]string) ([]entities.ListContainer, error)
	ContainerStatsF func(string) (*define.ContainerStats, error)
}

func (c *MockClient) Info() (*define.Info, error) {
	return c.InfoF()
}

func (c *MockClient) ContainerList(
	filters map[string][]string,
) ([]entities.ListContainer, error) {
	return c.ContainerListF(filters)
}

func (c *MockClient) ContainerStats(
	containerID string,
) (*define.ContainerStats, error) {
	return c.ContainerStatsF(containerID)
}

var baseClient = MockClient{
	InfoF: func() (*define.Info, error) {
		return &info, nil
	},
	ContainerListF: func(filters map[string][]string) ([]entities.ListContainer, error) {
		return containerList, nil
	},
	ContainerStatsF: func(containerID string) (*define.ContainerStats, error) {
		if containerID == container_test_1 {
			return &containerStats_nginx, nil
		}
		return &containerStats_blissful_lewin, nil
	},
}

func TestPodmanGatherContainerStats(t *testing.T) {

}

func TestPodman(t *testing.T) {
	// Get Podman socket location
	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	socket := "unix:" + sock_dir + "/podman/podman.sock"
	var acc testutil.Accumulator
	p := &Podman{
		Log:      testutil.Logger{},
		Endpoint: socket,
	}
	err := p.Gather(&acc)
	if err != nil {
		log.Fatal(err)
	}
	acc.Wait(1)
	log.Println(acc.GetTelegrafMetrics())
}
