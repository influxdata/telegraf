package nginx

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
)

/*
Structures built based on history of status module documentation
http://nginx.org/en/docs/http/ngx_http_status_module.html

Subsequent versions of status response structure available here:
1. http://web.archive.org/web/20130805111222/http://nginx.org/en/docs/http/ngx_http_status_module.html
2. http://web.archive.org/web/20131218101504/http://nginx.org/en/docs/http/ngx_http_status_module.html
3. not available
4. http://web.archive.org/web/20141218170938/http://nginx.org/en/docs/http/ngx_http_status_module.html
5. http://web.archive.org/web/20150414043916/http://nginx.org/en/docs/http/ngx_http_status_module.html
6. http://web.archive.org/web/20150918163811/http://nginx.org/en/docs/http/ngx_http_status_module.html
7. http://web.archive.org/web/20161107221028/http://nginx.org/en/docs/http/ngx_http_status_module.html
*/

type Status struct {
	Version       int    `json:"version"`
	NginxVersion  string `json:"nginx_version"`
	Address       string `json:"address"`
	Generation    *int   `json:"generation"`     // added in version 5
	LoadTimestamp *int64 `json:"load_timestamp"` // added in version 2
	Timestamp     int64  `json:"timestamp"`
	Pid           *int   `json:"pid"` // added in version 6

	Processes *struct { // added in version 5
		Respawned *int `json:"respawned"`
	} `json:"processes"`

	Connections struct {
		Accepted int `json:"accepted"`
		Dropped  int `json:"dropped"`
		Active   int `json:"active"`
		Idle     int `json:"idle"`
	} `json:"connections"`

	Ssl *struct { // added in version 6
		Handshakes       int64 `json:"handshakes"`
		HandshakesFailed int64 `json:"handshakes_failed"`
		SessionReuses    int64 `json:"session_reuses"`
	} `json:"ssl"`

	Requests struct {
		Total   int64 `json:"total"`
		Current int   `json:"current"`
	} `json:"requests"`

	ServerZones map[string]struct { // added in version 2
		Processing int   `json:"processing"`
		Requests   int64 `json:"requests"`
		Responses  struct {
			Responses1xx int64 `json:"1xx"`
			Responses2xx int64 `json:"2xx"`
			Responses3xx int64 `json:"3xx"`
			Responses4xx int64 `json:"4xx"`
			Responses5xx int64 `json:"5xx"`
			Total        int64 `json:"total"`
		} `json:"responses"`
		Discarded *int64 `json:"discarded"` // added in version 6
		Received  int64  `json:"received"`
		Sent      int64  `json:"sent"`
	} `json:"server_zones"`

	Upstreams map[string]struct {
		Peers []struct {
			ID        *int   `json:"id"` // added in version 3
			Server    string `json:"server"`
			Backup    bool   `json:"backup"`
			Weight    int    `json:"weight"`
			State     string `json:"state"`
			Active    int    `json:"active"`
			Keepalive *int   `json:"keepalive"` // removed in version 5
			MaxConns  *int   `json:"max_conns"` // added in version 3
			Requests  int64  `json:"requests"`
			Responses struct {
				Responses1xx int64 `json:"1xx"`
				Responses2xx int64 `json:"2xx"`
				Responses3xx int64 `json:"3xx"`
				Responses4xx int64 `json:"4xx"`
				Responses5xx int64 `json:"5xx"`
				Total        int64 `json:"total"`
			} `json:"responses"`
			Sent         int64 `json:"sent"`
			Received     int64 `json:"received"`
			Fails        int64 `json:"fails"`
			Unavail      int64 `json:"unavail"`
			HealthChecks struct {
				Checks     int64 `json:"checks"`
				Fails      int64 `json:"fails"`
				Unhealthy  int64 `json:"unhealthy"`
				LastPassed *bool `json:"last_passed"`
			} `json:"health_checks"`
			Downtime     int64  `json:"downtime"`
			Downstart    int64  `json:"downstart"`
			Selected     *int64 `json:"selected"`      // added in version 4
			HeaderTime   *int64 `json:"header_time"`   // added in version 5
			ResponseTime *int64 `json:"response_time"` // added in version 5
		} `json:"peers"`
		Keepalive int       `json:"keepalive"`
		Zombies   int       `json:"zombies"` // added in version 6
		Queue     *struct { // added in version 6
			Size      int   `json:"size"`
			MaxSize   int   `json:"max_size"`
			Overflows int64 `json:"overflows"`
		} `json:"queue"`
	} `json:"upstreams"`

	Caches map[string]struct { // added in version 2
		Size    int64 `json:"size"`
		MaxSize int64 `json:"max_size"`
		Cold    bool  `json:"cold"`
		Hit     struct {
			Responses int64 `json:"responses"`
			Bytes     int64 `json:"bytes"`
		} `json:"hit"`
		Stale struct {
			Responses int64 `json:"responses"`
			Bytes     int64 `json:"bytes"`
		} `json:"stale"`
		Updating struct {
			Responses int64 `json:"responses"`
			Bytes     int64 `json:"bytes"`
		} `json:"updating"`
		Revalidated *struct { // added in version 3
			Responses int64 `json:"responses"`
			Bytes     int64 `json:"bytes"`
		} `json:"revalidated"`
		Miss struct {
			Responses        int64 `json:"responses"`
			Bytes            int64 `json:"bytes"`
			ResponsesWritten int64 `json:"responses_written"`
			BytesWritten     int64 `json:"bytes_written"`
		} `json:"miss"`
		Expired struct {
			Responses        int64 `json:"responses"`
			Bytes            int64 `json:"bytes"`
			ResponsesWritten int64 `json:"responses_written"`
			BytesWritten     int64 `json:"bytes_written"`
		} `json:"expired"`
		Bypass struct {
			Responses        int64 `json:"responses"`
			Bytes            int64 `json:"bytes"`
			ResponsesWritten int64 `json:"responses_written"`
			BytesWritten     int64 `json:"bytes_written"`
		} `json:"bypass"`
	} `json:"caches"`

	Stream struct {
		ServerZones map[string]struct {
			Processing  int `json:"processing"`
			Connections int `json:"connections"`
			Sessions    *struct {
				Total       int64 `json:"total"`
				Sessions1xx int64 `json:"1xx"`
				Sessions2xx int64 `json:"2xx"`
				Sessions3xx int64 `json:"3xx"`
				Sessions4xx int64 `json:"4xx"`
				Sessions5xx int64 `json:"5xx"`
			} `json:"sessions"`
			Discarded *int64 `json:"discarded"` // added in version 7
			Received  int64  `json:"received"`
			Sent      int64  `json:"sent"`
		} `json:"server_zones"`
		Upstreams map[string]struct {
			Peers []struct {
				ID            int    `json:"id"`
				Server        string `json:"server"`
				Backup        bool   `json:"backup"`
				Weight        int    `json:"weight"`
				State         string `json:"state"`
				Active        int    `json:"active"`
				Connections   int64  `json:"connections"`
				ConnectTime   *int   `json:"connect_time"`
				FirstByteTime *int   `json:"first_byte_time"`
				ResponseTime  *int   `json:"response_time"`
				Sent          int64  `json:"sent"`
				Received      int64  `json:"received"`
				Fails         int64  `json:"fails"`
				Unavail       int64  `json:"unavail"`
				HealthChecks  struct {
					Checks     int64 `json:"checks"`
					Fails      int64 `json:"fails"`
					Unhealthy  int64 `json:"unhealthy"`
					LastPassed *bool `json:"last_passed"`
				} `json:"health_checks"`
				Downtime  int64 `json:"downtime"`
				Downstart int64 `json:"downstart"`
				Selected  int64 `json:"selected"`
			} `json:"peers"`
			Zombies int `json:"zombies"`
		} `json:"upstreams"`
	} `json:"stream"`
}

