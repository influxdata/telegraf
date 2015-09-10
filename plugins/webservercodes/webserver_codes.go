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

func (n Webservercodes) ParseRegex(regex string, f io.ReadSeeker) (*regexp.Regexp, error) {
    
    if rx, err := regexp.Compile(regex); err == nil {
        
        keys := rx.SubexpNames();
        if SearchStringInSlice("time", keys) && SearchStringInSlice("code", keys) {
            
            reader := reverse.NewScanner(f)
            if (reader.Scan()) {
                // we will check regexp validity by scan the last log line
                // and parse it, assuming that other lines will match as well
                parsedLine := rx.FindStringSubmatch(reader.Text())
                if len(parsedLine) >= 3 {
                    
                    logDt := parsedLine[1]
                    if _, err := time.Parse("02/Jan/2006:15:04:05 -0700", logDt); err == nil {
                        
                        return rx, nil
                    } else {
                        return nil, errors.New("Time must be in apache %t format. Example: '02/Jan/2006:15:04:05 -0700'")
                    }
                } else {
                    return nil, errors.New("Cannot find matches for regex in log line")
                }
            } else {
                // if not Scanned, file is empty, so we don't need to return regex error
                return rx, nil
            }
        } else {
            return nil, errors.New("Regexp must define 'time' and 'code' fields")
        }
    } else {
        return nil, err
    }
}

func (n *Webservercodes) ParseHttpCodes(file string, regex string, duration time.Duration) (*HttpStats, error) {
    
    stats := HttpStats{codes: make(map[int]int)}
    
    if f, err := os.Open(file); err == nil {
        
        defer f.Close()
        
        if rx, err := n.ParseRegex(regex, f); err == nil {
            var text, logDt string
            var parsedLine []string
            
            curTime := time.Now()
            reader := reverse.NewScanner(f)
            
            for reader.Scan() {
                text = reader.Text()
                
                parsedLine = rx.FindStringSubmatch(text)
                if len(parsedLine) > 0 {
                    logDt = parsedLine[1]
                    
                    if time, err := time.Parse("02/Jan/2006:15:04:05 -0700", logDt); err == nil {
                        if curTime.Sub(time) > duration {
                            break
                        }
                        if code, err := strconv.Atoi(parsedLine[2]); err == nil {
                            if _, ok := stats.codes[code]; ok {
                                stats.codes[code]++
                            } else {
                                stats.codes[code] = 1
                            }
                        }
                    }
                }
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
