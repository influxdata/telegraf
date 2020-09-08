// Package gtfs is an input plugin for collecting GTFS-realtime data.
package gtfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/errgroup"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type GTFS struct {
	Key                 string `toml:"key"`
	VehiclePositionsURL string `toml:"vehicle_positions_url"`
	TripUpdatesURL      string `toml:"trip_updates_url"`
	ServiceAlertsURL    string `toml:"service_alerts_url"`

	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Timeout  internal.Duration `toml:"timeout"`

	tls.ClientConfig

	client              *http.Client
	vehiclePositionsReq *http.Request
	tripUpdatesReq      *http.Request
	serviceAlertsReq    *http.Request
}

func (s *GTFS) Description() string {
	return "GTFS-realtime data"
}

func (s *GTFS) SampleConfig() string {
	return `
  ## API Key
  # key = "${GTFS_API_KEY}"
  ## URL for fetching vehicle positions
  # vehicle_positions_url = "https://host.test/VehiclePositions.pb"
  ## URL for fetching vehicle positions
  # trip_updates_url = "https://host.test/TripUpdates.pb"
  ## URL for fetching vehicle positions
  # service_alerts_url = "https://host.test/ServiceAlerts.pb"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Amount of testTimestamp allowed to complete the HTTP request
  # timeout = "5s"
`
}

func (g *GTFS) Init() error {
	tlsCfg, err := g.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	g.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: g.Timeout.Duration,
	}

	if g.VehiclePositionsURL == "" && g.TripUpdatesURL == "" && g.ServiceAlertsURL == "" {
		return errNoSourceConfigured
	}

	if g.VehiclePositionsURL != "" {
		r, err := g.newRequest(g.VehiclePositionsURL)
		if err != nil {
			return err
		}
		g.vehiclePositionsReq = r
	}

	if g.TripUpdatesURL != "" {
		r, err := g.newRequest(g.TripUpdatesURL)
		if err != nil {
			return err
		}
		g.tripUpdatesReq = r
	}

	if g.ServiceAlertsURL != "" {
		r, err := g.newRequest(g.ServiceAlertsURL)
		if err != nil {
			return err
		}
		g.serviceAlertsReq = r
	}

	return nil
}

func (g *GTFS) Gather(acc telegraf.Accumulator) error {
	var (
		work errgroup.Group
		now  = time.Now()
	)

	if g.vehiclePositionsReq != nil {
		work.Go(func() error { return g.gatherVehiclePositions(acc, now) })
	}

	if g.tripUpdatesReq != nil {
		work.Go(func() error { return g.gatherTripUpdates(acc, now) })
	}

	if g.serviceAlertsReq != nil {
		work.Go(func() error { return g.gatherServiceAlerts(acc, now) })
	}

	return work.Wait()
}

func (g *GTFS) gatherVehiclePositions(acc telegraf.Accumulator, t time.Time) error {
	entities, err := g.gather(g.vehiclePositionsReq)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		if entity.Vehicle == nil {
			continue
		}
		v := entity.Vehicle

		b, err := json.Marshal(v)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{
			"latitude":      v.Position.Latitude,
			"longitude":     v.Position.Longitude,
			"bearing":       v.Position.Bearing,
			"odometer":      v.Position.Odometer,
			"speed":         v.Position.Speed,
			"vehicle_id":    v.Vehicle.Id,
			"vehicle_label": v.Vehicle.Label,
			"trip_id":       v.Trip.TripId,
			"stop_id":       v.StopId,
			"status":        v.CurrentStatus,
			"congestion":    v.CongestionLevel,
			"json":          b,
		}

		tags := map[string]string{
			"route_id":     *v.Trip.RouteId,
			"direction_id": fmt.Sprintf("%d", v.Trip.DirectionId),
		}

		acc.AddFields("position", fields, tags, time.Unix(int64(*v.Timestamp), 0))
	}

	return nil
}

func (g *GTFS) gatherTripUpdates(acc telegraf.Accumulator, t time.Time) error {
	entities, err := g.gather(g.tripUpdatesReq)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		if entity.TripUpdate == nil {
			continue
		}
		u := entity.TripUpdate

		b, err := json.Marshal(u)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{
			"json": b,
		}

		tags := map[string]string{
			"route_id": *u.Trip.RouteId,
		}

		acc.AddFields("update", fields, tags, time.Unix(int64(*u.Timestamp), 0))
	}

	return nil
}

func (g *GTFS) gatherServiceAlerts(acc telegraf.Accumulator, t time.Time) error {
	entities, err := g.gather(g.serviceAlertsReq)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		if entity.Alert == nil {
			continue
		}
		a := entity.Alert

		b, err := json.Marshal(a)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{
			"cause": a.Cause.String(),
			"json":  b,
		}

		tags := map[string]string{
			"severity": a.SeverityLevel.String(),
		}

		acc.AddFields("alert", fields, tags, time.Now())
	}

	return nil
}

func (g *GTFS) gather(req *http.Request) ([]*gtfs.FeedEntity, error) {
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed gtfs.FeedMessage
	err = proto.Unmarshal(body, &feed)
	if err != nil {
		return nil, err
	}

	return feed.Entity, nil
}

func (g *GTFS) newRequest(s string) (*http.Request, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if g.Key != "" {
		q := u.Query()
		q.Set("key", g.Key)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if g.Username != "" && g.Password != "" {
		req.SetBasicAuth(g.Username, g.Password)
	}

	return req, nil
}

func init() {
	inputs.Add("gtfs", func() telegraf.Input { return &GTFS{} })
}

var errNoSourceConfigured = errors.New("No source url configured; at least one of vehicle_positions_url, trip_updates_url or service_alerts_url must be populated")
