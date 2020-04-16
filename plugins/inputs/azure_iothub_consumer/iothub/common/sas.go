package common

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParseConnectionString parses the given connection string into a key-value map,
// returns an error if at least one of required keys is missing.
func ParseConnectionString(cs string, require ...string) (map[string]string, error) {
	m := map[string]string{}
	for _, s := range strings.Split(cs, ";") {
		if s == "" {
			continue
		}
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("malformed connection string")
		}
		m[kv[0]] = kv[1]
	}
	for _, k := range require {
		if s := m[k]; s == "" {
			return nil, fmt.Errorf("%s is required", k)
		}
	}
	return m, nil
}

func GetEdgeModuleEnvironmentVariables() (map[string]string, error) {
	m := map[string]string{}

	require := []string{
		"ContainerHostName",
		"IOTHubHostName",
		"GatewayHostName",
		"DeviceID",
		"ModuleID",
		"GenerationID",
		"WorkloadAPI",
		"APIVersion",
	}

	m["ContainerHostName"] = os.Getenv("HOSTNAME")
	m["IOTHubHostName"] = os.Getenv("IOTEDGE_IOTHUBHOSTNAME")
	m["GatewayHostName"] = os.Getenv("IOTEDGE_GATEWAYHOSTNAME")
	m["DeviceID"] = os.Getenv("IOTEDGE_DEVICEID")
	m["ModuleID"] = os.Getenv("IOTEDGE_MODULEID")
	m["GenerationID"] = os.Getenv("IOTEDGE_MODULEGENERATIONID")
	m["WorkloadAPI"] = os.Getenv("IOTEDGE_WORKLOADURI")
	m["APIVersion"] = os.Getenv("IOTEDGE_APIVERSION")

	for _, k := range require {
		if s := m[k]; s == "" {
			return nil, fmt.Errorf("%s is required", k)
		}
	}
	return m, nil
}

// NewSharedAccessKey creates new shared access key for subsequent token generation.
func NewSharedAccessKey(hostname, policy, key string) *SharedAccessKey {
	return &SharedAccessKey{
		HostName:            hostname,
		SharedAccessKeyName: policy,
		SharedAccessKey:     key,
	}
}

// SharedAccessKey is SAS token generator.
type SharedAccessKey struct {
	HostName            string
	SharedAccessKeyName string
	SharedAccessKey     string
}

// Token generates a shared access signature for the named resource and lifetime.
func (c *SharedAccessKey) Token(
	resource string, lifetime time.Duration,
) (*SharedAccessSignature, error) {
	return NewSharedAccessSignature(
		resource, c.SharedAccessKeyName, c.SharedAccessKey, time.Now().Add(lifetime),
	)
}

// NewSharedAccessSignature initialized a new shared access signature
// and generates signature fields based on the given input.
func NewSharedAccessSignature(
	resource, policy, key string, expiry time.Time,
) (*SharedAccessSignature, error) {
	sig, err := mksig(resource, key, expiry)
	if err != nil {
		return nil, err
	}
	return &SharedAccessSignature{
		Sr:  resource,
		Sig: sig,
		Se:  expiry,
		Skn: policy,
	}, nil
}

func mksig(sr, key string, se time.Time) (string, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, b)
	if _, err := fmt.Fprintf(h, "%s\n%d", url.QueryEscape(sr), se.Unix()); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// SharedAccessSignature is a shared access signature instance.
type SharedAccessSignature struct {
	Sr  string
	Sig string
	Se  time.Time
	Skn string
}

// String converts the signature to a token string.
func (sas *SharedAccessSignature) String() string {
	s := "SharedAccessSignature " +
		"sr=" + url.QueryEscape(sas.Sr) +
		"&sig=" + url.QueryEscape(sas.Sig) +
		"&se=" + url.QueryEscape(strconv.FormatInt(sas.Se.Unix(), 10))
	if sas.Skn != "" {
		s += "&skn=" + url.QueryEscape(sas.Skn)
	}
	return s
}

// EDGE MODULE AUTOMATIC AUTHENTICATION

