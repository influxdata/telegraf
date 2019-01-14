package linux_fscache

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestFull(t *testing.T) {
	tmpfile := makeFakeFile([]byte(file_Full))
	defer os.Remove(tmpfile)

	f := FSCache{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := f.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
		"cookies_idx": int64(1), "cookies_dat": int64(0), "cookies_spc": int64(18),
		"objects_alc": int64(0), "objects_nal": int64(0), "objects_avl": int64(0), "objects_ded": int64(0),
		"chk_aux_non": int64(0), "chk_aux_ok": int64(0), "chk_aux_upd": int64(0), "chk_aux_obs": int64(0),
		"pages_mrk": int64(13), "pages_unc": int64(0),
		"acquire_n": int64(0), "acquire_nul": int64(0), "acquire_noc": int64(0), "acquire_ok": int64(0), "acquire_nbf": int64(0), "acquire_oom": int64(0),
		"lookups_n": int64(0), "lookups_neg": int64(0), "lookups_pos": int64(0), "lookups_crt": int64(0), "lookups_tmo": int64(0),
		"invals_n": int64(0), "invals_run": int64(0),
		"updates_n": int64(0), "updates_nul": int64(27), "updates_run": int64(0),
		"relinqs_n": int64(0), "relinqs_nul": int64(0), "relinqs_wcr": int64(0), "relinqs_rtr": int64(0),
		"attr_chg_n": int64(0), "attr_chg_ok": int64(0), "attr_chg_nbf": int64(0), "attr_chg_oom": int64(0), "attr_chg_run": int64(0),
		"allocs_n": int64(0), "allocs_ok": int64(0), "allocs_wt": int64(0), "allocs_nbf": int64(0), "allocs_int": int64(0),
		"allocs_ops": int64(0), "allocs_owt": int64(0), "allocs_abt": int64(0),
		"retrvls_n": int64(0), "retrvls_ok": int64(0), "retrvls_wt": int64(0), "retrvls_nod": int64(0), "retrvls_nbf": int64(0), "retrvls_int": int64(0), "retrvls_oom": int64(0),
		"retrvls_ops": int64(0), "retrvls_owt": int64(0), "retrvls_abt": int64(0),
		"stores_n": int64(45), "stores_ok": int64(0), "stores_agn": int64(0), "stores_nbf": int64(0), "stores_oom": int64(0), "stores_wrxd": int64(0), "stores_sol": int64(0),
		"stores_ops": int64(0), "stores_run": int64(0), "stores_pgs": int64(0), "stores_rxd": int64(0), "stores_irxd": int64(0), "stores_olm": int64(42), "stores_ipp": int64(0),
		"vm_scan_nos": int64(0), "vm_scan_gon": int64(0), "vm_scan_bsy": int64(0), "vm_scan_can": int64(0), "vm_scan_wt": int64(0),
		"ops_pend": int64(0), "ops_run": int64(0), "ops_enq": int64(0), "ops_can": int64(0), "ops_rej": int64(0),
		"ops_ini": int64(0), "ops_dfr": int64(0), "ops_rel": int64(0), "ops_gc": int64(0),
		"cache_op_alo": int64(0), "cache_op_luo": int64(92), "cache_op_luc": int64(0), "cache_op_gro": int64(0),
		"cache_op_inv": int64(12), "cache_op_upo": int64(0), "cache_op_dro": int64(0), "cache_op_pto": int64(0), "cache_op_atc": int64(0), "cache_op_syn": int64(0),
		"cache_op_rap": int64(0), "cache_op_ras": int64(0), "cache_op_alp": int64(0), "cache_op_als": int64(0), "cache_op_wrp": int64(0), "cache_op_ucp": int64(0), "cache_op_dsp": int64(0),
		"cache_ev_nsp": int64(0), "cache_ev_stl": int64(0), "cache_ev_rtr": int64(0), "cache_ev_cul": int64(89765121),
	}

	acc.AssertContainsFields(t, "linux_fscache", fields)
}

func TestEmpty(t *testing.T) {
	tmpfile := makeFakeFile([]byte(file_Empty))
	defer os.Remove(tmpfile)

	f := FSCache{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := f.Gather(&acc)
	assert.NoError(t, err)
}

func TestMissing(t *testing.T) {
	f := FSCache{
		statFile: "",
	}

	acc := testutil.Accumulator{}
	err := f.Gather(&acc)
	assert.Error(t, err)
}

const file_Full = `FS-Cache statistics(ver:1.0)
Cookies: idx=1 dat=0 spc=18
Objects: alc=0 nal=0 avl=0 ded=0
ChkAux : non=0 ok=0 upd=0 obs=0
Pages  : mrk=13 unc=0
Acquire: n=0 nul=0 noc=0 ok=0 nbf=0 oom=0
Lookups: n=0 neg=0 pos=0 crt=0 tmo=0
Invals : n=0 run=0
Updates: n=0 nul=27 run=0
Relinqs: n=0 nul=0 wcr=0 rtr=0
AttrChg: n=0 ok=0 nbf=0 oom=0 run=0
Allocs : n=0 ok=0 wt=0 nbf=0 int=0
Allocs : ops=0 owt=0 abt=0
Retrvls: n=0 ok=0 wt=0 nod=0 nbf=0 int=0 oom=0
Retrvls: ops=0 owt=0 abt=0
Stores : n=45 ok=0 agn=0 nbf=0 oom=0 wrxd=0 sol=0
Stores : ops=0 run=0 pgs=0 rxd=0 irxd=0 olm=42 ipp=0
VmScan : nos=0 gon=0 bsy=0 can=0 wt=0
Ops    : pend=0 run=0 enq=0 can=0 rej=0
Ops    : ini=0 dfr=0 rel=0 gc=0
CacheOp: alo=0 luo=92 luc=0 gro=0
CacheOp: inv=12 upo=0 dro=0 pto=0 atc=0 syn=0
CacheOp: rap=0 ras=0 alp=0 als=0 wrp=0 ucp=0 dsp=0
CacheEv: nsp=0 stl=0 rtr=0 cul=89765121`

const file_Empty = ``

func makeFakeFile(content []byte) string {
	tmpfile, err := ioutil.TempFile("", "fscache_test")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}
