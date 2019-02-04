package kube_inventory

import (
	"context"
	"time"

	"github.com/ericchiang/k8s"
	"github.com/ericchiang/k8s/apis/apps/v1beta1"
	"github.com/ericchiang/k8s/apis/apps/v1beta2"
	"github.com/ericchiang/k8s/apis/core/v1"

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

func (c *client) getDaemonSets(ctx context.Context) (*v1beta2.DaemonSetList, error) {
	list := new(v1beta2.DaemonSetList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}

func (c *client) getDeployments(ctx context.Context) (*v1beta1.DeploymentList, error) {
	list := &v1beta1.DeploymentList{}
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

func (c *client) getStatefulSets(ctx context.Context) (*v1beta1.StatefulSetList, error) {
	list := new(v1beta1.StatefulSetList)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return list, c.List(ctx, c.namespace, list)
}
