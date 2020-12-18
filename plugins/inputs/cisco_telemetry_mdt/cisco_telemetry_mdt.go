package cisco_telemetry_mdt

import (
	"bytes"
	"encoding/binary"
        "encoding/json"
	"fmt"
	"io"
	"net"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	dialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf"
        internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials" // Register GRPC gzip decoder to support compressed telemetry
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/peer"
)

const (
	// Maximum telemetry payload size (in bytes) to accept for GRPC dialout transport
	tcpMaxMsgLen uint32 = 1024 * 1024
)

// CiscoTelemetryMDT plugin for IOS XR, IOS XE and NXOS platforms
type CiscoTelemetryMDT struct {
	// Common configuration
	Transport      string
	ServiceAddress string            `toml:"service_address"`
	MaxMsgSize     int               `toml:"max_msg_size"`
	Aliases        map[string]string `toml:"aliases"`
        Dmes           map[string]string `toml:"dmes"`
	EmbeddedTags   []string          `toml:"embedded_tags"`

	Log telegraf.Logger

	// GRPC TLS settings
	internaltls.ServerConfig

	// Internal listener / client handle
	grpcServer *grpc.Server
	listener   net.Listener

	// Internal state
	aliases   map[string]string
        dmes      map[string]string
	warned    map[string]struct{}
	extraTags map[string]map[string]struct{}
	nxpathMap map[string]map[string]string //per path map
	nxpropMap map[string]int //Global property map
	mutex     sync.Mutex
	acc       telegraf.Accumulator
	wg        sync.WaitGroup
}

