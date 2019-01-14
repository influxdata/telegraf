# Linux FS-Cache Input Plugin

This plugin is only available on Linux.

The linux_fscache plugin gathers info about the statistics available in `/proc/fs/fscache/stats`.

The metrics are documented at `https://www.kernel.org/doc/Documentation/filesystems/caching/fscache.txt`
under the `STATISTICAL INFORMATION` section.

### Configuration:

```toml
# Get linux_fscache statistics from /proc/fs/fscache/stat
[[inputs.linux_fscache]]
  # no configuration
```

### Measurements & Fields:

- linux_fscache
	- Cookies_idx, integer, Number of index cookies allocated
	- Cookies_dat, integer, Number of data storage cookies allocated
	- Cookies_spc, integer, Number of special cookies allocated
	- Objects_alc, integer, Number of data storage cookies allocated
	- Objects_nal, integer, Number of special cookies allocated
	- Objects_avl, integer, Number of objects allocated
	- Objects_ded, integer, Number of object allocation failures
	- ChkAux_non, integer, Number of special cookies allocated
	- ChkAux_ok, integer, Number of objects allocated
	- ChkAux_upd, integer, Number of object allocation failures
	- ChkAux_obs, integer, Number of objects that reached the available state
	- Pages_mrk, integer, Number of objects allocated
	- Pages_unc, integer, Number of object allocation failures
	- Acquire_n, integer, Number of object allocation failures
	- Acquire_nul, integer, Number of objects that reached the available state
	- Acquire_noc, integer, Number of objects that reached the dead state
	- Acquire_ok, integer, Number of objects that didn't have a coherency check
	- Acquire_nbf, integer, Number of objects that passed a coherency check
	- Acquire_oom, integer, Number of objects that needed a coherency data update
	- Lookups_n, integer, Number of objects that reached the available state
	- Lookups_neg, integer, Number of objects that reached the dead state
	- Lookups_pos, integer, Number of objects that didn't have a coherency check
	- Lookups_crt, integer, Number of objects that passed a coherency check
	- Lookups_tmo, integer, Number of objects that needed a coherency data update
	- Invals_n, integer, Number of objects that reached the dead state
	- Invals_run, integer, Number of objects that didn't have a coherency check
	- Updates_n, integer, Number of objects that didn't have a coherency check
	- Updates_nul, integer, Number of objects that passed a coherency check
	- Updates_run, integer, Number of objects that needed a coherency data update
	- Relinqs_n, integer, Number of objects that passed a coherency check
	- Relinqs_nul, integer, Number of objects that needed a coherency data update
	- Relinqs_wcr, integer, Number of objects that were declared obsolete
	- Relinqs_rtr, integer, Number of pages marked as being cached
	- AttrChg_n, integer, Number of objects that needed a coherency data update
	- AttrChg_ok, integer, Number of objects that were declared obsolete
	- AttrChg_nbf, integer, Number of pages marked as being cached
	- AttrChg_oom, integer, Number of uncache page requests seen
	- AttrChg_run, integer, Number of acquire cookie requests seen
	- Allocs_n, integer, Number of objects that were declared obsolete
	- Allocs_ok, integer, Number of pages marked as being cached
	- Allocs_wt, integer, Number of uncache page requests seen
	- Allocs_nbf, integer, Number of acquire cookie requests seen
	- Allocs_int, integer, Number of acq reqs given a NULL parent
	- Allocs_ops, integer, Number of pages marked as being cached
	- Allocs_owt, integer, Number of uncache page requests seen
	- Allocs_abt, integer, Number of acquire cookie requests seen
	- Retrvls_n, integer, Number of uncache page requests seen
	- Retrvls_ok, integer, Number of acquire cookie requests seen
	- Retrvls_wt, integer, Number of acq reqs given a NULL parent
	- Retrvls_nod, integer, Number of acq reqs rejected due to no cache available
	- Retrvls_nbf, integer, Number of acq reqs succeeded
	- Retrvls_int, integer, Number of acq reqs rejected due to error
	- Retrvls_oom, integer, Number of acq reqs failed on ENOMEM
	- Retrvls_ops, integer, Number of acquire cookie requests seen
	- Retrvls_owt, integer, Number of acq reqs given a NULL parent
	- Retrvls_abt, integer, Number of acq reqs rejected due to no cache available
	- Stores_n, integer, Number of acq reqs given a NULL parent
	- Stores_ok, integer, Number of acq reqs rejected due to no cache available
	- Stores_agn, integer, Number of acq reqs succeeded
	- Stores_nbf, integer, Number of acq reqs rejected due to error
	- Stores_oom, integer, Number of acq reqs failed on ENOMEM
	- Stores_wrxd, integer, Number of lookup calls made on cache backends
	- Stores_sol, integer, Number of negative lookups made
	- Stores_ops, integer, Number of acq reqs rejected due to no cache available
	- Stores_run, integer, Number of acq reqs succeeded
	- Stores_pgs, integer, Number of acq reqs rejected due to error
	- Stores_rxd, integer, Number of acq reqs failed on ENOMEM
	- Stores_irxd, integer, Number of lookup calls made on cache backends
	- Stores_olm, integer, Number of negative lookups made
	- Stores_ipp, integer, Number of positive lookups made
	- VmScan_nos, integer, Number of acq reqs succeeded
	- VmScan_gon, integer, Number of acq reqs rejected due to error
	- VmScan_bsy, integer, Number of acq reqs failed on ENOMEM
	- VmScan_can, integer, Number of lookup calls made on cache backends
	- VmScan_wt, integer, Number of negative lookups made
	- Ops_pend, integer, Number of acq reqs rejected due to error
	- Ops_run, integer, Number of acq reqs failed on ENOMEM
	- Ops_enq, integer, Number of lookup calls made on cache backends
	- Ops_can, integer, Number of negative lookups made
	- Ops_rej, integer, Number of positive lookups made
	- Ops_ini, integer, Number of acq reqs failed on ENOMEM
	- Ops_dfr, integer, Number of lookup calls made on cache backends
	- Ops_rel, integer, Number of negative lookups made
	- Ops_gc, integer, Number of positive lookups made
	- CacheOp_alo, integer, Number of lookup calls made on cache backends
	- CacheOp_luo, integer, Number of negative lookups made
	- CacheOp_luc, integer, Number of positive lookups made
	- CacheOp_gro, integer, Number of objects created by lookup
	- CacheOp_inv, integer, Number of negative lookups made
	- CacheOp_upo, integer, Number of positive lookups made
	- CacheOp_dro, integer, Number of objects created by lookup
	- CacheOp_pto, integer, Number of lookups timed out and requeued
	- CacheOp_atc, integer, Number of update cookie requests seen
	- CacheOp_syn, integer, Number of upd reqs given a NULL parent
	- CacheOp_rap, integer, Number of positive lookups made
	- CacheOp_ras, integer, Number of objects created by lookup
	- CacheOp_alp, integer, Number of lookups timed out and requeued
	- CacheOp_als, integer, Number of update cookie requests seen
	- CacheOp_wrp, integer, Number of upd reqs given a NULL parent
	- CacheOp_ucp, integer, Number of upd reqs granted CPU time
	- CacheOp_dsp, integer, Number of relinquish cookie requests seen
	- CacheEv_nsp, integer, Number of objects created by lookup
	- CacheEv_stl, integer, Number of lookups timed out and requeued
	- CacheEv_rtr, integer, Number of update cookie requests seen
	- CacheEv_cul, integer, Number of upd reqs given a NULL parent

