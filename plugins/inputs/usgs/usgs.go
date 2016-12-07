package usgs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// USGS is the top level struct for this plugin
type USGS struct {
	Ok bool
}

// Description contains a decription of the Plugin's function
func (gs *USGS) Description() string {
	return "a plugin to gather USGS earthquake data"
}

// SampleConfig returns a sample configuration for the plugin
func (gs *USGS) SampleConfig() string {
	return ""
}

// Gather makes the HTTP call and converts the data
func (gs *USGS) Gather(acc telegraf.Accumulator) error {
	resp, err := http.Get("https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson")
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	eqs := &Resp{}
	err = json.Unmarshal(body, eqs)
	if err != nil {
		return err
	}
	meas := "usgsdata"
	for _, feat := range eqs.Features {
		fields := map[string]interface{}{
			// Event latitude
			"lat": feat.Geometry.Coordinates[0],
			// Event longitude
			"lng": feat.Geometry.Coordinates[1],
			// Event depth
			"depth": feat.Geometry.Coordinates[2],
			// Earthquake intensity: http://earthquake.usgs.gov/learn/topics/mag_vs_int.php
			"intensity": feat.Properties.Cdi,
			// Link to detail for this Feature
			"detail": feat.Properties.Detail,
			// Horizontal distance from the epicenter to the nearest station (in degrees). 1 degree is approximately 111.2 kilometers.
			"dmin": feat.Properties.Dmin,
			// The total number of felt reports submitted to the DYFI? system.
			"felt": feat.Properties.Felt,
			// The largest azimuthal gap between azimuthally adjacent stations (in degrees). In general, the smaller this number, the more reliable
			"gap": int(feat.Properties.Gap),
			// The magnitude for the event
			"magnitude": feat.Properties.Mag,
			// Method of magnitude calculation: https://earthquake.usgs.gov/data/comcat/data-eventterms.php#magType
			"magnitudeType": feat.Properties.MagType,
			// The maximum estimated instrumental intensity for the event.
			"maxIntensity": feat.Properties.Mmi,
			// Human readable place name
			"place": feat.Properties.Place,
			// A number describing how significant the event is. Larger numbers indicate a more significant event.
			"significance": int(feat.Properties.Sig),
			// Link to USGS Event Page for event.
			"usgsEventPage": feat.Properties.URL,
		}
		tags := map[string]string{
			"latInt": coordToString(feat.Geometry.Coordinates[0]),
			"lngInt": coordToString(feat.Geometry.Coordinates[1]),
			// Alert is “green”, “yellow”, “orange”, “red”
			"alert": toString(feat.Properties.Alert),
			// The total number of seismic stations used to determine earthquake location.
			"numStations": toString(feat.Properties.Nst),
			// Indicates whether the event has been reviewed by a human -> “automatic”, “reviewed”, “deleted”
			"reviewStatus": toString(feat.Properties.Status),
			// This flag is set to "1" for large events in oceanic regions and "0" otherwise.
			"tsunami": toString(feat.Properties.Tsunami),
			// Type of siesmic event “earthquake”, “quarry”
			"eventType": toString(feat.Properties.Type),
			// UTC offset for event Timezone
			"utcOffset": toString(feat.Properties.Tz),
		}

		var t time.Time
		// Convert interface to int64
		updated := feat.Properties.Updated
		// Convert interface to int64
		original := feat.Properties.Time
		// If the event has been more reciently updated use that as the timestamp
		if updated > original {
			t = time.Unix(0, updated*int64(time.Millisecond))
		} else {
			t = time.Unix(0, original*int64(time.Millisecond))
		}
		acc.AddFields(meas, fields, tags, t)
	}
	return nil
}

func init() {
	inputs.Add("usgs", func() telegraf.Input { return &USGS{} })
}

func toString(s interface{}) string {
	return fmt.Sprintf("%v", s)
}

func coordToString(coord float64) string {
	foo := math.Floor(coord)
	return fmt.Sprintf("%d", int(foo))
}

// Resp is used to unmarshal the response body from USGS
type Resp struct {
	Type     string `json:"type"`
	Metadata struct {
		Generated int64  `json:"generated"`
		URL       string `json:"url"`
		Title     string `json:"title"`
		Status    int    `json:"status"`
		API       string `json:"api"`
		Count     int    `json:"count"`
	} `json:"metadata"`
	Features []struct {
		Type       string `json:"type"`
		Properties struct {
			Mag     float64 `json:"mag"`
			Place   string  `json:"place"`
			Time    int64   `json:"time"`
			Updated int64   `json:"updated"`
			Tz      float64 `json:"tz"`
			URL     string  `json:"url"`
			Detail  string  `json:"detail"`
			Felt    float64 `json:"felt"`
			Cdi     float64 `json:"cdi"`
			Mmi     float64 `json:"mmi"`
			Alert   string  `json:"alert"`
			Status  string  `json:"status"`
			Tsunami float64 `json:"tsunami"`
			Sig     float64 `json:"sig"`
			Net     string  `json:"net"`
			Code    string  `json:"code"`
			Ids     string  `json:"ids"`
			Sources string  `json:"sources"`
			Types   string  `json:"types"`
			Nst     float64 `json:"nst"`
			Dmin    float64 `json:"dmin"`
			Rms     float64 `json:"rms"`
			Gap     float64 `json:"gap"`
			MagType string  `json:"magType"`
			Type    string  `json:"type"`
			Title   string  `json:"title"`
		} `json:"properties"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		ID string `json:"id"`
	} `json:"features"`
	Bbox []float64 `json:"bbox"`
}
