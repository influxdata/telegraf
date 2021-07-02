package podman

import (
	"log"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestPodmanGatherContainerStats(t *testing.T) {
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
