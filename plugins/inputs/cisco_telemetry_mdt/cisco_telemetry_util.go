package cisco_telemetry_mdt

import (
	"strconv"
	"strings"

	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
)

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
func nxosValueXformStringTofloat(field *telemetry.TelemetryField, _ interface{}) interface{} {
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
func nxosValueXformStringToUint64(field *telemetry.TelemetryField, _ interface{}) interface{} {
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
func nxosValueXformStringToInt64(field *telemetry.TelemetryField, _ interface{}) interface{} {
	//string to int64
	vals := field.GetStringValue()
	if vals != "" {
		if val64, err := strconv.ParseInt(vals, 10, 64); err == nil {
			return val64
		}
	}
	return nil
}

//auto-xform float properties
func nxosValueAutoXformFloatProp(field *telemetry.TelemetryField, _ interface{}) interface{} {
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
func nxosValueXformUint64ToString(field *telemetry.TelemetryField, _ interface{}) interface{} {
	switch val := field.ValueByType.(type) {
	case *telemetry.TelemetryField_StringValue:
		if len(val.StringValue) > 0 {
			return val.StringValue
		}
	case *telemetry.TelemetryField_Uint64Value:
		return strconv.FormatUint(val.Uint64Value, 10)
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
		return c.propMap[field.Name](field, value)
	}
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
			if ok {
				return vali
			}
		case *telemetry.TelemetryField_Uint64Value:
			vali, ok := value.(uint64)
			if ok {
				return vali
			}
		} //switch
		return nil
	//Xformation supported is only from String
	case "float":
		//nolint:revive // switch needed for `.(type)`
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
	return nil
}

func (c *CiscoTelemetryMDT) initMemPhys() {
	c.nxpathMap["show processes memory physical"] = map[string]string{"processname": "string"}
}

func (c *CiscoTelemetryMDT) initBgpV4() {
	key := "show bgp ipv4 unicast"
	c.nxpathMap[key] = make(map[string]string, 1)
	c.nxpathMap[key]["aspath"] = "string"
}

func (c *CiscoTelemetryMDT) initCPU() {
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

func (c *CiscoTelemetryMDT) initIPMroute() {
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
	c.initCPU()
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
	c.initIPMroute()
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
