// +build darwin freebsd linux netbsd openbsd

package prometheus

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

func (p *Prometheus) harvestSocket(acc telegraf.Accumulator) filepath.WalkFunc {
	return func(file string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// check if we have a socket (os.ModeSocket is a bit mask)
		if fileInfo.Mode()&os.ModeSocket == 0 {
			return nil
		}

		return p.gatherSocketURL(file, acc)
	}
}

func (p *Prometheus) createHTTPSocketClient(socket string) (*http.Client, error) {
	tlsCfg, err := internal.GetTLSConfig(
		p.SSLCert, p.SSLKey, p.SSLCA, p.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	// ingest the special UNIX socket net.Dial function via closure
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsCfg,
			DisableKeepAlives: true,
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.Dial("unix", socket)
				return c, err
			},
		},
		Timeout: p.ResponseTimeout.Duration,
	}
	return client, nil
}

func (p *Prometheus) gatherSocketURL(socket string, acc telegraf.Accumulator) error {
	url := p.SocketURLPath

	socketName := path.Base(socket)

	client, err := p.createHTTPSocketClient(socket)
	if err != nil {
		return err
	}

	if url == "" {
		url = "/metrics"
	}

	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	// Prepare the request
	req, err := http.NewRequest("GET", "http://"+socketName+url, nil)
	req.Header.Add("Accept", acceptHeader)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request to socket %s (URL: %s) returned HTTP status %s", socket, url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	metrics, err := Parse(body, resp.Header)
	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s",
			url, err)
	}
	// Add (or not) collected metrics
	for _, metric := range metrics {
		// prefix the metric names with the socket name
		name := socketName + "_" + metric.Name()
		acc.AddFields(name, metric.Fields(), metric.Tags(), metric.Time())
	}

	return nil
}
