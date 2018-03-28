package azuremonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/common/log"
)

// AzureInstanceMetadata is the proxy for accessing the instance metadata service on an Azure VM
type AzureInstanceMetadata struct {
}

// VirtualMachineMetadata contains information about a VM from the metadata service
type VirtualMachineMetadata struct {
	Raw             string
	AzureResourceID string
	Compute         struct {
		Location             string `json:"location"`
		Name                 string `json:"name"`
		Offer                string `json:"offer"`
		OsType               string `json:"osType"`
		PlacementGroupID     string `json:"placementGroupId"`
		PlatformFaultDomain  string `json:"platformFaultDomain"`
		PlatformUpdateDomain string `json:"platformUpdateDomain"`
		Publisher            string `json:"publisher"`
		ResourceGroupName    string `json:"resourceGroupName"`
		Sku                  string `json:"sku"`
		SubscriptionID       string `json:"subscriptionId"`
		Tags                 string `json:"tags"`
		Version              string `json:"version"`
		VMID                 string `json:"vmId"`
		VMScaleSetName       string `json:"vmScaleSetName"`
		VMSize               string `json:"vmSize"`
		Zone                 string `json:"zone"`
	} `json:"compute"`
	Network struct {
		Interface []struct {
			Ipv4 struct {
				IPAddress []struct {
					PrivateIPAddress string `json:"privateIpAddress"`
					PublicIPAddress  string `json:"publicIpAddress"`
				} `json:"ipAddress"`
				Subnet []struct {
					Address string `json:"address"`
					Prefix  string `json:"prefix"`
				} `json:"subnet"`
			} `json:"ipv4"`
			Ipv6 struct {
				IPAddress []interface{} `json:"ipAddress"`
			} `json:"ipv6"`
			MacAddress string `json:"macAddress"`
		} `json:"interface"`
	} `json:"network"`
}

// MsiToken is the managed service identity token
type MsiToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`

	expiresAt time.Time
	notBefore time.Time
	raw       string
}

func (m *MsiToken) parseTimes() {
	val, err := strconv.ParseInt(m.ExpiresOn, 10, 64)
	if err == nil {
		m.expiresAt = time.Unix(val, 0)
	}

	val, err = strconv.ParseInt(m.NotBefore, 10, 64)
	if err == nil {
		m.notBefore = time.Unix(val, 0)
	}
}

// ExpiresAt is the time at which the token expires
func (m *MsiToken) ExpiresAt() time.Time {
	return m.expiresAt
}

// ExpiresInDuration returns the duration until the token expires
func (m *MsiToken) ExpiresInDuration() time.Duration {
	expiresDuration := m.expiresAt.Sub(time.Now().UTC())
	return expiresDuration
}

// NotBeforeTime returns the time at which the token becomes valid
func (m *MsiToken) NotBeforeTime() time.Time {
	return m.notBefore
}

// GetMsiToken retrieves a managed service identity token from the specified port on the local VM
func (s *AzureInstanceMetadata) GetMsiToken(clientID string, resourceID string) (*MsiToken, error) {
	// Acquire an MSI token.  Documented at:
	// https://docs.microsoft.com/en-us/azure/active-directory/managed-service-identity/how-to-use-vm-token
	//
	//GET http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fmanagement.azure.com%2F&client_id=712eac09-e943-418c-9be6-9fd5c91078bl HTTP/1.1 Metadata: true

	// Create HTTP request for MSI token to access Azure Resource Manager
	var msiEndpoint *url.URL
	msiEndpoint, err := url.Parse(msiInstanceMetadataURL)
	if err != nil {
		return nil, err
	}

	// Resource ID defaults to https://management.azure.com
	if resourceID == "" {
		resourceID = "https://management.azure.com"
	}

	msiParameters := url.Values{}
	msiParameters.Add("resource", resourceID)
	msiParameters.Add("api-version", "2018-02-01")

	// Client id is optional
	if clientID != "" {
		msiParameters.Add("client_id", clientID)
	}

	msiEndpoint.RawQuery = msiParameters.Encode()
	req, err := http.NewRequest("GET", msiEndpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Metadata", "true")

	// Create the HTTP client and call the token service
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Complete reading the body
	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("Post Error. HTTP response code:%d message:%s, content: %s",
			resp.StatusCode, resp.Status, reply)
	}

	var token MsiToken
	if err := json.Unmarshal(reply, &token); err != nil {
		return nil, err
	}
	token.parseTimes()
	token.raw = string(reply)
	return &token, nil
}

const (
	vmInstanceMetadataURL  = "http://169.254.169.254/metadata/instance?api-version=2017-12-01"
	msiInstanceMetadataURL = "http://169.254.169.254/metadata/identity/oauth2/token"
)

// GetInstanceMetadata retrieves metadata about the current Azure VM
func (s *AzureInstanceMetadata) GetInstanceMetadata() (*VirtualMachineMetadata, error) {
	req, err := http.NewRequest("GET", vmInstanceMetadataURL, nil)
	if err != nil {
		log.Errorf("Error creating HTTP request")
		return nil, err
	}
	req.Header.Set("Metadata", "true")
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("Post Error. HTTP response code:%d message:%s reply:\n%s",
			resp.StatusCode, resp.Status, reply)
	}

	var metadata VirtualMachineMetadata
	if err := json.Unmarshal(reply, &metadata); err != nil {
		return nil, err
	}
	metadata.AzureResourceID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s",
		metadata.Compute.SubscriptionID, metadata.Compute.ResourceGroupName, metadata.Compute.Name)

	return &metadata, nil
}
