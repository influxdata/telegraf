/*
Package k8s implements a Kubernetes client.

	import (
		"context"

		"github.com/ericchiang/k8s"
		appsv1 "github.com/ericchiang/k8s/apis/apps/v1"
	)

	func listDeployments(ctx context.Context) (*appsv1.DeploymentList, error) {
		c, err := k8s.NewInClusterClient()
		if err != nil {
			return nil, err
		}

		var deployments appsv1.DeploymentList
		if err := c.List(ctx, "my-namespace", &deployments); err != nil {
			return nil, err
		}
		return deployments, nil
	}

*/
package k8s

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/http2"

	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

const (
	// AllNamespaces is given to list and watch operations to signify that the code should
	// list or watch resources in all namespaces.
	AllNamespaces = allNamespaces
	// Actual definition is private in case we want to change it later.
	allNamespaces = ""

	namespaceDefault = "default"
)

// String returns a pointer to a string. Useful for creating API objects
// that take pointers instead of literals.
func String(s string) *string { return &s }

// Int is a convenience for converting an int literal to a pointer to an int.
func Int(i int) *int { return &i }

// Int32 is a convenience for converting an int32 literal to a pointer to an int32.
func Int32(i int32) *int32 { return &i }

// Bool is a convenience for converting a bool literal to a pointer to a bool.
func Bool(b bool) *bool { return &b }

const (
	// Types for watch events.
	EventAdded    = "ADDED"
	EventDeleted  = "DELETED"
	EventModified = "MODIFIED"
	EventError    = "ERROR"
)

// Client is a Kuberntes client.
type Client struct {
	// The URL of the API server.
	Endpoint string

	// Namespace is the name fo the default reconciled from the client's config.
	// It is set when constructing a client using NewClient(), and defaults to
	// the value "default".
	//
	// This value should be used to access the client's default namespace. For
	// example, to create a configmap in the default namespace, use client.Namespace
	// when to fill the ObjectMeta:
	//
	//		client, err := k8s.NewClient(config)
	//		if err != nil {
	//			// handle error
	//		}
	//		cm := v1.ConfigMap{
	//			Metadata: &metav1.ObjectMeta{
	//				Name:      &k8s.String("my-configmap"),
	//				Namespace: &client.Namespace,
	//			},
	//			Data: map[string]string{"foo": "bar", "spam": "eggs"},
	//		}
	//		err := client.Create(ctx, cm)
	//
	Namespace string

	// SetHeaders provides a hook for modifying the HTTP headers of all requests.
	//
	//		client, err := k8s.NewClient(config)
	//		if err != nil {
	//			// handle error
	//		}
	//		client.SetHeaders = func(h http.Header) error {
	//			h.Set("Authorization", "Bearer "+mytoken)
	//			return nil
	//		}
	//
	SetHeaders func(h http.Header) error

	Client *http.Client
}

func (c *Client) newRequest(ctx context.Context, verb, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		return nil, err
	}
	if c.SetHeaders != nil {
		if err := c.SetHeaders(req.Header); err != nil {
			return nil, err
		}
	}
	return req.WithContext(ctx), nil
}

