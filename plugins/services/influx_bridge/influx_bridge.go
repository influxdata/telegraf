package influx_bridge

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"time"

	"github.com/influxdata/influxdb/query"
	"github.com/influxdata/influxdb/services/httpd"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/services"
	"github.com/influxdata/telegraf/pubsub"
)

type TimeFunc func() time.Time

type QueryMatchEntry struct {
	QueryMatch string
	MatchPass  []string
	MatchDrop  []string

	queryMatch *regexp.Regexp
	matchPass  []*regexp.Regexp
	matchDrop  []*regexp.Regexp
}

type InfluxBridge struct {
	Address         string
	InfluxDBAddress string `toml:"influx_db_address"`
	TLSCert         string `toml:"tls_cert"`
	TLSKey          string `toml:"tls_key"`
	QueryMatch      []QueryMatchEntry

	rp     *httputil.ReverseProxy
	scheme string
}

func (h *InfluxBridge) Description() string {
	return "Receive http web requests"
}

var sampleConfig = `
`

func (h *InfluxBridge) SampleConfig() string {
	return sampleConfig
}

func (h *InfluxBridge) Connect() error {
	fmt.Println("Http Connect")
	if len(h.TLSCert) > 0 && len(h.TLSKey) > 0 {
		h.scheme = "https"
	} else {
		h.scheme = "http"
	}

	for q := range h.QueryMatch {
		qme := &h.QueryMatch[q]
		qme.queryMatch = regexp.MustCompile(qme.QueryMatch)
		for _, mp := range qme.MatchPass {
			qme.matchPass = append(qme.matchPass, regexp.MustCompile(mp))
		}

		for _, md := range qme.MatchDrop {
			qme.matchDrop = append(qme.matchDrop, regexp.MustCompile(md))
		}
	}

	h.rp = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			req := r
			req.URL.Scheme = h.scheme
			req.URL.Host = h.InfluxDBAddress
		},
		ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode == http.StatusOK {
				gzReader, _ := gzip.NewReader(resp.Body)
				defer gzReader.Close()
				bodyBytes, _ := ioutil.ReadAll(gzReader)
				response := httpd.Response{Results: make([]*query.Result, 0)}
				err := response.UnmarshalJSON(bodyBytes)
				if err != nil {
					fmt.Println("JSON Parse Error", err)
				}

				query := resp.Request.URL.Query()["q"][0]
				var qm *QueryMatchEntry
				for _, qme := range h.QueryMatch {
					if qme.queryMatch.MatchString(query) {
						qm = &qme
						break
					}
				}

				for result := range response.Results {
					for ser := range response.Results[result].Series {
						var values [][]interface{}
						for val := range response.Results[result].Series[ser].Values {
							found := false
							for _, item := range response.Results[result].Series[ser].Values[val] {
								if qm != nil {
									for _, rePass := range qm.matchPass {
										if rePass.MatchString(item.(string)) {
											found = true
											break
										}
									}
								}
							}

							if found {
								for _, item := range response.Results[result].Series[ser].Values[val] {
									if qm != nil {
										for _, reDrop := range qm.matchDrop {
											if reDrop.MatchString(item.(string)) {
												found = false
											}
										}
									}
								}
							}

							if found {
								values = append(values, response.Results[result].Series[ser].Values[val])
							}

						}

						if len(values) > 0 {
							// We need to add a new series item.
							response.Results[result].Series[ser].Values = values
						}
					}

				}

				by, err := response.MarshalJSON()
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				if _, err := gz.Write(by); err != nil {
					panic(err)
				}
				if err := gz.Flush(); err != nil {
					panic(err)
				}
				if err := gz.Close(); err != nil {
					panic(err)
				}

				resp.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))
				resp.ContentLength = int64(len(b.Bytes()))
			}
			return nil
		},
	}
	return nil
}

func (h *InfluxBridge) Run(msgbus *pubsub.PubSub) error {
	//mux := gorouter.New()
	// mux.GET("/query", h.rp.ServeHTTP)
	var err error
	switch h.scheme {
	case "https":
		// err = http.ListenAndServeTLS(h.Address, h.TLSCert, h.TLSKey, mux)
		err = http.ListenAndServeTLS(h.Address, h.TLSCert, h.TLSKey, h.rp)
	default:
		err = http.ListenAndServe(h.Address, h.rp)
	}

	log.Println("exit")
	return err
}

func (h *InfluxBridge) Close() error {
	fmt.Println("Influx Bridge Close")
	return nil
}

func init() {
	services.Add("influx_bridge", func() telegraf.Service {
		return &InfluxBridge{}
	})
}
