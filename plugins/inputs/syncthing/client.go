package syncthing

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func (s *Syncthing) SystemConnections(ctx context.Context, host string) (*SystemConnections, error) {
	cons := new(SystemConnections)
	err := s.GetJSON(ctx, host, ConnectionsEndpoint, cons, nil)
	return cons, err
}

func (s *Syncthing) SystemStatus(ctx context.Context, host string) (*SystemStatus, error) {
	st := new(SystemStatus)
	err := s.GetJSON(ctx, host, SystemStatusEndpoint, st, nil)
	return st, err
}

// Need queries https://docs.syncthing.net/rest/db-need-get.html to find out how many files are needed
// to become in sync
func (s *Syncthing) Need(ctx context.Context, host string, folderID string) (*Need, error) {
	r := new(Need)
	q := url.Values{}
	q.Add("folder", folderID)

	err := s.GetJSON(ctx, host, NeedEndpoint, r, q)
	return r, err
}

func (s *Syncthing) SystemConfig(ctx context.Context, host string) (*SystemConfig, error) {
	r := new(SystemConfig)
	err := s.GetJSON(ctx, host, SysconfigEndpoint, r, nil)
	return r, err
}

func (s *Syncthing) GetJSON(ctx context.Context, host string, uri string, v interface{}, parms url.Values) error {
	rurl, err := url.Parse(host)
	if err != nil {
		return errors.Wrap(err, "failed to parse request URI")
	}
	rurl.Path = uri

	if len(parms) > 0 {
		rurl.RawQuery = parms.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rurl.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}
	req.Header.Set(AuthHeader, s.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("bad status code (%d) from http %q: body: %q", resp.StatusCode, rurl, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return errors.Wrapf(err, "failed to decode response body from %q", rurl.String())
	}
	return nil
}
