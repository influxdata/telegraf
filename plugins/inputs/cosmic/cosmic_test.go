package cosmic

import (
	"testing"

	"github.com/MissionCriticalCloud/go-cosmic/cosmic"
	"github.com/influxdata/telegraf/testutil"
)

func TestGatherVirtualMachineMetrics(t *testing.T) {
	var sampleVirtualMachines = []*cosmic.VirtualMachine{
		{
			Account:             "testaccount",
			Cpunumber:           1,
			Created:             "2018-01-01T00:00:00+0000",
			Displayname:         "testvm",
			Domain:              "testdomain",
			Domainid:            "00000000-0000-0000-0000-000000000001",
			Hostid:              "",
			Hostname:            "",
			Hypervisor:          "KVM",
			Id:                  "00000000-0000-0000-0000-000000000002",
			Instancename:        "i-1-12345-VM",
			Memory:              1024,
			Name:                "testvm",
			Serviceofferingid:   "00000000-0000-0000-0000-000000000003",
			Serviceofferingname: "testserviceofferingname",
			State:               "Stopped",
			Templatedisplaytext: "testtemplatetext",
			Templateid:          "00000000-0000-0000-0000-000000000004",
			Templatename:        "testtemplatename",
			Userid:              "00000000-0000-0000-0000-000000000005",
			Username:            "testuser",
			Zoneid:              "00000000-0000-0000-0000-000000000006",
			Zonename:            "testzone",
		},
	}

	expectedFields := map[string]interface{}{
		"cpunumber":           1,
		"memory":              1024,
		"state":               "Stopped",
		"hostid":              "",
		"hostname":            "",
		"serviceofferingid":   "00000000-0000-0000-0000-000000000003",
		"serviceofferingname": "testserviceofferingname",
	}

	expectedTags := map[string]string{
		"id":                  "00000000-0000-0000-0000-000000000002",
		"name":                "testvm",
		"account":             "testaccount",
		"created":             "2018-01-01T00:00:00+0000",
		"displayname":         "testvm",
		"domain":              "testdomain",
		"domainid":            "00000000-0000-0000-0000-000000000001",
		"hypervisor":          "KVM",
		"instancename":        "i-1-12345-VM",
		"templateid":          "00000000-0000-0000-0000-000000000004",
		"templatename":        "testtemplatename",
		"templatedisplaytext": "testtemplatetext",
		"userid":              "00000000-0000-0000-0000-000000000005",
		"username":            "testuser",
		"zoneid":              "00000000-0000-0000-0000-000000000006",
		"zonename":            "testzone",
	}

	var acc testutil.Accumulator

	cosmicT := &Cosmic{}

	cosmicT.ProcessVirtualMachineMetrics(&acc, sampleVirtualMachines)

	acc.AssertContainsTaggedFields(t, "cosmic_virtualmachine_metrics", expectedFields, expectedTags)
}

func TestGatherVolumeMetrics(t *testing.T) {
	var sampleVolumes = []*cosmic.Volume{
		{
			Account:                 "testaccount",
			Created:                 "2018-01-01T00:00:00+0000",
			Domain:                  "testdomain",
			Domainid:                "00000000-0000-0000-0000-000000000001",
			Hypervisor:              "KVM",
			Id:                      "00000000-0000-0000-0000-000000000002",
			Name:                    "testname",
			Zoneid:                  "00000000-0000-0000-0000-000000000003",
			Zonename:                "testzone",
			Attached:                "2018-01-01T00:00:00+0000",
			Destroyed:               false,
			Deviceid:                1,
			Diskofferingdisplaytext: "testdiskofferingtext",
			Diskofferingid:          "00000000-0000-0000-0000-000000000004",
			Diskofferingname:        "testdiskofferingname",
			Path:                    "00000000-0000-0000-0000-000000000005",
			Size:                    42949672960,
			State:                   "Ready",
			Storage:                 "teststorage",
			Storageid:               "00000000-0000-0000-0000-000000000006",
			Virtualmachineid:        "00000000-0000-0000-0000-000000000007",
			Vmdisplayname:           "testvmdisplayname",
			Vmname:                  "testvmname",
			Vmstate:                 "Running",
		},
	}

	expectedFields := map[string]interface{}{
		"attached":                "2018-01-01T00:00:00+0000",
		"destroyed":               false,
		"deviceid":                int64(1),
		"diskofferingdisplaytext": "testdiskofferingtext",
		"diskofferingid":          "00000000-0000-0000-0000-000000000004",
		"diskofferingname":        "testdiskofferingname",
		"path":                    "00000000-0000-0000-0000-000000000005",
		"size":                    int64(42949672960),
		"state":                   "Ready",
		"storage":                 "teststorage",
		"storageid":               "00000000-0000-0000-0000-000000000006",
		"virtualmachineid":        "00000000-0000-0000-0000-000000000007",
		"vmdisplayname":           "testvmdisplayname",
		"vmname":                  "testvmname",
		"vmstate":                 "Running",
	}

	expectedTags := map[string]string{
		"account":    "testaccount",
		"created":    "2018-01-01T00:00:00+0000",
		"domain":     "testdomain",
		"domainid":   "00000000-0000-0000-0000-000000000001",
		"hypervisor": "KVM",
		"id":         "00000000-0000-0000-0000-000000000002",
		"name":       "testname",
		"zoneid":     "00000000-0000-0000-0000-000000000003",
		"zonename":   "testzone",
	}

	var acc testutil.Accumulator

	cosmicT := &Cosmic{}

	cosmicT.ProcessVolumeMetrics(&acc, sampleVolumes)

	acc.AssertContainsTaggedFields(t, "cosmic_volume_metrics", expectedFields, expectedTags)
}

func TestGatherPublicIPMetricsMetrics(t *testing.T) {
	var samplePublicIpAddresses = []*cosmic.PublicIpAddress{
		{
			Account:          "testaccount",
			Aclid:            "00000000-0000-0000-0000-000000000004",
			Domain:           "testdomain",
			Domainid:         "00000000-0000-0000-0000-000000000001",
			Id:               "00000000-0000-0000-0000-000000000002",
			Ipaddress:        "1.2.3.4",
			State:            "Allocated",
			Virtualmachineid: "00000000-0000-0000-0000-000000000007",
			Vpcid:            "00000000-0000-0000-0000-000000000005",
			Zoneid:           "00000000-0000-0000-0000-000000000003",
			Zonename:         "testzone",
		},
	}

	expectedFields := map[string]interface{}{
		"aclid": "00000000-0000-0000-0000-000000000004",
		"state": "Allocated",
		"vpcid": "00000000-0000-0000-0000-000000000005",
	}

	expectedTags := map[string]string{
		"account":   "testaccount",
		"domain":    "testdomain",
		"domainid":  "00000000-0000-0000-0000-000000000001",
		"id":        "00000000-0000-0000-0000-000000000002",
		"ipaddress": "1.2.3.4",
		"zoneid":    "00000000-0000-0000-0000-000000000003",
		"zonename":  "testzone",
	}

	var acc testutil.Accumulator

	cosmicT := &Cosmic{}

	cosmicT.ProcessPublicIPMetrics(&acc, samplePublicIpAddresses)

	acc.AssertContainsTaggedFields(t, "cosmic_publicipaddress_metrics", expectedFields, expectedTags)
}
