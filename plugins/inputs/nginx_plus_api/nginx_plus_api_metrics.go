package nginx_plus_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

var (
	// errNotFound signals that the NGINX API routes does not exist.
	errNotFound = errors.New("not found")
)

func (n *NginxPlusAPI) gatherMetrics(addr *url.URL, acc telegraf.Accumulator) {
	addError(acc, n.gatherProcessesMetrics(addr, acc))
	addError(acc, n.gatherConnectionsMetrics(addr, acc))
	addError(acc, n.gatherSslMetrics(addr, acc))
	addError(acc, n.gatherHTTPRequestsMetrics(addr, acc))
	addError(acc, n.gatherHTTPServerZonesMetrics(addr, acc))
	addError(acc, n.gatherHTTPUpstreamsMetrics(addr, acc))
	addError(acc, n.gatherHTTPCachesMetrics(addr, acc))
	addError(acc, n.gatherStreamServerZonesMetrics(addr, acc))
	addError(acc, n.gatherStreamUpstreamsMetrics(addr, acc))

	if n.APIVersion >= 5 {
		addError(acc, n.gatherHTTPLocationZonesMetrics(addr, acc))
		addError(acc, n.gatherResolverZonesMetrics(addr, acc))
	}
}

func addError(acc telegraf.Accumulator, err error) {
	// This plugin has hardcoded API resource paths it checks that may not
	// be in the nginx.conf.  Currently, this is to prevent logging of
	// paths that are not configured.
	//
	// The correct solution is to do a GET to /api to get the available paths
	// on the server rather than simply ignore.
	if err != errNotFound {
		acc.AddError(err)
	}
}

func (n *NginxPlusAPI) gatherURL(addr *url.URL, path string) ([]byte, error) {
	address := fmt.Sprintf("%s/%d/%s", addr.String(), n.APIVersion, path)
	resp, err := n.client.Get(address)

	if err != nil {
		return nil, fmt.Errorf("error making HTTP request to %s: %s", address, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		// format as special error to catch and ignore as some nginx API
		// features are either optional, or only available in some versions
		return nil, errNotFound
	default:
		return nil, fmt.Errorf("%s returned HTTP status %s", address, resp.Status)
	}

	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	switch contentType {
	case "application/json":
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return body, nil
	default:
		return nil, fmt.Errorf("%s returned unexpected content type %s", address, contentType)
	}
}

func (n *NginxPlusAPI) gatherProcessesMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, processesPath)
	if err != nil {
		return err
	}

	var processes = &Processes{}

	if err := json.Unmarshal(body, processes); err != nil {
		return err
	}

	acc.AddFields(
		"nginx_plus_api_processes",
		map[string]interface{}{
			"respawned": processes.Respawned,
		},
		getTags(addr),
	)

	return nil
}

func (n *NginxPlusAPI) gatherConnectionsMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, connectionsPath)
	if err != nil {
		return err
	}

	var connections = &Connections{}

	if err := json.Unmarshal(body, connections); err != nil {
		return err
	}

	acc.AddFields(
		"nginx_plus_api_connections",
		map[string]interface{}{
			"accepted": connections.Accepted,
			"dropped":  connections.Dropped,
			"active":   connections.Active,
			"idle":     connections.Idle,
		},
		getTags(addr),
	)

	return nil
}

func (n *NginxPlusAPI) gatherSslMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, sslPath)
	if err != nil {
		return err
	}

	var ssl = &Ssl{}

	if err := json.Unmarshal(body, ssl); err != nil {
		return err
	}

	acc.AddFields(
		"nginx_plus_api_ssl",
		map[string]interface{}{
			"handshakes":        ssl.Handshakes,
			"handshakes_failed": ssl.HandshakesFailed,
			"session_reuses":    ssl.SessionReuses,
		},
		getTags(addr),
	)

	return nil
}

func (n *NginxPlusAPI) gatherHTTPRequestsMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, httpRequestsPath)
	if err != nil {
		return err
	}

	var httpRequests = &HTTPRequests{}

	if err := json.Unmarshal(body, httpRequests); err != nil {
		return err
	}

	acc.AddFields(
		"nginx_plus_api_http_requests",
		map[string]interface{}{
			"total":   httpRequests.Total,
			"current": httpRequests.Current,
		},
		getTags(addr),
	)

	return nil
}

