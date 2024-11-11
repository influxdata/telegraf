//go:generate ../../../tools/readme_config_includer/generator
package mikrotik

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs"

	common_tls "github.com/influxdata/telegraf/plugins/common/tls"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

//go:embed sample.conf
var sampleConfig string

type Mikrotik struct {
	Address         string          `toml:"address"`
	IgnoreCert      bool            `toml:"ignore_cert,omitempty"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	IgnoreComments  []string        `toml:"ignore_comments"`
	IncludeModules  []string        `toml:"include_modules"`
	Username        config.Secret   `toml:"username"`
	Password        config.Secret   `toml:"password"`
	Log             telegraf.Logger `toml:"-"`
	common_tls.ClientConfig

	tags map[string]string

	url []mikrotikEndpoint

	client *http.Client

	systemTagsURL []string
}

func (*Mikrotik) SampleConfig() string {
	return sampleConfig
}

func (h *Mikrotik) Start() error {
	return h.getSystemTags()
}

func (h *Mikrotik) Init() error {
	if h.Username.Empty() {
		return errors.New("mikrotik init -> username must be present")
	}

	if len(h.IncludeModules) == 0 {
		h.IncludeModules = append(h.IncludeModules, "system_resourses")
	}

	mainPropList, systemResourcesPropList, systemRouterBoardPropList := createPropLists()

	h.systemTagsURL = []string{
		h.Address + "/rest/system/resource?" + systemResourcesPropList,
		h.Address + "/rest/system/routerboard?" + systemRouterBoardPropList,
	}

	for _, selectedModule := range h.IncludeModules {
		if _, ok := modules[selectedModule]; !ok {
			return fmt.Errorf("mikrotik init -> module %s does not exist or has a typo. Correct modules are: %s", selectedModule, getModuleNames())
		}
		h.url = append(h.url, mikrotikEndpoint{name: selectedModule, url: fmt.Sprintf("%s%s?%s", h.Address, modules[selectedModule], mainPropList)})
	}

	ignoreCommentsFunction = basicCommentAndDisableFilter(h.IgnoreComments)

	return h.getClient()
}

func (h *Mikrotik) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, u := range h.url {
		wg.Add(1)
		go func(url mikrotikEndpoint) {
			defer wg.Done()

			if err := h.gatherURL(url, acc); err != nil {
				acc.AddError(fmt.Errorf("gather -> %w", err))
			}
		}(u)
	}
	wg.Wait()

	return nil
}

func (h *Mikrotik) getClient() (err error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("getClient -> %w", err)
	}

	h.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(h.ResponseTimeout),
	}

	return nil
}

func (h *Mikrotik) getSystemTags() error {
	h.tags = make(map[string]string)
	for _, tagURL := range h.systemTagsURL {
		request, err := http.NewRequest("GET", tagURL, nil)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}

		err = h.setRequestAuth(request)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}

		binaryData, err := h.queryData(request)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}

		err = json.Unmarshal(binaryData, &h.tags)
		if err != nil {
			return fmt.Errorf("getSystemTags -> %w", err)
		}
	}
	return nil
}

func (h *Mikrotik) gatherURL(endpoint mikrotikEndpoint, acc telegraf.Accumulator) error {
	request, err := http.NewRequest("GET", endpoint.url, nil)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	err = h.setRequestAuth(request)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	binaryData, err := h.queryData(request)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	timestamp := time.Now()

	result, err := binToCommon(binaryData)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	parsedData, err := parse(result)
	if err != nil {
		return fmt.Errorf("gatherURL -> %w", err)
	}

	for _, point := range parsedData {
		point.Tags["source-module"] = endpoint.name
		for n, v := range h.tags {
			point.Tags[n] = v
		}
		acc.AddFields("mikrotik", point.Fields, point.Tags, timestamp)
	}

	return nil
}

func (h *Mikrotik) setRequestAuth(request *http.Request) error {
	username, err := h.Username.Get()
	if err != nil {
		return fmt.Errorf("setRequestAuth: username -> %w", err)
	}
	defer username.Destroy()

	password, err := h.Password.Get()
	if err != nil {
		return fmt.Errorf("setRequestAuth: pasword -> %w", err)
	}
	defer password.Destroy()

	request.SetBasicAuth(username.String(), password.String())

	return nil
}

func (h *Mikrotik) queryData(request *http.Request) (data []byte, err error) {
	resp, err := h.client.Do(request)
	if err != nil {
		return data, fmt.Errorf("queryData -> %w", err)
	}

	defer resp.Body.Close()
	defer h.client.CloseIdleConnections()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("queryData -> received status code %d (%s), expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return data, fmt.Errorf("queryData -> %w", err)
	}

	return data, err
}

func init() {
	inputs.Add("mikrotik", func() telegraf.Input {
		return &Mikrotik{ResponseTimeout: config.Duration(time.Second * 5)}
	})
}
