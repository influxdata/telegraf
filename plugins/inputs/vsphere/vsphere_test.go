package vsphere

import (
	"github.com/vmware/govmomi/simulator"
	"crypto/tls"
	"testing"
	"fmt"
	"time"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func defaultVSphere() *VSphere {
	return &VSphere{
		GatherClusters:   true,
		ClusterMetrics:   nil,
		GatherHosts:      true,
		HostMetrics:      nil,
		GatherVms:        true,
		VmMetrics:        nil,
		GatherDatastores: true,
		DatastoreMetrics: nil,
		InsecureSkipVerify: true,

		ObjectsPerQuery:         256,
		ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
		Timeout:                 internal.Duration{Duration: time.Second * 20},
	}
}

func createSim() (*simulator.Model, *simulator.Server) {
	model := simulator.VPX()

	err := model.Create()
	if err != nil {
		fmt.Errorf("Error creating model: %s\n", err)
	}

	model.Service.TLS = new(tls.Config)

	s := model.Service.NewServer()
	fmt.Printf("Server created at: %s\n", s.URL)

	return model, s
}

func TestAll(t *testing.T) {
	m, s := createSim()
	defer m.Remove()
	defer s.Close()

	var acc testutil.Accumulator
	v := defaultVSphere()
	v.Vcenters = []string{s.URL.String()}
	require.NoError(t, v.Gather(&acc))
}