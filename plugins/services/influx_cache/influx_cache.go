package influx_cache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/services"
	"github.com/influxdata/telegraf/pubsub"
	"github.com/syndtr/goleveldb/leveldb"
)

type TimeFunc func() time.Time

type CachingTransport struct {
	transport       http.RoundTripper
	db              *leveldb.DB
	queryMatchEntry []QueryMatchEntry

	timerCache map[string]time.Time
	cache      map[string][]byte
	cacheMutex sync.RWMutex
}

type QueryMatchEntry struct {
	QueryMatch string
	ClearEvery string

	queryMatch *regexp.Regexp
	clearEvery time.Duration
}

type InfluxCache struct {
	Address         string
	InfluxDBAddress string `toml:"influx_db_address"`
	TLSCert         string `toml:"tls_cert"`
	TLSKey          string `toml:"tls_key"`
	CachePath       string
	QueryMatch      []QueryMatchEntry

	rp     *httputil.ReverseProxy
	scheme string
}

func (h *InfluxCache) Description() string {
	return "Receive http web requests"
}

var sampleConfig = `
`

func (t *CachingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.cacheMutex.RLock()
	timeSince, exists := t.timerCache[req.URL.RawQuery]
	t.cacheMutex.RUnlock()
	if exists {
		var queryMatch *QueryMatchEntry
		for i, qme := range t.queryMatchEntry {
			if qme.queryMatch.MatchString(req.URL.RawQuery) {
				queryMatch = &t.queryMatchEntry[i]
				break
			}
		}
		if queryMatch != nil {
			if time.Since(timeSince) <= queryMatch.clearEvery {
				fmt.Println("Cache HIT", timeSince, req.URL.RawQuery)

				byres, err := t.db.Get([]byte(req.URL.RawQuery), nil)
				if err != nil {
					log.Println("Error retrieving cache", req.URL.RawQuery, err)
				}

				res := &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
					Body:       ioutil.NopCloser(bytes.NewBuffer(byres)),
					Request:    req,
					Header:     make(http.Header, 0),
				}

				now := time.Now().UTC()

				res.Header.Add("Access-Control-Allow-Headers", "Accept, Accept-Encoding, Authorization, Content-Length, Content-Type, X-CSRF-Token, X-HTTP-Method-Override")
				res.Header.Add("Access-Control-Allow-Methods", "DELETE, GET, OPTIONS, POST, PUT")
				res.Header.Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
				res.Header.Add("Access-Control-Expose-Headers", "Date, X-InfluxDB-Version, X-InfluxDB-Build")
				res.Header.Add("Content-Type", "application/json")
				res.Header.Add("Content-Encoding", "gzip")
				res.Header.Add("Transfer-Encoding", "chunked")
				res.Header.Add("Date", now.Format("Mon, 02 Jan 2006 15:04:05 MST"))
				res.Header.Add("x-influxdb-build", "OSS")
				res.Header.Add("x-influxdb-version", "1.6.1")
				res.Header.Add("x-cache-status", "CACHE-HIT")
				res.Header.Add("x-cache-duration", queryMatch.ClearEvery)
				res.Header.Add("x-cache-time", timeSince.String())

				return res, nil
			} else {
				// This cache entry has expired. Clean it up.
				err := t.db.Delete([]byte(req.URL.RawQuery), nil)
				if err != nil {
					log.Println("Error deleting", req.URL.RawQuery, err)
				}
			}
		}
	}

	res, err := t.transport.RoundTrip(req)
	if err != nil {
		fmt.Println("Transport error", err)
		return nil, err
	}
	if res != nil && res.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		err = t.db.Put([]byte(req.URL.RawQuery), b, nil)
		if err != nil {
			return nil, err
		}

		now := time.Now()
		t.cacheMutex.Lock()
		t.timerCache[req.URL.RawQuery] = now
		t.cacheMutex.Unlock()
		res.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		res.Header.Add("x-cache-status", "CACHE-MISS")
		res.Header.Add("x-cache-time", now.String())

	}

	return res, err
}

func (h *InfluxCache) SampleConfig() string {
	return sampleConfig
}

func (h *InfluxCache) Connect() error {
	db, err := leveldb.OpenFile(h.CachePath, nil)
	if err != nil {
		log.Fatalf("Could not open leveldb", err)
	}

	if len(h.TLSCert) > 0 && len(h.TLSKey) > 0 {
		h.scheme = "https"
	} else {
		h.scheme = "http"
	}

	for i, qme := range h.QueryMatch {
		h.QueryMatch[i].queryMatch = regexp.MustCompile(qme.QueryMatch)
		h.QueryMatch[i].clearEvery, err = time.ParseDuration(qme.ClearEvery)
		if err != nil {
			log.Fatalf("Could not parse duration", err)
		}
	}

	h.rp = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			req := r
			req.URL.Scheme = h.scheme
			req.URL.Host = h.InfluxDBAddress
		},
		Transport: &CachingTransport{
			transport:       http.DefaultTransport,
			db:              db,
			queryMatchEntry: h.QueryMatch,
			timerCache:      make(map[string]time.Time),
			cache:           make(map[string][]byte),
			cacheMutex:      sync.RWMutex{},
		},
	}

	return nil
}

func (h *InfluxCache) Run(msgbus *pubsub.PubSub) error {
	var err error
	switch h.scheme {
	case "https":
		err = http.ListenAndServeTLS(h.Address, h.TLSCert, h.TLSKey, h.rp)
	default:
		err = http.ListenAndServe(h.Address, h.rp)
	}

	if err != nil {
		log.Fatalf("Could not start http server", err)
	}

	return err
}

func (h *InfluxCache) Close() error {
	fmt.Println("Influx Cache Close")
	return nil
}

func setCacheEntry() {

}

func getCacheEntry() {

}

func init() {
	services.Add("influx_cache", func() telegraf.Service {
		return &InfluxCache{}
	})
}