func gatherStatusUrl(r *bufio.Reader, tags map[string]string, acc telegraf.Accumulator) error {
	dec := json.NewDecoder(r)
	status := &Status{}
	if err := dec.Decode(status); err != nil {
		return fmt.Errorf("Error while decoding JSON response")
	}
	status.Gather(tags, acc)
	return nil
}

func (s *Status) Gather(tags map[string]string, acc telegraf.Accumulator) {
	s.gatherProcessesMetrics(tags, acc)
	s.gatherConnectionsMetrics(tags, acc)
	s.gatherSslMetrics(tags, acc)
	s.gatherRequestMetrics(tags, acc)
	s.gatherZoneMetrics(tags, acc)
	s.gatherUpstreamMetrics(tags, acc)
	s.gatherCacheMetrics(tags, acc)
	s.gatherStreamMetrics(tags, acc)
}

func (s *Status) gatherProcessesMetrics(tags map[string]string, acc telegraf.Accumulator) {
	acc.AddFields(
		"nginx.processes",
		map[string]interface{}{
			"respawned": s.Processes.Respawned,
		},
		tags,
	)

}

func (s *Status) gatherConnectionsMetrics(tags map[string]string, acc telegraf.Accumulator) {
	acc.AddFields(
		"nginx.connections",
		map[string]interface{}{
			"accepted": s.Connections.Accepted,
			"dropped":  s.Connections.Dropped,
			"active":   s.Connections.Active,
			"idle":     s.Connections.Idle,
		},
		tags,
	)
}