// TokenFromEdge generates a shared access signature for the named resource and lifetime using the Workload API sign endpoint
func (c *SharedAccessKey) TokenFromEdge(
	workloadURI, module, genid, resource string, lifetime time.Duration,
) (*SharedAccessSignature, error) {
	return NewSharedAccessSignatureFromEdge(
		workloadURI, module, genid, resource, time.Now().Add(lifetime),
	)
}

// NewSharedAccessSignature initialized a new shared access signature
// and generates signature fields based on the given input.
func NewSharedAccessSignatureFromEdge(
	workloadURI, module, genid, resource string, expiry time.Time,
) (*SharedAccessSignature, error) {
	sig, err := mksigViaEdge(workloadURI, resource, module, genid, expiry)
	if err != nil {
		return nil, err
	}
	return &SharedAccessSignature{
		Sr:  resource,
		Sig: sig,
		Se:  expiry,
	}, nil
}

func mksigViaEdge(workloadURI, resource, module, genid string, se time.Time) (string, error) {
	data := url.QueryEscape(resource) + "\n" + strconv.FormatInt(se.Unix(), 10)
	request := &EdgeSignRequestPayload{
		Data: base64.StdEncoding.EncodeToString([]byte(data)),
	}
	return edgeSignRequest(workloadURI, module, genid, request)
}

// EdgeSignRequestPayload is a placeholder object for sign requests.
type EdgeSignRequestPayload struct {
	KeyID string `json:"keyId"`
	Algo  string `json:"algo"`
	Data  string `json:"data"`
}

// Validate the properties on EdgeSignRequestPayload
func (esrp *EdgeSignRequestPayload) Validate() error {

	if len(esrp.Algo) < 1 {
		esrp.Algo = "HMACSHA256"
	}

	if len(esrp.KeyID) < 1 {
		esrp.KeyID = "primary"
	}

	if len(esrp.Data) < 1 {
		return fmt.Errorf("sign request: no data provided")
	}

	return nil
}

// EdgeSignRequestResponse is a container struct for the response.
type EdgeSignRequestResponse struct {
	Digest  string `json:"digest"`
	Message string `json:"message"`
}

func edgeSignRequest(workloadURI, name, genid string, payload *EdgeSignRequestPayload) (string, error) {

	esrr := EdgeSignRequestResponse{}

	// validate payload properties
	err := payload.Validate()
	if err != nil {
		return "", fmt.Errorf("sign: unable to sign request: %s", err.Error())
	}

	payloadJSON, _ := json.Marshal(payload)

	// catch unix domain sockets URIs
	if strings.Contains(workloadURI, "unix://") {

		addr, err := net.ResolveUnixAddr("unix", strings.TrimPrefix(workloadURI, "unix://"))
		if err != nil {
			fmt.Printf("Failed to resolve: %v\n", err)
			return "", err
		}

		httpc := http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", addr.Name)
				},
			},
		}

		var response *http.Response
		//var err error

		response, err = httpc.Post("http://iotedge"+fmt.Sprintf("/modules/%s/genid/%s/sign?api-version=2018-06-28", name, genid), "text/plain", bytes.NewBuffer(payloadJSON))
		if err != nil {
			return "", fmt.Errorf("sign: unable to sign request (resp): %s", err.Error())
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", fmt.Errorf("sign: unable to sign request (read): %s", err.Error())
		}

		err = json.Unmarshal(body, &esrr)
		if err != nil {
			return "", fmt.Errorf("sign: unable to sign request (unm): %s", err.Error())
		}

	} else {
		// format uri string for base uri
		uri := fmt.Sprintf("%smodules/%s/genid/%s/sign?api-version=2018-06-28", workloadURI, name, genid)

		// get http response and handle error
		resp, err := http.Post(uri, "text/plain", bytes.NewBuffer(payloadJSON))
		if err != nil {
			return "", fmt.Errorf("sign: unable to sign request (resp): %s", err.Error())
		}
		defer resp.Body.Close()

		// read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("sign: unable to sign request (read): %s", err.Error())
		}

		err = json.Unmarshal(body, &esrr)
		if err != nil {
			return "", fmt.Errorf("sign: unable to sign request (unm): %s", err.Error())
		}
	}

	// if error returned from WorkloadAPI
	if len(esrr.Message) > 0 {
		return "", fmt.Errorf("sign: unable to sign request: %s", esrr.Message)
	}

	return esrr.Digest, nil
}
