package k8s

import (
	"context"
	"path"

	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

type Version struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// Discovery is a client used to determine the API version and supported
// resources of the server.
type Discovery struct {
	client *Client
}

func NewDiscoveryClient(c *Client) *Discovery {
	return &Discovery{c}
}

func (d *Discovery) get(ctx context.Context, path string, resp interface{}) error {
	return d.client.do(ctx, "GET", urlForPath(d.client.Endpoint, path), nil, resp)
}

func (d *Discovery) Version(ctx context.Context) (*Version, error) {
	var v Version
	if err := d.get(ctx, "version", &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (d *Discovery) APIGroups(ctx context.Context) (*metav1.APIGroupList, error) {
	var groups metav1.APIGroupList
	if err := d.get(ctx, "apis", &groups); err != nil {
		return nil, err
	}
	return &groups, nil
}

func (d *Discovery) APIGroup(ctx context.Context, name string) (*metav1.APIGroup, error) {
	var group metav1.APIGroup
	if err := d.get(ctx, path.Join("apis", name), &group); err != nil {
		return nil, err
	}
	return &group, nil
}

func (d *Discovery) APIResources(ctx context.Context, groupName, groupVersion string) (*metav1.APIResourceList, error) {
	var list metav1.APIResourceList
	if err := d.get(ctx, path.Join("apis", groupName, groupVersion), &list); err != nil {
		return nil, err
	}
	return &list, nil
}
