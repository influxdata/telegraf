package vsan

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	vsanmethods "github.com/influxdata/telegraf/plugins/inputs/vsan/vsan-sdk/methods"
	vsantypes "github.com/influxdata/telegraf/plugins/inputs/vsan/vsan-sdk/types"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
	"time"
)

const (
	Namespace = "vsan"
	Path      = "/vsanHealth"
)

const (
	envURL      = "GOVMOMI_URL"
	envUserName = "GOVMOMI_USERNAME"
	envPassword = "GOVMOMI_PASSWORD"
	envInsecure = "GOVMOMI_INSECURE"
)

type Client struct {
	VCenter   string
	username  string
	password  string
	VimClient *vim25.Client
	Client    *soap.Client
}

var insecureDescription = fmt.Sprintf("Don't verify the server's certificate chain [%s]", envInsecure)

//var insecureFlag = flag.Bool("insecure", getEnvBool(envInsecure, false), insecureDescription)
var insecureFlag = flag.Bool("insecure", true, insecureDescription) //todo

// NewClient creates a Client for vSAN management
func NewClient(ctx context.Context, vcenter string, username string, password string) (*Client, error) {
	flag.Parse()
	var insecureFlag = true

	u, err := soap.ParseURL(vcenter + vim25.Path)
	if err != nil {
		return nil, err
	}

	// Override username and/or password as required
	u.User = url.UserPassword(username, password)

	// Connect and log in to ESX or vCenter
	govmomiClient, err := govmomi.NewClient(ctx, u, insecureFlag)
	if err != nil {
		return nil, err
	}
	soapClient := govmomiClient.Client.Client.NewServiceClient(Path, Namespace)
	client := &Client{
		VCenter:   vcenter,
		username:  username,
		password:  password,
		VimClient: govmomiClient.Client, // the vim client is used for querying cmmds
		Client:    soapClient,           // the soap client sends vsan request
	}
	return client, nil
}

func (c Client) QueryPerf(ctx context.Context, start time.Time, entityRefId string) (*vsantypes.VsanPerfQueryPerfResponse, error) {

	client := c.Client
	var perfSpecs []vsantypes.VsanPerfQuerySpec
	end := time.Now()
	perfSpec := vsantypes.VsanPerfQuerySpec{
		EntityRefId: fmt.Sprintf("%s:*", entityRefId),
		StartTime:   &start,
		EndTime:     &end,
	}
	perfSpecs = append(perfSpecs, perfSpec)

	cluster := vsantypes.ManagedObjectReference{
		Type:  "ClusterComputeResource",
		Value: "domain-c8",
	}

	perfManager := vsantypes.ManagedObjectReference{
		Type:  "VsanPerformanceManager",
		Value: "vsan-performance-manager",
	}

	perfRequest := vsantypes.VsanPerfQueryPerf{
		This:       perfManager,
		QuerySpecs: perfSpecs,
		Cluster:    &cluster,
	}

	res, err := vsanmethods.VsanPerfQueryPerf(ctx, client, &perfRequest)
	return res, err
}

func (c Client) QueryCmmds(ctx context.Context) (map[string]CmmdsEntity, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := c.VimClient

	clusterRef := types.ManagedObjectReference{
		Type:  "ClusterComputeResource",
		Value: "domain-c8",
	}

	clusterObj := object.NewClusterComputeResource(client, clusterRef)
	hosts, err := clusterObj.Hosts(ctx)

	if err != nil {
		log.Println("E! Error happen when get hosts: ", err)
		return nil, err
	}

	vis, err2 := hosts[0].ConfigManager().VsanInternalSystem(ctx)

	if err2 != nil {
		log.Println("E! Error happen when get VsanInternalSystem: ", err)
		return nil, err2
	}

	queries := make([]types.HostVsanInternalSystemCmmdsQuery, 2)
	hostnameCmmdsQuery := types.HostVsanInternalSystemCmmdsQuery{
		Type: "HOSTNAME",
	}

	diskCmmdsQuery := types.HostVsanInternalSystemCmmdsQuery{
		Type: "DISK",
	}

	queries = append(queries, hostnameCmmdsQuery)
	queries = append(queries, diskCmmdsQuery)

	request := types.QueryCmmds{
		This:    vis.Reference(),
		Queries: queries,
	}

	res, err := methods.QueryCmmds(ctx, client.RoundTripper, &request)
	if err != nil {
		log.Println("E! Query cmmds error: ", err)
		return nil, err
	}
	var clusterCmmds Cmmds

	err = json.Unmarshal([]byte(res.Returnval), &clusterCmmds)
	if err != nil {
		log.Println("E! Error when turning to json : ", err)
		return nil, err
	}

	cmmdsMap := make(map[string]CmmdsEntity)
	for _, entity := range clusterCmmds.Res {
		uuid := entity.UUID
		cmmdsMap[uuid] = entity
	}
	return cmmdsMap, nil
}

type CmmdsEntity struct {
	UUID    string      `json:"uuid"`
	Owner   string      `json:"owner"` // ESXi UUID
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Cmmds struct {
	Res []CmmdsEntity `json:"result"`
}
