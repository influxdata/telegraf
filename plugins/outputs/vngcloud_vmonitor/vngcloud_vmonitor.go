package vngcloud_vmonitor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/matishsiao/goInfo"
	"github.com/shirou/gopsutil/cpu"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	metricPath           = "/intake/v2/series"
	quotaPath            = "/intake/v2/check"
	defaultClientTimeout = 10 * time.Second
	defaultContentType   = "application/json"
	AgentVersion         = "1.26.0-2.0.0"
)

var defaultConfig = &VNGCloudvMonitor{
	URL:             "http://localhost:8081",
	Timeout:         config.Duration(10 * time.Second),
	Method:          http.MethodPost,
	IamURL:          "https://hcm-3.console.vngcloud.vn/iam/accounts-api/v2/auth/token",
	OutOfQuotaRetry: config.Duration(30 * time.Second),
	CheckQuotaRetry: config.Duration(30 * time.Second),
}

var sampleConfig = `
  ## URL is the address to send metrics to
  url = "http://localhost:8081"
  insecure_skip_verify = false
  data_format = "vngcloud_vmonitor"
  timeout = "30s"

  # From IAM service
  client_id = ""
  client_secret = ""
`

type Request struct {
	Method string
	Url    string
	Path   string
	Body   []byte
}

type Plugin struct {
	Name    string `json:"name"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type QuotaInfo struct {
	Checksum string    `json:"checksum"`
	Data     *infoHost `json:"data"`
}

type infoHost struct {
	Plugins []*Plugin `json:"plugins"`

	HashID string `json:"hash_id"`

	Kernel       string `json:"kernel"`
	Core         string `json:"core"`
	Platform     string `json:"platform"`
	OS           string `json:"os"`
	Hostname     string `json:"host_name"`
	CPUs         int    `json:"cpus"`
	ModelNameCPU string `json:"model_name_cpu"`
	Mem          uint64 `json:"mem"`
	Ip           string `json:"ip"`
	AgentVersion string `json:"agent_version"`
	UserAgent    string `toml:"user_agent"`
}

type VNGCloudvMonitor struct {
	URL             string            `toml:"url"`
	Timeout         config.Duration   `toml:"timeout"`
	Method          string            `toml:"method"`
	Headers         map[string]string `toml:"headers"`
	ContentEncoding string            `toml:"content_encoding"`
	Insecure        bool              `toml:"insecure_skip_verify"`
	ProxyStr        string            `toml:"proxy_url"`
	CheckQuotaRetry config.Duration   `toml:"check_quota_retry"`
	OutOfQuotaRetry config.Duration   `toml:"out_of_quota_retry"`

	IamURL       string `toml:"iam_url"`
	ClientId     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`

	Requests           *Request
	serializer         serializers.Serializer
	infoHost           *infoHost
	client_iam         *http.Client
	Oauth2ClientConfig *clientcredentials.Config

	dropCount int
	dropTime  time.Time

	dropByIam bool
}

func (h *VNGCloudvMonitor) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

func (h *VNGCloudvMonitor) initHTTPClient() error {
	log.Println("[vMonitor] Init client-iam ...")
	h.Oauth2ClientConfig = &clientcredentials.Config{
		ClientID:     h.ClientId,
		ClientSecret: h.ClientSecret,
		TokenURL:     h.IamURL,
	}

	token, err := h.Oauth2ClientConfig.TokenSource(context.Background()).Token()
	if err != nil {
		h.dropByIam = true
		return fmt.Errorf("[vMonitor] Failed to get token: %s", err.Error())
	}

	_, err = json.Marshal(token)
	if err != nil {
		h.dropByIam = true
		return fmt.Errorf("[vMonitor] Failed to Marshal token: %s", err.Error())
	}
	h.client_iam = h.Oauth2ClientConfig.Client(context.TODO())
	log.Println("[vMonitor] Init client-iam successfully")
	h.dropByIam = false
	return nil
}

