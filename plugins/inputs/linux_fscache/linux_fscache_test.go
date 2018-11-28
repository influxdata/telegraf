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
		"Cookies_idx": int64(1), "Cookies_dat": int64(0), "Cookies_spc": int64(18),
		"Objects_alc": int64(0), "Objects_nal": int64(0), "Objects_avl": int64(0), "Objects_ded": int64(0),
		"ChkAux_non": int64(0), "ChkAux_ok": int64(0), "ChkAux_upd": int64(0), "ChkAux_obs": int64(0),
		"Pages_mrk": int64(13), "Pages_unc": int64(0),
		"Acquire_n": int64(0), "Acquire_nul": int64(0), "Acquire_noc": int64(0), "Acquire_ok": int64(0), "Acquire_nbf": int64(0), "Acquire_oom": int64(0),
		"Lookups_n": int64(0), "Lookups_neg": int64(0), "Lookups_pos": int64(0), "Lookups_crt": int64(0), "Lookups_tmo": int64(0),
		"Invals_n": int64(0), "Invals_run": int64(0),
		"Updates_n": int64(0), "Updates_nul": int64(27), "Updates_run": int64(0),
		"Relinqs_n": int64(0), "Relinqs_nul": int64(0), "Relinqs_wcr": int64(0), "Relinqs_rtr": int64(0),
		"AttrChg_n": int64(0), "AttrChg_ok": int64(0), "AttrChg_nbf": int64(0), "AttrChg_oom": int64(0), "AttrChg_run": int64(0),
		"Allocs_n": int64(0), "Allocs_ok": int64(0), "Allocs_wt": int64(0), "Allocs_nbf": int64(0), "Allocs_int": int64(0),
		"Allocs_ops": int64(0), "Allocs_owt": int64(0), "Allocs_abt": int64(0),
		"Retrvls_n": int64(0), "Retrvls_ok": int64(0), "Retrvls_wt": int64(0), "Retrvls_nod": int64(0), "Retrvls_nbf": int64(0), "Retrvls_int": int64(0), "Retrvls_oom": int64(0),
		"Retrvls_ops": int64(0), "Retrvls_owt": int64(0), "Retrvls_abt": int64(0),
		"Stores_n": int64(45), "Stores_ok": int64(0), "Stores_agn": int64(0), "Stores_nbf": int64(0), "Stores_oom": int64(0), "Stores_wrxd": int64(0), "Stores_sol": int64(0),
		"Stores_ops": int64(0), "Stores_run": int64(0), "Stores_pgs": int64(0), "Stores_rxd": int64(0), "Stores_irxd": int64(0), "Stores_olm": int64(42), "Stores_ipp": int64(0),
		"VmScan_nos": int64(0), "VmScan_gon": int64(0), "VmScan_bsy": int64(0), "VmScan_can": int64(0), "VmScan_wt": int64(0),
		"Ops_pend": int64(0), "Ops_run": int64(0), "Ops_enq": int64(0), "Ops_can": int64(0), "Ops_rej": int64(0),
		"Ops_ini": int64(0), "Ops_dfr": int64(0), "Ops_rel": int64(0), "Ops_gc": int64(0),
		"CacheOp_alo": int64(0), "CacheOp_luo": int64(92), "CacheOp_luc": int64(0), "CacheOp_gro": int64(0),
		"CacheOp_inv": int64(12), "CacheOp_upo": int64(0), "CacheOp_dro": int64(0), "CacheOp_pto": int64(0), "CacheOp_atc": int64(0), "CacheOp_syn": int64(0),
		"CacheOp_rap": int64(0), "CacheOp_ras": int64(0), "CacheOp_alp": int64(0), "CacheOp_als": int64(0), "CacheOp_wrp": int64(0), "CacheOp_ucp": int64(0), "CacheOp_dsp": int64(0),
		"CacheEv_nsp": int64(0), "CacheEv_stl": int64(0), "CacheEv_rtr": int64(0), "CacheEv_cul": int64(89765121),
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
	assert.Error(t, err)
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
