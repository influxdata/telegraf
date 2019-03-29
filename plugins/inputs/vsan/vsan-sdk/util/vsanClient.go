package util

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	vimtypes "github.com/vmware/govmomi/vim25/types"
	"gitlab.eng.vmware.com/vsan-analytics/vsanPerfGo/pkg/vsphere"
	"gitlab.eng.vmware.com/vsan-analytics/vsanPerfGo/vsan/methods"
	vsantypes "gitlab.eng.vmware.com/vsan-analytics/vsanPerfGo/vsan/types"
	"os"
	"time"
)

const (
	Namespace = "vsan"
	Path      = "/vsanHealth"
)

var (
	VsanPerformanceManagerInstance = vimtypes.ManagedObjectReference{
		Type:  "VsanPerformanceManager",
		Value: "vsan-performance-manager",
	}

	VimClusterVsanVcDiskManagementSystemInstance = vimtypes.ManagedObjectReference{
		Type:  "VimClusterVsanVcDiskManagementSystem",
		Value: "vsan-disk-management-system",
	}
)

type VsanClient struct {
	*soap.Client
}

func buildVsanClient(ctx context.Context, c *vim25.Client) *VsanClient {
	sc := c.Client.NewServiceClient(Path, Namespace)
	return &VsanClient{sc}
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
	os.Exit(1)
}

func vcenterClient() vsphere.VirtualCenter {
	return vsphere.VirtualCenter{
		Config: buildVcenterConfig(),
	}
}

var vc = connectVcenter()

func queryVsanPerfData() *vsantypes.VsanPerfQueryPerfResponse {
	var querySpecs []vsantypes.VsanPerfQuerySpec
	startTime := time.Now()
	vsanPerfSpec := vsantypes.VsanPerfQuerySpec{
		EntityRefId: "host-domclient:*",
		StartTime:   &startTime,
		//EndTime     : time.Now().Unix()
	}
	querySpecs = append(querySpecs, vsanPerfSpec)

	cluster := vimtypes.ManagedObjectReference{
		Type:  "ClusterComputeResource",
		Value: "domain-c7",
	}
	vsanPerfQueryPerf := vsantypes.VsanPerfQueryPerf{
		This:       VsanPerformanceManagerInstance,
		QuerySpecs: querySpecs,
		Cluster:    cluster,
	}
	res, err := methods.VsanPerfQueryPerf(ctx, vsanClient.Client, &vsanPerfQueryPerf)
	if err != nil {
		exit(err)
	}
	fmt.Fprintf(os.Stdout, "res: %+v\n", res)
	return res
}

func buildVcenterConfig() *vsphere.VirtualCenterConfig {
	return &vsphere.VirtualCenterConfig{
		Host:     "10.172.199.160",
		Port:     443,
		Username: "Administrator@vsphere.local",
		Password: "Admin!23",
		Insecure: true,
	}
}

var vsanClient = buildVsanClient(ctx, vc.Client.Client)

func queryDiskMappings(hostMoRef vimtypes.ManagedObjectReference) *[]vsantypes.DiskMapInfoEx {
	queryDiskMappings := vsantypes.QueryDiskMappings{
		This: VimClusterVsanVcDiskManagementSystemInstance,
		Host: hostMoRef,
	}
	res, err := methods.QueryDiskMappings(ctx, vsanClient.Client, &queryDiskMappings)
	if err != nil {
		exit(err)
	}
	fmt.Fprintf(os.Stdout, "res: %+v\n", res)
	return &res.Returnval
}

var ctx, _ = context.WithCancel(context.Background())

func connectVcenter() vsphere.VirtualCenter {
	vcenter := vcenterClient()

	err := vcenter.Connect(ctx)
	if err != nil {
		exit(err)
	}

	vcenter.Client.Version = "6.7"
	return vcenter
}