### Tags:

None

### Example Output:

```
$ telegraf --config ~/ws/telegraf.conf --input-filter linux_fscache --test
* Plugin: linux_fscache, Collection 1
> linux_fscache cookies_idx=1i,cookies_dat=0i,cookies_spc=18i,
objects_alc=0i,objects_nal=0i,objects_avl=0i,objects_ded=0i,
chk_aux_non=0i,chk_aux_ok=0i,chk_aux_upd=0i,chk_aux_obs=0i,
pages_mrk=13i,pages_unc=0i,
acquire_n=0i,acquire_nul=0i,acquire_noc=0i,acquire_ok=0i,acquire_nbf=0i,acquire_oom=0i,
lookups_n=0i,lookups_neg=0i,lookups_pos=0i,lookups_crt=0i,lookups_tmo=0i,
invals_n=0i,invals_run=0i,
updates_n=0i,updates_nul=27i,updates_run=0i,
relinqs_n=0i,relinqs_nul=0i,relinqs_wcr=0i,relinqs_rtr=0i,
attr_chg_n=0i,attr_chg_ok=0i,attr_chg_nbf=0i,attr_chg_oom=0i,attr_chg_run=0i,
allocs_n=0i,allocs_ok=0i,allocs_wt=0i,allocs_nbf=0i,allocs_int=0i,
allocs_ops=0i,allocs_owt=0i,allocs_abt=0i,
retrvls_n=0i,retrvls_ok=0i,retrvls_wt=0i,retrvls_nod=0i,retrvls_nbf=0i,retrvls_int=0i,retrvls_oom=0i,
retrvls_ops=0i,retrvls_owt=0i,retrvls_abt=0i,
stores_n=45i,stores_ok=0i,stores_agn=0i,stores_nbf=0i,stores_oom=0i,stores_wrxd=0i,stores_sol=0i,
stores_ops=0i,stores_run=0i,stores_pgs=0i,stores_rxd=0i,stores_irxd=0i,stores_olm=42i,stores_ipp=0i,
vm_scan_nos=0i,vm_scan_gon=0i,vm_scan_bsy=0i,vm_scan_can=0i,vm_scan_wt=0i,
ops_pend=0i,ops_run=0i,ops_enq=0i,ops_can=0i,ops_rej=0i,
ops_ini=0i,ops_dfr=0i,ops_rel=0i,ops_gc=0i,
cache_op_alo=0i,cache_op_luo=92i,cache_op_luc=0i,cache_op_gro=0i,
cache_op_inv=12i,cache_op_upo=0i,cache_op_dro=0i,cache_op_pto=0i,cache_op_atc=0i,cache_op_syn=0i,
cache_op_rap=0i,cache_op_ras=0i,cache_op_alp=0i,cache_op_als=0i,cache_op_wrp=0i,cache_op_ucp=0i,cache_op_dsp=0i,
cache_ev_nsp=0i,cache_ev_stl=0i,cache_ev_rtr=0i,cache_ev_cul=89765121i,
```