func (n *NginxPlusAPI) gatherHTTPServerZonesMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, httpServerZonesPath)
	if err != nil {
		return err
	}

	var httpServerZones HTTPServerZones

	if err := json.Unmarshal(body, &httpServerZones); err != nil {
		return err
	}

	tags := getTags(addr)

	for zoneName, zone := range httpServerZones {
		zoneTags := map[string]string{}
		for k, v := range tags {
			zoneTags[k] = v
		}
		zoneTags["zone"] = zoneName
		acc.AddFields(
			"nginx_plus_api_http_server_zones",
			func() map[string]interface{} {
				result := map[string]interface{}{
					"processing":      zone.Processing,
					"requests":        zone.Requests,
					"responses_1xx":   zone.Responses.Responses1xx,
					"responses_2xx":   zone.Responses.Responses2xx,
					"responses_3xx":   zone.Responses.Responses3xx,
					"responses_4xx":   zone.Responses.Responses4xx,
					"responses_5xx":   zone.Responses.Responses5xx,
					"responses_total": zone.Responses.Total,
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

	return nil
}

// Added in 5 API version
func (n *NginxPlusAPI) gatherHTTPLocationZonesMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, httpLocationZonesPath)
	if err != nil {
		return err
	}

	var httpLocationZones HTTPLocationZones

	if err := json.Unmarshal(body, &httpLocationZones); err != nil {
		return err
	}

	tags := getTags(addr)

	for zoneName, zone := range httpLocationZones {
		zoneTags := map[string]string{}
		for k, v := range tags {
			zoneTags[k] = v
		}
		zoneTags["zone"] = zoneName
		acc.AddFields(
			"nginx_plus_api_http_location_zones",
			func() map[string]interface{} {
				result := map[string]interface{}{
					"requests":        zone.Requests,
					"responses_1xx":   zone.Responses.Responses1xx,
					"responses_2xx":   zone.Responses.Responses2xx,
					"responses_3xx":   zone.Responses.Responses3xx,
					"responses_4xx":   zone.Responses.Responses4xx,
					"responses_5xx":   zone.Responses.Responses5xx,
					"responses_total": zone.Responses.Total,
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

	return nil
}

func (n *NginxPlusAPI) gatherHTTPUpstreamsMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, httpUpstreamsPath)
	if err != nil {
		return err
	}

	var httpUpstreams HTTPUpstreams

	if err := json.Unmarshal(body, &httpUpstreams); err != nil {
		return err
	}

	tags := getTags(addr)

	for upstreamName, upstream := range httpUpstreams {
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
			upstreamFields["queue_size"] = upstream.Queue.Size
			upstreamFields["queue_max_size"] = upstream.Queue.MaxSize
			upstreamFields["queue_overflows"] = upstream.Queue.Overflows
		}
		acc.AddFields(
			"nginx_plus_api_http_upstreams",
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
				"responses_1xx":          peer.Responses.Responses1xx,
				"responses_2xx":          peer.Responses.Responses2xx,
				"responses_3xx":          peer.Responses.Responses3xx,
				"responses_4xx":          peer.Responses.Responses4xx,
				"responses_5xx":          peer.Responses.Responses5xx,
				"responses_total":        peer.Responses.Total,
				"sent":                   peer.Sent,
				"received":               peer.Received,
				"fails":                  peer.Fails,
				"unavail":                peer.Unavail,
				"healthchecks_checks":    peer.HealthChecks.Checks,
				"healthchecks_fails":     peer.HealthChecks.Fails,
				"healthchecks_unhealthy": peer.HealthChecks.Unhealthy,
				"downtime":               peer.Downtime,
				//"selected":               peer.Selected.toInt64,
				//"downstart":              peer.Downstart.toInt64,
			}
			if peer.HealthChecks.LastPassed != nil {
				peerFields["healthchecks_last_passed"] = *peer.HealthChecks.LastPassed
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
			peerTags["upstream_address"] = peer.Server
			if peer.ID != nil {
				peerTags["id"] = strconv.Itoa(*peer.ID)
			}
			acc.AddFields("nginx_plus_api_http_upstream_peers", peerFields, peerTags)
		}
	}
	return nil
}

func (n *NginxPlusAPI) gatherHTTPCachesMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, httpCachesPath)
	if err != nil {
		return err
	}

	var httpCaches HTTPCaches

	if err := json.Unmarshal(body, &httpCaches); err != nil {
		return err
	}

	tags := getTags(addr)

	for cacheName, cache := range httpCaches {
		cacheTags := map[string]string{}
		for k, v := range tags {
			cacheTags[k] = v
		}
		cacheTags["cache"] = cacheName
		acc.AddFields(
			"nginx_plus_api_http_caches",
			map[string]interface{}{
				"size":                      cache.Size,
				"max_size":                  cache.MaxSize,
				"cold":                      cache.Cold,
				"hit_responses":             cache.Hit.Responses,
				"hit_bytes":                 cache.Hit.Bytes,
				"stale_responses":           cache.Stale.Responses,
				"stale_bytes":               cache.Stale.Bytes,
				"updating_responses":        cache.Updating.Responses,
				"updating_bytes":            cache.Updating.Bytes,
				"revalidated_responses":     cache.Revalidated.Responses,
				"revalidated_bytes":         cache.Revalidated.Bytes,
				"miss_responses":            cache.Miss.Responses,
				"miss_bytes":                cache.Miss.Bytes,
				"miss_responses_written":    cache.Miss.ResponsesWritten,
				"miss_bytes_written":        cache.Miss.BytesWritten,
				"expired_responses":         cache.Expired.Responses,
				"expired_bytes":             cache.Expired.Bytes,
				"expired_responses_written": cache.Expired.ResponsesWritten,
				"expired_bytes_written":     cache.Expired.BytesWritten,
				"bypass_responses":          cache.Bypass.Responses,
				"bypass_bytes":              cache.Bypass.Bytes,
				"bypass_responses_written":  cache.Bypass.ResponsesWritten,
				"bypass_bytes_written":      cache.Bypass.BytesWritten,
			},
			cacheTags,
		)
	}

	return nil
}

