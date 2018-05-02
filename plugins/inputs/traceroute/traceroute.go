package traceroute

import (
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	tr_measurement  = "traceroute"
	hop_measurement = "traceroute_hop_data"
)

// Description will appear directly above the plugin definition in the config file
func (t *Traceroute) Description() string {
	return "Traceroutes given url(s) and return statistics"
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
const sampleConfig = `
# NOTE: this plugin forks the traceroute command. You may need to set capabilities
# via setcap cap_net_raw+p /bin/traceroute
  #
  ## List of urls to traceroute
  urls = ["www.google.com"] # required
  ## per-traceroute timeout, in s. 0 == no timeout
  ## it is highly recommended to set this value to match the telegraf interval
  # response_timeout = 0.0
  ## wait time per probe in seconds (traceroute -w <WAITTIME>)
  # waittime = 5.0
  ## starting TTL of packet (traceroute -f <FIRST_TTL>)
  # first_ttl = 1
  ## maximum number of hops (hence TTL) traceroute will probe (traceroute -m <MAX_TTL>)
  # max_ttl = 30
  ## number of probe packets sent per hop (traceroute -q <NQUERIES>)
  # nqueries = 3
  ## do not try to map IP addresses to host names (traceroute -n)
  # no_host_name = false
  ## use ICMP packets (traceroute -I)
  # icmp = false
  ## Lookup AS path in routes (traceroute -A)
  # as_path_lookups = false
  ## source interface/address to traceroute from (traceroute -i <INTERFACE/SRC_ADDR>)
  # interface = ""
`

func (t *Traceroute) SampleConfig() string {
	return sampleConfig
}

// Gather defines what data the plugin will gather.

func (t *Traceroute) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, host_url := range t.Urls {
		wg.Add(1)
		go func(target_fqdn string) {
			defer wg.Done()
			tags := map[string]string{"target_fqdn": target_fqdn}
			fields := make(map[string]interface{})

			_, err := net.LookupHost(target_fqdn)
			if err != nil {
				acc.AddError(err)
				fields["result_code"] = 1
				acc.AddFields(tr_measurement, fields, tags)
				return
			}

			tr_args := t.args(target_fqdn)
			output, err := t.tracerouteMethod(t.ResponseTimeout, tr_args...)

			//target_ip, number_of_hops, hop_info, err := parseTracerouteResults(output)
			results, err := parseTracerouteResults(output)
			tags["target_ip"] = results.Target_ip
			fields["result_code"] = 0
			fields["number_of_hops"] = results.Number_of_hops
			acc.AddFields(tr_measurement, fields, tags)

			for _, info := range results.Hop_info {
				hopTags := map[string]string{
					"target_fqdn":   results.Target_fqdn,
					"target_ip":     results.Target_ip,
					"column_number": strconv.Itoa(info.ColumnNum),
					"hop_fqdn":      info.Fqdn,
					"hop_ip":        info.Ip,
					"hop_number":    strconv.Itoa(info.HopNumber),
				}
				hopFields := map[string]interface{}{
					"hop_rtt_ms": info.RTT,
					"hop_asn":    info.Asn,
				}
				acc.AddFields(hop_measurement, hopFields, hopTags)
			}

		}(host_url)

	}
	return nil
}

func hostTracerouter(timeout float64, args ...string) (string, error) {
	var out []byte
	bin, err := exec.LookPath("traceroute")
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)
	if timeout == float64(0) {
		out, err = executeWithoutTimeout(c)
	} else {
		out, err = internal.CombinedOutputTimeout(c, time.Second*time.Duration(timeout+5))
	}
	return string(out), err
}

func init() {
	inputs.Add("traceroute", func() telegraf.Input {
		return &Traceroute{
			ResponseTimeout:  0,
			WaitTime:         5.0,
			FirstTTL:         1,
			MaxTTL:           30,
			Nqueries:         3,
			NoHostname:       false,
			UseICMP:          false,
			tracerouteMethod: hostTracerouter,
		}
	})
}
