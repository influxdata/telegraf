package haproxy

import (
        "encoding/csv"
        "fmt"
        "github.com/influxdata/telegraf"
        "github.com/influxdata/telegraf/plugins/inputs"
        "io"
        "net"
        "net/http"
        "net/url"
        "strconv"
        "strings"
        "sync"
        "time"
)

type haproxy struct {
        Servers []string

        client *http.Client
}

var sampleConfig = `
  ## An array of address to gather stats about. Specify an ip on hostname
  ## with optional port. ie localhost, 10.10.3.33:1936, etc.
  ## If no servers are specified, then default to 127.0.0.1:1936
  servers = ["http://myhaproxy.com:1936", "http://anotherhaproxy.com:1936"]
  ## Or you can also use local socket
  ## servers = ["socket:/run/haproxy/admin.sock"]
`

func (r *haproxy) SampleConfig() string {
        return sampleConfig
}

func (r *haproxy) Description() string {
        return "Read metrics of haproxy, via socket or csv stats page"
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *haproxy) Gather(acc telegraf.Accumulator) error {
        if len(g.Servers) == 0 {
                return g.gatherServer("http://127.0.0.1:1936", acc)
        }

        var wg sync.WaitGroup

        var outerr error

        for _, serv := range g.Servers {
                wg.Add(1)
                go func(serv string) {
                        defer wg.Done()
                        outerr = g.gatherServer(serv, acc)
                }(serv)
        }

        wg.Wait()

        return outerr
}

func (g *haproxy) gatherServerSocket(addr string, acc telegraf.Accumulator) error {
        var socketPath string
        socketAddr := strings.Split(addr, ":")

        if len(socketAddr) >= 2 {
                socketPath = socketAddr[1]
        } else {
                socketPath = socketAddr[0]
        }

        c, err := net.Dial("unix", socketPath)

        if err != nil {
                return fmt.Errorf("Could not connect to socket '%s': %s", addr, err)
        }

        _, errw := c.Write([]byte("show stat\n"))

        if errw != nil {
                return fmt.Errorf("Could not write to socket '%s': %s", addr, errw)
        }

        return importCsvResult(c, acc, socketPath)
}

func (g *haproxy) gatherServer(addr string, acc telegraf.Accumulator) error {
        if !strings.HasPrefix(addr, "http") {
                return g.gatherServerSocket(addr, acc)
        }

        if g.client == nil {
                tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
                client := &http.Client{
                        Transport: tr,
                        Timeout:   time.Duration(4 * time.Second),
                }
                g.client = client
        }

        u, err := url.Parse(addr)
        if err != nil {
                return fmt.Errorf("Unable parse server address '%s': %s", addr, err)
        }

        req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s%s/;csv", u.Scheme, u.Host, u.Path), nil)
        if u.User != nil {
                p, _ := u.User.Password()
                req.SetBasicAuth(u.User.Username(), p)
        }

        res, err := g.client.Do(req)
        if err != nil {
                return fmt.Errorf("Unable to connect to haproxy server '%s': %s", addr, err)
        }

        if res.StatusCode != 200 {
                return fmt.Errorf("Unable to get valid stat result from '%s': %s", addr, err)
        }

        return importCsvResult(res.Body, acc, u.Host)
}

func importCsvResult(r io.Reader, acc telegraf.Accumulator, host string) error {
        csv := csv.NewReader(r)
        result, err := csv.ReadAll()
        now := time.Now()

        var keys []string
        px_sv_status := make(map[string]map[string]int)
        non_stat_fields := map[string]bool{
                "pxname":       true,
                "svname":       true,
                "status":       true,
                "tracked":      true,
                "check_status": true,
                "last_chk":     true,
                "pid":          true,
                "iid":          true,
                "sid":          true,
                "lastchg":      true,
                "type":         true,
                "check_code":   true,
        }

        for i := range result {
                if i == 0 {
                        keys = result[i]
                        keys[0] = strings.Replace(keys[0], "# ", "", -1)
                } else {

                        row := make(map[string]string, len(result[i]))

                        for f, v := range result[i] {
                                row[keys[f]] = v
                        }

                        tags := map[string]string{
                                "server": host,
                                "proxy":  row["pxname"],
                                "sv":     row["svname"],
                        }

                        if row["svname"] != "BACKEND" && row["svname"] != "FRONTEND" {
                                if len(px_sv_status[row["pxname"]]) == 0 {
                                        px_sv_status[row["pxname"]] = map[string]int{
                                                "_status_act":       0,
                                                "_status_act_up":    0,
                                                "_status_act_down":  0,
                                                "_status_act_maint": 0,
                                                "_status_act_drain": 0,
                                                "_status_act_other": 0,
                                                "_status_bck":       0,
                                                "_status_bck_up":    0,
                                                "_status_bck_down":  0,
                                                "_status_bck_maint": 0,
                                                "_status_bck_drain": 0,
                                                "_status_bck_other": 0,
                                        }
                                }

                                if row["act"] == "1" {
                                        px_sv_status[row["pxname"]]["_status_act"]++
                                        switch row["status"] {
                                        case "UP":
                                                px_sv_status[row["pxname"]]["_status_act_up"]++
                                        case "DOWN":
                                                px_sv_status[row["pxname"]]["_status_act_down"]++
                                        case "MAINT":
                                                px_sv_status[row["pxname"]]["_status_act_maint"]++
                                        case "DRAIN":
                                                px_sv_status[row["pxname"]]["_status_act_drain"]++
                                        default:
                                                px_sv_status[row["pxname"]]["_status_act_other"]++
                                        }

                                } else {
                                        px_sv_status[row["pxname"]]["_status_bck"]++
                                        switch row["status"] {
                                        case "UP":
                                                px_sv_status[row["pxname"]]["_status_bck_up"]++
                                        case "DOWN":
                                                px_sv_status[row["pxname"]]["_status_bck_down"]++
                                        case "MAINT":
                                                px_sv_status[row["pxname"]]["_status_bck_maint"]++
                                        case "DRAIN":
                                                px_sv_status[row["pxname"]]["_status_bck_drain"]++
                                        default:
                                                px_sv_status[row["pxname"]]["_status_bck_other"]++
                                        }
                                }
                        }

                        if row["svname"] == "BACKEND" && len(px_sv_status[row["pxname"]]) > 0 {
                                for s, c := range px_sv_status[row["pxname"]] {
                                        row[s] = strconv.Itoa(c)
                                }
                        }

                        fields := make(map[string]interface{})
                        for field, v := range row {
                                if non_stat_fields[field] {
                                        continue
                                }

                                ival, err := strconv.ParseUint(v, 10, 64)
                                if err == nil {
                                        fields[field] = ival
                                }
                        }
                        acc.AddFields("haproxy", fields, tags, now)
                }
        }
        return err
}

func init() {
        inputs.Add("haproxy", func() telegraf.Input {
                return &haproxy{}
        })
}
