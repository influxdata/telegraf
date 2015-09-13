package webservercodes

import (
    "net/http"
    "strconv"
    "sync"
    "time"
    "os"
    "io"
    "regexp"
    "errors"
    
    "github.com/rogpeppe/rog-go/reverse"
    "github.com/influxdb/telegraf/plugins"
)

type Vhost struct {
    Host string
    AccessLog string
    RegexParsestring string
    ParseInterval string
}

type Webservercodes struct {
    Vhosts []*Vhost
}

type HttpStats struct {
    codes map[int]int
}

type CombinedEntry struct {
    time time.Time
    code int
}

var sampleConfig = `
# List of virtualhosts for http codes collecting
# (each section for one virtualhost, none for disable collecting codes)
[[webservercodes.vhosts]]
# 'host' field should match hostname in appropriate status url
host = "defaulthost"

# Telegraf user must have read permissions to this log file
# (shell command 'sudo adduser telegraf adm' for apache on Ubuntu)
access_log = "/var/log/apache2/access.log"

# Regular expression for fetching codes from log file strings.
# You can adjust this pattern for your custom log format
# Example for apache "common" and "combined" log formats, nginx default log format ("combined"):
#   '\[(?P<time>[^\]]+)\] ".*?" (?P<code>\d{3})'
# This pattern matches for strings like (example):
# 127.0.0.1 - - [30/Aug/2015:05:59:36 +0000] "GET / HTTP/1.1" 404 379 "-" "-"
regex_parsestring = '\[(?P<time>[^\]]+)\] ".*?" (?P<code>\d{3})'

# Plugin will parse log for http codes from 'now' till 'now - parse_interval' moments.
# parse_interval must be in time.Duration format (see https://golang.org/pkg/time/#ParseDuration)
parse_interval = "10s"
`

func (n *Webservercodes) SampleConfig() string {
    return sampleConfig
}

func (n *Webservercodes) Description() string {
    return "Read webserver access log files and count http return codes found"
}

func (n *Webservercodes) Gather(acc plugins.Accumulator) error {
    var wg sync.WaitGroup
    
    hostStats := map[string]HttpStats{}
    errChan := make(chan error)
    successChan := make(chan bool)
    remainingItems := len(n.Vhosts)
    
    for _, vhost := range n.Vhosts {
        if duration, err := time.ParseDuration(vhost.ParseInterval); err == nil {
            wg.Add(1)
            go func(host string, logfile string, regex string, duration time.Duration, successChan chan bool, errChan chan error) {
                defer wg.Done()
                if stats, err := n.ParseHttpCodes(logfile, regex, duration); err == nil {
                    hostStats[host] = *stats
                    successChan <- true
                } else {
                    errChan <- err
                }
            }(vhost.Host, vhost.AccessLog, vhost.RegexParsestring, duration, successChan, errChan)
        } else {
            return err
        }
    }
    for {
        select {
            case _ = <-successChan:
                remainingItems--
            case err := <-errChan:
                return err
        }
        if remainingItems == 0 {
            break;
        }
    }
    
    wg.Wait()
    
    for vhost, stats := range hostStats {
        n.gatherCodes(vhost, stats, acc)
    }
    
    return nil
}

func SearchStringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func (n Webservercodes) CombineKeysValues(keys []string, values []string) (*CombinedEntry, error) {
    
    if len(values) < len(keys) {
        return nil, errors.New("Not enough substrings")
    }
    
    items := map[string]string{}
    for k, v := range keys {
        items[v] = values[k]
    }
    
    combined := CombinedEntry{}
    if logDt, ok := items["time"]; ok {
        if time, err := time.Parse("02/Jan/2006:15:04:05 -0700", logDt); err == nil {
            combined.time = time
        } else {
            return nil, errors.New("Time must be in apache %t format. Example: '02/Jan/2006:15:04:05 -0700'")
        }
    } else {
        return nil, errors.New("Time is absent in log line")
    }
    
    if _, ok := items["code"]; ok {
        code, _ := strconv.Atoi(items["code"])
        combined.code = code
    } else {
        return nil, errors.New("Http code is absent in log line")
    }
    
    return &combined, nil
}

