package aurora

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type RoleType int

const (
	Unknown RoleType = iota
	Leader
	Follower
)

func (r RoleType) String() string {
	switch r {
	case Leader:
		return "leader"
	case Follower:
		return "follower"
	default:
		return "unknown"
	}
}

var (
	defaultTimeout = 5 * time.Second
	defaultRoles   = []string{"leader", "follower"}
)

type Vars map[string]interface{}

type Aurora struct {
	Schedulers []string        `toml:"schedulers"`
	Roles      []string        `toml:"roles"`
	Timeout    config.Duration `toml:"timeout"`
	Username   string          `toml:"username"`
	Password   string          `toml:"password"`
	tls.ClientConfig

	client *http.Client
	urls   []*url.URL
}

func (a *Aurora) Gather(acc telegraf.Accumulator) error {
	if a.client == nil {
		err := a.initialize()
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	var wg sync.WaitGroup
	for _, u := range a.urls {
		wg.Add(1)
		go func(u *url.URL) {
			defer wg.Done()
			role, err := a.gatherRole(ctx, u)
			if err != nil {
				acc.AddError(fmt.Errorf("%s: %v", u, err))
				return
			}

			if !a.roleEnabled(role) {
				return
			}

			err = a.gatherScheduler(ctx, u, role, acc)
			if err != nil {
				acc.AddError(fmt.Errorf("%s: %v", u, err))
			}
		}(u)
	}
	wg.Wait()

	return nil
}

func (a *Aurora) initialize() error {
	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsCfg,
		},
	}

	urls := make([]*url.URL, 0, len(a.Schedulers))
	for _, s := range a.Schedulers {
		loc, err := url.Parse(s)
		if err != nil {
			return err
		}

		urls = append(urls, loc)
	}

	if a.Timeout < config.Duration(time.Second) {
		a.Timeout = config.Duration(defaultTimeout)
	}

	if len(a.Roles) == 0 {
		a.Roles = defaultRoles
	}

	a.client = client
	a.urls = urls
	return nil
}

func (a *Aurora) roleEnabled(role RoleType) bool {
	if len(a.Roles) == 0 {
		return true
	}

	for _, v := range a.Roles {
		if role.String() == v {
			return true
		}
	}
	return false
}

func (a *Aurora) gatherRole(ctx context.Context, origin *url.URL) (RoleType, error) {
	loc := *origin
	loc.Path = "leaderhealth"
	req, err := http.NewRequest("GET", loc.String(), nil)
	if err != nil {
		return Unknown, err
	}

	if a.Username != "" || a.Password != "" {
		req.SetBasicAuth(a.Username, a.Password)
	}
	req.Header.Add("Accept", "text/plain")

	resp, err := a.client.Do(req.WithContext(ctx))
	if err != nil {
		return Unknown, err
	}
	if err := resp.Body.Close(); err != nil {
		return Unknown, fmt.Errorf("closing body failed: %v", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return Leader, nil
	case http.StatusBadGateway:
		fallthrough
	case http.StatusServiceUnavailable:
		return Follower, nil
	default:
		return Unknown, fmt.Errorf("%v", resp.Status)
	}
}

func (a *Aurora) gatherScheduler(
	ctx context.Context, origin *url.URL, role RoleType, acc telegraf.Accumulator,
) error {
	loc := *origin
	loc.Path = "vars.json"
	req, err := http.NewRequest("GET", loc.String(), nil)
	if err != nil {
		return err
	}

	if a.Username != "" || a.Password != "" {
		req.SetBasicAuth(a.Username, a.Password)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := a.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%v", resp.Status)
	}

	var vars Vars
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	err = decoder.Decode(&vars)
	if err != nil {
		return fmt.Errorf("decoding response: %v", err)
	}

	var fields = make(map[string]interface{}, len(vars))
	for k, v := range vars {
		switch v := v.(type) {
		case json.Number:
			// Aurora encodes numbers as you would specify them as a literal,
			// use this to determine if a value is a float or int.
			if strings.ContainsAny(v.String(), ".eE") {
				fv, err := v.Float64()
				if err != nil {
					acc.AddError(err)
					continue
				}
				fields[k] = fv
			} else {
				fi, err := v.Int64()
				if err != nil {
					acc.AddError(err)
					continue
				}
				fields[k] = fi
			}
		default:
			continue
		}
	}

	acc.AddFields("aurora",
		fields,
		map[string]string{
			"scheduler": origin.String(),
			"role":      role.String(),
		},
	)
	return nil
}

func init() {
	inputs.Add("aurora", func() telegraf.Input {
		return &Aurora{}
	})
}