func (n *NginxPlusAPI) gatherStreamServerZonesMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, streamServerZonesPath)
	if err != nil {
		return err
	}

	var streamServerZones StreamServerZones

	if err := json.Unmarshal(body, &streamServerZones); err != nil {
		return err
	}

	tags := getTags(addr)

	for zoneName, zone := range streamServerZones {
		zoneTags := map[string]string{}
		for k, v := range tags {
			zoneTags[k] = v
		}
		zoneTags["zone"] = zoneName
		acc.AddFields(
			"nginx_plus_api_stream_server_zones",
			map[string]interface{}{
				"processing":  zone.Processing,
				"connections": zone.Connections,
				"received":    zone.Received,
				"sent":        zone.Sent,
			},
			zoneTags,
		)
	}

	return nil
}

// Added in 5 API version
func (n *NginxPlusAPI) gatherResolverZonesMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, resolverZonesPath)
	if err != nil {
		return err
	}

	var resolverZones ResolverZones

	if err := json.Unmarshal(body, &resolverZones); err != nil {
		return err
	}

	tags := getTags(addr)

	for zoneName, resolver := range resolverZones {
		zoneTags := map[string]string{}
		for k, v := range tags {
			zoneTags[k] = v
		}
		zoneTags["zone"] = zoneName
		acc.AddFields(
			"nginx_plus_api_resolver_zones",
			map[string]interface{}{
				"name": resolver.Requests.Name,
				"srv":  resolver.Requests.Srv,
				"addr": resolver.Requests.Addr,

				"noerror":  resolver.Responses.Noerror,
				"formerr":  resolver.Responses.Formerr,
				"servfail": resolver.Responses.Servfail,
				"nxdomain": resolver.Responses.Nxdomain,
				"notimp":   resolver.Responses.Notimp,
				"refused":  resolver.Responses.Refused,
				"timedout": resolver.Responses.Timedout,
				"unknown":  resolver.Responses.Unknown,
			},
			zoneTags,
		)
	}

	return nil
}

func (n *NginxPlusAPI) gatherStreamUpstreamsMetrics(addr *url.URL, acc telegraf.Accumulator) error {
	body, err := n.gatherURL(addr, streamUpstreamsPath)
	if err != nil {
		return err
	}

	var streamUpstreams StreamUpstreams

	if err := json.Unmarshal(body, &streamUpstreams); err != nil {
		return err
	}

	tags := getTags(addr)

	for upstreamName, upstream := range streamUpstreams {
		upstreamTags := map[string]string{}
		for k, v := range tags {
			upstreamTags[k] = v
		}
		upstreamTags["upstream"] = upstreamName
		acc.AddFields(
			"nginx_plus_api_stream_upstreams",
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
				"healthchecks_checks":    peer.HealthChecks.Checks,
				"healthchecks_fails":     peer.HealthChecks.Fails,
				"healthchecks_unhealthy": peer.HealthChecks.Unhealthy,
				"downtime":               peer.Downtime,
			}
			if peer.HealthChecks.LastPassed != nil {
				peerFields["healthchecks_last_passed"] = *peer.HealthChecks.LastPassed
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
			peerTags["upstream_address"] = peer.Server
			peerTags["id"] = strconv.Itoa(peer.ID)

			acc.AddFields("nginx_plus_api_stream_upstream_peers", peerFields, peerTags)
		}
	}

	return nil
}

func getTags(addr *url.URL) map[string]string {
	h := addr.Host
	host, port, err := net.SplitHostPort(h)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}
	return map[string]string{"source": host, "port": port}
}
