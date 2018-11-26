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

var parsedFullOutput = map[string]interface{}{
	"server0_queries":     float64(32),
	"num_queries":         float64(32),
	"time_boot":           float64(340867.515436),
	"time_elapsed":        float64(3522.901971),
	"size_db_disk":        float64(11275648),
	"size_db_mem":         float64(5910672),
	"size_xfrd_mem":       float64(83979048),
	"size_config_disk":    float64(0),
	"size_config_mem":     float64(15600),
	"num_type_A":          float64(24),
	"num_type_NS":         float64(1),
	"num_type_MD":         float64(0),
	"num_type_MF":         float64(0),
	"num_type_CNAME":      float64(0),
	"num_type_SOA":        float64(0),
	"num_type_MB":         float64(0),
	"num_type_MG":         float64(0),
	"num_type_MR":         float64(0),
	"num_type_NULL":       float64(0),
	"num_type_WKS":        float64(0),
	"num_type_PTR":        float64(0),
	"num_type_HINFO":      float64(0),
	"num_type_MINFO":      float64(0),
	"num_type_MX":         float64(0),
	"num_type_TXT":        float64(0),
	"num_type_RP":         float64(0),
	"num_type_AFSDB":      float64(0),
	"num_type_X25":        float64(0),
	"num_type_ISDN":       float64(0),
	"num_type_RT":         float64(0),
	"num_type_NSAP":       float64(0),
	"num_type_SIG":        float64(0),
	"num_type_KEY":        float64(0),
	"num_type_PX":         float64(0),
	"num_type_AAAA":       float64(0),
	"num_type_LOC":        float64(0),
	"num_type_NXT":        float64(0),
	"num_type_SRV":        float64(0),
	"num_type_NAPTR":      float64(0),
	"num_type_KX":         float64(0),
	"num_type_CERT":       float64(0),
	"num_type_DNAME":      float64(0),
	"num_type_OPT":        float64(0),
	"num_type_APL":        float64(0),
	"num_type_DS":         float64(5),
	"num_type_SSHFP":      float64(0),
	"num_type_IPSECKEY":   float64(0),
	"num_type_RRSIG":      float64(0),
	"num_type_NSEC":       float64(0),
	"num_type_DNSKEY":     float64(2),
	"num_type_DHCID":      float64(0),
	"num_type_NSEC3":      float64(0),
	"num_type_NSEC3PARAM": float64(0),
	"num_type_TLSA":       float64(0),
	"num_type_SMIMEA":     float64(0),
	"num_type_CDS":        float64(0),
	"num_type_CDNSKEY":    float64(0),
	"num_type_OPENPGPKEY": float64(0),
	"num_type_CSYNC":      float64(0),
	"num_type_SPF":        float64(0),
	"num_type_NID":        float64(0),
	"num_type_L32":        float64(0),
	"num_type_L64":        float64(0),
	"num_type_LP":         float64(0),
	"num_type_EUI48":      float64(0),
	"num_type_EUI64":      float64(0),
	"num_opcode_QUERY":    float64(32),
	"num_class_IN":        float64(32),
	"num_rcode_NOERROR":   float64(16),
	"num_rcode_FORMERR":   float64(0),
	"num_rcode_SERVFAIL":  float64(0),
	"num_rcode_NXDOMAIN":  float64(16),
	"num_rcode_NOTIMP":    float64(0),
	"num_rcode_REFUSED":   float64(0),
	"num_rcode_YXDOMAIN":  float64(0),
	"num_edns":            float64(32),
	"num_ednserr":         float64(0),
	"num_udp":             float64(32),
	"num_udp6":            float64(0),
	"num_tcp":             float64(0),
	"num_tcp6":            float64(0),
	"num_answer_wo_aa":    float64(8),
	"num_rxerr":           float64(0),
	"num_txerr":           float64(0),
	"num_raxfr":           float64(0),
	"num_truncated":       float64(0),
	"num_dropped":         float64(0),
	"zone_master":         float64(0),
	"zone_slave":          float64(8),
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
