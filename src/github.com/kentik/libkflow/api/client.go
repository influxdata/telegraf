package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

type Client struct {
	Email     string
	Token     string
	deviceURL string
	updateURL string
	*http.Client
}

type ClientConfig struct {
	Email   string
	Token   string
	Timeout time.Duration
	API     *url.URL
	Proxy   *url.URL
}

func NewClient(config ClientConfig) *Client {
	transport := *(http.DefaultTransport.(*http.Transport))
	transport.Proxy = nil

	client := &http.Client{
		Transport: &transport,
		Timeout:   config.Timeout,
	}

	if config.Proxy != nil {
		transport.Proxy = http.ProxyURL(config.Proxy)
	}

	return &Client{
		Email:     config.Email,
		Token:     config.Token,
		deviceURL: config.API.String() + "/device/%v",
		updateURL: config.API.String() + "/company/%v/device/%v/tags/snmp",
		Client:    client,
	}
}

func (c *Client) GetDeviceByID(did int) (*Device, error) {
	return c.getdevice(fmt.Sprintf(c.deviceURL, did))
}

func (c *Client) GetDeviceByName(name string) (*Device, error) {
	return c.getdevice(fmt.Sprintf(c.deviceURL, NormalizeName(name)))
}

func (c *Client) GetDeviceByIP(ip net.IP) (*Device, error) {
	return c.getdevice(fmt.Sprintf(c.deviceURL, ip))
}

func (c *Client) GetDeviceByIF(name string) (*Device, error) {
	nif, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := nif.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
			dev, err := c.GetDeviceByIP(ip)
			if err == nil {
				return dev, err
			}
		}
	}

	return nil, &Error{StatusCode: 404}
}

func (c *Client) getdevice(url string) (*Device, error) {
	r, err := c.do("GET", url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, &Error{StatusCode: r.StatusCode}
	}

	dr := &DeviceResponse{}
	if err := json.NewDecoder(r.Body).Decode(dr); err != nil {
		return nil, err
	}

	return dr.Device, nil
}

func (c *Client) GetInterfaces(did int) ([]Interface, error) {
	url := fmt.Sprintf(c.deviceURL+"/interfaces", did)

	r, err := c.do("GET", url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, &Error{StatusCode: r.StatusCode}
	}

	interfaces := []Interface{}
	err = json.NewDecoder(r.Body).Decode(&interfaces)

	return interfaces, err
}

func (c *Client) UpdateInterfaces(dev *Device, nif *net.Interface) error {
	difs, err := c.GetInterfaces(dev.ID)
	if err != nil {
		return err
	}

	updates, err := GetInterfaceUpdates(nif)
	if err != nil {
		return err
	}

	if0 := InterfaceUpdate{
		Index: 0,
		Desc:  "kernel",
	}
	updates[if0.Desc] = if0

	for _, dif := range difs {
		name := dif.Desc
		if nif, ok := updates[name]; ok {
			if nif.Index == dif.Index &&
				nif.Desc == dif.Desc &&
				nif.Address == dif.Address &&
				nif.Netmask == dif.Netmask &&
				reflect.DeepEqual(nif.Addrs, dif.Addrs) {
				delete(updates, name)
			}
		}
	}

	if len(updates) == 0 {
		return nil
	}

	url := fmt.Sprintf(c.updateURL, dev.CompanyID, dev.ID)

	body, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	r, err := c.do("PUT", url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	io.Copy(ioutil.Discard, r.Body)

	if r.StatusCode != 200 {
		return &Error{StatusCode: r.StatusCode}
	}

	return nil
}

func (c *Client) SendFlow(url string, buf *bytes.Buffer) error {
	r, err := c.do("POST", url, "application/binary", buf)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	io.Copy(ioutil.Discard, r.Body)

	if r.StatusCode != 200 {
		return fmt.Errorf("api: HTTP status code %d", r.StatusCode)
	}

	return nil
}

func (c *Client) do(method, url, ctype string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("X-CH-Auth-Email", c.Email)
	r.Header.Set("X-CH-Auth-API-Token", c.Token)
	r.Header.Set("Content-Type", ctype)

	return c.Client.Do(r)
}