func getIp(address, port string) (string, error) {
	log.Printf("[vMonitor] Dial %s %s", address, port)
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(address, port), 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

func getModelNameCPU() (string, error) {
	a, err := cpu.Info()
	if err != nil {
		return "", err
	}
	return a[0].ModelName, nil
}

func (h *VNGCloudvMonitor) getHostInfo() (*infoHost, error) {
	getHostPort := func(urlStr string) (string, error) {
		u, err := url.Parse(urlStr)
		if err != nil {
			return "", fmt.Errorf("[vMonitor] proxy invalid %s", h.ProxyStr)
		}

		host, port, err := net.SplitHostPort(u.Host)

		if err != nil {
			return "", err
		}

		ipLocal, err := getIp(host, port)
		if err != nil {
			return "", err
		}
		return ipLocal, nil
	}

	var ipLocal string
	var err error
	// get ip local
	if h.ProxyStr != "" {
		ipLocal, err = getHostPort(h.ProxyStr)
	} else {
		ipLocal, err = getHostPort(h.URL)
	}

	if err != nil {
		return nil, fmt.Errorf("[vMonitor] err getting ip address %s", err.Error())
	}
	if err != nil {
		return nil, fmt.Errorf("[vMonitor] err getting mac_address %s", err.Error())
	}
	// get ip local

	gi, err := goInfo.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("[vMonitor] error getting os info: %s", err)
	}
	ps := system.NewSystemPS()
	vm, err := ps.VMStat()

	hashCode := sha256.New()
	hashCode.Write([]byte(gi.Hostname))
	hashedID := hex.EncodeToString(hashCode.Sum(nil))

	if err != nil {
		return nil, fmt.Errorf("[vMonitor] error getting virtual memory info: %s", err)
	}

	modelNameCPU, err := getModelNameCPU()

	if err != nil {
		return nil, fmt.Errorf("[vMonitor] error getting cpu model name: %s", err)
	}

	h.infoHost.HashID = hashedID
	h.infoHost.Kernel = gi.Kernel
	h.infoHost.Core = gi.Core
	h.infoHost.Platform = gi.Platform
	h.infoHost.OS = gi.OS
	h.infoHost.CPUs = gi.CPUs
	h.infoHost.ModelNameCPU = modelNameCPU
	h.infoHost.Mem = vm.Total
	h.infoHost.Ip = ipLocal
	h.infoHost.AgentVersion = AgentVersion
	h.infoHost.UserAgent = fmt.Sprintf("%s/%s (%s)", "vMonitorAgent", AgentVersion, h.infoHost.OS)

	return h.infoHost, nil
}