func (n Webservercodes) ValidateRegexp(regex string, f io.ReadSeeker) (*regexp.Regexp, []string, error) {
    
    keys := []string{}
    var rx *regexp.Regexp
    var err error
    
    if rx, err = regexp.Compile(regex); err != nil {
        // error in case of malformed regexp
        return nil, keys, err
    }
    
    keys = rx.SubexpNames();
    if !(SearchStringInSlice("time", keys) && SearchStringInSlice("code", keys)) {
        // error if fields 'time' or 'code' are defined
        return nil, keys, errors.New("Regexp must define 'time' and 'code' fields")
    }
    
    // we will check regexp validity by scan the last log line
    // and parse it, assuming that other lines will match as well
    reader := reverse.NewScanner(f)
    if (!reader.Scan()) {
        // if not Scanned, file is empty, so we don't need to return regex error
        return rx, keys, nil
    }
    
    strings := rx.FindStringSubmatch(reader.Text())
    if len(strings) == 0 {
        // error if regexp mismatch
        return nil, keys, errors.New("Log entries are not match regexp")
    }
    if _, err := n.CombineKeysValues(keys, strings); err != nil {
        // error if no values for 'time' or 'code' are found in parsed log line
        return nil, keys, err
    }
    
    return rx, keys, nil
}

func (n *Webservercodes) ParseHttpCodes(file string, regex string, duration time.Duration) (*HttpStats, error) {
    
    stats := HttpStats{codes: make(map[int]int)}
    
    if f, err := os.Open(file); err == nil {
        defer f.Close()
        
        if rx, keys, err := n.ValidateRegexp(regex, f); err == nil {
            curTime := time.Now()
            errorsCounter := 0
            errorsMax := 100 // there is something wrong if more than errorsMax parse errors
            var vastedLoop bool
            var strings []string
            
            reader := reverse.NewScanner(f)
            for reader.Scan() {
                vastedLoop = false
                strings = rx.FindStringSubmatch(reader.Text())
                if len(strings) > 0 {
                    if parsedLine, err := n.CombineKeysValues(keys, strings); err == nil {
                        if curTime.Sub(parsedLine.time) > duration {
                            break
                        }
                        if _, ok := stats.codes[parsedLine.code]; ok {
                            stats.codes[parsedLine.code]++
                        } else {
                            stats.codes[parsedLine.code] = 1
                        }
                    } else {
                        vastedLoop = true
                    }
                } else {
                    vastedLoop = true
                }
                if vastedLoop {
                    errorsCounter++
                }
                if errorsCounter >= errorsMax {
                    break
                }
            }
            if errorsCounter >= errorsMax {
                return nil, errors.New("Too many entries with wrong format in log file. Check regex_parsestring")
            }
        } else {
            return nil, err
        }
    } else {
        return nil, err
    }
    
    return &stats, nil
}

func (n *Webservercodes) gatherCodes(vhost string, stats HttpStats, acc plugins.Accumulator) {
    
    tags := map[string]string{"virtualhost" : vhost}
    total := 0
    var num int
    
    for _, i := range []int{http.StatusContinue,
                    http.StatusSwitchingProtocols,
                    http.StatusOK,
                    http.StatusCreated,
                    http.StatusAccepted,
                    http.StatusNonAuthoritativeInfo,
                    http.StatusNoContent,
                    http.StatusResetContent,
                    http.StatusPartialContent,
                    http.StatusMultipleChoices,
                    http.StatusMovedPermanently,
                    http.StatusFound,
                    http.StatusSeeOther,
                    http.StatusNotModified,
                    http.StatusUseProxy,
                    http.StatusTemporaryRedirect,
                    http.StatusBadRequest,
                    http.StatusUnauthorized,
                    http.StatusPaymentRequired,
                    http.StatusForbidden,
                    http.StatusNotFound,
                    http.StatusMethodNotAllowed,
                    http.StatusNotAcceptable,
                    http.StatusProxyAuthRequired,
                    http.StatusRequestTimeout,
                    http.StatusConflict,
                    http.StatusGone,
                    http.StatusLengthRequired,
                    http.StatusPreconditionFailed,
                    http.StatusRequestEntityTooLarge,
                    http.StatusRequestURITooLong,
                    http.StatusUnsupportedMediaType,
                    http.StatusRequestedRangeNotSatisfiable,
                    http.StatusExpectationFailed,
                    http.StatusTeapot,
                    http.StatusInternalServerError,
                    http.StatusNotImplemented,
                    http.StatusBadGateway,
                    http.StatusServiceUnavailable,
                    http.StatusGatewayTimeout,
                    http.StatusHTTPVersionNotSupported} {
        if _, ok := stats.codes[i]; ok {
            num = stats.codes[i]
        } else {
            num = 0
        }
        total += num
        acc.Add(strconv.Itoa(i), num, tags);
    }
    acc.Add("total", total, tags);
}

func init() {
    plugins.Add("webservercodes", func() plugins.Plugin {
        return &Webservercodes{}
    })
}
