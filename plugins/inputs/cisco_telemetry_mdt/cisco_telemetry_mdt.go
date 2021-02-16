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
	"github.com/influxdata/telegraf/metric"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
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
	propMap   map[string]func(field *telemetry.TelemetryField, value interface{}) interface{}
	mutex     sync.Mutex
	acc       telegraf.Accumulator
	wg        sync.WaitGroup
}

type NxPayloadXfromStructure struct {
	Name string `json:"Name"`
	Prop []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"prop"`
}

//xform Field to string
func xformValueString(field *telemetry.TelemetryField) string {
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
	return ""
}

//xform Uint64 to int64
func nxosValueXformUint64Toint64(field *telemetry.TelemetryField, value interface{}) interface{} {
	if field.GetUint64Value() != 0 {
		return int64(value.(uint64))
	}
	return nil
}

//xform string to float
func nxosValueXformStringTofloat(field *telemetry.TelemetryField, value interface{}) interface{} {
	//convert property to float from string.
	vals := field.GetStringValue()
	if vals != "" {
		if valf, err := strconv.ParseFloat(vals, 64); err == nil {
			return valf
		}
	}
	return nil
}

//xform string to uint64
func nxosValueXformStringToUint64(field *telemetry.TelemetryField, value interface{}) interface{} {
	//string to uint64
	vals := field.GetStringValue()
	if vals != "" {
		if val64, err := strconv.ParseUint(vals, 10, 64); err == nil {
			return val64
		}
	}
	return nil
}

//xform string to int64
func nxosValueXformStringToInt64(field *telemetry.TelemetryField, value interface{}) interface{} {
	//string to int64
	vals := field.GetStringValue()
	if vals != "" {
		if val64, err := strconv.ParseInt(vals, 10, 64); err == nil {
			return val64
		}
	}
	return nil
}

//auto-xform
func nxosValueAutoXform(field *telemetry.TelemetryField, value interface{}) interface{} {
	//check if we want auto xformation
	vals := field.GetStringValue()
	if vals != "" {
		if val64, err := strconv.ParseUint(vals, 10, 64); err == nil {
			return val64
		}
		if valf, err := strconv.ParseFloat(vals, 64); err == nil {
			return valf
		}
		if val64, err := strconv.ParseInt(vals, 10, 64); err == nil {
			return val64
		}
	} // switch
	return nil
}

//auto-xform float properties
func nxosValueAutoXformFloatProp(field *telemetry.TelemetryField, value interface{}) interface{} {
	//check if we want auto xformation
	vals := field.GetStringValue()
	if vals != "" {
		if valf, err := strconv.ParseFloat(vals, 64); err == nil {
			return valf
		}
	} // switch
	return nil
}

//xform uint64 to string
func nxosValueXformUint64ToString(field *telemetry.TelemetryField, value interface{}) interface{} {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_StringValue:
		if len(val.StringValue) > 0 {
			return val.StringValue
		}
	case *telemetry.TelemetryField_Uint64Value:
		return (strconv.FormatUint(val.Uint64Value, 10))
	}
	return nil
}

//Xform value field.
func (c *CiscoTelemetryMDT) nxosValueXform(field *telemetry.TelemetryField, value interface{}, path string) interface{} {
	if strings.ContainsRune(path, ':') {
		// not NXOS
		return nil
	}
	if _, ok := c.propMap[field.Name]; ok {
		return (c.propMap[field.Name](field, value))
	} else {
		//check if we want auto xformation
		if _, ok := c.propMap["auto-prop-xfromi"]; ok {
                        return c.propMap["auto-prop-xfrom"](field, value)
		}
		//Now check path based conversion.
		//If mapping is found then do the required transformation.
                if c.nxpathMap[path] == nil {
		        return nil
	        }
		switch c.nxpathMap[path][field.Name] {
		//Xformation supported is only from String, Uint32 and Uint64
		case "integer":
			switch val := field.ValueByType.(type) {
			case *telemetry.TelemetryField_StringValue:
				if vali, err := strconv.ParseInt(val.StringValue, 10, 32); err == nil {
					return vali
				}
			case *telemetry.TelemetryField_Uint32Value:
                                vali, ok := value.(uint32)
                                if ok == true {
                                    return vali
                                }
			case *telemetry.TelemetryField_Uint64Value:
                                vali, ok := value.(uint64)
                                if ok == true {
                                    return vali
                                }
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
			return xformValueString(field)
		case "int64":
			switch val := field.ValueByType.(type) {
			case *telemetry.TelemetryField_StringValue:
				if vali, err := strconv.ParseInt(val.StringValue, 10, 64); err == nil {
					return vali
				}
			case *telemetry.TelemetryField_Uint64Value:
				return int64(value.(uint64))
			} //switch
		} //switch
	} // else
	return nil
}

func (c *CiscoTelemetryMDT) initPower() {
	key := "show environment power"
	c.nxpathMap[key] = make(map[string]string, 100)
	c.nxpathMap[key]["reserve_sup"] = "string"
	c.nxpathMap[key]["det_volt"] = "string"
	c.nxpathMap[key]["heatsink_temp"] = "string"
	c.nxpathMap[key]["det_pintot"] = "string"
	c.nxpathMap[key]["det_iinb"] = "string"
	c.nxpathMap[key]["ps_input_current"] = "string"
	c.nxpathMap[key]["modnum"] = "string"
	c.nxpathMap[key]["trayfannum"] = "string"
	c.nxpathMap[key]["modstatus_3k"] = "string"
	c.nxpathMap[key]["fan2rpm"] = "string"
	c.nxpathMap[key]["amps_alloced"] = "string"
	c.nxpathMap[key]["all_inlets_connected"] = "string"
	c.nxpathMap[key]["tot_pow_out_actual_draw"] = "string"
	c.nxpathMap[key]["ps_redun_op_mode"] = "string"
	c.nxpathMap[key]["curtemp"] = "string"
	c.nxpathMap[key]["mod_model"] = "string"
	c.nxpathMap[key]["fanmodel"] = "string"
	c.nxpathMap[key]["ps_output_current"] = "string"
	c.nxpathMap[key]["majthres"] = "string"
	c.nxpathMap[key]["input_type"] = "string"
	c.nxpathMap[key]["allocated"] = "string"
	c.nxpathMap[key]["fanhwver"] = "string"
	c.nxpathMap[key]["clkhwver"] = "string"
	c.nxpathMap[key]["fannum"] = "string"
	c.nxpathMap[key]["watts_requested"] = "string"
	c.nxpathMap[key]["cumulative_power"] = "string"
	c.nxpathMap[key]["tot_gridB_capacity"] = "string"
	c.nxpathMap[key]["pow_used_by_mods"] = "string"
	c.nxpathMap[key]["tot_pow_alloc_budgeted"] = "string"
	c.nxpathMap[key]["psumod"] = "string"
	c.nxpathMap[key]["ps_status_3k"] = "string"
	c.nxpathMap[key]["temptype"] = "string"
	c.nxpathMap[key]["regval"] = "string"
	c.nxpathMap[key]["inlet_temp"] = "string"
	c.nxpathMap[key]["det_cord"] = "string"
	c.nxpathMap[key]["reserve_fan"] = "string"
	c.nxpathMap[key]["det_pina"] = "string"
	c.nxpathMap[key]["minthres"] = "string"
	c.nxpathMap[key]["actual_draw"] = "string"
	c.nxpathMap[key]["sensor"] = "string"
	c.nxpathMap[key]["zone"] = "string"
	c.nxpathMap[key]["det_iin"] = "string"
	c.nxpathMap[key]["det_iout"] = "string"
	c.nxpathMap[key]["det_vin"] = "string"
	c.nxpathMap[key]["fan1rpm"] = "string"
	c.nxpathMap[key]["tot_gridA_capacity"] = "string"
	c.nxpathMap[key]["fanperc"] = "string"
	c.nxpathMap[key]["det_pout"] = "string"
	c.nxpathMap[key]["alarm_str"] = "string"
	c.nxpathMap[key]["zonespeed"] = "string"
	c.nxpathMap[key]["det_total_cap"] = "string"
	c.nxpathMap[key]["reserve_xbar"] = "string"
	c.nxpathMap[key]["det_vout"] = "string"
	c.nxpathMap[key]["watts_alloced"] = "string"
	c.nxpathMap[key]["ps_in_power"] = "string"
	c.nxpathMap[key]["tot_pow_input_actual_draw"] = "string"
	c.nxpathMap[key]["ps_output_voltage"] = "string"
	c.nxpathMap[key]["det_name"] = "string"
	c.nxpathMap[key]["tempmod"] = "string"
	c.nxpathMap[key]["clockname"] = "string"
	c.nxpathMap[key]["fanname"] = "string"
	c.nxpathMap[key]["regnumstr"] = "string"
	c.nxpathMap[key]["bitnumstr"] = "string"
	c.nxpathMap[key]["ps_slot"] = "string"
	c.nxpathMap[key]["actual_out"] = "string"
	c.nxpathMap[key]["ps_input_voltage"] = "string"
	c.nxpathMap[key]["psmodel"] = "string"
	c.nxpathMap[key]["speed"] = "string"
	c.nxpathMap[key]["clkmodel"] = "string"
	c.nxpathMap[key]["ps_redun_mode_3k"] = "string"
	c.nxpathMap[key]["tot_pow_capacity"] = "string"
	c.nxpathMap[key]["amps"] = "string"
	c.nxpathMap[key]["available_pow"] = "string"
	c.nxpathMap[key]["reserve_supxbarfan"] = "string"
	c.nxpathMap[key]["watts"] = "string"
	c.nxpathMap[key]["det_pinb"] = "string"
	c.nxpathMap[key]["det_vinb"] = "string"
	c.nxpathMap[key]["ps_state"] = "string"
	c.nxpathMap[key]["det_sw_alarm"] = "string"
	c.nxpathMap[key]["regnum"] = "string"
	c.nxpathMap[key]["amps_requested"] = "string"
	c.nxpathMap[key]["fanrpm"] = "string"
	c.nxpathMap[key]["actual_input"] = "string"
	c.nxpathMap[key]["outlet_temp"] = "string"
	c.nxpathMap[key]["tot_capa"] = "string"
}

func (c *CiscoTelemetryMDT) initMemPhys() {
        c.nxpathMap["show processes memory physical"] = map[string]string{"processname": "string"}
}

func (c *CiscoTelemetryMDT) initBgpV4() {
	key := "show bgp ipv4 unicast"
	c.nxpathMap[key] = make(map[string]string, 1)
	c.nxpathMap[key]["aspath"] = "string"
}

func (c *CiscoTelemetryMDT) initCpu() {
	key := "show processes cpu"
	c.nxpathMap[key] = make(map[string]string, 5)
	c.nxpathMap[key]["kernel_percent"] = "float"
	c.nxpathMap[key]["idle_percent"] = "float"
	c.nxpathMap[key]["process"] = "string"
	c.nxpathMap[key]["user_percent"] = "float"
	c.nxpathMap[key]["onesec"] = "float"
}

func (c *CiscoTelemetryMDT) initResources() {
	key := "show system resources"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["cpu_state_user"] = "float"
	c.nxpathMap[key]["kernel"] = "float"
	c.nxpathMap[key]["current_memory_status"] = "string"
	c.nxpathMap[key]["load_avg_15min"] = "float"
	c.nxpathMap[key]["idle"] = "float"
	c.nxpathMap[key]["load_avg_1min"] = "float"
	c.nxpathMap[key]["user"] = "float"
	c.nxpathMap[key]["cpu_state_idle"] = "float"
	c.nxpathMap[key]["load_avg_5min"] = "float"
	c.nxpathMap[key]["cpu_state_kernel"] = "float"
}

func (c *CiscoTelemetryMDT) initPtpCorrection() {
	key := "show ptp corrections"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["sup-time"] = "string"
	c.nxpathMap[key]["correction-val"] = "int64"
	c.nxpathMap[key]["ptp-header"] = "string"
	c.nxpathMap[key]["intf-name"] = "string"
	c.nxpathMap[key]["ptp-end"] = "string"
}

func (c *CiscoTelemetryMDT) initTrans() {
	key := "show interface transceiver details"
	c.nxpathMap[key] = make(map[string]string, 100)
	c.nxpathMap[key]["uncorrect_ber_alrm_hi"] = "string"
	c.nxpathMap[key]["uncorrect_ber_cur_warn_lo"] = "string"
	c.nxpathMap[key]["current_warn_lo"] = "float"
	c.nxpathMap[key]["pre_fec_ber_max_alrm_hi"] = "string"
	c.nxpathMap[key]["serialnum"] = "string"
	c.nxpathMap[key]["pre_fec_ber_acc_warn_lo"] = "string"
	c.nxpathMap[key]["pre_fec_ber_max_warn_lo"] = "string"
	c.nxpathMap[key]["laser_temp_warn_hi"] = "float"
	c.nxpathMap[key]["type"] = "string"
	c.nxpathMap[key]["rx_pwr_0"] = "float"
	c.nxpathMap[key]["rx_pwr_warn_hi"] = "float"
	c.nxpathMap[key]["uncorrect_ber_warn_hi"] = "string"
	c.nxpathMap[key]["qsfp_or_cfp"] = "string"
	c.nxpathMap[key]["protocol_type"] = "string"
	c.nxpathMap[key]["uncorrect_ber"] = "string"
	c.nxpathMap[key]["uncorrect_ber_cur_alrm_hi"] = "string"
	c.nxpathMap[key]["tec_current"] = "float"
	c.nxpathMap[key]["pre_fec_ber"] = "string"
	c.nxpathMap[key]["uncorrect_ber_max_warn_lo"] = "string"
	c.nxpathMap[key]["uncorrect_ber_min"] = "string"
	c.nxpathMap[key]["current_alrm_lo"] = "float"
	c.nxpathMap[key]["uncorrect_ber_acc_warn_lo"] = "string"
	c.nxpathMap[key]["snr_warn_lo"] = "float"
	c.nxpathMap[key]["rev"] = "string"
	c.nxpathMap[key]["laser_temp_alrm_lo"] = "float"
	c.nxpathMap[key]["current"] = "float"
	c.nxpathMap[key]["rx_pwr_1"] = "float"
	c.nxpathMap[key]["tec_current_warn_hi"] = "float"
	c.nxpathMap[key]["pre_fec_ber_cur_warn_lo"] = "string"
	c.nxpathMap[key]["cisco_part_number"] = "string"
	c.nxpathMap[key]["uncorrect_ber_acc_warn_hi"] = "string"
	c.nxpathMap[key]["temp_warn_hi"] = "float"
	c.nxpathMap[key]["laser_freq_warn_lo"] = "float"
	c.nxpathMap[key]["uncorrect_ber_max_alrm_lo"] = "string"
	c.nxpathMap[key]["snr_alrm_hi"] = "float"
	c.nxpathMap[key]["pre_fec_ber_cur_alrm_lo"] = "string"
	c.nxpathMap[key]["tx_pwr_alrm_hi"] = "float"
	c.nxpathMap[key]["pre_fec_ber_min_warn_lo"] = "string"
	c.nxpathMap[key]["pre_fec_ber_min_warn_hi"] = "string"
	c.nxpathMap[key]["rx_pwr_alrm_hi"] = "float"
	c.nxpathMap[key]["tec_current_warn_lo"] = "float"
	c.nxpathMap[key]["uncorrect_ber_acc_alrm_hi"] = "string"
	c.nxpathMap[key]["rx_pwr_4"] = "float"
	c.nxpathMap[key]["uncorrect_ber_cur"] = "string"
	c.nxpathMap[key]["pre_fec_ber_alrm_hi"] = "string"
	c.nxpathMap[key]["rx_pwr_warn_lo"] = "float"
	c.nxpathMap[key]["bit_encoding"] = "string"
	c.nxpathMap[key]["pre_fec_ber_acc"] = "string"
	c.nxpathMap[key]["sfp"] = "string"
	c.nxpathMap[key]["pre_fec_ber_acc_alrm_hi"] = "string"
	c.nxpathMap[key]["pre_fec_ber_min"] = "string"
	c.nxpathMap[key]["current_warn_hi"] = "float"
	c.nxpathMap[key]["pre_fec_ber_max_alrm_lo"] = "string"
	c.nxpathMap[key]["uncorrect_ber_cur_warn_hi"] = "string"
	c.nxpathMap[key]["current_alrm_hi"] = "float"
	c.nxpathMap[key]["pre_fec_ber_acc_alrm_lo"] = "string"
	c.nxpathMap[key]["snr_alrm_lo"] = "float"
	c.nxpathMap[key]["uncorrect_ber_acc"] = "string"
	c.nxpathMap[key]["tx_len"] = "string"
	c.nxpathMap[key]["uncorrect_ber_alrm_lo"] = "string"
	c.nxpathMap[key]["pre_fec_ber_alrm_lo"] = "string"
	c.nxpathMap[key]["txcvr_type"] = "string"
	c.nxpathMap[key]["tec_current_alrm_lo"] = "float"
	c.nxpathMap[key]["volt_alrm_lo"] = "float"
	c.nxpathMap[key]["temp_alrm_hi"] = "float"
	c.nxpathMap[key]["uncorrect_ber_min_warn_lo"] = "string"
	c.nxpathMap[key]["laser_freq"] = "float"
	c.nxpathMap[key]["uncorrect_ber_min_warn_hi"] = "string"
	c.nxpathMap[key]["uncorrect_ber_cur_alrm_lo"] = "string"
	c.nxpathMap[key]["pre_fec_ber_max_warn_hi"] = "string"
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["fiber_type_byte0"] = "string"
	c.nxpathMap[key]["laser_freq_alrm_lo"] = "float"
	c.nxpathMap[key]["pre_fec_ber_cur_warn_hi"] = "string"
	c.nxpathMap[key]["partnum"] = "string"
	c.nxpathMap[key]["snr"] = "float"
	c.nxpathMap[key]["volt_alrm_hi"] = "float"
	c.nxpathMap[key]["connector_type"] = "string"
	c.nxpathMap[key]["tx_medium"] = "string"
	c.nxpathMap[key]["tx_pwr_warn_hi"] = "float"
	c.nxpathMap[key]["cisco_vendor_id"] = "string"
	c.nxpathMap[key]["cisco_ext_id"] = "string"
	c.nxpathMap[key]["uncorrect_ber_max_warn_hi"] = "string"
	c.nxpathMap[key]["pre_fec_ber_max"] = "string"
	c.nxpathMap[key]["uncorrect_ber_min_alrm_hi"] = "string"
	c.nxpathMap[key]["pre_fec_ber_warn_hi"] = "string"
	c.nxpathMap[key]["tx_pwr_alrm_lo"] = "float"
	c.nxpathMap[key]["uncorrect_ber_warn_lo"] = "string"
	c.nxpathMap[key]["10gbe_code"] = "string"
	c.nxpathMap[key]["cable_type"] = "string"
	c.nxpathMap[key]["laser_freq_alrm_hi"] = "float"
	c.nxpathMap[key]["rx_pwr_3"] = "float"
	c.nxpathMap[key]["rx_pwr"] = "float"
	c.nxpathMap[key]["volt_warn_hi"] = "float"
	c.nxpathMap[key]["pre_fec_ber_cur_alrm_hi"] = "string"
	c.nxpathMap[key]["temperature"] = "float"
	c.nxpathMap[key]["voltage"] = "float"
	c.nxpathMap[key]["tx_pwr"] = "float"
	c.nxpathMap[key]["laser_temp_alrm_hi"] = "float"
	c.nxpathMap[key]["tx_speeds"] = "string"
	c.nxpathMap[key]["uncorrect_ber_min_alrm_lo"] = "string"
	c.nxpathMap[key]["pre_fec_ber_min_alrm_hi"] = "string"
	c.nxpathMap[key]["ciscoid"] = "string"
	c.nxpathMap[key]["tx_pwr_warn_lo"] = "float"
	c.nxpathMap[key]["cisco_product_id"] = "string"
	c.nxpathMap[key]["info_not_available"] = "string"
	c.nxpathMap[key]["laser_temp"] = "float"
	c.nxpathMap[key]["pre_fec_ber_cur"] = "string"
	c.nxpathMap[key]["fiber_type_byte1"] = "string"
	c.nxpathMap[key]["tx_type"] = "string"
	c.nxpathMap[key]["pre_fec_ber_min_alrm_lo"] = "string"
	c.nxpathMap[key]["pre_fec_ber_warn_lo"] = "string"
	c.nxpathMap[key]["temp_alrm_lo"] = "float"
	c.nxpathMap[key]["volt_warn_lo"] = "float"
	c.nxpathMap[key]["rx_pwr_alrm_lo"] = "float"
	c.nxpathMap[key]["rx_pwr_2"] = "float"
	c.nxpathMap[key]["tec_current_alrm_hi"] = "float"
	c.nxpathMap[key]["uncorrect_ber_acc_alrm_lo"] = "string"
	c.nxpathMap[key]["uncorrect_ber_max_alrm_hi"] = "string"
	c.nxpathMap[key]["temp_warn_lo"] = "float"
	c.nxpathMap[key]["snr_warn_hi"] = "float"
	c.nxpathMap[key]["laser_temp_warn_lo"] = "float"
	c.nxpathMap[key]["pre_fec_ber_acc_warn_hi"] = "string"
	c.nxpathMap[key]["laser_freq_warn_hi"] = "float"
	c.nxpathMap[key]["uncorrect_ber_max"] = "string"
}

func (c *CiscoTelemetryMDT) initIgmp() {
	key := "show ip igmp groups vrf all"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["group-type"] = "string"
	c.nxpathMap[key]["translate"] = "string"
	c.nxpathMap[key]["sourceaddress"] = "string"
	c.nxpathMap[key]["vrf-cntxt"] = "string"
	c.nxpathMap[key]["expires"] = "string"
	c.nxpathMap[key]["group-addr"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
}

func (c *CiscoTelemetryMDT) initVrfAll() {
	key := "show ip igmp interface vrf all"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["if-name"] = "string"
	c.nxpathMap[key]["static-group-map"] = "string"
	c.nxpathMap[key]["rll"] = "string"
	c.nxpathMap[key]["host-proxy"] = "string"
	c.nxpathMap[key]["il"] = "string"
	c.nxpathMap[key]["join-group-map"] = "string"
	c.nxpathMap[key]["expires"] = "string"
	c.nxpathMap[key]["host-proxy-group-map"] = "string"
	c.nxpathMap[key]["next-query"] = "string"
	c.nxpathMap[key]["q-ver"] = "string"
	c.nxpathMap[key]["if-status"] = "string"
	c.nxpathMap[key]["un-solicited"] = "string"
	c.nxpathMap[key]["ip-sum"] = "string"
}

func (c *CiscoTelemetryMDT) initIgmpSnoop() {
	key := "show ip igmp snooping"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["repsup"] = "string"
	c.nxpathMap[key]["omf_enabled"] = "string"
	c.nxpathMap[key]["v3repsup"] = "string"
	c.nxpathMap[key]["grepsup"] = "string"
	c.nxpathMap[key]["lkupmode"] = "string"
	c.nxpathMap[key]["description"] = "string"
	c.nxpathMap[key]["vlinklocalgrpsup"] = "string"
	c.nxpathMap[key]["gv3repsup"] = "string"
	c.nxpathMap[key]["reportfloodall"] = "string"
	c.nxpathMap[key]["leavegroupaddress"] = "string"
	c.nxpathMap[key]["enabled"] = "string"
	c.nxpathMap[key]["omf"] = "string"
	c.nxpathMap[key]["sq"] = "string"
	c.nxpathMap[key]["sqr"] = "string"
	c.nxpathMap[key]["eht"] = "string"
	c.nxpathMap[key]["fl"] = "string"
	c.nxpathMap[key]["reportfloodenable"] = "string"
	c.nxpathMap[key]["snoop-on"] = "string"
	c.nxpathMap[key]["glinklocalgrpsup"] = "string"
}

func (c *CiscoTelemetryMDT) initIgmpSnoopGroups() {
	key := "show ip igmp snooping groups"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["src-uptime"] = "string"
	c.nxpathMap[key]["source"] = "string"
	c.nxpathMap[key]["dyn-if-name"] = "string"
	c.nxpathMap[key]["raddr"] = "string"
	c.nxpathMap[key]["old-host"] = "string"
	c.nxpathMap[key]["snoop-enabled"] = "string"
	c.nxpathMap[key]["expires"] = "string"
	c.nxpathMap[key]["omf-enabled"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["src-expires"] = "string"
	c.nxpathMap[key]["addr"] = "string"
}

func (c *CiscoTelemetryMDT) initIgmpSnoopGroupDetails() {
	key := "show ip igmp snooping groups detail"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["src-uptime"] = "string"
	c.nxpathMap[key]["source"] = "string"
	c.nxpathMap[key]["dyn-if-name"] = "string"
	c.nxpathMap[key]["raddr"] = "string"
	c.nxpathMap[key]["old-host"] = "string"
	c.nxpathMap[key]["snoop-enabled"] = "string"
	c.nxpathMap[key]["expires"] = "string"
	c.nxpathMap[key]["omf-enabled"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["src-expires"] = "string"
	c.nxpathMap[key]["addr"] = "string"
}

func (c *CiscoTelemetryMDT) initIgmpSnoopGroupsSumm() {
	key := "show ip igmp snooping groups summary"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["src-uptime"] = "string"
	c.nxpathMap[key]["source"] = "string"
	c.nxpathMap[key]["dyn-if-name"] = "string"
	c.nxpathMap[key]["raddr"] = "string"
	c.nxpathMap[key]["old-host"] = "string"
	c.nxpathMap[key]["snoop-enabled"] = "string"
	c.nxpathMap[key]["expires"] = "string"
	c.nxpathMap[key]["omf-enabled"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["src-expires"] = "string"
	c.nxpathMap[key]["addr"] = "string"
}

func (c *CiscoTelemetryMDT) initMrouter() {
	key := "show ip igmp snooping mrouter"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["expires"] = "string"
}

func (c *CiscoTelemetryMDT) initSnoopStats() {
	key := "show ip igmp snooping statistics"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["ut"] = "string"
}

func (c *CiscoTelemetryMDT) initPimInterface() {
	key := "show ip pim interface vrf all"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["if-is-border"] = "string"
	c.nxpathMap[key]["cached_if_status"] = "string"
	c.nxpathMap[key]["genid"] = "string"
	c.nxpathMap[key]["if-name"] = "string"
	c.nxpathMap[key]["last-cleared"] = "string"
	c.nxpathMap[key]["is-pim-vpc-svi"] = "string"
	c.nxpathMap[key]["if-addr"] = "string"
	c.nxpathMap[key]["is-pim-enabled"] = "string"
	c.nxpathMap[key]["pim-dr-address"] = "string"
	c.nxpathMap[key]["hello-timer"] = "string"
	c.nxpathMap[key]["pim-bfd-enabled"] = "string"
	c.nxpathMap[key]["vpc-peer-nbr"] = "string"
	c.nxpathMap[key]["nbr-policy-name"] = "string"
	c.nxpathMap[key]["is-auto-enabled"] = "string"
	c.nxpathMap[key]["if-status"] = "string"
	c.nxpathMap[key]["jp-out-policy-name"] = "string"
	c.nxpathMap[key]["if-addr-summary"] = "string"
	c.nxpathMap[key]["if-dr"] = "string"
	c.nxpathMap[key]["jp-in-policy-name"] = "string"
}

func (c *CiscoTelemetryMDT) initPimNeigh() {
	key := "show ip pim neighbor vrf all"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["longest-hello-intvl"] = "string"
	c.nxpathMap[key]["if-name"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["expires"] = "string"
	c.nxpathMap[key]["bfd-state"] = "string"
}

func (c *CiscoTelemetryMDT) initPimRoute() {
	key := "show ip pim route vrf all"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["rpf-nbr-1"] = "string"
	c.nxpathMap[key]["rpf-nbr-addr"] = "string"
	c.nxpathMap[key]["register"] = "string"
	c.nxpathMap[key]["sgexpire"] = "string"
	c.nxpathMap[key]["oif-bf-str"] = "string"
	c.nxpathMap[key]["mcast-addrs"] = "string"
	c.nxpathMap[key]["rp-addr"] = "string"
	c.nxpathMap[key]["immediate-bf-str"] = "string"
	c.nxpathMap[key]["sgr-prune-list-bf-str"] = "string"
	c.nxpathMap[key]["context-name"] = "string"
	c.nxpathMap[key]["intf-name"] = "string"
	c.nxpathMap[key]["immediate-timeout-bf-str"] = "string"
	c.nxpathMap[key]["rp-local"] = "string"
	c.nxpathMap[key]["sgrexpire"] = "string"
	c.nxpathMap[key]["timeout-bf-str"] = "string"
	c.nxpathMap[key]["timeleft"] = "string"
}

func (c *CiscoTelemetryMDT) initPimRp() {
	key := "show ip pim rp vrf all"
	c.nxpathMap[key] = make(map[string]string, 20)
	c.nxpathMap[key]["is-bsr-forward-only"] = "string"
	c.nxpathMap[key]["is-rpaddr-local"] = "string"
	c.nxpathMap[key]["bsr-expires"] = "string"
	c.nxpathMap[key]["autorp-expire-time"] = "string"
	c.nxpathMap[key]["rp-announce-policy-name"] = "string"
	c.nxpathMap[key]["rp-cand-policy-name"] = "string"
	c.nxpathMap[key]["is-autorp-forward-only"] = "string"
	c.nxpathMap[key]["rp-uptime"] = "string"
	c.nxpathMap[key]["rp-owner-flags"] = "string"
	c.nxpathMap[key]["df-bits-recovered"] = "string"
	c.nxpathMap[key]["bs-timer"] = "string"
	c.nxpathMap[key]["rp-discovery-policy-name"] = "string"
	c.nxpathMap[key]["arp-rp-addr"] = "string"
	c.nxpathMap[key]["auto-rp-addr"] = "string"
	c.nxpathMap[key]["autorp-expires"] = "string"
	c.nxpathMap[key]["is-autorp-enabled"] = "string"
	c.nxpathMap[key]["is-bsr-local"] = "string"
	c.nxpathMap[key]["is-autorp-listen-only"] = "string"
	c.nxpathMap[key]["autorp-dis-timer"] = "string"
	c.nxpathMap[key]["bsr-rp-expires"] = "string"
	c.nxpathMap[key]["static-rp-group-map"] = "string"
	c.nxpathMap[key]["rp-source"] = "string"
	c.nxpathMap[key]["autorp-cand-address"] = "string"
	c.nxpathMap[key]["autorp-up-time"] = "string"
	c.nxpathMap[key]["is-bsr-enabled"] = "string"
	c.nxpathMap[key]["bsr-uptime"] = "string"
	c.nxpathMap[key]["is-bsr-listen-only"] = "string"
	c.nxpathMap[key]["rpf-nbr-address"] = "string"
	c.nxpathMap[key]["is-rp-local"] = "string"
	c.nxpathMap[key]["is-autorp-local"] = "string"
	c.nxpathMap[key]["bsr-policy-name"] = "string"
	c.nxpathMap[key]["grange-grp"] = "string"
	c.nxpathMap[key]["rp-addr"] = "string"
	c.nxpathMap[key]["anycast-rp-addr"] = "string"
}

func (c *CiscoTelemetryMDT) initPimStats() {
	key := "show ip pim statistics vrf all"
	c.nxpathMap[key] = make(map[string]string, 1)
	c.nxpathMap[key]["vrf-name"] = "string"
}

func (c *CiscoTelemetryMDT) initIntfBrief() {
	key := "show interface brief"
	c.nxpathMap[key] = make(map[string]string, 2)
	c.nxpathMap[key]["speed"] = "string"
	c.nxpathMap[key]["vlan"] = "string"
}

func (c *CiscoTelemetryMDT) initPimVrf() {
	key := "show ip pim vrf all"
	c.nxpathMap[key] = make(map[string]string, 1)
	c.nxpathMap[key]["table-id"] = "string"
}

func (c *CiscoTelemetryMDT) initIpMroute() {
	key := "show ip mroute summary vrf all"
	c.nxpathMap[key] = make(map[string]string, 40)
	c.nxpathMap[key]["nat-mode"] = "string"
	c.nxpathMap[key]["oif-name"] = "string"
	c.nxpathMap[key]["nat-route-type"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["mofrr-nbr"] = "string"
	c.nxpathMap[key]["extranet_addr"] = "string"
	c.nxpathMap[key]["stale-route"] = "string"
	c.nxpathMap[key]["pending"] = "string"
	c.nxpathMap[key]["bidir"] = "string"
	c.nxpathMap[key]["expry_timer"] = "string"
	c.nxpathMap[key]["mofrr-iif"] = "string"
	c.nxpathMap[key]["group_addrs"] = "string"
	c.nxpathMap[key]["mpib-name"] = "string"
	c.nxpathMap[key]["rpf"] = "string"
	c.nxpathMap[key]["mcast-addrs"] = "string"
	c.nxpathMap[key]["route-mdt-iod"] = "string"
	c.nxpathMap[key]["sr-oif"] = "string"
	c.nxpathMap[key]["stats-rate-buf"] = "string"
	c.nxpathMap[key]["source_addr"] = "string"
	c.nxpathMap[key]["route-iif"] = "string"
	c.nxpathMap[key]["rpf-nbr"] = "string"
	c.nxpathMap[key]["translated-route-src"] = "string"
	c.nxpathMap[key]["group_addr"] = "string"
	c.nxpathMap[key]["lisp-src-rloc"] = "string"
	c.nxpathMap[key]["stats-pndg"] = "string"
	c.nxpathMap[key]["rate_buf"] = "string"
	c.nxpathMap[key]["extranet_vrf_name"] = "string"
	c.nxpathMap[key]["fabric-interest"] = "string"
	c.nxpathMap[key]["translated-route-grp"] = "string"
	c.nxpathMap[key]["internal"] = "string"
	c.nxpathMap[key]["oif-mpib-name"] = "string"
	c.nxpathMap[key]["oif-uptime"] = "string"
	c.nxpathMap[key]["omd-vpc-svi"] = "string"
	c.nxpathMap[key]["source_addrs"] = "string"
	c.nxpathMap[key]["stale-oif"] = "string"
	c.nxpathMap[key]["core-interest"] = "string"
	c.nxpathMap[key]["oif-list-bitfield"] = "string"
}

func (c *CiscoTelemetryMDT) initIpv6Mroute() {
	key := "show ipv6 mroute summary vrf all"
	c.nxpathMap[key] = make(map[string]string, 40)
	c.nxpathMap[key]["nat-mode"] = "string"
	c.nxpathMap[key]["oif-name"] = "string"
	c.nxpathMap[key]["nat-route-type"] = "string"
	c.nxpathMap[key]["uptime"] = "string"
	c.nxpathMap[key]["mofrr-nbr"] = "string"
	c.nxpathMap[key]["extranet_addr"] = "string"
	c.nxpathMap[key]["stale-route"] = "string"
	c.nxpathMap[key]["pending"] = "string"
	c.nxpathMap[key]["bidir"] = "string"
	c.nxpathMap[key]["expry_timer"] = "string"
	c.nxpathMap[key]["mofrr-iif"] = "string"
	c.nxpathMap[key]["group_addrs"] = "string"
	c.nxpathMap[key]["mpib-name"] = "string"
	c.nxpathMap[key]["rpf"] = "string"
	c.nxpathMap[key]["mcast-addrs"] = "string"
	c.nxpathMap[key]["route-mdt-iod"] = "string"
	c.nxpathMap[key]["sr-oif"] = "string"
	c.nxpathMap[key]["stats-rate-buf"] = "string"
	c.nxpathMap[key]["source_addr"] = "string"
	c.nxpathMap[key]["route-iif"] = "string"
	c.nxpathMap[key]["rpf-nbr"] = "string"
	c.nxpathMap[key]["translated-route-src"] = "string"
	c.nxpathMap[key]["group_addr"] = "string"
	c.nxpathMap[key]["lisp-src-rloc"] = "string"
	c.nxpathMap[key]["stats-pndg"] = "string"
	c.nxpathMap[key]["rate_buf"] = "string"
	c.nxpathMap[key]["extranet_vrf_name"] = "string"
	c.nxpathMap[key]["fabric-interest"] = "string"
	c.nxpathMap[key]["translated-route-grp"] = "string"
	c.nxpathMap[key]["internal"] = "string"
	c.nxpathMap[key]["oif-mpib-name"] = "string"
	c.nxpathMap[key]["oif-uptime"] = "string"
	c.nxpathMap[key]["omd-vpc-svi"] = "string"
	c.nxpathMap[key]["source_addrs"] = "string"
	c.nxpathMap[key]["stale-oif"] = "string"
	c.nxpathMap[key]["core-interest"] = "string"
	c.nxpathMap[key]["oif-list-bitfield"] = "string"
}

func (c *CiscoTelemetryMDT) initVpc() {
	key := "sys/vpc"
	c.nxpathMap[key] = make(map[string]string, 5)
	c.nxpathMap[key]["type2CompatQualStr"] = "string"
	c.nxpathMap[key]["compatQualStr"] = "string"
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["issuFromVer"] = "string"
	c.nxpathMap[key]["issuToVer"] = "string"
}

func (c *CiscoTelemetryMDT) initBgp() {
	key := "sys/bgp"
	c.nxpathMap[key] = make(map[string]string, 18)
	c.nxpathMap[key]["dynRtMap"] = "string"
	c.nxpathMap[key]["nhRtMap"] = "string"
	c.nxpathMap[key]["epePeerSet"] = "string"
	c.nxpathMap[key]["asn"] = "string"
	c.nxpathMap[key]["peerImp"] = "string"
	c.nxpathMap[key]["wght"] = "string"
	c.nxpathMap[key]["assocDom"] = "string"
	c.nxpathMap[key]["tblMap"] = "string"
	c.nxpathMap[key]["unSupprMap"] = "string"
	c.nxpathMap[key]["sessionContImp"] = "string"
	c.nxpathMap[key]["allocLblRtMap"] = "string"
	c.nxpathMap[key]["defMetric"] = "string"
	c.nxpathMap[key]["password"] = "string"
	c.nxpathMap[key]["retainRttRtMap"] = "string"
	c.nxpathMap[key]["clusterId"] = "string"
	c.nxpathMap[key]["localAsn"] = "string"
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["defOrgRtMap"] = "string"
}

func (c *CiscoTelemetryMDT) initCh() {
	key := "sys/ch"
	c.nxpathMap[key] = make(map[string]string, 10)
	c.nxpathMap[key]["fanName"] = "string"
	c.nxpathMap[key]["typeCordConnected"] = "string"
	c.nxpathMap[key]["vendor"] = "string"
	c.nxpathMap[key]["model"] = "string"
	c.nxpathMap[key]["rev"] = "string"
	c.nxpathMap[key]["vdrId"] = "string"
	c.nxpathMap[key]["hardwareAlarm"] = "string"
	c.nxpathMap[key]["unit"] = "string"
	c.nxpathMap[key]["hwVer"] = "string"
}

func (c *CiscoTelemetryMDT) initIntf() {
	key := "sys/intf"
	c.nxpathMap[key] = make(map[string]string, 10)
	c.nxpathMap[key]["descr"] = "string"
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["lastStCause"] = "string"
	c.nxpathMap[key]["description"] = "string"
	c.nxpathMap[key]["unit"] = "string"
	c.nxpathMap[key]["operFECMode"] = "string"
	c.nxpathMap[key]["operBitset"] = "string"
	c.nxpathMap[key]["mdix"] = "string"
}

func (c *CiscoTelemetryMDT) initProcsys() {
	key := "sys/procsys"
	c.nxpathMap[key] = make(map[string]string, 10)
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["id"] = "string"
	c.nxpathMap[key]["upTs"] = "string"
	c.nxpathMap[key]["interval"] = "string"
	c.nxpathMap[key]["memstatus"] = "string"
}

func (c *CiscoTelemetryMDT) initProc() {
	key := "sys/proc"
	c.nxpathMap[key] = make(map[string]string, 2)
	c.nxpathMap[key]["processName"] = "string"
	c.nxpathMap[key]["procArg"] = "string"
}

func (c *CiscoTelemetryMDT) initBfd() {
	key := "sys/bfd/inst"
	c.nxpathMap[key] = make(map[string]string, 4)
	c.nxpathMap[key]["descr"] = "string"
	c.nxpathMap[key]["vrfName"] = "string"
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["name"] = "string"
}

func (c *CiscoTelemetryMDT) initLldp() {
	key := "sys/lldp"
	c.nxpathMap[key] = make(map[string]string, 7)
	c.nxpathMap[key]["sysDesc"] = "string"
	c.nxpathMap[key]["portDesc"] = "string"
	c.nxpathMap[key]["portIdV"] = "string"
	c.nxpathMap[key]["chassisIdV"] = "string"
	c.nxpathMap[key]["sysName"] = "string"
	c.nxpathMap[key]["name"] = "string"
	c.nxpathMap[key]["id"] = "string"
}

func (c *CiscoTelemetryMDT) initDb() {
	c.nxpathMap = make(map[string]map[string]string, 200)

	c.initPower()
	c.initMemPhys()
	c.initBgpV4()
	c.initCpu()
	c.initResources()
	c.initPtpCorrection()
	c.initTrans()
	c.initIgmp()
	c.initVrfAll()
	c.initIgmpSnoop()
	c.initIgmpSnoopGroups()
	c.initIgmpSnoopGroupDetails()
	c.initIgmpSnoopGroupsSumm()
	c.initMrouter()
	c.initSnoopStats()
	c.initPimInterface()
	c.initPimNeigh()
	c.initPimRoute()
	c.initPimRp()
	c.initPimStats()
	c.initIntfBrief()
	c.initPimVrf()
	c.initIpMroute()
	c.initIpv6Mroute()
	c.initVpc()
	c.initBgp()
	c.initCh()
	c.initIntf()
	c.initProcsys()
	c.initProc()
	c.initBfd()
	c.initLldp()
}

// Start the Cisco MDT service
func (c *CiscoTelemetryMDT) Start(acc telegraf.Accumulator) error {
	var err error
	c.acc = acc
	c.listener, err = net.Listen("tcp", c.ServiceAddress)
	if err != nil {
		return err
	}

	c.propMap = make(map[string]func(field *telemetry.TelemetryField, value interface{}) interface{}, 100)
	c.propMap["test"] = nxosValueXformUint64Toint64
	c.propMap["asn"] = nxosValueXformUint64ToString            //uint64 to string.
	c.propMap["subscriptionId"] = nxosValueXformUint64ToString //uint64 to string.
	c.propMap["operState"] = nxosValueXformUint64ToString      //uint64 to string.

	// Invert aliases list
	c.warned = make(map[string]struct{})
	c.aliases = make(map[string]string, len(c.Aliases))
	for alias, path := range c.Aliases {
		c.aliases[path] = alias
	}
	c.initDb()

	c.dmes = make(map[string]string, len(c.Dmes))
	for dme, path := range c.Dmes {
		c.dmes[path] = dme
		if path == "uint64 to int" {
			c.propMap[dme] = nxosValueXformUint64Toint64
		} else if path == "uint64 to string" {
			c.propMap[dme] = nxosValueXformUint64ToString
		} else if path == "string to float64" {
			c.propMap[dme] = nxosValueXformStringTofloat
		} else if path == "string to uint64" {
			c.propMap[dme] = nxosValueXformStringToUint64
		} else if path == "string to int64" {
			c.propMap[dme] = nxosValueXformStringToInt64
		} else if path == "auto-float-xfrom" {
			c.propMap[dme] = nxosValueAutoXformFloatProp
		} else if dme[0:6] == "dnpath" { //path based property map
			js := []byte(path)
			var jsStruct NxPayloadXfromStructure

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
		switch subfield.Name {
		case "vrfName", "address", "maskLen":
			tags[subfield.Name] = decodeTag(subfield)
		}
		if value := decodeValue(subfield); value != nil {
			grouper.Add(measurement, tags, timestamp, subfield.Name, value)
		}
		if subfield.Name != "nextHop" {
			continue
		}
		//For next hop table fill the keys in the tag - which is address and vrfname
		for _, subf := range subfield.Fields {
			for _, ff := range subf.Fields {
				switch ff.Name {
				case "address", "vrfName":
					key := "nextHop/" + ff.Name
					tags[key] = decodeTag(ff)
				}
				if value := decodeValue(ff); value != nil {
					name := "nextHop/" + ff.Name
					grouper.Add(measurement, tags, timestamp, name, value)
				}
			}
		}
	}
}

func (c *CiscoTelemetryMDT) parseClassAttributeField(grouper *metric.SeriesGrouper, field *telemetry.TelemetryField, prefix string, path string, tags map[string]string, timestamp time.Time) {
	// DME structure: https://developer.cisco.com/site/nxapi-dme-model-reference-api/
	var nxAttributes *telemetry.TelemetryField
	isDme := strings.Contains(path, "sys/")
	if path == "rib" {
		//handle native data path rib
		c.parseRib(grouper, field, prefix, path, tags, timestamp)
		return
	}
	if field == nil || !isDme {
		return
	}

	if len(field.Fields) >= 1 && field.Fields[0] != nil && field.Fields[0].Fields != nil && field.Fields[0].Fields[0] != nil && field.Fields[0].Fields[0].Fields[0].Name != "attributes" {
		return
	}
	nxAttributes = field.Fields[0].Fields[0].Fields[0].Fields[0]

	for _, subfield := range nxAttributes.Fields {
		if subfield.Name == "dn" {
			tags["dn"] = decodeTag(subfield)
		} else {
			c.parseContentField(grouper, subfield, "", path, tags, timestamp)
		}
	}
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
	isEVENT := isNXOS && strings.Contains(path, "EVENT-LIST")
        nxChildren = nil
        nxAttributes = nil
	for _, subfield := range field.Fields {
		if isNXOS && subfield.Name == "attributes" && len(subfield.Fields) > 0 {
			nxAttributes = subfield.Fields[0]
		} else if isNXOS && subfield.Name == "children" && len(subfield.Fields) > 0 {
			if !isEVENT {
				nxChildren = subfield
			} else {
				sub := subfield.Fields
				if sub[0] != nil && sub[0].Fields[0].Name == "subscriptionId" && len(sub[0].Fields) >= 2 {
					nxAttributes = sub[0].Fields[1].Fields[0].Fields[0].Fields[0].Fields[0].Fields[0]
				}
			}
			//if nxAttributes == NULL then class based query.
			if nxAttributes == nil {
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
					if len(row.Fields) == 1 {
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
 ##Define Property Xformation, please refer README and https://pubhub.devnetcloud.com/media/dme-docs-9-3-3/docs/appendix/ for Model details.
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
