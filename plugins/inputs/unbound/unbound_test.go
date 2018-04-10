package unbound

import (
	"bytes"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

var TestTimeout = internal.Duration{Duration: time.Second}

func UnboundControl(output string, Timeout internal.Duration, useSudo bool, Server string) func(string, internal.Duration, bool, string) (*bytes.Buffer, error) {
	return func(string, internal.Duration, bool, string) (*bytes.Buffer, error) {
		return bytes.NewBuffer([]byte(output)), nil
	}
}

func TestParseFullOutput(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Unbound{
		run: UnboundControl(fullOutput, TestTimeout, true, ""),
	}
	err := v.Gather(acc)

	assert.NoError(t, err)

	assert.True(t, acc.HasMeasurement("unbound"))

	assert.Len(t, acc.Metrics, 1)
	assert.Equal(t, acc.NFields(), 63)

	acc.AssertContainsFields(t, "unbound", parsedFullOutput)
}

var parsedFullOutput = map[string]interface{}{
	"thread0_num_queries":              float64(11907596),
	"thread0_num_cachehits":            float64(11489288),
	"thread0_num_cachemiss":            float64(418308),
	"thread0_num_prefetch":             float64(0),
	"thread0_num_recursivereplies":     float64(418308),
	"thread0_requestlist_avg":          float64(0.400229),
	"thread0_requestlist_max":          float64(11),
	"thread0_requestlist_overwritten":  float64(0),
	"thread0_requestlist_exceeded":     float64(0),
	"thread0_requestlist_current_all":  float64(0),
	"thread0_requestlist_current_user": float64(0),
	"thread0_recursion_time_avg":       float64(0.015020),
	"thread0_recursion_time_median":    float64(0.00292343),
	"total_num_queries":                float64(11907596),
	"total_num_cachehits":              float64(11489288),
	"total_num_cachemiss":              float64(418308),
	"total_num_prefetch":               float64(0),
	"total_num_recursivereplies":       float64(418308),
	"total_requestlist_avg":            float64(0.400229),
	"total_requestlist_max":            float64(11),
	"total_requestlist_overwritten":    float64(0),
	"total_requestlist_exceeded":       float64(0),
	"total_requestlist_current_all":    float64(0),
	"total_requestlist_current_user":   float64(0),
	"total_recursion_time_avg":         float64(0.015020),
	"total_recursion_time_median":      float64(0.00292343),
	"time_now":                         float64(1509968734.735180),
	"time_up":                          float64(1472897.672099),
	"time_elapsed":                     float64(1472897.672099),
	"mem_total_sbrk":                   float64(7462912),
	"mem_cache_rrset":                  float64(285056),
	"mem_cache_message":                float64(320000),
	"mem_mod_iterator":                 float64(16532),
	"mem_mod_validator":                float64(112097),
	"num_query_type_A":                 float64(7062688),
	"num_query_type_PTR":               float64(43097),
	"num_query_type_TXT":               float64(2998),
	"num_query_type_AAAA":              float64(4499711),
	"num_query_type_SRV":               float64(5691),
	"num_query_type_ANY":               float64(293411),
	"num_query_class_IN":               float64(11907596),
	"num_query_opcode_QUERY":           float64(11907596),
	"num_query_tcp":                    float64(293411),
	"num_query_ipv6":                   float64(0),
	"num_query_flags_QR":               float64(0),
	"num_query_flags_AA":               float64(0),
	"num_query_flags_TC":               float64(0),
	"num_query_flags_RD":               float64(11907596),
	"num_query_flags_RA":               float64(0),
	"num_query_flags_Z":                float64(0),
	"num_query_flags_AD":               float64(1),
	"num_query_flags_CD":               float64(0),
	"num_query_edns_present":           float64(6202),
	"num_query_edns_DO":                float64(6201),
	"num_answer_rcode_NOERROR":         float64(11857463),
	"num_answer_rcode_SERVFAIL":        float64(17),
	"num_answer_rcode_NXDOMAIN":        float64(50116),
	"num_answer_rcode_nodata":          float64(3914360),
	"num_answer_secure":                float64(44289),
	"num_answer_bogus":                 float64(1),
	"num_rrset_bogus":                  float64(0),
	"unwanted_queries":                 float64(0),
	"unwanted_replies":                 float64(0),
}

var fullOutput = `thread0.num.queries=11907596
thread0.num.cachehits=11489288
thread0.num.cachemiss=418308
thread0.num.prefetch=0
thread0.num.recursivereplies=418308
thread0.requestlist.avg=0.400229
thread0.requestlist.max=11
thread0.requestlist.overwritten=0
thread0.requestlist.exceeded=0
thread0.requestlist.current.all=0
thread0.requestlist.current.user=0
thread0.recursion.time.avg=0.015020
thread0.recursion.time.median=0.00292343
total.num.queries=11907596
total.num.cachehits=11489288
total.num.cachemiss=418308
total.num.prefetch=0
total.num.recursivereplies=418308
total.requestlist.avg=0.400229
total.requestlist.max=11
total.requestlist.overwritten=0
total.requestlist.exceeded=0
total.requestlist.current.all=0
total.requestlist.current.user=0
total.recursion.time.avg=0.015020
total.recursion.time.median=0.00292343
time.now=1509968734.735180
time.up=1472897.672099
time.elapsed=1472897.672099
mem.total.sbrk=7462912
mem.cache.rrset=285056
mem.cache.message=320000
mem.mod.iterator=16532
mem.mod.validator=112097
histogram.000000.000000.to.000000.000001=20
histogram.000000.000001.to.000000.000002=5
histogram.000000.000002.to.000000.000004=13
histogram.000000.000004.to.000000.000008=18
histogram.000000.000008.to.000000.000016=67
histogram.000000.000016.to.000000.000032=94
histogram.000000.000032.to.000000.000064=113
histogram.000000.000064.to.000000.000128=190
histogram.000000.000128.to.000000.000256=369
histogram.000000.000256.to.000000.000512=1034
histogram.000000.000512.to.000000.001024=5503
histogram.000000.001024.to.000000.002048=155724
histogram.000000.002048.to.000000.004096=107623
histogram.000000.004096.to.000000.008192=17739
histogram.000000.008192.to.000000.016384=4177
histogram.000000.016384.to.000000.032768=82021
histogram.000000.032768.to.000000.065536=33772
histogram.000000.065536.to.000000.131072=7159
histogram.000000.131072.to.000000.262144=1109
histogram.000000.262144.to.000000.524288=295
histogram.000000.524288.to.000001.000000=890
histogram.000001.000000.to.000002.000000=136
histogram.000002.000000.to.000004.000000=233
histogram.000004.000000.to.000008.000000=2
histogram.000008.000000.to.000016.000000=0
histogram.000016.000000.to.000032.000000=2
histogram.000032.000000.to.000064.000000=0
histogram.000064.000000.to.000128.000000=0
histogram.000128.000000.to.000256.000000=0
histogram.000256.000000.to.000512.000000=0
histogram.000512.000000.to.001024.000000=0
histogram.001024.000000.to.002048.000000=0
histogram.002048.000000.to.004096.000000=0
histogram.004096.000000.to.008192.000000=0
histogram.008192.000000.to.016384.000000=0
histogram.016384.000000.to.032768.000000=0
histogram.032768.000000.to.065536.000000=0
histogram.065536.000000.to.131072.000000=0
histogram.131072.000000.to.262144.000000=0
histogram.262144.000000.to.524288.000000=0
num.query.type.A=7062688
num.query.type.PTR=43097
num.query.type.TXT=2998
num.query.type.AAAA=4499711
num.query.type.SRV=5691
num.query.type.ANY=293411
num.query.class.IN=11907596
num.query.opcode.QUERY=11907596
num.query.tcp=293411
num.query.ipv6=0
num.query.flags.QR=0
num.query.flags.AA=0
num.query.flags.TC=0
num.query.flags.RD=11907596
num.query.flags.RA=0
num.query.flags.Z=0
num.query.flags.AD=1
num.query.flags.CD=0
num.query.edns.present=6202
num.query.edns.DO=6201
num.answer.rcode.NOERROR=11857463
num.answer.rcode.SERVFAIL=17
num.answer.rcode.NXDOMAIN=50116
num.answer.rcode.nodata=3914360
num.answer.secure=44289
num.answer.bogus=1
num.rrset.bogus=0
unwanted.queries=0
unwanted.replies=0`
