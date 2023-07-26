package netflow

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
)

// From https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
//
//go:embed layer4_protocol_numbers.csv
var l4ProtoFile []byte

// From https://www.iana.org/assignments/ip-parameters/ip-parameters.xhtml
//
//go:embed ipv4_options.csv
var ip4OptionFile []byte

var l4ProtoMapping map[uint8]string
var ipv4OptionMapping []string

func initL4ProtoMapping() error {
	buf := bytes.NewBuffer(l4ProtoFile)
	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return errors.New("empty file")
	}

	l4ProtoMapping = make(map[uint8]string)
	for _, r := range records[1:] {
		if len(r) != 2 {
			return fmt.Errorf("invalid record: %v", r)
		}
		name := strings.ToLower(r[1])
		if name == "" {
			continue
		}
		id, err := strconv.ParseUint(r[0], 10, 8)
		if err != nil {
			return fmt.Errorf("%w: %v", err, r)
		}
		l4ProtoMapping[uint8(id)] = name
	}

	return nil
}

func initIPv4OptionMapping() error {
	buf := bytes.NewBuffer(ip4OptionFile)
	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return errors.New("empty file")
	}

	ipv4OptionMapping = make([]string, 32)
	for _, r := range records[1:] {
		if len(r) != 2 {
			return fmt.Errorf("invalid record: %v", r)
		}
		idx, err := strconv.ParseUint(r[0], 10, 8)
		if err != nil {
			return fmt.Errorf("%w: %v", err, r)
		}
		ipv4OptionMapping[idx] = r[1]
	}

	return nil
}

func decodeInt32(b []byte) interface{} {
	return int64(int32(binary.BigEndian.Uint32(b)))
}

func decodeUint(b []byte) interface{} {
	switch len(b) {
	case 1:
		return uint64(b[0])
	case 2:
		return uint64(binary.BigEndian.Uint16(b))
	case 4:
		return uint64(binary.BigEndian.Uint32(b))
	case 8:
		return binary.BigEndian.Uint64(b)
	}
	panic(fmt.Errorf("invalid length for uint buffer %v", b))
}

func decodeFloat64(b []byte) interface{} {
	raw := binary.BigEndian.Uint64(b)
	return math.Float64frombits(raw)
}

// According to https://www.rfc-editor.org/rfc/rfc5101#section-6.1.5
func decodeBool(b []byte) interface{} {
	if b[0] == 1 {
		return true
	}
	if b[0] == 2 {
		return false
	}
	return b[0]
}

func decodeHex(b []byte) interface{} {
	if len(b) == 0 {
		return ""
	}
	return "0x" + hex.EncodeToString(b)
}

func decodeString(b []byte) interface{} {
	return string(b)
}

func decodeMAC(b []byte) interface{} {
	mac := net.HardwareAddr(b)
	return mac.String()
}

func decodeIP(b []byte) interface{} {
	ip := net.IP(b)
	return ip.String()
}

func decodeIPFromUint32(a uint32) interface{} {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, a)
	return decodeIP(b)
}

func decodeL4Proto(b []byte) interface{} {
	return mapL4Proto(b[0])
}

func mapL4Proto(id uint8) string {
	name, found := l4ProtoMapping[id]
	if found {
		return name
	}
	return strconv.FormatUint(uint64(id), 10)
}

func decodeIPv4Options(b []byte) interface{} {
	flags := binary.BigEndian.Uint32(b)

	var result []string
	for i := 0; i < 32; i++ {
		name := ipv4OptionMapping[i]
		if name == "" {
			name = fmt.Sprintf("UA%d", i)
		}
		if (flags>>i)&0x01 != 0 {
			result = append(result, name)
		}
	}

	return strings.Join(result, ",")
}

func decodeTCPFlags(b []byte) interface{} {
	if len(b) < 1 {
		return ""
	}

	if len(b) == 1 {
		return mapTCPFlags(b[0])
	}

	// IPFIX has more flags
	results := make([]string, 0, 8)
	for i := 7; i >= 0; i-- {
		if (b[0]>>i)&0x01 != 0 {
			// Currently all flags are reserved so denote the bit set
			results = append(results, "*")
		} else {
			results = append(results, ".")
		}
	}
	return strings.Join(results, "") + mapTCPFlags(b[1])
}

func mapTCPFlags(flags uint8) string {
	flagMapping := []string{
		"F", // FIN
		"S", // SYN
		"R", // RST
		"P", // PSH
		"A", // ACK
		"U", // URG
		"E", // ECE
		"C", // CWR
	}

	result := make([]string, 0, 8)

	for i := 7; i >= 0; i-- {
		if (flags>>i)&0x01 != 0 {
			result = append(result, flagMapping[i])
		} else {
			result = append(result, ".")
		}
	}

	return strings.Join(result, "")
}