// NewClient initializes a client from a client config.
func NewClient(config *Config) (*Client, error) {
	if len(config.Contexts) == 0 {
		if config.CurrentContext != "" {
			return nil, fmt.Errorf("no contexts with name %q", config.CurrentContext)
		}

		if n := len(config.Clusters); n == 0 {
			return nil, errors.New("no clusters provided")
		} else if n > 1 {
			return nil, errors.New("multiple clusters but no current context")
		}
		if n := len(config.AuthInfos); n == 0 {
			return nil, errors.New("no users provided")
		} else if n > 1 {
			return nil, errors.New("multiple users but no current context")
		}

		return newClient(config.Clusters[0].Cluster, config.AuthInfos[0].AuthInfo, namespaceDefault)
	}

	var ctx Context
	if config.CurrentContext == "" {
		if n := len(config.Contexts); n == 0 {
			return nil, errors.New("no contexts provided")
		} else if n > 1 {
			return nil, errors.New("multiple contexts but no current context")
		}
		ctx = config.Contexts[0].Context
	} else {
		for _, c := range config.Contexts {
			if c.Name == config.CurrentContext {
				ctx = c.Context
				goto configFound
			}
		}
		return nil, fmt.Errorf("no config named %q", config.CurrentContext)
	configFound:
	}

	if ctx.Cluster == "" {
		return nil, fmt.Errorf("context doesn't have a cluster")
	}
	if ctx.AuthInfo == "" {
		return nil, fmt.Errorf("context doesn't have a user")
	}
	var (
		user    AuthInfo
		cluster Cluster
	)

	for _, u := range config.AuthInfos {
		if u.Name == ctx.AuthInfo {
			user = u.AuthInfo
			goto userFound
		}
	}
	return nil, fmt.Errorf("no user named %q", ctx.AuthInfo)
userFound:

	for _, c := range config.Clusters {
		if c.Name == ctx.Cluster {
			cluster = c.Cluster
			goto clusterFound
		}
	}
	return nil, fmt.Errorf("no cluster named %q", ctx.Cluster)
clusterFound:

	namespace := ctx.Namespace
	if namespace == "" {
		namespace = namespaceDefault
	}

	return newClient(cluster, user, namespace)
}

// NewInClusterClient returns a client that uses the service account bearer token mounted
// into Kubernetes pods.
func NewInClusterClient() (*Client, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, err
	}

	cluster := Cluster{
		Server:               "https://" + host + ":" + port,
		CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
	}
	user := AuthInfo{TokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token"}
	return newClient(cluster, user, string(namespace))
}

func load(filepath string, data []byte) (out []byte, err error) {
	if filepath != "" {
		data, err = ioutil.ReadFile(filepath)
	}
	return data, err
}

