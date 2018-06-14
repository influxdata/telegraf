package kube_state

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/apps/v1beta2"
	autoscaling "k8s.io/api/autoscaling/v2beta1"
	v1batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type client struct {
	baseURL     string
	httpClient  *http.Client
	bearerToken string
	semaphore   chan struct{}
}

func newClient(baseURL string, timeout time.Duration, maxConns int, bearerToken string, tlsConfig *tls.Config) *client {
	return &client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    maxConns,
				TLSClientConfig: tlsConfig,
			},
			Timeout: timeout,
		},
		bearerToken: bearerToken,
		semaphore:   make(chan struct{}, maxConns),
	}
}

func (c *client) getAPIResourceList(ctx context.Context) (rList *metav1.APIResourceList, err error) {
	rList = new(metav1.APIResourceList)
	if err = c.doGet(ctx, "", rList); err != nil {
		return nil, err
	}
	if rList.GroupVersion == "" {
		return nil, &APIError{
			URL:        c.baseURL,
			StatusCode: http.StatusOK,
			Title:      "empty group version",
		}
	}
	return rList, nil
}

func (c *client) getConfigMaps(ctx context.Context) (list *v1.ConfigMapList, err error) {
	list = new(v1.ConfigMapList)
	if err = c.doGet(ctx, "/configmaps/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getCronJobs(ctx context.Context) (list *batchv1beta1.CronJobList, err error) {
	list = new(batchv1beta1.CronJobList)
	if err = c.doGet(ctx, "/cronjobs/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getDaemonSets(ctx context.Context) (list *v1beta2.DaemonSetList, err error) {
	list = new(v1beta2.DaemonSetList)
	if err = c.doGet(ctx, "/daemonsets/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getDeployments(ctx context.Context) (list *v1beta1.DeploymentList, err error) {
	list = new(v1beta1.DeploymentList)
	if err = c.doGet(ctx, "/deployments/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getEndpoints(ctx context.Context) (list *v1.EndpointsList, err error) {
	list = new(v1.EndpointsList)
	if err = c.doGet(ctx, "/endpoints/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getHorizontalPodAutoScalers(ctx context.Context) (list *autoscaling.HorizontalPodAutoscalerList, err error) {
	list = new(autoscaling.HorizontalPodAutoscalerList)
	if err = c.doGet(ctx, "/horizontalpodautoscalers/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getJobs(ctx context.Context) (list *v1batch.JobList, err error) {
	list = new(v1batch.JobList)
	if err = c.doGet(ctx, "/jobs/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getLimitRanges(ctx context.Context) (list *v1.LimitRangeList, err error) {
	list = new(v1.LimitRangeList)
	if err = c.doGet(ctx, "/limitranges/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getNamespaces(ctx context.Context) (list *v1.NamespaceList, err error) {
	list = new(v1.NamespaceList)
	if err = c.doGet(ctx, "/namespaces/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getNodes(ctx context.Context) (list *v1.NodeList, err error) {
	list = new(v1.NodeList)
	if err = c.doGet(ctx, "/nodes/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getPersistentVolumes(ctx context.Context) (list *v1.PersistentVolumeList, err error) {
	list = new(v1.PersistentVolumeList)
	if err = c.doGet(ctx, "/persistentvolumes/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getPersistentVolumeClaims(ctx context.Context) (list *v1.PersistentVolumeClaimList, err error) {
	list = new(v1.PersistentVolumeClaimList)
	if err = c.doGet(ctx, "/persistentvolumeclaims/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getPods(ctx context.Context) (list *v1.PodList, err error) {
	list = new(v1.PodList)
	if err = c.doGet(ctx, "/pods/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getResourceQuotas(ctx context.Context) (list *v1.ResourceQuotaList, err error) {
	list = new(v1.ResourceQuotaList)
	if err = c.doGet(ctx, "/resourcequotas/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getReplicaSets(ctx context.Context) (list *v1beta2.ReplicaSetList, err error) {
	list = new(v1beta2.ReplicaSetList)
	if err = c.doGet(ctx, "/replicasets/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getReplicationControllers(ctx context.Context) (list *v1.ReplicationControllerList, err error) {
	list = new(v1.ReplicationControllerList)
	if err = c.doGet(ctx, "/replicationcontrollers/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getSecrets(ctx context.Context) (list *v1.SecretList, err error) {
	list = new(v1.SecretList)
	if err = c.doGet(ctx, "/secrets/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getServices(ctx context.Context) (list *v1.ServiceList, err error) {
	list = new(v1.ServiceList)
	if err = c.doGet(ctx, "/services/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) getStatefulSets(ctx context.Context) (list *v1beta1.StatefulSetList, err error) {
	list = new(v1beta1.StatefulSetList)
	if err = c.doGet(ctx, "/statefulsets/", list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) doGet(ctx context.Context, url string, v interface{}) error {
	req, err := createGetRequest(c.baseURL+url, c.bearerToken)
	if err != nil {
		return err
	}
	select {
	case c.semaphore <- struct{}{}:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		<-c.semaphore
		return err
	}
	defer func() {
		resp.Body.Close()
		<-c.semaphore
	}()

	// Clear invalid token if unauthorized
	if resp.StatusCode == http.StatusUnauthorized {
		c.bearerToken = ""
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func createGetRequest(url string, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Add("Accept", "application/json")

	return req, nil
}

type APIError struct {
	URL         string
	StatusCode  int
	Title       string
	Description string
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("[%s] %s: %s", e.URL, e.Title, e.Description)
	}
	return fmt.Sprintf("[%s] %s", e.URL, e.Title)
}