func decodeFragmentFlags(b []byte) interface{} {
	flagMapping := []string{
		"*", // do not care
		"*", // do not care
		"*", // do not care
		"*", // do not care
		"*", // do not care
		"M", // MF -- more fragments
		"D", // DF -- don't fragment
		"R", // RS -- reserved
	}

	flags := b[0]
	result := make([]string, 0, 8)
	for i := 7; i >= 0; i-- {
		if (flags>>i)&0x01 != 0 {
			result = append(result, flagMapping[i])
		} else {
			result = append(result, ".")
		}
	}

	return strings.Join(result, "")
}

func decodeSampleAlgo(b []byte) interface{} {
	switch b[0] {
	case 1:
		return "deterministic"
	case 2:
		return "random"
	}
	return strconv.FormatUint(uint64(b[0]), 10)
}

func decodeEngineType(b []byte) interface{} {
	return mapEngineType(b[0])
}

func mapEngineType(b uint8) string {
	switch b {
	case 0:
		return "RP"
	case 1:
		return "VIP/linecard"
	case 2:
		return "PFC/DFC"
	}
	return strconv.FormatUint(uint64(b), 10)
}

func decodeMPLSType(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "unknown"
	case 1:
		return "TE-MIDPT"
	case 2:
		return "Pseudowire"
	case 3:
		return "VPN"
	case 4:
		return "BGP"
	case 5:
		return "LDP"
	case 6:
		return "Path computation element"
	case 7:
		return "OSPFv2"
	case 8:
		return "OSPFv3"
	case 9:
		return "IS-IS"
	case 10:
		return "BGP segment routing Prefix-SID"
	}
	return strconv.FormatUint(uint64(b[0]), 10)
}

func decodeIPVersion(b []byte) interface{} {
	switch b[0] {
	case 4:
		return "IPv4"
	case 6:
		return "IPv6"
	}
	return strconv.FormatUint(uint64(b[0]), 10)
}

