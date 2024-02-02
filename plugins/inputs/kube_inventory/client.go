package kube_inventory

import (
	"context"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

type client struct {
	namespace string
	timeout   time.Duration
	*kubernetes.Clientset
}

func newClient(baseURL, namespace, bearerTokenFile string, bearerToken string, timeout time.Duration, tlsConfig tls.ClientConfig) (*client, error) {
	var clientConfig *rest.Config
	var err error

	if baseURL == "" {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		clientConfig = &rest.Config{
			TLSClientConfig: rest.TLSClientConfig{
				ServerName: tlsConfig.ServerName,
				Insecure:   tlsConfig.InsecureSkipVerify,
				CAFile:     tlsConfig.TLSCA,
				CertFile:   tlsConfig.TLSCert,
				KeyFile:    tlsConfig.TLSKey,
			},
			Host:          baseURL,
			ContentConfig: rest.ContentConfig{},
		}

		if bearerTokenFile != "" {
			clientConfig.BearerTokenFile = bearerTokenFile
		} else if bearerToken != "" {
			clientConfig.BearerToken = bearerToken
		}
	}

	c, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		Clientset: c,
		timeout:   timeout,
		namespace: namespace,
	}, nil
}

func newHTTPClient(tlsConfig tls.ClientConfig, bearerTokenFile string, responseTimeout config.Duration) (*http.Client, error) {
	tlsCfg, err := tlsConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	clientConfig := &rest.Config{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		ContentConfig:   rest.ContentConfig{},
		Timeout:         time.Duration(responseTimeout),
		BearerTokenFile: bearerTokenFile,
	}
	return rest.HTTPClientFor(clientConfig)
}
func (c *client) getDaemonSets(ctx context.Context) (*appsv1.DaemonSetList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.AppsV1().DaemonSets(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getDeployments(ctx context.Context) (*appsv1.DeploymentList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getEndpoints(ctx context.Context) (*corev1.EndpointsList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.CoreV1().Endpoints(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getIngress(ctx context.Context) (*netv1.IngressList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.NetworkingV1().Ingresses(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getNodes(ctx context.Context, name string) (*corev1.NodeList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	var fieldSelector string
	if name != "" {
		fieldSelector = "metadata.name=" + name
	}
	return c.CoreV1().Nodes().List(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
}

func (c *client) getPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
}

func (c *client) getPersistentVolumeClaims(ctx context.Context) (*corev1.PersistentVolumeClaimList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.CoreV1().PersistentVolumeClaims(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getPods(ctx context.Context, nodeName string) (*corev1.PodList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var fieldSelector string
	if nodeName != "" {
		fieldSelector = "spec.nodeName=" + nodeName
	}
	return c.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
}

func (c *client) getServices(ctx context.Context) (*corev1.ServiceList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getStatefulSets(ctx context.Context) (*appsv1.StatefulSetList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.AppsV1().StatefulSets(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getResourceQuotas(ctx context.Context) (*corev1.ResourceQuotaList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.CoreV1().ResourceQuotas(c.namespace).List(ctx, metav1.ListOptions{})
}

func (c *client) getTLSSecrets(ctx context.Context) (*corev1.SecretList, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"type": "kubernetes.io/tls"}}
	return c.CoreV1().Secrets(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: labels.Set(labelSelector.MatchLabels).String(),
	})
}
