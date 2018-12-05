package kube_lite

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"
	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
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
	if *rList.GroupVersion == "" {
		return nil, &APIError{
			URL:        c.baseURL,
			StatusCode: http.StatusOK,
			Title:      "empty group version",
		}
	}
	return rList, nil
}

func (c *client) getDeployments(ctx context.Context) (list *v1beta1.DeploymentList, err error) {
	list = new(v1beta1.DeploymentList)
	if err = c.doGet(ctx, "/deployments/", list); err != nil {
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

func (c *client) getPods(ctx context.Context) (list *v1.PodList, err error) {
	list = new(v1.PodList)
	if err = c.doGet(ctx, "/pods/", list); err != nil {
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
