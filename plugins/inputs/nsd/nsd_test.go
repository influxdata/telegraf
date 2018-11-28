package nsd

import (
	"bytes"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var TestTimeout = internal.Duration{Duration: time.Second}

func NsdControl(output string, Timeout internal.Duration, useSudo bool, Server string, ThreadAsTag bool) func(string, internal.Duration, bool, string, bool) (*bytes.Buffer, error) {
	return func(string, internal.Duration, bool, string, bool) (*bytes.Buffer, error) {
		return bytes.NewBuffer([]byte(output)), nil
	}
}

func TestParseFullOutput(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Nsd{
		run: NsdControl(fullOutput, TestTimeout, true, "", false),
	}
	err := v.Gather(acc)

	assert.NoError(t, err)

	assert.True(t, acc.HasMeasurement("nsd"))

	assert.Len(t, acc.Metrics, 1)
	assert.Equal(t, acc.NFields(), 89)

	acc.AssertContainsFields(t, "nsd", parsedFullOutput)
}

func TestParseFullOutputThreadAsTag(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Nsd{
		run:         NsdControl(fullOutput, TestTimeout, true, "", true),
		ThreadAsTag: true,
	}
	err := v.Gather(acc)

	assert.NoError(t, err)

	assert.True(t, acc.HasMeasurement("nsd"))
	assert.True(t, acc.HasMeasurement("nsd_threads"))

	assert.Len(t, acc.Metrics, 2)
	assert.Equal(t, acc.NFields(), 89)

	acc.AssertContainsFields(t, "nsd", parsedFullOutputThreadAsTagMeasurementNsd)
	acc.AssertContainsFields(t, "nsd_threads", parsedFullOutputThreadAsTagMeasurementNsdThreads)
}

var parsedFullOutput = map[string]interface{}{
	"server0_queries":     uint64(32),
	"num_queries":         uint64(32),
	"time_boot":           float64(340867.515436),
	"time_elapsed":        float64(3522.901971),
	"size_db_disk":        uint64(11275648),
	"size_db_mem":         uint64(5910672),
	"size_xfrd_mem":       uint64(83979048),
	"size_config_disk":    uint64(0),
	"size_config_mem":     uint64(15600),
	"num_type_A":          uint64(24),
	"num_type_NS":         uint64(1),
	"num_type_MD":         uint64(0),
	"num_type_MF":         uint64(0),
	"num_type_CNAME":      uint64(0),
	"num_type_SOA":        uint64(0),
	"num_type_MB":         uint64(0),
	"num_type_MG":         uint64(0),
	"num_type_MR":         uint64(0),
	"num_type_NULL":       uint64(0),
	"num_type_WKS":        uint64(0),
	"num_type_PTR":        uint64(0),
	"num_type_HINFO":      uint64(0),
	"num_type_MINFO":      uint64(0),
	"num_type_MX":         uint64(0),
	"num_type_TXT":        uint64(0),
	"num_type_RP":         uint64(0),
	"num_type_AFSDB":      uint64(0),
	"num_type_X25":        uint64(0),
	"num_type_ISDN":       uint64(0),
	"num_type_RT":         uint64(0),
	"num_type_NSAP":       uint64(0),
	"num_type_SIG":        uint64(0),
	"num_type_KEY":        uint64(0),
	"num_type_PX":         uint64(0),
	"num_type_AAAA":       uint64(0),
	"num_type_LOC":        uint64(0),
	"num_type_NXT":        uint64(0),
	"num_type_SRV":        uint64(0),
	"num_type_NAPTR":      uint64(0),
	"num_type_KX":         uint64(0),
	"num_type_CERT":       uint64(0),
	"num_type_DNAME":      uint64(0),
	"num_type_OPT":        uint64(0),
	"num_type_APL":        uint64(0),
	"num_type_DS":         uint64(5),
	"num_type_SSHFP":      uint64(0),
	"num_type_IPSECKEY":   uint64(0),
	"num_type_RRSIG":      uint64(0),
	"num_type_NSEC":       uint64(0),
	"num_type_DNSKEY":     uint64(2),
	"num_type_DHCID":      uint64(0),
	"num_type_NSEC3":      uint64(0),
	"num_type_NSEC3PARAM": uint64(0),
	"num_type_TLSA":       uint64(0),
	"num_type_SMIMEA":     uint64(0),
	"num_type_CDS":        uint64(0),
	"num_type_CDNSKEY":    uint64(0),
	"num_type_OPENPGPKEY": uint64(0),
	"num_type_CSYNC":      uint64(0),
	"num_type_SPF":        uint64(0),
	"num_type_NID":        uint64(0),
	"num_type_L32":        uint64(0),
	"num_type_L64":        uint64(0),
	"num_type_LP":         uint64(0),
	"num_type_EUI48":      uint64(0),
	"num_type_EUI64":      uint64(0),
	"num_opcode_QUERY":    uint64(32),
	"num_class_IN":        uint64(32),
	"num_rcode_NOERROR":   uint64(16),
	"num_rcode_FORMERR":   uint64(0),
	"num_rcode_SERVFAIL":  uint64(0),
	"num_rcode_NXDOMAIN":  uint64(16),
	"num_rcode_NOTIMP":    uint64(0),
	"num_rcode_REFUSED":   uint64(0),
	"num_rcode_YXDOMAIN":  uint64(0),
	"num_edns":            uint64(32),
	"num_ednserr":         uint64(0),
	"num_udp":             uint64(32),
	"num_udp6":            uint64(0),
	"num_tcp":             uint64(0),
	"num_tcp6":            uint64(0),
	"num_answer_wo_aa":    uint64(8),
	"num_rxerr":           uint64(0),
	"num_txerr":           uint64(0),
	"num_raxfr":           uint64(0),
	"num_truncated":       uint64(0),
	"num_dropped":         uint64(0),
	"zone_master":         uint64(0),
	"zone_slave":          uint64(8),
}

var parsedFullOutputThreadAsTagMeasurementNsdThreads = map[string]interface{}{
	"queries": uint64(32),
}

var parsedFullOutputThreadAsTagMeasurementNsd = map[string]interface{}{
	"num_queries":         uint64(32),
	"time_boot":           float64(340867.515436),
	"time_elapsed":        float64(3522.901971),
	"size_db_disk":        uint64(11275648),
	"size_db_mem":         uint64(5910672),
	"size_xfrd_mem":       uint64(83979048),
	"size_config_disk":    uint64(0),
	"size_config_mem":     uint64(15600),
	"num_type_A":          uint64(24),
	"num_type_NS":         uint64(1),
	"num_type_MD":         uint64(0),
	"num_type_MF":         uint64(0),
	"num_type_CNAME":      uint64(0),
	"num_type_SOA":        uint64(0),
	"num_type_MB":         uint64(0),
	"num_type_MG":         uint64(0),
	"num_type_MR":         uint64(0),
	"num_type_NULL":       uint64(0),
	"num_type_WKS":        uint64(0),
	"num_type_PTR":        uint64(0),
	"num_type_HINFO":      uint64(0),
	"num_type_MINFO":      uint64(0),
	"num_type_MX":         uint64(0),
	"num_type_TXT":        uint64(0),
	"num_type_RP":         uint64(0),
	"num_type_AFSDB":      uint64(0),
	"num_type_X25":        uint64(0),
	"num_type_ISDN":       uint64(0),
	"num_type_RT":         uint64(0),
	"num_type_NSAP":       uint64(0),
	"num_type_SIG":        uint64(0),
	"num_type_KEY":        uint64(0),
	"num_type_PX":         uint64(0),
	"num_type_AAAA":       uint64(0),
	"num_type_LOC":        uint64(0),
	"num_type_NXT":        uint64(0),
	"num_type_SRV":        uint64(0),
	"num_type_NAPTR":      uint64(0),
	"num_type_KX":         uint64(0),
	"num_type_CERT":       uint64(0),
	"num_type_DNAME":      uint64(0),
	"num_type_OPT":        uint64(0),
	"num_type_APL":        uint64(0),
	"num_type_DS":         uint64(5),
	"num_type_SSHFP":      uint64(0),
	"num_type_IPSECKEY":   uint64(0),
	"num_type_RRSIG":      uint64(0),
	"num_type_NSEC":       uint64(0),
	"num_type_DNSKEY":     uint64(2),
	"num_type_DHCID":      uint64(0),
	"num_type_NSEC3":      uint64(0),
	"num_type_NSEC3PARAM": uint64(0),
	"num_type_TLSA":       uint64(0),
	"num_type_SMIMEA":     uint64(0),
	"num_type_CDS":        uint64(0),
	"num_type_CDNSKEY":    uint64(0),
	"num_type_OPENPGPKEY": uint64(0),
	"num_type_CSYNC":      uint64(0),
	"num_type_SPF":        uint64(0),
	"num_type_NID":        uint64(0),
	"num_type_L32":        uint64(0),
	"num_type_L64":        uint64(0),
	"num_type_LP":         uint64(0),
	"num_type_EUI48":      uint64(0),
	"num_type_EUI64":      uint64(0),
	"num_opcode_QUERY":    uint64(32),
	"num_class_IN":        uint64(32),
	"num_rcode_NOERROR":   uint64(16),
	"num_rcode_FORMERR":   uint64(0),
	"num_rcode_SERVFAIL":  uint64(0),
	"num_rcode_NXDOMAIN":  uint64(16),
	"num_rcode_NOTIMP":    uint64(0),
	"num_rcode_REFUSED":   uint64(0),
	"num_rcode_YXDOMAIN":  uint64(0),
	"num_edns":            uint64(32),
	"num_ednserr":         uint64(0),
	"num_udp":             uint64(32),
	"num_udp6":            uint64(0),
	"num_tcp":             uint64(0),
	"num_tcp6":            uint64(0),
	"num_answer_wo_aa":    uint64(8),
	"num_rxerr":           uint64(0),
	"num_txerr":           uint64(0),
	"num_raxfr":           uint64(0),
	"num_truncated":       uint64(0),
	"num_dropped":         uint64(0),
	"zone_master":         uint64(0),
	"zone_slave":          uint64(8),
}

var fullOutput = `
server0.queries=32
num.queries=32
time.boot=340867.515436
time.elapsed=3522.901971
size.db.disk=11275648
size.db.mem=5910672
size.xfrd.mem=83979048
size.config.disk=0
size.config.mem=15600
num.type.A=24
num.type.NS=1
num.type.MD=0
num.type.MF=0
num.type.CNAME=0
num.type.SOA=0
num.type.MB=0
num.type.MG=0
num.type.MR=0
num.type.NULL=0
num.type.WKS=0
num.type.PTR=0
num.type.HINFO=0
num.type.MINFO=0
num.type.MX=0
num.type.TXT=0
num.type.RP=0
num.type.AFSDB=0
num.type.X25=0
num.type.ISDN=0
num.type.RT=0
num.type.NSAP=0
num.type.SIG=0
num.type.KEY=0
num.type.PX=0
num.type.AAAA=0
num.type.LOC=0
num.type.NXT=0
num.type.SRV=0
num.type.NAPTR=0
num.type.KX=0
num.type.CERT=0
num.type.DNAME=0
num.type.OPT=0
num.type.APL=0
num.type.DS=5
num.type.SSHFP=0
num.type.IPSECKEY=0
num.type.RRSIG=0
num.type.NSEC=0
num.type.DNSKEY=2
num.type.DHCID=0
num.type.NSEC3=0
num.type.NSEC3PARAM=0
num.type.TLSA=0
num.type.SMIMEA=0
num.type.CDS=0
num.type.CDNSKEY=0
num.type.OPENPGPKEY=0
num.type.CSYNC=0
num.type.SPF=0
num.type.NID=0
num.type.L32=0
num.type.L64=0
num.type.LP=0
num.type.EUI48=0
num.type.EUI64=0
num.opcode.QUERY=32
num.class.IN=32
num.rcode.NOERROR=16
num.rcode.FORMERR=0
num.rcode.SERVFAIL=0
num.rcode.NXDOMAIN=16
num.rcode.NOTIMP=0
num.rcode.REFUSED=0
num.rcode.YXDOMAIN=0
num.edns=32
num.ednserr=0
num.udp=32
num.udp6=0
num.tcp=0
num.tcp6=0
num.answer_wo_aa=8
num.rxerr=0
num.txerr=0
num.raxfr=0
num.truncated=0
num.dropped=0
zone.master=0
zone.slave=8
`
