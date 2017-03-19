package libvirt

import (
	"fmt"
	"reflect"
	"testing"

	"errors"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const mockListStr = ` Id    Name                           State
----------------------------------------------------
 1     vm1                           running
 2     vm2                           running

`

const mockListNoVmsStr = ` Id    Name                           State
----------------------------------------------------

`

const mockListNotInstalled = `The program 'virsh' is currently not installed. You can install it by typing:
sudo apt install libvirt-bin
`

const mockDomInfo1 = `Id:             1
Name:           vm1
UUID:           6695eb01-f6a4-8304-79aa-97f2502e193f
OS Type:        linux
State:          running
CPU(s):         2
CPU time:       1489945053.2s
Max memory:     8388608 KiB
Used memory:    2097152 KiB
Persistent:     yes
Autostart:      disable
Managed save:   no

`

const mockDomInfo2 = `Id:             2
Name:           vm2
UUID:           3ff44288-6039-49ae-c803-4e399a52779d
OS Type:        linux
State:          running
CPU(s):         8
CPU time:       11234.6s
Max memory:     4194304 KiB
Used memory:    4194304 KiB
Persistent:     yes
Autostart:      disable
Managed save:   no

`

func mockVirshNormal(_ string, cmd ...string) (string, error) {
	if reflect.DeepEqual([]string{"list"}, cmd) {
		return mockListStr, nil
	} else if reflect.DeepEqual([]string{"dominfo", "vm1"}, cmd) {
		return mockDomInfo1, nil
	} else if reflect.DeepEqual([]string{"dominfo", "vm2"}, cmd) {
		return mockDomInfo2, nil
	}

	return "", fmt.Errorf("unknown cmd: %q", cmd)
}

func mockVirshNoVms(_ string, cmd ...string) (string, error) {
	if reflect.DeepEqual([]string{"list"}, cmd) {
		return mockListNoVmsStr, nil
	}

	return "", fmt.Errorf("unexpected cmd: %q", cmd)
}

func mockVirshNotInstalled(_ string, cmd ...string) (string, error) {
	if reflect.DeepEqual([]string{"list"}, cmd) {
		return mockListNotInstalled, errors.New("exec: executable file not found in $PATH")
	}

	return "", fmt.Errorf("unexpected cmd: %q", cmd)
}

func TestLibvirtNormal(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default", virsh: mockVirshNormal}

	err := lv.Gather(&acc)
	require.NoError(t, err)

	vm1_tags := map[string]string{
		"domain": "vm1",
		"state":  "running",
	}
	vm1_fields := map[string]interface{}{
		"cpu_time":    1489945053.2,
		"max_memory":  uint64(8388608),
		"used_memory": uint64(2097152),
		"n_vcpu":      uint64(2),
	}

	vm2_tags := map[string]string{
		"domain": "vm2",
		"state":  "running",
	}
	vm2_fields := map[string]interface{}{
		"cpu_time":    11234.6,
		"max_memory":  uint64(4194304),
		"used_memory": uint64(4194304),
		"n_vcpu":      uint64(8),
	}

	acc.AssertContainsTaggedFields(t, "libvirt", vm1_fields, vm1_tags)
	acc.AssertContainsTaggedFields(t, "libvirt", vm2_fields, vm2_tags)
}

func TestLibvirtNoVms(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default", virsh: mockVirshNoVms}

	err := lv.Gather(&acc)
	require.NoError(t, err)

	acc.AssertDoesNotContainMeasurement(t, "libvirt")
}

func TestLibvirtNotInstalled(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default", virsh: mockVirshNotInstalled}

	err := lv.Gather(&acc)
	require.EqualError(t, err, "exec: executable file not found in $PATH", "expect not installed to fail")

	acc.AssertDoesNotContainMeasurement(t, "libvirt")
}
