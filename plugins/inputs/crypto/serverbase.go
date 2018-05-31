package crypto

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

// required tags (addad by minerBase):
// * address		// address of the miner instance
// * source			// miner source
// * name			// miner friendly name
// * algorithm		// mining algorithm
// required tags:
// * version		// miner version
// * pool			// used pool
// required fields:
// * hashrate	    // integer, total hashrate in H/s or Sol/s etc.
// * uptime   		// integer, uptime of the miner in second
// recommended field names:
// * shares_total		// integer
// * shares_accepted	// integer
// * shares_rejected	// integer
// * shares_discarded	// integer
// * shares_rate		// float64
// per GPU/Unit based fields:
// * hashrate		// integer
// * temperature	// integer
// * fan			// integer
type serverGather interface {
	serverCount() int
	getAddress(i int) string
	getFriendlyName(i int) string
	serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error
}

var serverSampleConf = `
  interval = "1m"
  ## Servers addresses, names
  servers = ["localhost:80"]
  names   = ["hostname"]
`

type serverBase struct {
	Servers []string `toml:"servers"`
	Names   []string `toml:"names"`
}

func (m *serverBase) serverCount() int {
	return len(m.Servers)
}

func (m *serverBase) getAddress(i int) string {
	return m.Servers[i]
}

func (m *serverBase) getFriendlyName(i int) string {
	return m.Names[i]
}

// MinerGather for minerBase
func (m *serverBase) minerGather(acc telegraf.Accumulator, server serverGather) error {
	var wg sync.WaitGroup
	wg.Add(server.serverCount())
	for i := 0; i < server.serverCount(); i++ {
		tags := map[string]string{
			"address": server.getAddress(i),
			"name":    server.getFriendlyName(i),
		}
		go func(i int, tags map[string]string) {
			defer wg.Done()
			/*
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				host, port, err := net.SplitHostPort(address)
				if err != nil {
					acc.AddError(err)
					return
				}
				if len(port) > 0 {
					_, err := strconv.Atoi(port)
					if err != nil {
						acc.AddError(err)
						return
					}
				} else {
					port = "80"
				}
			*/
			acc.AddError(server.serverGather(acc, i, tags))
		}(i, tags)
	}
	wg.Wait()
	return nil
}

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: networkTimeout * time.Second,
		// KeepAlive: 0,
		DualStack: true,
	}).Dial,
	// MaxIdleConns:        100,
	// MaxIdleConnsPerHost: 2,
	// IdleConnTimeout:     90 * time.Second,
	// DisableKeepAlives:   true,
	TLSHandshakeTimeout: networkTimeout * time.Second,
}

func httpReader(URL string) (io.ReadCloser, error) {
	var httpClient = &http.Client{
		Timeout:   networkTimeout * time.Second,
		Transport: netTransport,
	}
	response, err := httpClient.Get(URL)
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}

func getResponse(url string, reply interface{}, errorPrefix string) bool {
	reader, err := httpReader(url)
	if err != nil { // we skip network problems
		log.Println(errorPrefix+" error: ", err, url)
		return false
	}
	defer reader.Close()

	err = json.NewDecoder(reader).Decode(&reply)
	if err != nil { // we skip network problems
		log.Println(errorPrefix+" error: ", err, url)
		return false
	}
	io.Copy(ioutil.Discard, reader)
	return true
}

func jsonReader(address string, command string, buf *bytes.Buffer) error {
	conn, err := net.DialTimeout("tcp", address, networkTimeout*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(networkTimeout * time.Second))

	w := bufio.NewWriter(conn)
	if _, err = w.WriteString(command); err != nil {
		return err
	}
	w.Flush()

	_, err = io.Copy(buf, conn)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