type JsonStructure struct {
	Name string `json:"Name"`
	Prop []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"prop"`
}

//xform Field to string
func xformValueString(field *telemetry.TelemetryField) interface{} {
	var str string
        switch val := field.ValueByType.(type) {
        case *telemetry.TelemetryField_StringValue:
                if len(val.StringValue) > 0 {
                        return val.StringValue
                }
        case *telemetry.TelemetryField_Uint32Value:
		str = strconv.FormatUint(uint64(val.Uint32Value), 10)
		return str
        case *telemetry.TelemetryField_Uint64Value:
		str = strconv.FormatUint(val.Uint64Value, 10)
                return str
        case *telemetry.TelemetryField_Sint32Value:
		str = strconv.FormatInt(int64(val.Sint32Value), 10)
		return str
        case *telemetry.TelemetryField_Sint64Value:
		str = strconv.FormatInt(val.Sint64Value, 10)
		return str
        }
        return nil
}

//Xform value field.
func (c *CiscoTelemetryMDT) nxosValueXform(field *telemetry.TelemetryField, value interface{}, path string) interface{} {
     isNXOS := !strings.ContainsRune(path, ':')
     if (isNXOS) {
        if xfrom, ok := c.nxpropMap[field.Name]; ok {
           if xfrom == 1 {
               switch field.ValueByType.(type) {
               case *telemetry.TelemetryField_Uint64Value:
                    return int64(value.(uint64))
               default:
                    return nil
               }
           } else if xfrom == 2 { //if type is uint64 convert it to string.
                switch val := field.ValueByType.(type) {
                case *telemetry.TelemetryField_StringValue:
                        if len(val.StringValue) > 0 {
                           return val.StringValue
                        }
                case *telemetry.TelemetryField_Uint64Value:
                        str := strconv.FormatUint(val.Uint64Value, 10)
                        return str
                }
                return nil
           } else if xfrom == 3 {
                //convert property to float from string.
                switch val := field.ValueByType.(type) {
                case *telemetry.TelemetryField_StringValue:
                        if len(val.StringValue) > 0 {
                             if valf, err := strconv.ParseFloat(val.StringValue, 64); err == nil {
                                  return valf
                             }
                       }
                }
                return nil
           } else if xfrom == 4 {
                //string to uint64
                switch val := field.ValueByType.(type) {
                case *telemetry.TelemetryField_StringValue:
                        if len(val.StringValue) > 0 {
                             if val64, err := strconv.ParseUint(val.StringValue, 10, 64); err == nil {
                                  return val64
                             }
                       }
                }
                return nil
           } else if xfrom == 5 {
                //string to int64
                switch val := field.ValueByType.(type) {
                case *telemetry.TelemetryField_StringValue:
                        if len(val.StringValue) > 0 {
                             if val64, err := strconv.ParseInt(val.StringValue, 10, 64); err == nil {
                                  return val64
                             }
                       }
                }
                return nil
           } else {
                return nil
           }
        } else {
                //check if we want auto xformation
                switch val := field.ValueByType.(type) {
                case *telemetry.TelemetryField_StringValue:
                    if c.nxpropMap["auto-xfrom"] == 6 {
                         if val64, err := strconv.ParseUint(val.StringValue,10,64); err == nil {
                             return val64
                         }
                         if valf, err := strconv.ParseFloat(val.StringValue, 64); err == nil {
                             return valf
                         }
                         if val64, err := strconv.ParseInt(val.StringValue,10,64); err == nil {
                             return val64
                         }
                    } else if c.nxpropMap["auto-prop-xfrom"] == 7 {
                        if valf, err := strconv.ParseFloat(val.StringValue, 64); err == nil {
                             return valf
                        }
                    }
                } // switch
                //Now check path based conversion.
                //If mapping is found then do the required transformation.
                if c.nxpathMap[path] != nil {
                    switch c.nxpathMap[path][field.Name] {
                    //Xformation supported is only from String, Uint32 and Uint64			    
                    case "integer":
                         switch val := field.ValueByType.(type) {
                         case *telemetry.TelemetryField_StringValue:
                             if vali, err := strconv.ParseInt(val.StringValue,10,32); err == nil {
                                 return vali
                             }
                         case *telemetry.TelemetryField_Uint32Value:
                             return int(value.(uint32))
                         case *telemetry.TelemetryField_Uint64Value:
                             return int64(value.(uint64))
                         } //switch
                         return nil
                    //Xformation supported is only from String
                    case "float":
                         switch val := field.ValueByType.(type) {
                         case *telemetry.TelemetryField_StringValue:
                             if valf, err := strconv.ParseFloat(val.StringValue, 64); err == nil {
                                 return valf
                             }
                         } //switch
                         return nil
                    case "string":
			    return (xformValueString(field))
                    case "int64":
                         switch val := field.ValueByType.(type) {
                         case *telemetry.TelemetryField_StringValue:
                             if vali, err := strconv.ParseInt(val.StringValue,10,64); err == nil {
                                 return vali
                             }
                         case *telemetry.TelemetryField_Uint64Value:
                             return int64(value.(uint64))
                         } //switch
                    } //switch
                } //if
        } //else
     } //Nxos
     return nil
}

// Start the Cisco MDT service
func (c *CiscoTelemetryMDT) Start(acc telegraf.Accumulator) error {
	var err error
	c.acc = acc
	c.listener, err = net.Listen("tcp", c.ServiceAddress)
	if err != nil {
		return err
	}

        //Create Nexus property transform list.
        c.nxpropMap = make(map[string]int, 100 + len(c.Dmes))
        c.nxpropMap["asn"] = 2 //check if it's uint64 type then convert to string.
        c.nxpropMap["subscriptionId"] = 2 //check if it's uint64 type then convert to string.
	c.nxpropMap["operState"] = 2 //check if it's uint64 type then convert to string.

	// Invert aliases list
	c.warned = make(map[string]struct{})
	c.aliases = make(map[string]string, len(c.Aliases))
	for alias, path := range c.Aliases {
		c.aliases[path] = alias
	}
        //Load Init Json structure
jsonData := []byte(`[{"Name": "show environment power","prop": [{"Key": "reserve_sup","Value": "string"}, {"Key": "det_volt","Value": "string"}, {"Key": "heatsink_temp","Value": "string"}, {"Key": "det_pintot","Value": "string"}, {"Key": "det_iinb","Value": "string"}, {"Key": "ps_input_current","Value": "string"}, {"Key": "modnum","Value": "string"}, {"Key": "trayfannum","Value": "string"}, {"Key": "modstatus_3k","Value": "string"}, {"Key": "fan2rpm","Value": "string"}, {"Key": "amps_alloced","Value": "string"}, {"Key": "all_inlets_connected","Value": "string"}, {"Key": "tot_pow_out_actual_draw","Value": "string"}, {"Key": "ps_redun_op_mode","Value": "string"}, {"Key": "curtemp","Value": "string"}, {"Key": "mod_model","Value": "string"}, {"Key": "fanmodel","Value": "string"}, {"Key": "ps_output_current","Value": "string"}, {"Key": "majthres","Value": "string"}, {"Key": "input_type","Value": "string"}, {"Key": "allocated","Value": "string"}, {"Key": "fanhwver","Value": "string"}, {"Key": "clkhwver","Value": "string"}, {"Key": "fannum","Value": "string"}, {"Key": "watts_requested","Value": "string"}, {"Key": "cumulative_power","Value": "string"}, {"Key": "tot_gridB_capacity","Value": "string"}, {"Key": "pow_used_by_mods","Value": "string"}, {"Key": "tot_pow_alloc_budgeted","Value": "string"}, {"Key": "psumod","Value": "string"}, {"Key": "ps_status_3k","Value": "string"}, {"Key": "temptype","Value": "string"}, {"Key": "regval","Value": "string"}, {"Key": "inlet_temp","Value": "string"}, {"Key": "det_cord","Value": "string"}, {"Key": "reserve_fan","Value": "string"}, {"Key": "det_pina","Value": "string"}, {"Key": "minthres","Value": "string"}, {"Key": "actual_draw","Value": "string"}, {"Key": "sensor","Value": "string"}, {"Key": "zone","Value": "string"}, {"Key": "det_iin","Value": "string"}, {"Key": "det_iout","Value": "string"}, {"Key": "det_vin","Value": "string"}, {"Key": "fan1rpm","Value": "string"}, {"Key": "tot_gridA_capacity","Value": "string"}, {"Key": "fanperc","Value": "string"}, {"Key": "det_pout","Value": "string"}, {"Key": "alarm_str","Value": "string"}, {"Key": "zonespeed","Value": "string"}, {"Key": "det_total_cap","Value": "string"}, {"Key": "reserve_xbar","Value": "string"}, {"Key": "det_vout","Value": "string"}, {"Key": "watts_alloced","Value": "string"}, {"Key": "ps_in_power","Value": "string"}, {"Key": "tot_pow_input_actual_draw","Value": "string"}, {"Key": "ps_output_voltage","Value": "string"}, {"Key": "det_name","Value": "string"}, {"Key": "tempmod","Value": "string"}, {"Key": "clockname","Value": "string"}, {"Key": "fanname","Value": "string"}, {"Key": "regnumstr","Value": "string"}, {"Key": "bitnumstr","Value": "string"}, {"Key": "ps_slot","Value": "string"}, {"Key": "actual_out","Value": "string"}, {"Key": "ps_input_voltage","Value": "string"}, {"Key": "psmodel","Value": "string"}, {"Key": "speed","Value": "string"}, {"Key": "clkmodel","Value": "string"}, {"Key": "ps_redun_mode_3k","Value": "string"}, {"Key": "tot_pow_capacity","Value": "string"}, {"Key": "amps","Value": "string"}, {"Key": "available_pow","Value": "string"}, {"Key": "reserve_supxbarfan","Value": "string"}, {"Key": "watts","Value": "string"}, {"Key": "det_pinb","Value": "string"}, {"Key": "det_vinb","Value": "string"}, {"Key": "ps_state","Value": "string"}, {"Key": "det_sw_alarm","Value": "string"}, {"Key": "regnum","Value": "string"}, {"Key": "amps_requested","Value": "string"}, {"Key": "fanrpm","Value": "string"}, {"Key": "actual_input","Value": "string"}, {"Key": "outlet_temp","Value": "string"}, {"Key": "tot_capa","Value": "string"}]},
{"Name": "show processes memory physical","prop": [{"Key": "processname","Value": "string"}]},
{"Name": "show bgp ipv4 unicast","prop": [{"Key": "aspath","Value": "string"}]},
{"Name": "show processes cpu","prop": [{"Key": "kernel_percent","Value": "float"}, {"Key": "idle_percent","Value": "float"}, {"Key": "process","Value": "string"}, {"Key": "user_percent","Value": "float"}, {"Key": "onesec","Value": "float"}]},
{"Name": "show system resources","prop": [{"Key": "cpu_state_user","Value": "float"}, {"Key": "kernel","Value": "float"}, {"Key": "current_memory_status","Value": "string"}, {"Key": "load_avg_15min","Value": "float"}, {"Key": "idle","Value": "float"}, {"Key": "load_avg_1min","Value": "float"}, {"Key": "user","Value": "float"}, {"Key": "cpu_state_idle","Value": "float"}, {"Key": "load_avg_5min","Value": "float"}, {"Key": "cpu_state_kernel","Value": "float"}]},
{"Name": "show ptp corrections","prop": [{"Key": "sup-time","Value": "string"}, {"Key": "correction-val","Value": "int64"}, {"Key": "ptp-header","Value": "string"}, {"Key": "intf-name","Value": "string"}, {"Key": "ptp-end","Value": "string"}]},
{"Name": "show interface transceiver details","prop": [{"Key": "uncorrect_ber_alrm_hi","Value": "string"}, {"Key": "uncorrect_ber_cur_warn_lo","Value": "string"}, {"Key": "current_warn_lo","Value": "float"}, {"Key": "pre_fec_ber_max_alrm_hi","Value": "string"}, {"Key": "serialnum","Value": "string"}, {"Key": "pre_fec_ber_acc_warn_lo","Value": "string"}, {"Key": "pre_fec_ber_max_warn_lo","Value": "string"}, {"Key": "laser_temp_warn_hi","Value": "float"}, {"Key": "type","Value": "string"}, {"Key": "rx_pwr_0","Value": "float"}, {"Key": "rx_pwr_warn_hi","Value": "float"}, {"Key": "uncorrect_ber_warn_hi","Value": "string"}, {"Key": "qsfp_or_cfp","Value": "string"}, {"Key": "protocol_type","Value": "string"}, {"Key": "uncorrect_ber","Value": "string"}, {"Key": "uncorrect_ber_cur_alrm_hi","Value": "string"}, {"Key": "tec_current","Value": "float"}, {"Key": "pre_fec_ber","Value": "string"}, {"Key": "uncorrect_ber_max_warn_lo","Value": "string"}, {"Key": "uncorrect_ber_min","Value": "string"}, {"Key": "current_alrm_lo","Value": "float"}, {"Key": "uncorrect_ber_acc_warn_lo","Value": "string"}, {"Key": "snr_warn_lo","Value": "float"}, {"Key": "rev","Value": "string"}, {"Key": "laser_temp_alrm_lo","Value": "float"}, {"Key": "current","Value": "float"}, {"Key": "rx_pwr_1","Value": "float"}, {"Key": "tec_current_warn_hi","Value": "float"}, {"Key": "pre_fec_ber_cur_warn_lo","Value": "string"}, {"Key": "cisco_part_number","Value": "string"}, {"Key": "uncorrect_ber_acc_warn_hi","Value": "string"}, {"Key": "temp_warn_hi","Value": "float"}, {"Key": "laser_freq_warn_lo","Value": "float"}, {"Key": "uncorrect_ber_max_alrm_lo","Value": "string"}, {"Key": "snr_alrm_hi","Value": "float"}, {"Key": "pre_fec_ber_cur_alrm_lo","Value": "string"}, {"Key": "tx_pwr_alrm_hi","Value": "float"}, {"Key": "pre_fec_ber_min_warn_lo","Value": "string"}, {"Key": "pre_fec_ber_min_warn_hi","Value": "string"}, {"Key": "rx_pwr_alrm_hi","Value": "float"}, {"Key": "tec_current_warn_lo","Value": "float"}, {"Key": "uncorrect_ber_acc_alrm_hi","Value": "string"}, {"Key": "rx_pwr_4","Value": "float"}, {"Key": "uncorrect_ber_cur","Value": "string"}, {"Key": "pre_fec_ber_alrm_hi","Value": "string"}, {"Key": "rx_pwr_warn_lo","Value": "float"}, {"Key": "bit_encoding","Value": "string"}, {"Key": "pre_fec_ber_acc","Value": "string"}, {"Key": "sfp","Value": "string"}, {"Key": "pre_fec_ber_acc_alrm_hi","Value": "string"}, {"Key": "pre_fec_ber_min","Value": "string"}, {"Key": "current_warn_hi","Value": "float"}, {"Key": "pre_fec_ber_max_alrm_lo","Value": "string"}, {"Key": "uncorrect_ber_cur_warn_hi","Value": "string"}, {"Key": "current_alrm_hi","Value": "float"}, {"Key": "pre_fec_ber_acc_alrm_lo","Value": "string"}, {"Key": "snr_alrm_lo","Value": "float"}, {"Key": "uncorrect_ber_acc","Value": "string"}, {"Key": "tx_len","Value": "string"}, {"Key": "uncorrect_ber_alrm_lo","Value": "string"}, {"Key": "pre_fec_ber_alrm_lo","Value": "string"}, {"Key": "txcvr_type","Value": "string"}, {"Key": "tec_current_alrm_lo","Value": "float"}, {"Key": "volt_alrm_lo","Value": "float"}, {"Key": "temp_alrm_hi","Value": "float"}, {"Key": "uncorrect_ber_min_warn_lo","Value": "string"}, {"Key": "laser_freq","Value": "float"}, {"Key": "uncorrect_ber_min_warn_hi","Value": "string"}, {"Key": "uncorrect_ber_cur_alrm_lo","Value": "string"}, {"Key": "pre_fec_ber_max_warn_hi","Value": "string"}, {"Key": "name","Value": "string"}, {"Key": "fiber_type_byte0","Value": "string"}, {"Key": "laser_freq_alrm_lo","Value": "float"}, {"Key": "pre_fec_ber_cur_warn_hi","Value": "string"}, {"Key": "partnum","Value": "string"}, {"Key": "snr","Value": "float"}, {"Key": "volt_alrm_hi","Value": "float"}, {"Key": "connector_type","Value": "string"}, {"Key": "tx_medium","Value": "string"}, {"Key": "tx_pwr_warn_hi","Value": "float"}, {"Key": "cisco_vendor_id","Value": "string"}, {"Key": "cisco_ext_id","Value": "string"}, {"Key": "uncorrect_ber_max_warn_hi","Value": "string"}, {"Key": "pre_fec_ber_max","Value": "string"}, {"Key": "uncorrect_ber_min_alrm_hi","Value": "string"}, {"Key": "pre_fec_ber_warn_hi","Value": "string"}, {"Key": "tx_pwr_alrm_lo","Value": "float"}, {"Key": "uncorrect_ber_warn_lo","Value": "string"}, {"Key": "10gbe_code","Value": "string"}, {"Key": "cable_type","Value": "string"}, {"Key": "laser_freq_alrm_hi","Value": "float"}, {"Key": "rx_pwr_3","Value": "float"}, {"Key": "rx_pwr","Value": "float"}, {"Key": "volt_warn_hi","Value": "float"}, {"Key": "pre_fec_ber_cur_alrm_hi","Value": "string"}, {"Key": "temperature","Value": "float"}, {"Key": "voltage","Value": "float"}, {"Key": "tx_pwr","Value": "float"}, {"Key": "laser_temp_alrm_hi","Value": "float"}, {"Key": "tx_speeds","Value": "string"}, {"Key": "uncorrect_ber_min_alrm_lo","Value": "string"}, {"Key": "pre_fec_ber_min_alrm_hi","Value": "string"}, {"Key": "ciscoid","Value": "string"}, {"Key": "tx_pwr_warn_lo","Value": "float"}, {"Key": "cisco_product_id","Value": "string"}, {"Key": "info_not_available","Value": "string"}, {"Key": "laser_temp","Value": "float"}, {"Key": "pre_fec_ber_cur","Value": "string"}, {"Key": "fiber_type_byte1","Value": "string"}, {"Key": "tx_type","Value": "string"}, {"Key": "pre_fec_ber_min_alrm_lo","Value": "string"}, {"Key": "pre_fec_ber_warn_lo","Value": "string"}, {"Key": "temp_alrm_lo","Value": "float"}, {"Key": "volt_warn_lo","Value": "float"}, {"Key": "rx_pwr_alrm_lo","Value": "float"}, {"Key": "rx_pwr_2","Value": "float"}, {"Key": "tec_current_alrm_hi","Value": "float"}, {"Key": "uncorrect_ber_acc_alrm_lo","Value": "string"}, {"Key": "uncorrect_ber_max_alrm_hi","Value": "string"}, {"Key": "temp_warn_lo","Value": "float"}, {"Key": "snr_warn_hi","Value": "float"}, {"Key": "laser_temp_warn_lo","Value": "float"}, {"Key": "pre_fec_ber_acc_warn_hi","Value": "string"}, {"Key": "laser_freq_warn_hi","Value": "float"}, {"Key": "uncorrect_ber_max","Value": "string"}]},
{"Name": "show ip igmp groups vrf all","prop": [{"Key": "group-type","Value": "string"}, {"Key": "translate","Value": "string"}, {"Key": "sourceaddress","Value": "string"}, {"Key": "vrf-cntxt","Value": "string"}, {"Key": "expires","Value": "string"}, {"Key": "group-addr","Value": "string"}, {"Key": "uptime","Value": "string"}]},
{"Name": "show ip igmp interface vrf all","prop": [{"Key": "if-name","Value": "string"}, {"Key": "static-group-map","Value": "string"}, {"Key": "rll","Value": "string"}, {"Key": "host-proxy","Value": "string"}, {"Key": "il","Value": "string"}, {"Key": "join-group-map","Value": "string"}, {"Key": "expires","Value": "string"}, {"Key": "host-proxy-group-map","Value": "string"}, {"Key": "next-query","Value": "string"}, {"Key": "q-ver","Value": "string"}, {"Key": "if-status","Value": "string"}, {"Key": "un-solicited","Value": "string"}, {"Key": "ip-sum","Value": "string"}]},
{"Name": "show ip igmp snooping","prop": [{"Key": "repsup","Value": "string"}, {"Key": "omf_enabled","Value": "string"}, {"Key": "v3repsup","Value": "string"}, {"Key": "grepsup","Value": "string"}, {"Key": "lkupmode","Value": "string"}, {"Key": "description","Value": "string"}, {"Key": "vlinklocalgrpsup","Value": "string"}, {"Key": "gv3repsup","Value": "string"}, {"Key": "reportfloodall","Value": "string"}, {"Key": "leavegroupaddress","Value": "string"}, {"Key": "enabled","Value": "string"}, {"Key": "omf","Value": "string"}, {"Key": "sq","Value": "string"}, {"Key": "sqr","Value": "string"}, {"Key": "eht","Value": "string"}, {"Key": "fl","Value": "string"}, {"Key": "reportfloodenable","Value": "string"}, {"Key": "snoop-on","Value": "string"}, {"Key": "glinklocalgrpsup","Value": "string"}]},
{"Name": "show ip igmp snooping groups","prop": [{"Key": "src-uptime","Value": "string"}, {"Key": "source","Value": "string"}, {"Key": "dyn-if-name","Value": "string"}, {"Key": "raddr","Value": "string"}, {"Key": "old-host","Value": "string"}, {"Key": "snoop-enabled","Value": "string"}, {"Key": "expires","Value": "string"}, {"Key": "omf-enabled","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "src-expires","Value": "string"}, {"Key": "addr","Value": "string"}]},
{"Name": "show ip igmp snooping groups detail","prop": [{"Key": "src-uptime","Value": "string"}, {"Key": "source","Value": "string"}, {"Key": "dyn-if-name","Value": "string"}, {"Key": "raddr","Value": "string"}, {"Key": "old-host","Value": "string"}, {"Key": "snoop-enabled","Value": "string"}, {"Key": "expires","Value": "string"}, {"Key": "omf-enabled","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "src-expires","Value": "string"}, {"Key": "addr","Value": "string"}]},
{"Name": "show ip igmp snooping groups summary","prop": [{"Key": "src-uptime","Value": "string"}, {"Key": "source","Value": "string"}, {"Key": "dyn-if-name","Value": "string"}, {"Key": "raddr","Value": "string"}, {"Key": "old-host","Value": "string"}, {"Key": "snoop-enabled","Value": "string"}, {"Key": "expires","Value": "string"}, {"Key": "omf-enabled","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "src-expires","Value": "string"}, {"Key": "addr","Value": "string"}]},
{"Name": "show ip igmp snooping mrouter","prop": [{"Key": "uptime","Value": "string"}, {"Key": "expires","Value": "string"}]},
{"Name": "show ip igmp snooping statistics","prop": [{"Key": "ut","Value": "string"}]},
{"Name": "show ip pim interface vrf all","prop": [{"Key": "if-is-border","Value": "string"}, {"Key": "cached_if_status","Value": "string"}, {"Key": "genid","Value": "string"}, {"Key": "if-name","Value": "string"}, {"Key": "last-cleared","Value": "string"}, {"Key": "is-pim-vpc-svi","Value": "string"}, {"Key": "if-addr","Value": "string"}, {"Key": "is-pim-enabled","Value": "string"}, {"Key": "pim-dr-address","Value": "string"}, {"Key": "hello-timer","Value": "string"}, {"Key": "pim-bfd-enabled","Value": "string"}, {"Key": "vpc-peer-nbr","Value": "string"}, {"Key": "nbr-policy-name","Value": "string"}, {"Key": "is-auto-enabled","Value": "string"}, {"Key": "if-status","Value": "string"}, {"Key": "jp-out-policy-name","Value": "string"}, {"Key": "if-addr-summary","Value": "string"}, {"Key": "if-dr","Value": "string"}, {"Key": "jp-in-policy-name","Value": "string"}]},
{"Name": "show ip pim neighbor vrf all","prop": [{"Key": "longest-hello-intvl","Value": "string"}, {"Key": "if-name","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "expires","Value": "string"}, {"Key": "bfd-state","Value": "string"}]},
{"Name": "show ip pim route vrf all","prop": [{"Key": "rpf-nbr-1","Value": "string"}, {"Key": "rpf-nbr-addr","Value": "string"}, {"Key": "register","Value": "string"}, {"Key": "sgexpire","Value": "string"}, {"Key": "oif-bf-str","Value": "string"}, {"Key": "mcast-addrs","Value": "string"}, {"Key": "rp-addr","Value": "string"}, {"Key": "immediate-bf-str","Value": "string"}, {"Key": "sgr-prune-list-bf-str","Value": "string"}, {"Key": "context-name","Value": "string"}, {"Key": "intf-name","Value": "string"}, {"Key": "immediate-timeout-bf-str","Value": "string"}, {"Key": "rp-local","Value": "string"}, {"Key": "sgrexpire","Value": "string"}, {"Key": "timeout-bf-str","Value": "string"}, {"Key": "timeleft","Value": "string"}]},
{"Name": "show ip pim rp vrf all","prop": [{"Key": "is-bsr-forward-only","Value": "string"}, {"Key": "is-rpaddr-local","Value": "string"}, {"Key": "bsr-expires","Value": "string"}, {"Key": "autorp-expire-time","Value": "string"}, {"Key": "rp-announce-policy-name","Value": "string"}, {"Key": "rp-cand-policy-name","Value": "string"}, {"Key": "is-autorp-forward-only","Value": "string"}, {"Key": "rp-uptime","Value": "string"}, {"Key": "rp-owner-flags","Value": "string"}, {"Key": "df-bits-recovered","Value": "string"}, {"Key": "bs-timer","Value": "string"}, {"Key": "rp-discovery-policy-name","Value": "string"}, {"Key": "arp-rp-addr","Value": "string"}, {"Key": "auto-rp-addr","Value": "string"}, {"Key": "autorp-expires","Value": "string"}, {"Key": "is-autorp-enabled","Value": "string"}, {"Key": "is-bsr-local","Value": "string"}, {"Key": "is-autorp-listen-only","Value": "string"}, {"Key": "autorp-dis-timer","Value": "string"}, {"Key": "bsr-rp-expires","Value": "string"}, {"Key": "static-rp-group-map","Value": "string"}, {"Key": "rp-source","Value": "string"}, {"Key": "autorp-cand-address","Value": "string"}, {"Key": "autorp-up-time","Value": "string"}, {"Key": "is-bsr-enabled","Value": "string"}, {"Key": "bsr-uptime","Value": "string"}, {"Key": "is-bsr-listen-only","Value": "string"}, {"Key": "rpf-nbr-address","Value": "string"}, {"Key": "is-rp-local","Value": "string"}, {"Key": "is-autorp-local","Value": "string"}, {"Key": "bsr-policy-name","Value": "string"}, {"Key": "grange-grp","Value": "string"}, {"Key": "rp-addr","Value": "string"}, {"Key": "anycast-rp-addr","Value": "string"}]},
{"Name": "show ip pim statistics vrf all","prop": [{"Key": "vrf-name","Value": "string"}]},
{"Name": "show interface brief","prop": [{"Key": "speed","Value": "string"},{"Key": "vlan","Value": "string"}]},
{"Name": "show ip pim vrf all","prop": [{"Key": "table-id","Value": "string"}]},
{"Name": "show ip mroute summary vrf all","prop": [{"Key": "nat-mode","Value": "string"}, {"Key": "oif-name","Value": "string"}, {"Key": "nat-route-type","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "mofrr-nbr","Value": "string"}, {"Key": "extranet_addr","Value": "string"}, {"Key": "stale-route","Value": "string"}, {"Key": "pending","Value": "string"}, {"Key": "bidir","Value": "string"}, {"Key": "expry_timer","Value": "string"}, {"Key": "mofrr-iif","Value": "string"}, {"Key": "group_addrs","Value": "string"}, {"Key": "mpib-name","Value": "string"}, {"Key": "rpf","Value": "string"}, {"Key": "mcast-addrs","Value": "string"}, {"Key": "route-mdt-iod","Value": "string"}, {"Key": "sr-oif","Value": "string"}, {"Key": "stats-rate-buf","Value": "string"}, {"Key": "source_addr","Value": "string"}, {"Key": "route-iif","Value": "string"}, {"Key": "rpf-nbr","Value": "string"}, {"Key": "translated-route-src","Value": "string"}, {"Key": "group_addr","Value": "string"}, {"Key": "lisp-src-rloc","Value": "string"}, {"Key": "stats-pndg","Value": "string"}, {"Key": "rate_buf","Value": "string"}, {"Key": "extranet_vrf_name","Value": "string"}, {"Key": "fabric-interest","Value": "string"}, {"Key": "translated-route-grp","Value": "string"}, {"Key": "internal","Value": "string"}, {"Key": "oif-mpib-name","Value": "string"}, {"Key": "oif-uptime","Value": "string"}, {"Key": "omd-vpc-svi","Value": "string"}, {"Key": "source_addrs","Value": "string"}, {"Key": "stale-oif","Value": "string"}, {"Key": "core-interest","Value": "string"}, {"Key": "oif-list-bitfield","Value": "string"}]},
{"Name": "show ipv6 mroute summary vrf all","prop": [{"Key": "nat-mode","Value": "string"}, {"Key": "oif-name","Value": "string"}, {"Key": "nat-route-type","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "mofrr-nbr","Value": "string"}, {"Key": "extranet_addr","Value": "string"}, {"Key": "stale-route","Value": "string"}, {"Key": "pending","Value": "string"}, {"Key": "bidir","Value": "string"}, {"Key": "expry_timer","Value": "string"}, {"Key": "mofrr-iif","Value": "string"}, {"Key": "group_addrs","Value": "string"}, {"Key": "mpib-name","Value": "string"}, {"Key": "rpf","Value": "string"}, {"Key": "mcast-addrs","Value": "string"}, {"Key": "route-mdt-iod","Value": "string"}, {"Key": "sr-oif","Value": "string"}, {"Key": "stats-rate-buf","Value": "string"}, {"Key": "source_addr","Value": "string"}, {"Key": "route-iif","Value": "string"}, {"Key": "rpf-nbr","Value": "string"}, {"Key": "translated-route-src","Value": "string"}, {"Key": "group_addr","Value": "string"}, {"Key": "lisp-src-rloc","Value": "string"}, {"Key": "stats-pndg","Value": "string"}, {"Key": "rate_buf","Value": "string"}, {"Key": "extranet_vrf_name","Value": "string"}, {"Key": "fabric-interest","Value": "string"}, {"Key": "translated-route-grp","Value": "string"}, {"Key": "internal","Value": "string"}, {"Key": "oif-mpib-name","Value": "string"}, {"Key": "oif-uptime","Value": "string"}, {"Key": "omd-vpc-svi","Value": "string"}, {"Key": "source_addrs","Value": "string"}, {"Key": "stale-oif","Value": "string"}, {"Key": "core-interest","Value": "string"}, {"Key": "oif-list-bitfield","Value": "string"}]},
{"Name": "show ip mroute summary vrf all","prop": [{"Key": "nat-mode","Value": "string"}, {"Key": "oif-name","Value": "string"}, {"Key": "nat-route-type","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "mofrr-nbr","Value": "string"}, {"Key": "extranet_addr","Value": "string"}, {"Key": "stale-route","Value": "string"}, {"Key": "pending","Value": "string"}, {"Key": "bidir","Value": "string"}, {"Key": "expry_timer","Value": "string"}, {"Key": "mofrr-iif","Value": "string"}, {"Key": "group_addrs","Value": "string"}, {"Key": "mpib-name","Value": "string"}, {"Key": "rpf","Value": "string"}, {"Key": "mcast-addrs","Value": "string"}, {"Key": "route-mdt-iod","Value": "string"}, {"Key": "sr-oif","Value": "string"}, {"Key": "stats-rate-buf","Value": "string"}, {"Key": "source_addr","Value": "string"}, {"Key": "route-iif","Value": "string"}, {"Key": "rpf-nbr","Value": "string"}, {"Key": "translated-route-src","Value": "string"}, {"Key": "group_addr","Value": "string"}, {"Key": "lisp-src-rloc","Value": "string"}, {"Key": "stats-pndg","Value": "string"}, {"Key": "rate_buf","Value": "string"}, {"Key": "extranet_vrf_name","Value": "string"}, {"Key": "fabric-interest","Value": "string"}, {"Key": "translated-route-grp","Value": "string"}, {"Key": "internal","Value": "string"}, {"Key": "oif-mpib-name","Value": "string"}, {"Key": "oif-uptime","Value": "string"}, {"Key": "omd-vpc-svi","Value": "string"}, {"Key": "source_addrs","Value": "string"}, {"Key": "stale-oif","Value": "string"}, {"Key": "core-interest","Value": "string"}, {"Key": "oif-list-bitfield","Value": "string"}]},
{"Name": "show ipv6 mroute summary vrf all","prop": [{"Key": "nat-mode","Value": "string"}, {"Key": "oif-name","Value": "string"}, {"Key": "nat-route-type","Value": "string"}, {"Key": "uptime","Value": "string"}, {"Key": "mofrr-nbr","Value": "string"}, {"Key": "extranet_addr","Value": "string"}, {"Key": "stale-route","Value": "string"}, {"Key": "pending","Value": "string"}, {"Key": "bidir","Value": "string"}, {"Key": "expry_timer","Value": "string"}, {"Key": "mofrr-iif","Value": "string"}, {"Key": "group_addrs","Value": "string"}, {"Key": "mpib-name","Value": "string"}, {"Key": "rpf","Value": "string"}, {"Key": "mcast-addrs","Value": "string"}, {"Key": "route-mdt-iod","Value": "string"}, {"Key": "sr-oif","Value": "string"}, {"Key": "stats-rate-buf","Value": "string"}, {"Key": "source_addr","Value": "string"}, {"Key": "route-iif","Value": "string"}, {"Key": "rpf-nbr","Value": "string"}, {"Key": "translated-route-src","Value": "string"}, {"Key": "group_addr","Value": "string"}, {"Key": "lisp-src-rloc","Value": "string"}, {"Key": "stats-pndg","Value": "string"}, {"Key": "rate_buf","Value": "string"}, {"Key": "extranet_vrf_name","Value": "string"}, {"Key": "fabric-interest","Value": "string"}, {"Key": "translated-route-grp","Value": "string"}, {"Key": "internal","Value": "string"}, {"Key": "oif-mpib-name","Value": "string"}, {"Key": "oif-uptime","Value": "string"}, {"Key": "omd-vpc-svi","Value": "string"}, {"Key": "source_addrs","Value": "string"}, {"Key": "stale-oif","Value": "string"}, {"Key": "core-interest","Value": "string"}, {"Key": "oif-list-bitfield","Value": "string"}]},{"Name": "sys/vpc","prop": [{"Key": "type2CompatQualStr","Value": "string"}, {"Key": "compatQualStr","Value": "string"}, {"Key": "name","Value": "string"}, {"Key": "issuFromVer","Value": "string"}, {"Key": "issuToVer","Value": "string"}]},
{"Name": "sys/bgp","prop": [{"Key": "dynRtMap","Value": "string"}, {"Key": "nhRtMap","Value": "string"}, {"Key": "epePeerSet","Value": "string"}, {"Key": "asn","Value": "string"}, {"Key": "peerImp","Value": "string"}, {"Key": "wght","Value": "string"}, {"Key": "assocDom","Value": "string"}, {"Key": "tblMap","Value": "string"}, {"Key": "unSupprMap","Value": "string"}, {"Key": "sessionContImp","Value": "string"}, {"Key": "allocLblRtMap","Value": "string"}, {"Key": "defMetric","Value": "string"}, {"Key": "password","Value": "string"}, {"Key": "retainRttRtMap","Value": "string"}, {"Key": "clusterId","Value": "string"}, {"Key": "localAsn","Value": "string"}, {"Key": "name","Value": "string"}, {"Key": "defOrgRtMap","Value": "string"}]},
{"Name": "sys/ch","prop": [{"Key": "fanName","Value": "string"}, {"Key": "typeCordConnected","Value": "string"}, {"Key": "vendor","Value": "string"}, {"Key": "model","Value": "string"}, {"Key": "rev","Value": "string"}, {"Key": "vdrId","Value": "string"}, {"Key": "hardwareAlarm","Value": "string"}, {"Key": "unit","Value": "string"}, {"Key": "hwVer","Value": "string"}]},
{"Name": "sys/intf","prop": [{"Key": "descr","Value": "string"}, {"Key": "name","Value": "string"}, {"Key": "lastStCause","Value": "string"}, {"Key": "description","Value": "string"}, {"Key": "unit","Value": "string"}, {"Key": "operFECMode","Value": "string"}, {"Key": "operBitset","Value": "string"}, {"Key": "mdix","Value": "string"}]},
{"Name": "sys/procsys","prop": [{"Key": "name","Value": "string"}, {"Key": "id","Value": "string"}, {"Key": "upTs","Value": "string"}, {"Key": "interval","Value": "string"}, {"Key": "memstatus","Value": "string"}]},
{"Name": "sys/proc","prop": [{"Key": "processName","Value": "string"}, {"Key": "procArg","Value": "string"}]},
{"Name": "sys/bfd/inst","prop": [{"Key": "descr","Value": "string"}, {"Key": "vrfName","Value": "string"}, {"Key": "name","Value": "string"}, {"Key": "name","Value": "string"}]},
{"Name": "sys/lldp","prop": [{"Key": "sysDesc","Value": "string"}, {"Key": "portDesc","Value": "string"}, {"Key": "portIdV","Value": "string"}, {"Key": "chassisIdV","Value": "string"}, {"Key": "sysName","Value": "string"}, {"Key": "name","Value": "string"}, {"Key": "id","Value": "string"}]}]`)
        var jdata []JsonStructure

        c.dmes = make(map[string]string, len(c.Dmes))
        c.nxpathMap = make(map[string]map[string]string, len(c.Dmes)+len(jdata))
        if err = json.Unmarshal(jsonData, &jdata); err == nil {
            for _, jd := range jdata {
                c.nxpathMap[jd.Name] = make(map[string]string, len(jd.Prop))
                for _, prop := range jd.Prop {
                    c.nxpathMap[jd.Name][prop.Key] = prop.Value
                }
            }
        }
        c.dmes = make(map[string]string, len(c.Dmes))
        for dme, path := range c.Dmes {
                c.dmes[path] = dme
                if path == "uint64 to int" {
                    c.nxpropMap[dme] = 1
                } else if path == "uint64 to string" {
                    c.nxpropMap[dme] = 2
                } else if path == "string to float64" {
                    c.nxpropMap[dme] = 3
                } else if path == "string to uint64" {
                    c.nxpropMap[dme] = 4
                } else if path == "string to int64" {
                    c.nxpropMap[dme] = 5
                } else if path == "true" {
                    c.nxpropMap[dme] = 6
                } else if path == "auto-float-xfrom" {
                    c.nxpropMap[dme] = 7
                } else if dme[0:6] == "dnpath" { //path based property map
                    js := []byte(path)
                    var jsStruct JsonStructure

                    err := json.Unmarshal(js, &jsStruct)
                    if err == nil {
                        //Build 2 level Hash nxpathMap Key = jsStruct.Name, Value = map of jsStruct.Prop
                        //It will override the default of code if same path is provided in configuration.
                        c.nxpathMap[jsStruct.Name] = make(map[string]string, len(jsStruct.Prop))
                        for _, prop := range jsStruct.Prop {
                            c.nxpathMap[jsStruct.Name][prop.Key] = prop.Value
                        }
                    }
                }
        }

	// Fill extra tags
	c.extraTags = make(map[string]map[string]struct{})
	for _, tag := range c.EmbeddedTags {
		dir := strings.Replace(path.Dir(tag), "-", "_", -1)
		if _, hasKey := c.extraTags[dir]; !hasKey {
			c.extraTags[dir] = make(map[string]struct{})
		}
		c.extraTags[dir][path.Base(tag)] = struct{}{}
	}

	switch c.Transport {
	case "tcp":
		// TCP dialout server accept routine
		c.wg.Add(1)
		go func() {
			c.acceptTCPClients()
			c.wg.Done()
		}()

	case "grpc":
		var opts []grpc.ServerOption
		tlsConfig, err := c.ServerConfig.TLSConfig()
		if err != nil {
			c.listener.Close()
			return err
		} else if tlsConfig != nil {
			opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		}

		if c.MaxMsgSize > 0 {
			opts = append(opts, grpc.MaxRecvMsgSize(c.MaxMsgSize))
		}

		c.grpcServer = grpc.NewServer(opts...)
		dialout.RegisterGRPCMdtDialoutServer(c.grpcServer, c)

		c.wg.Add(1)
		go func() {
			c.grpcServer.Serve(c.listener)
			c.wg.Done()
		}()

	default:
		c.listener.Close()
		return fmt.Errorf("invalid Cisco MDT transport: %s", c.Transport)
	}

	return nil
}

// AcceptTCPDialoutClients defines the TCP dialout server main routine
func (c *CiscoTelemetryMDT) acceptTCPClients() {
	// Keep track of all active connections, so we can close them if necessary
	var mutex sync.Mutex
	clients := make(map[net.Conn]struct{})

	for {
		conn, err := c.listener.Accept()
		if neterr, ok := err.(*net.OpError); ok && (neterr.Timeout() || neterr.Temporary()) {
			continue
		} else if err != nil {
			break // Stop() will close the connection so Accept() will fail here
		}

		mutex.Lock()
		clients[conn] = struct{}{}
		mutex.Unlock()

		// Individual client connection routine
		c.wg.Add(1)
		go func() {
			c.Log.Debugf("Accepted Cisco MDT TCP dialout connection from %s", conn.RemoteAddr())
			if err := c.handleTCPClient(conn); err != nil {
				c.acc.AddError(err)
			}
			c.Log.Debugf("Closed Cisco MDT TCP dialout connection from %s", conn.RemoteAddr())

			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()

			conn.Close()
			c.wg.Done()
		}()
	}

	// Close all remaining client connections
	mutex.Lock()
	for client := range clients {
		if err := client.Close(); err != nil {
			c.Log.Errorf("Failed to close TCP dialout client: %v", err)
		}
	}
	mutex.Unlock()
}

// Handle a TCP telemetry client
func (c *CiscoTelemetryMDT) handleTCPClient(conn net.Conn) error {
	// TCP Dialout telemetry framing header
	var hdr struct {
		MsgType       uint16
		MsgEncap      uint16
		MsgHdrVersion uint16
		MsgFlags      uint16
		MsgLen        uint32
	}

	var payload bytes.Buffer

	for {
		// Read and validate dialout telemetry header
		if err := binary.Read(conn, binary.BigEndian, &hdr); err != nil {
			return err
		}

		maxMsgSize := tcpMaxMsgLen
		if c.MaxMsgSize > 0 {
			maxMsgSize = uint32(c.MaxMsgSize)
		}

		if hdr.MsgLen > maxMsgSize {
			return fmt.Errorf("dialout packet too long: %v", hdr.MsgLen)
		} else if hdr.MsgFlags != 0 {
			return fmt.Errorf("invalid dialout flags: %v", hdr.MsgFlags)
		}

		// Read and handle telemetry packet
		payload.Reset()
		if size, err := payload.ReadFrom(io.LimitReader(conn, int64(hdr.MsgLen))); size != int64(hdr.MsgLen) {
			if err != nil {
				return err
			}
			return fmt.Errorf("TCP dialout premature EOF")
		}

		c.handleTelemetry(payload.Bytes())
	}
}

// MdtDialout RPC server method for grpc-dialout transport
func (c *CiscoTelemetryMDT) MdtDialout(stream dialout.GRPCMdtDialout_MdtDialoutServer) error {
	peer, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		c.Log.Debugf("Accepted Cisco MDT GRPC dialout connection from %s", peer.Addr)
	}

	var chunkBuffer bytes.Buffer

	for {
		packet, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				c.acc.AddError(fmt.Errorf("GRPC dialout receive error: %v", err))
			}
			break
		}

		if len(packet.Data) == 0 && len(packet.Errors) != 0 {
			c.acc.AddError(fmt.Errorf("GRPC dialout error: %s", packet.Errors))
			break
		}

		// Reassemble chunked telemetry data received from NX-OS
		if packet.TotalSize == 0 {
			c.handleTelemetry(packet.Data)
		} else if int(packet.TotalSize) <= c.MaxMsgSize {
			chunkBuffer.Write(packet.Data)
			if chunkBuffer.Len() >= int(packet.TotalSize) {
				c.handleTelemetry(chunkBuffer.Bytes())
				chunkBuffer.Reset()
			}
		} else {
			c.acc.AddError(fmt.Errorf("dropped too large packet: %dB > %dB", packet.TotalSize, c.MaxMsgSize))
		}
	}

	if peerOK {
		c.Log.Debugf("Closed Cisco MDT GRPC dialout connection from %s", peer.Addr)
	}

	return nil
}

// Handle telemetry packet from any transport, decode and add as measurement
func (c *CiscoTelemetryMDT) handleTelemetry(data []byte) {
	msg := &telemetry.Telemetry{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		c.acc.AddError(fmt.Errorf("Cisco MDT failed to decode: %v", err))
		return
	}

	grouper := metric.NewSeriesGrouper()
	for _, gpbkv := range msg.DataGpbkv {
		// Produce metadata tags
		var tags map[string]string

		// Top-level field may have measurement timestamp, if not use message timestamp
		measured := gpbkv.Timestamp
		if measured == 0 {
			measured = msg.MsgTimestamp
		}

		timestamp := time.Unix(int64(measured/1000), int64(measured%1000)*1000000)

		// Find toplevel GPBKV fields "keys" and "content"
		var keys, content *telemetry.TelemetryField = nil, nil
		for _, field := range gpbkv.Fields {
			if field.Name == "keys" {
				keys = field
			} else if field.Name == "content" {
				content = field
			}
		}

		if keys == nil || content == nil {
			c.Log.Infof("Message from %s missing keys or content", msg.GetNodeIdStr())
			continue
		}

		// Parse keys
		tags = make(map[string]string, len(keys.Fields)+3)
		tags["source"] = msg.GetNodeIdStr()
                if msg.GetSubscriptionIdStr() != "" {
		    tags["subscription"] = msg.GetSubscriptionIdStr()
                }
		tags["path"] = msg.GetEncodingPath()

		for _, subfield := range keys.Fields {
			c.parseKeyField(tags, subfield, "")
		}

		// Parse values
		for _, subfield := range content.Fields {
			c.parseContentField(grouper, subfield, "", msg.EncodingPath, tags, timestamp)
		}
	}

	for _, metric := range grouper.Metrics() {
		c.acc.AddMetric(metric)
	}
        fmt.Println("Deepak: end", c.acc)
}

func decodeValue(field *telemetry.TelemetryField) interface{} {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_BytesValue:
		return val.BytesValue
	case *telemetry.TelemetryField_StringValue:
                if len(val.StringValue) > 0 {
                        return val.StringValue
                }
	case *telemetry.TelemetryField_BoolValue:
		return val.BoolValue
	case *telemetry.TelemetryField_Uint32Value:
		return val.Uint32Value
	case *telemetry.TelemetryField_Uint64Value:
		return val.Uint64Value
	case *telemetry.TelemetryField_Sint32Value:
		return val.Sint32Value
	case *telemetry.TelemetryField_Sint64Value:
		return val.Sint64Value
	case *telemetry.TelemetryField_DoubleValue:
		return val.DoubleValue
	case *telemetry.TelemetryField_FloatValue:
		return val.FloatValue
	}
	return nil
}

func decodeTag(field *telemetry.TelemetryField) string {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_BytesValue:
		return string(val.BytesValue)
	case *telemetry.TelemetryField_StringValue:
		return val.StringValue
	case *telemetry.TelemetryField_BoolValue:
		if val.BoolValue {
			return "true"
		}
		return "false"
	case *telemetry.TelemetryField_Uint32Value:
		return strconv.FormatUint(uint64(val.Uint32Value), 10)
	case *telemetry.TelemetryField_Uint64Value:
		return strconv.FormatUint(val.Uint64Value, 10)
	case *telemetry.TelemetryField_Sint32Value:
		return strconv.FormatInt(int64(val.Sint32Value), 10)
	case *telemetry.TelemetryField_Sint64Value:
		return strconv.FormatInt(val.Sint64Value, 10)
	case *telemetry.TelemetryField_DoubleValue:
		return strconv.FormatFloat(val.DoubleValue, 'f', -1, 64)
	case *telemetry.TelemetryField_FloatValue:
		return strconv.FormatFloat(float64(val.FloatValue), 'f', -1, 32)
	default:
		return ""
	}
}

// Recursively parse tag fields
func (c *CiscoTelemetryMDT) parseKeyField(tags map[string]string, field *telemetry.TelemetryField, prefix string) {
	localname := strings.Replace(field.Name, "-", "_", -1)
	name := localname
	if len(localname) == 0 {
		name = prefix
	} else if len(prefix) > 0 {
		name = prefix + "/" + localname
	}

	if tag := decodeTag(field); len(name) > 0 && len(tag) > 0 {
		if _, exists := tags[localname]; !exists { // Use short keys whenever possible
			tags[localname] = tag
		} else {
			tags[name] = tag
		}
	}

	for _, subfield := range field.Fields {
		c.parseKeyField(tags, subfield, name)
	}
}

func (c *CiscoTelemetryMDT) parseRib(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField, prefix string, path string, tags map[string]string, timestamp time.Time) {
        // RIB 
        measurement := path
        for _, subfield := range field.Fields {
            //For Every table fill the keys which are vrfName, address and masklen
            if subfield.Name == "vrfName" {
                tags["vrfName"] = decodeTag(subfield)
            }
            if subfield.Name == "address" {
                tags["address"] = decodeTag(subfield)
            }
            if subfield.Name == "maskLen" {
                tags["maskLen"] = decodeTag(subfield)
            }
            if value := decodeValue(subfield); value != nil {
                grouper.Add(measurement, tags, timestamp, subfield.Name, value)
            }
            if subfield.Name == "nextHop" {
                //For next hop table fill the keys in the tag - which is address and vrfname
                for _, subf := range subfield.Fields {
                    for _, ff := range subf.Fields {
                        if ff.Name == "address" {
                            tags["nextHop/address"] = decodeTag(ff)
                        }
                        if ff.Name == "vrfName" {
                            tags["nextHop/vrfName"] = decodeTag(ff)
                        }
                        if value := decodeValue(ff); value != nil {
                            name := "nextHop/" + ff.Name
                            grouper.Add(measurement, tags, timestamp, name, value)
                        }
                    }
                }
            }
        }
}
 

func (c *CiscoTelemetryMDT) parseClassAttributeField(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField, prefix string, path string, tags map[string]string, timestamp time.Time) {
        // DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
        rn := ""
        dn := false

        var nxAttributes *telemetry.TelemetryField
        nxAttributes = field
        isDme := strings.Contains(path, "sys/")
        if path == "rib" {
           //handle native data path rib
           c.parseRib(grouper, field, prefix, path, tags, timestamp)
           return
        }
        if field == nil {
            return
        }
        if !isDme {
            return
        }

        //TODO: Need to protect it from causing panic
        if field.Fields[0].Fields[0].Fields[0].Name != "attributes" {
           return
        }
        nxAttributes = field.Fields[0].Fields[0].Fields[0].Fields[0]
         
        for _, subfield := range nxAttributes.Fields {
                if subfield.Name == "rn" {
                        rn = decodeTag(subfield)
                } else if subfield.Name == "dn" {
                        dn = true
                        rn = decodeTag(subfield)
                }
        }

        if !dn { // Check for distinguished name being present
                c.acc.AddError(fmt.Errorf("NX-OS decoding failed: missing dn field"))
                return
        }
        tags["dn"] = rn

        for _, subfield := range nxAttributes.Fields {
                if subfield.Name != "dn" {
                        c.parseContentField(grouper, subfield, "", path, tags, timestamp)
                }
        }

        delete(tags, prefix)
}

func (c *CiscoTelemetryMDT) parseContentField(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField, prefix string,
	path string, tags map[string]string, timestamp time.Time) {
	name := strings.Replace(field.Name, "-", "_", -1)

        if (name == "modTs" || name == "createTs") && decodeValue(field) == "never" {
           return
        }
	if len(name) == 0 {
		name = prefix
	} else if len(prefix) > 0 {
		name = prefix + "/" + name
	}

	extraTags := c.extraTags[strings.Replace(path, "-", "_", -1)+"/"+name]

	if value := decodeValue(field); value != nil {
		// Do alias lookup, to shorten measurement names
		measurement := path
		if alias, ok := c.aliases[path]; ok {
			measurement = alias
		} else {
			c.mutex.Lock()
			if _, haveWarned := c.warned[path]; !haveWarned {
				c.Log.Debugf("No measurement alias for encoding path: %s", path)
				c.warned[path] = struct{}{}
			}
			c.mutex.Unlock()
		}

                if val := c.nxosValueXform(field, value, path); val != nil {
                    grouper.Add(measurement, tags, timestamp, name, val)
                } else {
                    grouper.Add(measurement, tags, timestamp, name, value)
                }
		return
	}

	if len(extraTags) > 0 {
		for _, subfield := range field.Fields {
			if _, isExtraTag := extraTags[subfield.Name]; isExtraTag {
				tags[name+"/"+strings.Replace(subfield.Name, "-", "_", -1)] = decodeTag(subfield)
			}
		}
	}

	var nxAttributes, nxChildren, nxRows *telemetry.TelemetryField
	isNXOS := !strings.ContainsRune(path, ':') // IOS-XR and IOS-XE have a colon in their encoding path, NX-OS does not
        var isEVENT bool
        if isNXOS {
           isEVENT = strings.Contains(path, "EVENT-LIST")
        }
	for _, subfield := range field.Fields {
		if isNXOS && subfield.Name == "attributes" && len(subfield.Fields) > 0 {
			nxAttributes = subfield.Fields[0]
		} else if isNXOS && subfield.Name == "children" && len(subfield.Fields) > 0 {
                        if !isEVENT {
			    nxChildren = subfield
                        } else {
                            sub := subfield.Fields
                            if sub[0].Fields[0].Name == "subscriptionId" && len(sub[0].Fields) >= 2  {
                                nxAttributes = sub[0].Fields[1].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
                            }
                        }
                        //if nxAttributes == NULL then class based query.
                        if (nxAttributes == nil) {
                            //call function walking over walking list.
                            for _, sub := range subfield.Fields {
                                c.parseClassAttributeField(grouper, sub, name, path, tags, timestamp) 
                            }
                        }
		} else if isNXOS && strings.HasPrefix(subfield.Name, "ROW_") {
			nxRows = subfield
		} else if _, isExtraTag := extraTags[subfield.Name]; !isExtraTag { // Regular telemetry decoding
			c.parseContentField(grouper, subfield, name, path, tags, timestamp)
		}
	}

	if nxAttributes == nil && nxRows == nil {
		return
	} else if nxRows != nil {
		// NXAPI structure: https://developer.cisco.com/docs/cisco-nexus-9000-series-nx-api-cli-reference-release-9-2x/
		for _, row := range nxRows.Fields {
			for i, subfield := range row.Fields {
				if i == 0 { // First subfield contains the index, promote it from value to tag
					tags[prefix] = decodeTag(subfield)
                                        //We can have subfield so recursively handle it.
                                        if (len(row.Fields) == 1) {
                                            tags["row_number"] = strconv.FormatInt(int64(i), 10)
                                            c.parseContentField(grouper, subfield, "", path, tags, timestamp)
                                        }
				} else {
					c.parseContentField(grouper, subfield, "", path, tags, timestamp)
				}
                                // Nxapi we can't identify keys always from prefix
                                tags["row_number"] = strconv.FormatInt(int64(i), 10)
			}
			delete(tags, prefix)
		}
		return
	}

	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	rn := ""
	dn := false

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "rn" {
			rn = decodeTag(subfield)
		} else if subfield.Name == "dn" {
			dn = true
		}
	}

	if len(rn) > 0 {
		tags[prefix] = rn
	} else if !dn { // Check for distinguished name being present
		c.acc.AddError(fmt.Errorf("NX-OS decoding failed: missing dn field"))
		return
	}

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name != "rn" {
			c.parseContentField(grouper, subfield, "", path, tags, timestamp)
		}
	}

	if nxChildren != nil {
		// This is a nested structure, children will inherit relative name keys of parent
		for _, subfield := range nxChildren.Fields {
			c.parseContentField(grouper, subfield, prefix, path, tags, timestamp)
		}
	}
	delete(tags, prefix)
}

func (c *CiscoTelemetryMDT) Address() net.Addr {
	return c.listener.Addr()
}

// Stop listener and cleanup
func (c *CiscoTelemetryMDT) Stop() {
	if c.grpcServer != nil {
		// Stop server and terminate all running dialout routines
		c.grpcServer.Stop()
	}
	if c.listener != nil {
		c.listener.Close()
	}
	c.wg.Wait()
}

const sampleConfig = `
 ## Telemetry transport can be "tcp" or "grpc".  TLS is only supported when
 ## using the grpc transport.
 transport = "grpc"
 ## Address and port to host telemetry listener
 service_address = ":57000"
 ## Enable TLS; grpc transport only.
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"
 ## Enable TLS client authentication and define allowed CA certificates; grpc
 ##  transport only.
 # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]
 ## Define (for certain nested telemetry measurements with embedded tags) which fields are tags
 # embedded_tags = ["Cisco-IOS-XR-qos-ma-oper:qos/interface-table/interface/input/service-policy-names/service-policy-instance/statistics/class-stats/class-name"]
 ## Define aliases to map telemetry encoding paths to simple measurement names
 [inputs.cisco_telemetry_mdt.aliases]
   ifstats = "ietf-interfaces:interfaces-state/interface/statistics"
 [inputs.cisco_telemetry_mdt.dmes]
   ModTs = "ignore"
   CreateTs = "ignore"
`

// SampleConfig of plugin
func (c *CiscoTelemetryMDT) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (c *CiscoTelemetryMDT) Description() string {
	return "Cisco model-driven telemetry (MDT) input plugin for IOS XR, IOS XE and NX-OS platforms"
}

// Gather plugin measurements (unused)
func (c *CiscoTelemetryMDT) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("cisco_telemetry_mdt", func() telegraf.Input {
		return &CiscoTelemetryMDT{
			Transport:      "grpc",
			ServiceAddress: "127.0.0.1:57000",
		}
	})
}