func newClient(cluster Cluster, user AuthInfo, namespace string) (*Client, error) {
	if cluster.Server == "" {
		// NOTE: kubectl defaults to localhost:8080, but it's probably better to just
		// be strict.
		return nil, fmt.Errorf("no cluster endpoint provided")
	}

	ca, err := load(cluster.CertificateAuthority, cluster.CertificateAuthorityData)
	if err != nil {
		return nil, fmt.Errorf("loading certificate authority: %v", err)
	}

	clientCert, err := load(user.ClientCertificate, user.ClientCertificateData)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %v", err)
	}
	clientKey, err := load(user.ClientKey, user.ClientKeyData)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %v", err)
	}

	// See https://github.com/gtank/cryptopasta
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: cluster.InsecureSkipTLSVerify,
	}

	if len(ca) != 0 {
		tlsConfig.RootCAs = x509.NewCertPool()
		if !tlsConfig.RootCAs.AppendCertsFromPEM(ca) {
			return nil, errors.New("certificate authority doesn't contain any certificates")
		}
	}
	if len(clientCert) != 0 {
		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, fmt.Errorf("invalid client cert and key pair: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	token := user.Token
	if user.TokenFile != "" {
		data, err := ioutil.ReadFile(user.TokenFile)
		if err != nil {
			return nil, fmt.Errorf("load token file: %v", err)
		}
		token = string(data)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       tlsConfig,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	client := &Client{
		Endpoint:  cluster.Server,
		Namespace: namespace,
		Client: &http.Client{
			Transport: transport,
		},
	}

	if token != "" {
		client.SetHeaders = func(h http.Header) error {
			h.Set("Authorization", "Bearer "+token)
			return nil
		}
	}
	if user.Username != "" && user.Password != "" {
		auth := user.Username + ":" + user.Password
		auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		client.SetHeaders = func(h http.Header) error {
			h.Set("Authorization", auth)
			return nil
		}
	}
	return client, nil
}

// APIError is an error from a unexpected status code.
type APIError struct {
	// The status object returned by the Kubernetes API,
	Status *metav1.Status

	// Status code returned by the HTTP request.
	//
	// NOTE: For some reason the value set in Status.Code
	// doesn't correspond to the HTTP status code. Possibly
	// a bug?
	Code int
}

func (e *APIError) Error() string {
	if e.Status != nil && e.Status.Message != nil && e.Status.Status != nil {
		return fmt.Sprintf("kubernetes api: %s %d %s", *e.Status.Status, e.Code, *e.Status.Message)
	}
	return fmt.Sprintf("%#v", e)
}

func checkStatusCode(contentType string, statusCode int, body []byte) error {
	if statusCode/100 == 2 {
		return nil
	}

	return newAPIError(contentType, statusCode, body)
}

func newAPIError(contentType string, statusCode int, body []byte) error {
	status := new(metav1.Status)
	if err := unmarshal(body, contentType, status); err != nil {
		return fmt.Errorf("decode error status %d: %v", statusCode, err)
	}
	return &APIError{status, statusCode}
}

func (c *Client) client() *http.Client {
	if c.Client == nil {
		return http.DefaultClient
	}
	return c.Client
}

// Create creates a resource of a registered type. The API version and resource
// type is determined by the type of the req argument. The result is unmarshaled
// into req.
//
//		configMap := corev1.ConfigMap{
//			Metadata: &metav1.ObjectMeta{
//				Name:      k8s.String("my-configmap"),
//				Namespace: k8s.String("my-namespace"),
//			},
//			Data: map[string]string{
//				"my-key": "my-val",
//			},
//		}
//		if err := client.Create(ctx, &configMap); err != nil {
//			// handle error
//		}
//		// resource is updated with response of create request
//		fmt.Println(conifgMap.Metaata.GetCreationTimestamp())
//
func (c *Client) Create(ctx context.Context, req Resource, options ...Option) error {
	url, err := resourceURL(c.Endpoint, req, false, options...)
	if err != nil {
		return err
	}
	return c.do(ctx, "POST", url, req, req)
}

func (c *Client) Delete(ctx context.Context, req Resource, options ...Option) error {
	url, err := resourceURL(c.Endpoint, req, true, options...)
	if err != nil {
		return err
	}
	o := &deleteOptions{
		Kind:              "DeleteOptions",
		APIVersion:        "v1",
		PropagationPolicy: "Background",
	}
	for _, option := range options {
		option.updateDelete(req, o)
	}

	return c.do(ctx, "DELETE", url, o, nil)
}

func (c *Client) Update(ctx context.Context, req Resource, options ...Option) error {
	url, err := resourceURL(c.Endpoint, req, true, options...)
	if err != nil {
		return err
	}
	return c.do(ctx, "PUT", url, req, req)
}

func (c *Client) Get(ctx context.Context, namespace, name string, resp Resource, options ...Option) error {
	url, err := resourceGetURL(c.Endpoint, namespace, name, resp, options...)
	if err != nil {
		return err
	}
	return c.do(ctx, "GET", url, nil, resp)
}

func (c *Client) List(ctx context.Context, namespace string, resp ResourceList, options ...Option) error {
	url, err := resourceListURL(c.Endpoint, namespace, resp, options...)
	if err != nil {
		return err
	}
	return c.do(ctx, "GET", url, nil, resp)
}

func (c *Client) do(ctx context.Context, verb, url string, req, resp interface{}) error {
	var (
		contentType string
		body        io.Reader
	)
	if req != nil {
		ct, data, err := marshal(req)
		if err != nil {
			return fmt.Errorf("encoding object: %v", err)
		}
		contentType = ct
		body = bytes.NewReader(data)
	}
	r, err := http.NewRequest(verb, url, body)
	if err != nil {
		return fmt.Errorf("new request: %v", err)
	}
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
		r.Header.Set("Accept", contentType)
	} else if resp != nil {
		r.Header.Set("Accept", contentTypeFor(resp))
	}
	if c.SetHeaders != nil {
		c.SetHeaders(r.Header)
	}

	re, err := c.client().Do(r)
	if err != nil {
		return fmt.Errorf("performing request: %v", err)
	}
	defer re.Body.Close()

	respBody, err := ioutil.ReadAll(re.Body)
	if err != nil {
		return fmt.Errorf("read body: %v", err)
	}

	respCT := re.Header.Get("Content-Type")
	if err := checkStatusCode(respCT, re.StatusCode, respBody); err != nil {
		return err
	}
	if resp != nil {
		if err := unmarshal(respBody, respCT, resp); err != nil {
			return fmt.Errorf("decode response: %v", err)
		}
	}
	return nil
}
