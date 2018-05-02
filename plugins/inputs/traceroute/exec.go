package traceroute

import (
	"bytes"
	"os/exec"
	"strconv"
)

type HostTracerouter func(timeout float64, args ...string) (string, error)

// Traceroute struct should be named the same as the Plugin
type Traceroute struct {

	// URLs to traceroute
	Urls []string

	// Total timeout duration each traceroute call, in seconds. 0 means no timeout
	// Type: float
	// Default: 0.0
	ResponseTimeout float64 `toml:"response_timeout"`

	// Wait time per probe in seconds (traceroute -w <WAITTIME>)
	// Type: float
	// Default: 5.0 sec
	WaitTime float64 `toml:"waittime"`

	// Starting TTL of packet (traceroute -f <FIRST_TTL>)
	// Type: int
	// Default: 1
	FirstTTL int `toml:"first_ttl"`

	// Maximum number of hops (hence TTL) traceroute will probe (traceroute -m <MAX_TTL>)
	// Type: int
	// Default: 30
	MaxTTL int `toml:"max_ttl"`

	// Number of probe packets sent per hop (traceroute -q <NQUERIES>)
	// Type: int
	// Default: 3
	Nqueries int `toml:"nqueries"`

	// Do not try to map IP addresses to host names (traceroute -n)
	// Default: false
	NoHostname bool `toml:"no_host_name"`

	// Use ICMP packets (traceroute -I)
	// Default: false
	UseICMP bool `toml:"icmp"`

	// Lookup AS path in routes (traceroute -A)
	// Default: false
	ASPathLookups bool `toml:"as_path_lookups"`

	// Source interface/address (traceroute -i <INTERFACE/SRC_ADDR>)
	// Type: string
	Interface string `toml:"interface"`

	// host traceroute function
	tracerouteMethod HostTracerouter
}

func executeWithoutTimeout(c *exec.Cmd) ([]byte, error) {
	var b bytes.Buffer
	c.Stderr = &b
	out, err := c.Output()
	if err != nil {
		out = b.Bytes()
	}
	return out, err
}

func (t *Traceroute) args(url string) []string {
	args := []string{url}
	//args = append(args, url)
	if t.WaitTime > 0.0 {
		args = append(args, "-w", strconv.FormatFloat(t.WaitTime, 'f', -1, 64))
	}
	if t.FirstTTL > 0 {
		args = append(args, "-f", strconv.Itoa(t.FirstTTL))
	}
	if t.MaxTTL > 0 && t.MaxTTL >= t.FirstTTL {
		args = append(args, "-m", strconv.Itoa(t.MaxTTL))
	}
	if t.Nqueries > 0 {
		args = append(args, "-q", strconv.Itoa(t.Nqueries))
	}
	if t.NoHostname {
		args = append(args, "-n")
	}
	if t.UseICMP {
		args = append(args, "-I")
	}
	if t.ASPathLookups {
		args = append(args, "-A")
	}
	if t.Interface != "" {
		args = append(args, "-i", t.Interface)
	}
	return args
}
