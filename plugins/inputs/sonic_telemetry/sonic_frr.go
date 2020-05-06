package sonic_telemetry_gnmi

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/influxdata/telegraf"
)

const sock_addr = "/etc/sonic/frr/bgpd_client_sock"

type SonicFRR struct {
	Vrfs []Vrf `toml:"vrf"`
}

type Vrf struct {
	Name    string   `toml:"name"`
	AfiSafi []string `toml:"address_family"`
}

var sampleConfigSonicFrr = `
  ## Define the VRF and Address Families for the BGP Neighbors from FRR on SONiC OS
  #  [[inputs.sonic_frr.vrf]]
  #    name = "default"
  #    ## Address Family one of : "ipv4", "ipv6"
  #    address_family = ["ipv4", "ipv6"]
  #  [[inputs.sonic_frr.vrf]]
  #    name = "Vrf_blue"
  #    address_family = ["ipv4", "ipv6"]
`

func (sf *SonicFRR) Description() string {
	return "Collect BGP Neighbors in a given Address Family and VRF from FRR on SONiC OS"
}

func (sf *SonicFRR) SampleConfig() string {
	return sampleConfigSonicFrr
}

func (sf *SonicFRR) Gather(acc telegraf.Accumulator) error {
	vrfs, err := sf.listVrfs()
	if err != nil {
		return err
	}

	for _, vrf := range vrfs {
		vrf_name := vrf.Name
		for _, afisafi := range vrf.AfiSafi {
			// Issue vtysh command
			if afisafi != "ipv4" && afisafi != "ipv6" {
				acc.AddError(fmt.Errorf(" Unsupported Address Family vrf:'%s' afisafi:'%s'", vrf_name, afisafi))
				continue
			}
			vtysh_cmd := "show ip bgp vrf " + vrf_name + " " + afisafi + " neighbors json"
			bgp_nbr_json, err := exec_vtysh_cmd(vtysh_cmd, acc)
			if err != nil {
				acc.AddError(fmt.Errorf("error executing vtysh cmd '%s' err: '%s'", vtysh_cmd, err))
				continue
			}

			var nbr_cnt, nbr_up_cnt int32
			for nbr_key := range bgp_nbr_json {
				if nbr_data_json, ok := bgp_nbr_json[nbr_key].(map[string]interface{}); ok {
					err, nbr_state := sf.gatherNbrInfo(vrf_name, nbr_key, afisafi, nbr_data_json, acc)
					if err != nil {
						acc.AddError(fmt.Errorf("Unable to Gather NbrInfo for vrf:'%s' nbr:'%s' afisafi:'%s' err: '%s'",
							vrf_name, nbr_key, afisafi, err))
					}
					//fmt.Println("vtysh_cmd: ", vtysh_cmd)
					nbr_cnt++
					if nbr_state == true {
						nbr_up_cnt++
					}
				}
			}
			fields := map[string]interface{}{
				"totalNbrs":   nbr_cnt,
				"totalNbrsUp": nbr_up_cnt,
			}

			tags := map[string]string{
				"Vrf":        vrf_name,
				"AddrFamily": afisafi,
			}
			acc.AddFields("BGPNbrCount", fields, tags)
			//fmt.Println("tags: ", tags)
			//fmt.Println("fields: ", fields)
		}
	}
	return nil
}

func (sf *SonicFRR) gatherNbrInfo(vrf_name string, nbr_key string, afisafi string, nbr_data_json interface{}, acc telegraf.Accumulator) (error, bool) {
	var nbrState bool = false

	fields := make(map[string]interface{})

	nbr_data_val := nbr_data_json.(map[string]interface{})
	if value, ok := nbr_data_val["remoteAs"]; ok {
		fields["remoteAs"] = value
	}
	if value, ok := nbr_data_val["localAs"]; ok {
		fields["localAs"] = value
	}
	if value, ok := nbr_data_val["remoteRouterId"]; ok {
		fields["remoteRouterID"] = value
	}
	if value, ok := nbr_data_val["localRouterId"]; ok {
		fields["localRouterID"] = value
	}
	if value, ok := nbr_data_val["bgpState"]; ok {
		fields["state"] = value
		if value == "Established" {
			nbrState = true
		}
	}
	if value, ok := nbr_data_val["bgpTimerUpString"]; ok {
		fields["uptime"] = value
	}

	if bgp_msg_stats, ok := nbr_data_val["messageStats"].(map[string]interface{}); ok {
		if value, ok := bgp_msg_stats["totalSent"]; ok {
			total_msgs_sent := value
			fields["totalMsgsSent"] = total_msgs_sent
		}

		if value, ok := bgp_msg_stats["totalRecv"]; ok {
			total_msgs_recv := value
			fields["totalMsgsRecv"] = total_msgs_recv
		}
	}

	if addr_family_info, ok := nbr_data_val["addressFamilyInfo"].(map[string]interface{}); ok {
		if ipv4_info, ok := addr_family_info["ipv4Unicast"].(map[string]interface{}); ok {
			if value, ok := ipv4_info["acceptedPrefixCounter"]; ok {
				recv_pfxs := value
				fields["ipv4PrefixRecv"] = recv_pfxs
			}

			if value, ok := ipv4_info["sentPrefixCounter"]; ok {
				sent_pfxs := value
				fields["ipv4PrefixSent"] = sent_pfxs
			}
		}
		if ipv6_info, ok := addr_family_info["ipv6Unicast"].(map[string]interface{}); ok {
			if value, ok := ipv6_info["acceptedPrefixCounter"]; ok {
				recv_pfxs := value
				fields["ipv6PrefixRecv"] = recv_pfxs
			}

			if value, ok := ipv6_info["sentPrefixCounter"]; ok {
				sent_pfxs := value
				fields["ipv6PrefixSent"] = sent_pfxs
			}
		}
	}
	tags := map[string]string{
		"Vrf":         vrf_name,
		"AddrFamily":  afisafi,
		"BGPNeighbor": nbr_key,
	}
	acc.AddFields("BGPNeighbors", fields, tags)
	//fmt.Println("tags: ", tags)
	//fmt.Println("fields: ", fields)

	return nil, nbrState
}

func (sf *SonicFRR) listVrfs() ([]Vrf, error) {
	var vrfs []Vrf
	if len(sf.Vrfs) > 0 {
		vrfs = sf.Vrfs
	} else {
		return nil, nil
	}
	return vrfs, nil
}

func exec_vtysh_cmd(vtysh_cmd string, acc telegraf.Accumulator) (map[string]interface{}, error) {
	var err error
	oper_err := errors.New("Operational error")

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: sock_addr, Net: "unix"})
	if err != nil {
		acc.AddError(fmt.Errorf("Failed to connect proxy server: %s\n", err))
		return nil, err
	}
	defer conn.Close()
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(len(vtysh_cmd)))
	_, err = conn.Write(bs)
	if err != nil {
		acc.AddError(fmt.Errorf("Failed to write command length to server: %s\n", err))
		return nil, err
	}
	_, err = conn.Write([]byte(vtysh_cmd))
	if err != nil {
		acc.AddError(fmt.Errorf("Failed to write command length to server: %s\n", err))
		return nil, oper_err
	}

	var outputJson map[string]interface{}
	err = json.NewDecoder(conn).Decode(&outputJson)
	if err != nil {
		acc.AddError(fmt.Errorf("Not able to decode vtysh json output: %s\n", err))
		return nil, oper_err
	}

	if outputJson == nil {
		acc.AddError(fmt.Errorf("output empty\n"))
		return nil, oper_err
	}

	return outputJson, err
}
