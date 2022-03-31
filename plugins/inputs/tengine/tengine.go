package tengine

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"io"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Tengine struct {
	Urls            []string
	ResponseTimeout config.Duration
	tls.ClientConfig

	client *http.Client
}

func (n *Tengine) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Create an HTTP client that is re-used for each
	// collection interval
	if n.client == nil {
		client, err := n.createHTTPClient()
		if err != nil {
			return err
		}
		n.client = client
	}

	for _, u := range n.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(n.gatherURL(addr, acc))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (n *Tengine) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if n.ResponseTimeout < config.Duration(time.Second) {
		n.ResponseTimeout = config.Duration(time.Second * 5)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(n.ResponseTimeout),
	}

	return client, nil
}

type TengineStatus struct {
	host                  string
	bytesIn               uint64
	bytesOut              uint64
	connTotal             uint64
	reqTotal              uint64
	http2xx               uint64
	http3xx               uint64
	http4xx               uint64
	http5xx               uint64
	httpOtherStatus       uint64
	rt                    uint64
	upsReq                uint64
	upsRt                 uint64
	upsTries              uint64
	http200               uint64
	http206               uint64
	http302               uint64
	http304               uint64
	http403               uint64
	http404               uint64
	http416               uint64
	http499               uint64
	http500               uint64
	http502               uint64
	http503               uint64
	http504               uint64
	http508               uint64
	httpOtherDetailStatus uint64
	httpUps4xx            uint64
	httpUps5xx            uint64
}

func (n *Tengine) gatherURL(addr *url.URL, acc telegraf.Accumulator) error {
	var tengineStatus TengineStatus
	resp, err := n.client.Get(addr.String())
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}
	r := bufio.NewReader(resp.Body)

	for {
		line, err := r.ReadString('\n')

		if err != nil || io.EOF == err {
			break
		}
		lineSplit := strings.Split(strings.TrimSpace(line), ",")
		if len(lineSplit) != 30 {
			continue
		}
		tengineStatus.host = lineSplit[0]
		if err != nil {
			return err
		}
		tengineStatus.bytesIn, err = strconv.ParseUint(lineSplit[1], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.bytesOut, err = strconv.ParseUint(lineSplit[2], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.connTotal, err = strconv.ParseUint(lineSplit[3], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.reqTotal, err = strconv.ParseUint(lineSplit[4], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http2xx, err = strconv.ParseUint(lineSplit[5], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http3xx, err = strconv.ParseUint(lineSplit[6], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http4xx, err = strconv.ParseUint(lineSplit[7], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http5xx, err = strconv.ParseUint(lineSplit[8], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.httpOtherStatus, err = strconv.ParseUint(lineSplit[9], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.rt, err = strconv.ParseUint(lineSplit[10], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.upsReq, err = strconv.ParseUint(lineSplit[11], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.upsRt, err = strconv.ParseUint(lineSplit[12], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.upsTries, err = strconv.ParseUint(lineSplit[13], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http200, err = strconv.ParseUint(lineSplit[14], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http206, err = strconv.ParseUint(lineSplit[15], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http302, err = strconv.ParseUint(lineSplit[16], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http304, err = strconv.ParseUint(lineSplit[17], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http403, err = strconv.ParseUint(lineSplit[18], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http404, err = strconv.ParseUint(lineSplit[19], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http416, err = strconv.ParseUint(lineSplit[20], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http499, err = strconv.ParseUint(lineSplit[21], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http500, err = strconv.ParseUint(lineSplit[22], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http502, err = strconv.ParseUint(lineSplit[23], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http503, err = strconv.ParseUint(lineSplit[24], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http504, err = strconv.ParseUint(lineSplit[25], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.http508, err = strconv.ParseUint(lineSplit[26], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.httpOtherDetailStatus, err = strconv.ParseUint(lineSplit[27], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.httpUps4xx, err = strconv.ParseUint(lineSplit[28], 10, 64)
		if err != nil {
			return err
		}
		tengineStatus.httpUps5xx, err = strconv.ParseUint(lineSplit[29], 10, 64)
		if err != nil {
			return err
		}
		tags := getTags(addr, tengineStatus.host)
		fields := map[string]interface{}{
			"bytes_in":                 tengineStatus.bytesIn,
			"bytes_out":                tengineStatus.bytesOut,
			"conn_total":               tengineStatus.connTotal,
			"req_total":                tengineStatus.reqTotal,
			"http_2xx":                 tengineStatus.http2xx,
			"http_3xx":                 tengineStatus.http3xx,
			"http_4xx":                 tengineStatus.http4xx,
			"http_5xx":                 tengineStatus.http5xx,
			"http_other_status":        tengineStatus.httpOtherStatus,
			"rt":                       tengineStatus.rt,
			"ups_req":                  tengineStatus.upsReq,
			"ups_rt":                   tengineStatus.upsRt,
			"ups_tries":                tengineStatus.upsTries,
			"http_200":                 tengineStatus.http200,
			"http_206":                 tengineStatus.http206,
			"http_302":                 tengineStatus.http302,
			"http_304":                 tengineStatus.http304,
			"http_403":                 tengineStatus.http403,
			"http_404":                 tengineStatus.http404,
			"http_416":                 tengineStatus.http416,
			"http_499":                 tengineStatus.http499,
			"http_500":                 tengineStatus.http500,
			"http_502":                 tengineStatus.http502,
			"http_503":                 tengineStatus.http503,
			"http_504":                 tengineStatus.http504,
			"http_508":                 tengineStatus.http508,
			"http_other_detail_status": tengineStatus.httpOtherDetailStatus,
			"http_ups_4xx":             tengineStatus.httpUps4xx,
			"http_ups_5xx":             tengineStatus.httpUps5xx,
		}
		acc.AddFields("tengine", fields, tags)
	}

	// Return the potential error of the loop-read
	return err
}

// Get tag(s) for the tengine plugin
func getTags(addr *url.URL, serverName string) map[string]string {
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
	return map[string]string{"server": host, "port": port, "server_name": serverName}
}

func init() {
	inputs.Add("tengine", func() telegraf.Input {
		return &Tengine{}
	})
}