func (s *Status) gatherSslMetrics(tags map[string]string, acc telegraf.Accumulator) {
	acc.AddFields(
		"nginx.ssl",
		map[string]interface{}{
			"handshakes":        s.Ssl.Handshakes,
			"handshakes_failed": s.Ssl.HandshakesFailed,
			"session_reuses":    s.Ssl.SessionReuses,
		},
		tags,
	)
}

func (s *Status) gatherRequestMetrics(tags map[string]string, acc telegraf.Accumulator) {
	acc.AddFields(
		"nginx.requests",
		map[string]interface{}{
			"total":   s.Requests.Total,
			"current": s.Requests.Current,
		},
		tags,
	)
}

func (s *Status) gatherZoneMetrics(tags map[string]string, acc telegraf.Accumulator) {
	for zoneName, zone := range s.ServerZones {
		zoneTags := map[string]string{}
		for k, v := range tags {
			zoneTags[k] = v
		}
		zoneTags["zone"] = zoneName
		acc.AddFields(
			"nginx.zone",
			func() map[string]interface{} {
				result := map[string]interface{}{
					"processing":      zone.Processing,
					"requests":        zone.Requests,
					"responses.1xx":   zone.Responses.Responses1xx,
					"responses.2xx":   zone.Responses.Responses2xx,
					"responses.3xx":   zone.Responses.Responses3xx,
					"responses.4xx":   zone.Responses.Responses4xx,
					"responses.5xx":   zone.Responses.Responses5xx,
					"responses.total": zone.Responses.Total,
					"received":        zone.Received,
					"sent":            zone.Sent,
				}
				if zone.Discarded != nil {
					result["discarded"] = *zone.Discarded
				}
				return result
			}(),
			zoneTags,
		)
	}
}

func (s *Status) gatherUpstreamMetrics(tags map[string]string, acc telegraf.Accumulator) {
	for upstreamName, upstream := range s.Upstreams {
		upstreamTags := map[string]string{}
		for k, v := range tags {
			upstreamTags[k] = v
		}
		upstreamTags["upstream"] = upstreamName
		upstreamFields := map[string]interface{}{
			"keepalive": upstream.Keepalive,
			"zombies":   upstream.Zombies,
		}
		if upstream.Queue != nil {
			upstreamFields["queue.size"] = upstream.Queue.Size
			upstreamFields["queue.max_size"] = upstream.Queue.MaxSize
			upstreamFields["queue.overflows"] = upstream.Queue.Overflows
		}
		acc.AddFields(
			"nginx.upstream",
			upstreamFields,
			upstreamTags,
		)
		for _, peer := range upstream.Peers {
			peerFields := map[string]interface{}{
				"backup":                 peer.Backup,
				"weight":                 peer.Weight,
				"state":                  peer.State,
				"active":                 peer.Active,
				"requests":               peer.Requests,
				"responses.1xx":          peer.Responses.Responses1xx,
				"responses.2xx":          peer.Responses.Responses2xx,
				"responses.3xx":          peer.Responses.Responses3xx,
				"responses.4xx":          peer.Responses.Responses4xx,
				"responses.5xx":          peer.Responses.Responses5xx,
				"responses.total":        peer.Responses.Total,
				"sent":                   peer.Sent,
				"received":               peer.Received,
				"fails":                  peer.Fails,
				"unavail":                peer.Unavail,
				"healthchecks.checks":    peer.HealthChecks.Checks,
				"healthchecks.fails":     peer.HealthChecks.Fails,
				"healthchecks.unhealthy": peer.HealthChecks.Unhealthy,
				"downtime":               peer.Downtime,
				"downstart":              peer.Downstart,
				"selected":               peer.Selected,
			}
			if peer.HealthChecks.LastPassed != nil {
				peerFields["healthchecks.last_passed"] = *peer.HealthChecks.LastPassed
			}
			if peer.HeaderTime != nil {
				peerFields["header_time"] = *peer.HeaderTime
			}
			if peer.ResponseTime != nil {
				peerFields["response_time"] = *peer.ResponseTime
			}
			if peer.MaxConns != nil {
				peerFields["max_conns"] = *peer.MaxConns
			}
			peerTags := map[string]string{}
			for k, v := range upstreamTags {
				peerTags[k] = v
			}
			peerTags["serverAddress"] = peer.Server
			if peer.ID != nil {
				peerTags["id"] = strconv.Itoa(*peer.ID)
			}
			acc.AddFields("nginx.upstream.peer", peerFields, peerTags)
		}
	}
}