func decodeDirection(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "ingress"
	case 1:
		return "egress"
	}
	return strconv.FormatUint(uint64(b[0]), 10)
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-forwarding-status
func decodeFwdStatus(b []byte) interface{} {
	switch b[0] >> 6 {
	case 0:
		return "unknown"
	case 1:
		return "forwarded"
	case 2:
		return "dropped"
	case 3:
		return "consumed"
	}
	return "invalid"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-forwarding-status
func decodeFwdReason(b []byte) interface{} {
	switch b[0] {
	// unknown
	case 0:
		return "unknown"
	// forwarded
	case 64:
		return "unknown"
	case 65:
		return "fragmented"
	case 66:
		return "not fragmented"
	// dropped
	case 128:
		return "unknown"
	case 129:
		return "ACL deny"
	case 130:
		return "ACL drop"
	case 131:
		return "unroutable"
	case 132:
		return "adjacency"
	case 133:
		return "fragmentation and DF set"
	case 134:
		return "bad header checksum"
	case 135:
		return "bad total length"
	case 136:
		return "bad header length"
	case 137:
		return "bad TTL"
	case 138:
		return "policer"
	case 139:
		return "WRED"
	case 140:
		return "RPF"
	case 141:
		return "for us"
	case 142:
		return "bad output interface"
	case 143:
		return "hardware"
	// consumed
	case 192:
		return "unknown"
	case 193:
		return "terminate punt adjacency"
	case 194:
		return "terminate incomplete adjacency"
	case 195:
		return "terminate for us"
	case 14:
		return ""
	}
	return "invalid"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-firewall-event
func decodeFWEvent(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "ignore"
	case 1:
		return "flow created"
	case 2:
		return "flow deleted"
	case 3:
		return "flow denied"
	case 4:
		return "flow alert"
	case 5:
		return "flow update"
	}
	return strconv.FormatUint(uint64(b[0]), 10)
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-flow-end-reason
func decodeFlowEndReason(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "reserved"
	case 1:
		return "idle timeout"
	case 2:
		return "active timeout"
	case 3:
		return "end of flow"
	case 4:
		return "forced end"
	case 5:
		return "lack of resources"
	}
	return "unassigned"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-biflow-direction
func decodeBiflowDirection(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "arbitrary"
	case 1:
		return "initiator"
	case 2:
		return "reverse initiator"
	case 3:
		return "perimeter"
	}
	return "unassigned"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-observation-point-type
func decodeOpsPointType(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "invalid"
	case 1:
		return "physical port"
	case 2:
		return "port channel"
	case 3:
		return "vlan"
	}
	return "unassigned"
}

func decodeAnonStabilityClass(b []byte) interface{} {
	switch b[1] & 0x03 {
	case 1:
		return "session"
	case 2:
		return "exporter-collector"
	case 3:
		return "stable"
	}
	return "undefined"
}

func decodeAnonFlags(b []byte) interface{} {
	var result []string
	if b[0]&(1<<2) != 0 {
		result = append(result, "PmA")
	}

	if b[0]&(1<<3) != 0 {
		result = append(result, "LOR")
	}

	return strings.Join(result, ",")
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-anonymization-technique
func decodeAnonTechnique(b []byte) interface{} {
	tech := binary.BigEndian.Uint16(b)
	switch tech {
	case 0:
		return "undefined"
	case 1:
		return "none"
	case 2:
		return "precision degradation"
	case 3:
		return "binning"
	case 4:
		return "enumeration"
	case 5:
		return "permutation"
	case 6:
		return "structure permutation"
	case 7:
		return "reverse truncation"
	case 8:
		return "noise"
	case 9:
		return "offset"
	}
	return "unassigned"
}

func decodeTechnology(b []byte) interface{} {
	switch string(b) {
	case "yes", "y", "1":
		return "yes"
	case "no", "n", "2":
		return "no"
	case "unassigned", "u", "0":
		return "unassigned"
	}
	switch b[0] {
	case 0:
		return "unassigned"
	case 1:
		return "yes"
	case 2:
		return "no"
	}
	return "undefined"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-nat-type
func decodeIPNatType(b []byte) interface{} {
	tech := binary.BigEndian.Uint16(b)
	switch tech {
	case 0:
		return "unknown"
	case 1:
		return "NAT44"
	case 2:
		return "NAT64"
	case 3:
		return "NAT46"
	case 4:
		return "IPv4 no NAT"
	case 5:
		return "NAT66"
	case 6:
		return "IPv6 no NAT"
	}
	return "unassigned"
}

// https://www.iana.org/assignments/psamp-parameters/psamp-parameters.xhtml
func decodeSelectorAlgorithm(b []byte) interface{} {
	tech := binary.BigEndian.Uint16(b)
	switch tech {
	case 0:
		return "reserved"
	case 1:
		return "systematic count-based sampling"
	case 2:
		return "systematic time-based sampling"
	case 3:
		return "random n-out-of-N sampling"
	case 4:
		return "uniform probabilistic sampling"
	case 5:
		return "property match filtering"
	case 6:
		return "hash based filtering using BOB"
	case 7:
		return "hash based filtering using IPSX"
	case 8:
		return "hash based filtering using CRC"
	case 9:
		return "flow-state dependent"
	}
	return "unassigned"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-value-distribution-method
func decodeValueDistMethod(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "unspecified"
	case 1:
		return "start interval"
	case 2:
		return "end interval"
	case 3:
		return "mid interval"
	case 4:
		return "simple uniform distribution"
	case 5:
		return "proportional uniform distribution"
	case 6:
		return "simulated process"
	case 7:
		return "direct"
	}
	return "unassigned"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-data-link-frame-type
func decodeDataLinkFrameType(b []byte) interface{} {
	switch binary.BigEndian.Uint16(b) {
	case 0x0001:
		return "IEEE802.3 ethernet"
	case 0x0002:
		return "IEEE802.11 MAC"
	}
	return "unassigned"
}

// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-mib-capture-time-semantics
func decodeCaptureTimeSemantics(b []byte) interface{} {
	switch b[0] {
	case 0:
		return "undefined"
	case 1:
		return "begin"
	case 2:
		return "end"
	case 3:
		return "export"
	case 4:
		return "average"
	}
	return "unassigned"
}

func decodeSflowIPVersion(v uint32) string {
	switch v {
	case 0:
		return "unknown"
	case 1:
		return "IPv4"
	case 2:
		return "IPv6"
	}
	return strconv.FormatUint(uint64(v), 10)
}

func decodeSflowSourceInterface(t uint32) string {
	switch t {
	case 0:
		return "in_snmp"
	case 1:
		return "in_vlan_id"
	case 2:
		return "in_phy_interface"
	}
	return ""
}

func decodeSflowHeaderProtocol(t uint32) string {
	switch t {
	case 1:
		return "ETHERNET-ISO8023"
	case 2:
		return "ISO88024-TOKENBUS"
	case 3:
		return "ISO88025-TOKENRING"
	case 4:
		return "FDDI"
	case 5:
		return "FRAME-RELAY"
	case 6:
		return "X25"
	case 7:
		return "PPP"
	case 8:
		return "SMDS"
	case 9:
		return "AAL5"
	case 10:
		return "AAL5-IP"
	case 11:
		return "IPv4"
	case 12:
		return "IPv6"
	case 13:
		return "MPLS"
	}
	return "unassigned"
}