func (h *VNGCloudvMonitor) setDefault() error {
	if h.Method == "" {
		h.Method = http.MethodPost
	}

	h.Method = strings.ToUpper(h.Method)
	if h.Method != http.MethodPost && h.Method != http.MethodPut {
		return fmt.Errorf("[vMonitor] Invalid method [%s] %s", h.URL, h.Method)
	}

	if h.Timeout == 0 {
		h.Timeout = config.Duration(defaultClientTimeout)
	}
	return nil
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (h *VNGCloudvMonitor) CheckConfig() error {
	ok := isUrl(h.URL)
	if !ok {
		return fmt.Errorf("[vMonitor] URL Invalid %s", h.URL)
	}
	return nil
}

func (h *VNGCloudvMonitor) Connect() error {

	if err := h.CheckConfig(); err != nil {
		return err
	}

	if err := h.setDefault(); err != nil {
		return err
	}

	// h.client_iam = client_iam
	err := h.initHTTPClient()
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = h.getHostInfo()
	if err != nil {
		return err
	}

	return nil
}

func (h *VNGCloudvMonitor) Close() error {
	return nil
}

func (h *VNGCloudvMonitor) Description() string {
	return "Configuration for vMonitor output."
}

func (h *VNGCloudvMonitor) SampleConfig() string {
	//log.Print(sampleConfig)
	return sampleConfig
}

func (h *VNGCloudvMonitor) setPlugins(metrics []telegraf.Metric) error {
	a := h.infoHost.Plugins
	nameTemp := ""
	hostname := ""

	existCheck := func(name string) bool {
		for _, e := range a {
			if name == e.Name {
				return true
			}
		}
		return false
	}
	for _, element := range metrics {
		if element.Name() != nameTemp || nameTemp == "" {
			if !existCheck(element.Name()) {
				hostTemp, ok := element.GetTag("host")

				if ok {
					hostname = hostTemp
				}

				msg := "running"
				a = append(a, &Plugin{
					Name:    element.Name(),
					Status:  0,
					Message: msg,
				})
				nameTemp = element.Name()
			}
		}
	}

	if hostname == "" && h.infoHost.Hostname == "" {
		hostnameTemp, err := os.Hostname()
		if err != nil {
			return err
		}
		h.infoHost.Hostname = hostnameTemp
	}
	if hostname != "" {
		h.infoHost.Hostname = hostname
	}
	h.infoHost.Plugins = a
	return nil
}

func (h *VNGCloudvMonitor) Write(metrics []telegraf.Metric) error {
	if h.dropCount > 0 && time.Now().Before(h.dropTime) {
		log.Printf("[vMonitor] Drop %d metrics because OUT_OF_QUOTA.", len(metrics))
		return nil
	}

	if h.dropByIam {
		err := h.initHTTPClient()
		if err != nil {
			log.Print(err)
		}
		log.Printf("[vMonitor] Drop %d metrics because of IAM.", len(metrics))
		return nil
	}

	if err := h.setPlugins(metrics); err != nil {
		return err
	}

	reqBody, err := h.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	if err := h.write(reqBody); err != nil {
		return err
	}

	return nil
}

func (h *VNGCloudvMonitor) write(reqBody []byte) error {

	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)

	var err error
	if h.ContentEncoding == "gzip" {
		rc := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

	req, err := http.NewRequest(h.Method, fmt.Sprintf("%s%s", h.URL, metricPath), reqBodyBuffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", defaultContentType)
	req.Header.Set("checksum", h.infoHost.HashID)
	req.Header.Set("User-Agent", h.infoHost.UserAgent)

	if h.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}
	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	resp, err := h.client_iam.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dataRsp, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	log.Printf("[vMonitor] Request-ID: %s with body length %d byte and response body %s", resp.Header.Get("Api-Request-ID"), len(reqBody), dataRsp)

	if err := h.handleResponse(resp.StatusCode); err != nil {
		return err
	}

	return nil
}

func (h *VNGCloudvMonitor) handleResponse(respCode int) error {
	if respCode == 201 {
		return nil
	} else if respCode == 400 {
		return fmt.Errorf("[vMonitor] Bad request")
	} else if respCode == 401 {
		log.Printf("[vMonitor] IAM Unauthorized")
		err := h.initHTTPClient()
		if err != nil {
			log.Print(err)
			return err
		}
		return fmt.Errorf("[vMonitor] IAM Unauthorized")
	} else if respCode == 403 {
		h.dropByIam = true
		return fmt.Errorf("[vMonitor] IAM Forbidden")
	} else if respCode == 428 {
		if err := h.checkQuota(); err != nil {
			return err
		}
		return fmt.Errorf("[vMonitor] Checking quota success, try to send metric again")
	} else if respCode == 503 || respCode == 504 {
		return fmt.Errorf("[vMonitor] Gateway Timeout or Service Unavailable (%d)", respCode)
	} else if respCode == 408 {
		return fmt.Errorf("[vMonitor] Request Time-out (%d)", respCode)
	} else {
		return fmt.Errorf("[vMonitor] Unhandled Status Code %d", respCode)
	}
}

func (h *VNGCloudvMonitor) checkQuota() error {

	quotaStruct := &QuotaInfo{
		Checksum: h.infoHost.HashID,
		Data:     h.infoHost,
	}
	quotaJson, err := json.Marshal(quotaStruct)
	if err != nil {
		return fmt.Errorf("[vMonitor] Can not marshal quota struct: %s", err)
	}

	const retryTime = 8

	for i := 0; i < retryTime; i++ {

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", h.URL, quotaPath), bytes.NewBuffer(quotaJson))
		if err != nil {
			log.Println("[vMonitor] Error create new request: ", err)
			continue
		}
		req.Header.Set("checksum", h.infoHost.HashID)
		req.Header.Set("Content-Type", defaultContentType)
		req.Header.Set("User-Agent", h.infoHost.UserAgent)
		resp, err := h.client_iam.Do(req)

		if err != nil {
			log.Printf("[vMonitor] Send request checking quota failed, error: (%s), try to send again", err)
			continue
		}
		defer resp.Body.Close()
		dataRsp, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Printf("[vMonitor] Request-ID: %s. An error occurred when reading response body (%s), error: (%s), try to send again", resp.Header.Get("Api-Request-ID"), dataRsp, err.Error())
			continue
		}

		if resp.StatusCode == 200 {
			log.Printf("[vMonitor] Request-ID: %s. Checking quota success. Continue send metric.", resp.Header.Get("Api-Request-ID"))
			h.dropCount = 0
			h.dropTime = time.Now()
			return nil
		} else if resp.StatusCode == 409 {
			if h.dropCount < retryTime {
				h.dropCount++
			}
			dropDuration := time.Duration(int(math.Pow(2, float64(h.dropCount))) * int(h.OutOfQuotaRetry))
			h.dropTime = time.Now().Add(dropDuration)
			log.Printf("[vMonitor] Request-ID: %s. Package out of quota. Check quota again in %.0fm ", resp.Header.Get("Api-Request-ID"), dropDuration.Minutes())
			return fmt.Errorf("OUT_OF_QUOTA")

		} else if resp.StatusCode != 428 && resp.StatusCode != 503 && resp.StatusCode != 504 {
			log.Printf("[vMonitor] Request-ID: %s. Receive an unhandled StatusCode = %d.", resp.Header.Get("Api-Request-ID"), resp.StatusCode)
		}

		if i == retryTime-1 {
			return fmt.Errorf("[vMonitor] Can not check quota, max retry exceed")
		} else {
			log.Printf("[vMonitor] Request-ID: %s. Checking quota fail (%s), sleep in %.0fs and retry", resp.Header.Get("Api-Request-ID"), dataRsp, time.Duration(int(h.CheckQuotaRetry)*int(math.Pow(2, float64(i)))).Seconds())
			time.Sleep(time.Duration(int(h.CheckQuotaRetry) * int(math.Pow(2, float64(i)))))
		}
	}

	return nil
}

func init() {
	outputs.Add("vngcloud_vmonitor", func() telegraf.Output {
		infoHosts := &infoHost{
			Plugins:  []*Plugin{},
			HashID:   "",
			Kernel:   "",
			Core:     "",
			Platform: "",
			OS:       "",
			Hostname: "",
			CPUs:     0,
			Mem:      0,
		}
		log.Print("#################### Welcome to vMonitor (VNGCLOUD) ####################")
		return &VNGCloudvMonitor{
			Timeout:         defaultConfig.Timeout,
			Method:          defaultConfig.Method,
			URL:             defaultConfig.URL,
			IamURL:          defaultConfig.IamURL,
			OutOfQuotaRetry: defaultConfig.OutOfQuotaRetry,
			CheckQuotaRetry: defaultConfig.CheckQuotaRetry,
			infoHost:        infoHosts,

			dropCount: 0,
			dropTime:  time.Now(),

			dropByIam: false,
		}
	})
}
