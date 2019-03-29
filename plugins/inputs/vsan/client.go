package vsan

import (
	"context"
	"flag"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"net/url"
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
	VCenter  string
	username string
	password string
	Client   *soap.Client
	Perf     *performance.Manager
}

var insecureDescription = fmt.Sprintf("Don't verify the server's certificate chain [%s]", envInsecure)
//var insecureFlag = flag.Bool("insecure", getEnvBool(envInsecure, false), insecureDescription)
var insecureFlag = flag.Bool("insecure", true, insecureDescription)

// NewClient creates a soap.Client for vSAN management
func NewVSANClient(ctx context.Context, vcenter string, username string, password string) (*Client, error) {
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
		VCenter:  vcenter,
		username: username,
		password: password,
		Client:   soapClient,
		Perf:     performance.NewManager(govmomiClient.Client),
	}
	return client, nil
}