func (s *Status) gatherCacheMetrics(tags map[string]string, acc telegraf.Accumulator) {
	for cacheName, cache := range s.Caches {
		cacheTags := map[string]string{}
		for k, v := range tags {
			cacheTags[k] = v
		}
		cacheTags["cache"] = cacheName
		acc.AddFields(
			"nginx.cache",
			map[string]interface{}{
				"size":                      cache.Size,
				"max_size":                  cache.MaxSize,
				"cold":                      cache.Cold,
				"hit.responses":             cache.Hit.Responses,
				"hit.bytes":                 cache.Hit.Bytes,
				"stale.responses":           cache.Stale.Responses,
				"stale.bytes":               cache.Stale.Bytes,
				"updating.responses":        cache.Updating.Responses,
				"updating.bytes":            cache.Updating.Bytes,
				"revalidated.responses":     cache.Revalidated.Responses,
				"revalidated.bytes":         cache.Revalidated.Bytes,
				"miss.responses":            cache.Miss.Responses,
				"miss.bytes":                cache.Miss.Bytes,
				"miss.responses_written":    cache.Miss.ResponsesWritten,
				"miss.bytes_written":        cache.Miss.BytesWritten,
				"expired.responses":         cache.Expired.Responses,
				"expired.bytes":             cache.Expired.Bytes,
				"expired.responses_written": cache.Expired.ResponsesWritten,
				"expired.bytes_written":     cache.Expired.BytesWritten,
				"bypass.responses":          cache.Bypass.Responses,
				"bypass.bytes":              cache.Bypass.Bytes,
				"bypass.responses_written":  cache.Bypass.ResponsesWritten,
				"bypass.bytes_written":      cache.Bypass.BytesWritten,
			},
			cacheTags,
		)
	}
}

func (s *Status) gatherStreamMetrics(tags map[string]string, acc telegraf.Accumulator) {
	for zoneName, zone := range s.Stream.ServerZones {
		zoneTags := map[string]string{}
		for k, v := range tags {
			zoneTags[k] = v
		}
		zoneTags["zone"] = zoneName
		acc.AddFields(
			"nginx.stream.zone",
			map[string]interface{}{
				"processing":  zone.Processing,
				"connections": zone.Connections,
				"received":    zone.Received,
				"sent":        zone.Sent,
			},
			zoneTags,
		)
	}
	for upstreamName, upstream := range s.Stream.Upstreams {
		upstreamTags := map[string]string{}
		for k, v := range tags {
			upstreamTags[k] = v
		}
		upstreamTags["upstream"] = upstreamName
		acc.AddFields(
			"nginx.stream.upstream",
			map[string]interface{}{
				"zombies": upstream.Zombies,
			},
			upstreamTags,
		)
		for _, peer := range upstream.Peers {
			peerFields := map[string]interface{}{
				"backup":                 peer.Backup,
				"weight":                 peer.Weight,
				"state":                  peer.State,
				"active":                 peer.Active,
				"connections":            peer.Connections,
				"sent":                   peer.Sent,
				"received":               peer.Received,
				"fails":                  peer.Fails,
				"unavail":                peer.Unavail,
				"healthchecks.checks":    peer.HealthChecks.Checks,
				"healthchecks.fails":     peer.HealthChecks.Fails,
				"healthchecks.unhealthy": peer.HealthChecks.Unhealthy,
				"downtime":               peer.Downtime,
				"downstart":              peer.Downstart,
				"selected":               peer.Selected,
			}
			if peer.HealthChecks.LastPassed != nil {
				peerFields["healthchecks.last_passed"] = *peer.HealthChecks.LastPassed
			}
			if peer.ConnectTime != nil {
				peerFields["connect_time"] = *peer.ConnectTime
			}
			if peer.FirstByteTime != nil {
				peerFields["first_byte_time"] = *peer.FirstByteTime
			}
			if peer.ResponseTime != nil {
				peerFields["response_time"] = *peer.ResponseTime
			}
			peerTags := map[string]string{}
			for k, v := range upstreamTags {
				peerTags[k] = v
			}
			peerTags["serverAddress"] = peer.Server
			peerTags["id"] = strconv.Itoa(peer.ID)
			acc.AddFields("nginx.stream.upstream.peer", peerFields, peerTags)
		}
	}
}
