package kube_inventory

import (
	"context"
	"time"

	"github.com/ericchiang/k8s"
	v1APPS "github.com/ericchiang/k8s/apis/apps/v1"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
	v1beta1EXT "github.com/ericchiang/k8s/apis/extensions/v1beta1"

	"github.com/influxdata/telegraf/internal/tls"
)

type client struct {
	namespace string
	timeout   time.Duration
	*k8s.Client
}

func newClient(baseURL, namespace, bearerToken string, timeout time.Duration, tlsConfig tls.ClientConfig) (*client, error) {
	c, err := k8s.NewClient(&k8s.Config{
		Clusters: []k8s.NamedCluster{{Name: "cluster", Cluster: k8s.Cluster{
			Server:                baseURL,
			InsecureSkipTLSVerify: tlsConfig.InsecureSkipVerify,
			CertificateAuthority:  tlsConfig.TLSCA,
		}}},
		Contexts: []k8s.NamedContext{{Name: "context", Context: k8s.Context{
			Cluster:   "cluster",
			AuthInfo:  "auth",
			Namespace: namespace,
		}}},
		AuthInfos: []k8s.NamedAuthInfo{{Name: "auth", AuthInfo: k8s.AuthInfo{
			Token:             bearerToken,
			ClientCertificate: tlsConfig.TLSCert,
			ClientKey:         tlsConfig.TLSKey,
		}}},
	})
	if err != nil {
		return nil, err
	}

	return &client{
		Client:    c,
		timeout:   timeout,
		namespace: namespace,
	}, nil
}

func (c *client) getDaemonSets(ctx context.Context) (*v1APPS.DaemonSetList, error) {
	list := new(v1APPS.DaemonSetList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getDeployments(ctx context.Context) (*v1APPS.DeploymentList, error) {
	list := &v1APPS.DeploymentList{}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getEndpoints(ctx context.Context) (*v1.EndpointsList, error) {
	list := new(v1.EndpointsList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getIngress(ctx context.Context) (*v1beta1EXT.IngressList, error) {
	list := new(v1beta1EXT.IngressList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getNodes(ctx context.Context) (*v1.NodeList, error) {
	list := new(v1.NodeList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, "", list)
}

func (c *client) getPersistentVolumes(ctx context.Context) (*v1.PersistentVolumeList, error) {
	list := new(v1.PersistentVolumeList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, "", list)
}

func (c *client) getPersistentVolumeClaims(ctx context.Context) (*v1.PersistentVolumeClaimList, error) {
	list := new(v1.PersistentVolumeClaimList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getPods(ctx context.Context) (*v1.PodList, error) {
	list := new(v1.PodList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getServices(ctx context.Context) (*v1.ServiceList, error) {
	list := new(v1.ServiceList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getStatefulSets(ctx context.Context) (*v1APPS.StatefulSetList, error) {
	list := new(v1APPS.StatefulSetList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}
