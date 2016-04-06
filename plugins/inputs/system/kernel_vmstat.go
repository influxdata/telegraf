// +build linux

package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// /proc/vmstat file line prefixes to gather stats on.
// This is currently not being used as we are retrieving all the stats. Left here for references.
var (
	nr_free_pages                 = []byte("nr_free_pages")
	nr_inactive_anon              = []byte("nr_inactive_anon")
	nr_active_anon                = []byte("nr_active_anon")
	nr_inactive_file              = []byte("nr_inactive_file")
	nr_active_file                = []byte("nr_active_file")
	nr_unevictable                = []byte("nr_unevictable")
	nr_mlock                      = []byte("nr_mlock")
	nr_anon_pages                 = []byte("nr_anon_pages")
	nr_mapped                     = []byte("nr_mapped")
	nr_file_pages                 = []byte("nr_file_pages")
	nr_dirty                      = []byte("nr_dirty")
	nr_writeback                  = []byte("nr_writeback")
	nr_slab_reclaimable           = []byte("nr_slab_reclaimable")
	nr_slab_unreclaimable         = []byte("nr_slab_unreclaimable")
	nr_page_table_pages           = []byte("nr_page_table_pages")
	nr_kernel_stack               = []byte("nr_kernel_stack")
	nr_unstable                   = []byte("nr_unstable")
	nr_bounce                     = []byte("nr_bounce")
	nr_vmscan_write               = []byte("nr_vmscan_write")
	nr_writeback_temp             = []byte("nr_writeback_temp")
	nr_isolated_anon              = []byte("nr_isolated_anon")
	nr_isolated_file              = []byte("nr_isolated_file")
	nr_shmem                      = []byte("nr_shmem")
	numa_hit                      = []byte("numa_hit")
	numa_miss                     = []byte("numa_miss")
	numa_foreign                  = []byte("numa_foreign")
	numa_interleave               = []byte("numa_interleave")
	numa_local                    = []byte("numa_local")
	numa_other                    = []byte("numa_other")
	nr_anon_transparent_hugepages = []byte("nr_anon_transparent_hugepages")
	pgpgin                        = []byte("pgpgin")
	pgpgout                       = []byte("pgpgout")
	pswpin                        = []byte("pswpin")
	pswpout                       = []byte("pswpout")
	pgalloc_dma                   = []byte("pgalloc_dma")
	pgalloc_dma32                 = []byte("pgalloc_dma32")
	pgalloc_normal                = []byte("pgalloc_normal")
	pgalloc_movable               = []byte("pgalloc_movable")
	pgfree                        = []byte("pgfree")
	pgactivate                    = []byte("pgactivate")
	pgdeactivate                  = []byte("pgdeactivate")
	pgfault                       = []byte("pgfault")
	pgmajfault                    = []byte("pgmajfault")
	pgrefill_dma                  = []byte("pgrefill_dma")
	pgrefill_dma32                = []byte("pgrefill_dma32")
	pgrefill_normal               = []byte("pgrefill_normal")
	pgrefill_movable              = []byte("pgrefill_movable")
	pgsteal_dma                   = []byte("pgsteal_dma")
	pgsteal_dma32                 = []byte("pgsteal_dma32")
	pgsteal_normal                = []byte("pgsteal_normal")
	pgsteal_movable               = []byte("pgsteal_movable")
	pgscan_kswapd_dma             = []byte("pgscan_kswapd_dma")
	pgscan_kswapd_dma32           = []byte("pgscan_kswapd_dma32")
	pgscan_kswapd_normal          = []byte("pgscan_kswapd_normal")
	pgscan_kswapd_movable         = []byte("pgscan_kswapd_movable")
	pgscan_direct_dma             = []byte("pgscan_direct_dma")
	pgscan_direct_dma32           = []byte("pgscan_direct_dma32")
	pgscan_direct_normal          = []byte("pgscan_direct_normal")
	pgscan_direct_movable         = []byte("pgscan_direct_movable")
	zone_reclaim_failed           = []byte("zone_reclaim_failed")
	pginodesteal                  = []byte("pginodesteal")
	slabs_scanned                 = []byte("slabs_scanned")
	kswapd_steal                  = []byte("kswapd_steal")
	kswapd_inodesteal             = []byte("kswapd_inodesteal")
	kswapd_low_wmark_hit_quickly  = []byte("kswapd_low_wmark_hit_quickly")
	kswapd_high_wmark_hit_quickly = []byte("kswapd_high_wmark_hit_quickly")
	kswapd_skip_congestion_wait   = []byte("kswapd_skip_congestion_wait")
	pageoutrun                    = []byte("pageoutrun")
	allocstall                    = []byte("allocstall")
	pgrotated                     = []byte("pgrotated")
	compact_blocks_moved          = []byte("compact_blocks_moved")
	compact_pages_moved           = []byte("compact_pages_moved")
	compact_pagemigrate_failed    = []byte("compact_pagemigrate_failed")
	compact_stall                 = []byte("compact_stall")
	compact_fail                  = []byte("compact_fail")
	compact_success               = []byte("compact_success")
	htlb_buddy_alloc_success      = []byte("htlb_buddy_alloc_success")
	htlb_buddy_alloc_fail         = []byte("htlb_buddy_alloc_fail")
	unevictable_pgs_culled        = []byte("unevictable_pgs_culled")
	unevictable_pgs_scanned       = []byte("unevictable_pgs_scanned")
	unevictable_pgs_rescued       = []byte("unevictable_pgs_rescued")
	unevictable_pgs_mlocked       = []byte("unevictable_pgs_mlocked")
	unevictable_pgs_munlocked     = []byte("unevictable_pgs_munlocked")
	unevictable_pgs_cleared       = []byte("unevictable_pgs_cleared")
	unevictable_pgs_stranded      = []byte("unevictable_pgs_stranded")
	unevictable_pgs_mlockfreed    = []byte("unevictable_pgs_mlockfreed")
	thp_fault_alloc               = []byte("thp_fault_alloc")
	thp_fault_fallback            = []byte("thp_fault_fallback")
	thp_collapse_alloc            = []byte("thp_collapse_alloc")
	thp_collapse_alloc_failed     = []byte("thp_collapse_alloc_failed")
	thp_split                     = []byte("thp_split")
)

type KernelVmstat struct {
	statFile string
}

func (k *KernelVmstat) Description() string {
	return "Get kernel statistics from /proc/vmstat"
}

func (k *KernelVmstat) SampleConfig() string {
	return `[[inputs.kernel_vmstat]]`
}

func (k *KernelVmstat) Gather(acc telegraf.Accumulator) error {
	data, err := k.getProcVmstat()
	if err != nil {
		return err
	}

	fields := make(map[string]interface{})

	dataFields := bytes.Fields(data)
	for i, field := range dataFields {

		// dataFields is an array of {"stat1_name", "stat1_value", "stat2_name", "stat2_value", ...}
		// We only want the even number index as that contain the stat name.
		if i%2 == 0 {
			// Convert the stat value into an integer.
			m, err := strconv.Atoi(string(dataFields[i+1]))
			if err != nil {
				return err
			}

			fields[string(field)] = int64(m)
		}
	}

	acc.AddFields("kernel_vmstat", fields, map[string]string{})
	return nil
}

func (k *KernelVmstat) getProcVmstat() ([]byte, error) {
	if _, err := os.Stat(k.statFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("kernel_vmstat: %s does not exist!", k.statFile)
	} else if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(k.statFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func init() {
	inputs.Add("kernel_vmstat", func() telegraf.Input {
		return &KernelVmstat{
			statFile: "/proc/vmstat",
		}
	})
}
