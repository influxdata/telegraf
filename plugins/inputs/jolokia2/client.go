package jolokia2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type Client struct {
	URL    string
	client *http.Client
	config *ClientConfig
}

type ClientConfig struct {
	ResponseTimeout time.Duration
	Username        string
	Password        string
	ProxyConfig     *ProxyConfig
	tls.ClientConfig
}

type ProxyConfig struct {
	DefaultTargetUsername string
	DefaultTargetPassword string
	Targets               []ProxyTargetConfig
}

type ProxyTargetConfig struct {
	Username string
	Password string
	URL      string
}

type ReadRequest struct {
	Mbean      string
	Attributes []string
	Path       string
}

type ReadResponse struct {
	Status            int
	Value             interface{}
	RequestMbean      string
	RequestAttributes []string
	RequestPath       string
	RequestTarget     string
}

// Jolokia JSON request object. Example: {
//   "type": "read",
//   "mbean: "java.lang:type="Runtime",
//   "attribute": "Uptime",
//   "target": {
//     "url: "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//   }
// }
type jolokiaRequest struct {
	Type      string         `json:"type"`
	Mbean     string         `json:"mbean"`
	Attribute interface{}    `json:"attribute,omitempty"`
	Path      string         `json:"path,omitempty"`
	Target    *jolokiaTarget `json:"target,omitempty"`
}

type jolokiaTarget struct {
	URL      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// Jolokia JSON response object. Example: {
//   "request": {
//     "type": "read"
//     "mbean": "java.lang:type=Runtime",
//     "attribute": "Uptime",
//     "target": {
//       "url": "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//     }
//   },
//   "value": 1214083,
//   "timestamp": 1488059309,
//   "status": 200
// }
type jolokiaResponse struct {
	Request jolokiaRequest `json:"request"`
	Value   interface{}    `json:"value"`
	Status  int            `json:"status"`
}

func NewClient(url string, config *ClientConfig) (*Client, error) {
	tlsConfig, err := config.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: config.ResponseTimeout,
		TLSClientConfig:       tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.ResponseTimeout,
	}

	return &Client{
		URL:    url,
		config: config,
		client: client,
	}, nil
}

func (c *Client) read(requests []ReadRequest) ([]ReadResponse, error) {
	jrequests := makeJolokiaRequests(requests, c.config.ProxyConfig)
	requestBody, err := json.Marshal(jrequests)
	if err != nil {
		return nil, err
	}

	requestUrl, err := formatReadUrl(c.URL, c.config.Username, c.config.Password)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("unable to create new request '%s': %s", requestUrl, err)
	}

	req.Header.Add("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			c.URL, resp.StatusCode, http.StatusText(resp.StatusCode), http.StatusOK, http.StatusText(http.StatusOK))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jresponses []jolokiaResponse
	if err = json.Unmarshal([]byte(responseBody), &jresponses); err != nil {
		return nil, fmt.Errorf("Error decoding JSON response: %s: %s", err, responseBody)
	}

	return makeReadResponses(jresponses), nil
}

func makeJolokiaRequests(rrequests []ReadRequest, proxyConfig *ProxyConfig) []jolokiaRequest {
	jrequests := make([]jolokiaRequest, 0)
	if proxyConfig == nil {
		for _, rr := range rrequests {
			jrequests = append(jrequests, makeJolokiaRequest(rr, nil))
		}
	} else {
		for _, t := range proxyConfig.Targets {
			if t.Username == "" {
				t.Username = proxyConfig.DefaultTargetUsername
			}
			if t.Password == "" {
				t.Password = proxyConfig.DefaultTargetPassword
			}

			for _, rr := range rrequests {
				jtarget := &jolokiaTarget{
					URL:      t.URL,
					User:     t.Username,
					Password: t.Password,
				}

				jrequests = append(jrequests, makeJolokiaRequest(rr, jtarget))
			}
		}
	}

	return jrequests
}

func makeJolokiaRequest(rrequest ReadRequest, jtarget *jolokiaTarget) jolokiaRequest {
	jrequest := jolokiaRequest{
		Type:   "read",
		Mbean:  rrequest.Mbean,
		Path:   rrequest.Path,
		Target: jtarget,
	}

	if len(rrequest.Attributes) == 1 {
		jrequest.Attribute = rrequest.Attributes[0]
	}
	if len(rrequest.Attributes) > 1 {
		jrequest.Attribute = rrequest.Attributes
	}

	return jrequest
}

func makeReadResponses(jresponses []jolokiaResponse) []ReadResponse {
	rresponses := make([]ReadResponse, 0)

	for _, jr := range jresponses {
		rrequest := ReadRequest{
			Mbean:      jr.Request.Mbean,
			Path:       jr.Request.Path,
			Attributes: []string{},
		}

		attrValue := jr.Request.Attribute
		if attrValue != nil {
			attribute, ok := attrValue.(string)
			if ok {
				rrequest.Attributes = []string{attribute}
			} else {
				attributes, _ := attrValue.([]interface{})
				rrequest.Attributes = make([]string, len(attributes))
				for i, attr := range attributes {
					rrequest.Attributes[i] = attr.(string)
				}
			}
		}
		rresponse := ReadResponse{
			Value:             jr.Value,
			Status:            jr.Status,
			RequestMbean:      rrequest.Mbean,
			RequestAttributes: rrequest.Attributes,
			RequestPath:       rrequest.Path,
		}
		if jtarget := jr.Request.Target; jtarget != nil {
			rresponse.RequestTarget = jtarget.URL
		}

		rresponses = append(rresponses, rresponse)
	}

	return rresponses
}

func formatReadUrl(configUrl, username, password string) (string, error) {
	parsedUrl, err := url.Parse(configUrl)
	if err != nil {
		return "", err
	}

	readUrl := url.URL{
		Host:   parsedUrl.Host,
		Scheme: parsedUrl.Scheme,
	}

	if username != "" || password != "" {
		readUrl.User = url.UserPassword(username, password)
	}

	readUrl.Path = path.Join(parsedUrl.Path, "read")
	readUrl.Query().Add("ignoreErrors", "true")
	return readUrl.String(), nil
}
