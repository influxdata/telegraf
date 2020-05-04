package becs

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## BECS server.
  ## Default = "localhost:4490".
  # server = "localhost:4490"

  ## BECS login credentials.
  username = "becs"
  password = "becs"
  # namespace = ""

  ## Resources to collect active clients from.
  ## Example = ["10.0.0.0/8"].
  # resources = []

  ## Include memory pools from applications.
  ## Default = false.
  # include_pools = false
`

var soapRequest = envelope{
	Env:     "http://schemas.xmlsoap.org/soap/envelope/",
	Xsi:     "http://www.w3.org/2001/XMLSchema-instance",
	Xsd:     "http://www.w3.org/2001/XMLSchema",
	Becs:    "urn:packetfront_becs",
	SoapEnc: "http://schemas.xmlsoap.org/soap/encoding/",
	Body:    body{Env: "http://schemas.xmlsoap.org/soap/encoding/", ID: "_0"},
}

//Becs object.
type Becs struct {
	BecsServer    string   `toml:"server"`
	Username      string   `toml:"username"`
	Password      string   `toml:"password"`
	Namespace     string   `toml:"namespace"`
	Resources     []string `toml:"resources"`
	IncludePools  bool     `toml:"include_pools"`
	Log           telegraf.Logger
	url           url.URL
	httpClient    http.Client
	sessionActive bool
}

//SampleConfig returns the default BECS configuration.
func (b *Becs) SampleConfig() string {
	return sampleConfig
}

//Description returns the plugin description.
func (b *Becs) Description() string {
	return "Read metrics from given BECS server"
}

//Init sets proper values.
func (b *Becs) Init() error {
	if len(b.BecsServer) == 0 {
		b.BecsServer = "localhost:4490"
	}

	b.url.Scheme = "http"
	b.url.Host = b.BecsServer
	b.httpClient.Timeout = time.Second * 5

	return nil
}

//Gather metrics from given BECS server.
func (b *Becs) Gather(acc telegraf.Accumulator) error {
	/*BECS will drop the session after 10 minutes if the connection is lost.
	The counter is reset every time a call is made*/
	if !b.sessionActive {
		err := b.sessionLogin()
		if err != nil {
			return err
		}

		b.sessionActive = true
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.applicationStatusGet(acc)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.metricGet(acc)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.clientFind(acc)
	}()

	wg.Wait()

	return nil
}

//sessionLogin logs in to BECS and adds the sessionid to soapRequest.
func (b *Becs) sessionLogin() error {
	method := sessionLogin{}
	method.In.Type = "becs:sessionLoginIn"
	method.In.Username = b.Username
	method.In.Password = b.Password
	method.In.Namespace = b.Namespace

	soapRequest.Body.Method = method
	req, err := xml.Marshal(soapRequest)
	if err != nil {
		return err
	}

	resp := sessionLoginResponse{}
	err = b.soapCall(req, &resp)
	if err != nil {
		return err
	}

	if resp.Body.Response.Out.Err != 0 {
		return errors.New(resp.Body.Response.Out.ErrTxt)
	}

	soapRequest.Header.Request.Type = "becs:requestHeader"
	soapRequest.Header.Request.SessionID.Type = "xsd:string"
	soapRequest.Header.Request.SessionID.ID = resp.Body.Response.Out.SessionID

	return nil
}

//applicationList returns a list of BECS applications.
func (b *Becs) applicationList() ([]string, error) {
	method := applicationList{}
	method.In.Type = "becs:applicationListIn"

	req := soapRequest
	req.Body.Method = method
	data, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp := applicationListResponse{}
	if err := b.soapCall(data, &resp); err != nil {
		return nil, err
	}

	if resp.Body.Response.Out.Err != 0 {
		if resp.Body.Response.Out.Err == 8 { //Not authenticated.
			b.sessionActive = false
		}

		return nil, errors.New(resp.Body.Response.Out.ErrTxt)
	}

	return resp.Body.Response.Out.Names.Items, nil
}

//applicationStatusGet collects metrics from BECS applications.
func (b *Becs) applicationStatusGet(acc telegraf.Accumulator) {
	applicationList, err := b.applicationList()
	if err != nil {
		b.Log.Error(err)
		return
	}

	for _, applicationName := range applicationList {
		method := applicationStatusGet{}
		method.In.Type = "becs:applicationStatusGetIn"
		method.In.Name = applicationName
		method.In.IncludePools = b.IncludePools

		req := soapRequest
		req.Body.Method = method
		data, err := xml.Marshal(req)
		if err != nil {
			b.Log.Error(err)
			continue
		}

		resp := applicationStatusGetResponse{}
		if err := b.soapCall(data, &resp); err != nil {
			b.Log.Error(err)
			continue
		}

		if resp.Body.Response.Out.Err != 0 {
			if resp.Body.Response.Out.Err == 8 { //Not authenticated.
				b.sessionActive = false
			}

			b.Log.Error(resp.Body.Response.Out.ErrTxt)
			continue
		}

		tags := map[string]string{
			"application": resp.Body.Response.Out.Displayname,
			"server":      resp.Body.Response.Out.Hostname,
		}

		fields := map[string]interface{}{
			"uptime":       resp.Body.Response.Out.UpTime,
			"cpuusage":     resp.Body.Response.Out.CPUUsage,
			"cpuaverage60": resp.Body.Response.Out.CPUAverage60,
		}

		acc.AddFields("becs_applications", fields, tags)

		if b.IncludePools {
			pools := resp.Body.Response.Out.MemoryPools.Items
			for _, pool := range pools {
				tags["memorypool"] = pool.Name

				poolFields := map[string]interface{}{
					"size":       pool.Size,
					"out":        pool.Out,
					"pages":      pool.Pages,
					"emptypages": pool.EmptyPages,
				}

				acc.AddFields("becs_applications", poolFields, tags)
			}
		}
	}
}

//metricGet collects elements from em_elements.
func (b *Becs) metricGet(acc telegraf.Accumulator) {
	method := metricGet{}
	method.In.Type = "becs:metricGetIn"
	method.In.Name = "em_elements"

	req := soapRequest
	req.Body.Method = method
	data, err := xml.Marshal(req)
	if err != nil {
		b.Log.Error(err)
		return
	}

	resp := metricGetResponse{}
	if err = b.soapCall(data, &resp); err != nil {
		b.Log.Error(err)
		return
	}

	if resp.Body.Response.Out.Err != 0 {
		if resp.Body.Response.Out.Err == 8 { //Not authenticated.
			b.sessionActive = false
		}

		b.Log.Error(resp.Body.Response.Out.ErrTxt)
		return
	}

	for _, metric := range resp.Body.Response.Out.Metrics.Items {
		tags := map[string]string{
			"server": b.url.Hostname(),
			"metric": metric.Name,
		}

		for _, metricValue := range metric.Values.Items {
			for _, metricLabel := range metricValue.Labels.Items {
				tags[metricLabel.Name] = metricLabel.Value
			}

			acc.AddGauge("becs_metrics", map[string]interface{}{"elements": metricValue.Value}, tags)
		}
	}
}

//clientFind collects the number of clients within a resource.
func (b *Becs) clientFind(acc telegraf.Accumulator) {
	for _, resource := range b.Resources {
		if _, _, err := net.ParseCIDR(resource); err != nil {
			b.Log.Error(err)
			continue
		}

		method := clientFind{}
		method.In.Type = "becs:clientFindIn"
		method.In.IP = resource
		method.In.Limit = 1 //Limit is 1 because we only want the actual number, not the clients.

		req := soapRequest
		req.Body.Method = method
		data, err := xml.Marshal(req)
		if err != nil {
			b.Log.Error(err)
			continue
		}

		resp := clientFindResponse{}
		if err = b.soapCall(data, &resp); err != nil {
			b.Log.Error(err)
			continue
		}

		if resp.Body.Response.Out.Err != 0 {
			if resp.Body.Response.Out.Err == 8 { //Not authenticated.
				b.sessionActive = false
			}

			b.Log.Error(resp.Body.Response.Out.ErrTxt)
			continue
		}

		tags := map[string]string{
			"server":   b.url.Hostname(),
			"resource": resource,
		}

		acc.AddFields("becs_clients", map[string]interface{}{"clients": resp.Body.Response.Out.Actual}, tags)
	}
}

//soapCall sends a soap request to BECS and decode the response.
func (b *Becs) soapCall(request []byte, response interface{}) error {
	resp, err := b.httpClient.Post(b.url.String(), "text/xml; charset=utf-8", bytes.NewBuffer(request))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return errors.New("Decode: " + err.Error())
	}

	return nil
}

func init() {
	inputs.Add("becs", func() telegraf.Input { return &Becs{} })
}
