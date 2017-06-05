// +build linux

package linux_mem

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var slabinfoSample = `slabinfo - version: 2.1
# name            <active_objs> <num_objs> <objsize> <objperslab> <pagesperslab> : tunables <limit> <batchcount> <sharedfactor> : slabdata <active_slabs> <num_slabs> <sharedavail>
isofs_inode_cache      0    140    584   28    4 : tunables    0    0    0 : slabdata      5      5      0
fuse_inode         14559  14628    704   23    4 : tunables    0    0    0 : slabdata    636    636      0
bio-3                442    442    960   34    8 : tunables    0    0    0 : slabdata     13     13      0
cifs_request           5     12  16512    1    8 : tunables    0    0    0 : slabdata     12     12      0
cifs_inode_cache      24    144    680   24    4 : tunables    0    0    0 : slabdata      6      6      0
nfs_direct_cache       0      0    352   23    2 : tunables    0    0    0 : slabdata      0      0      0
nfs_commit_data       23    138    704   23    4 : tunables    0    0    0 : slabdata      6      6      0
nfs_inode_cache     1086   1221    976   33    8 : tunables    0    0    0 : slabdata     37     37      0
rpc_inode_cache        0      0    576   28    4 : tunables    0    0    0 : slabdata      0      0      0
btrfs_delayed_extent_op      0      0     40  102    1 : tunables    0    0    0 : slabdata      0      0      0
btrfs_delayed_data_ref      0      0     96   42    1 : tunables    0    0    0 : slabdata      0      0      0
btrfs_delayed_tree_ref      0      0     88   46    1 : tunables    0    0    0 : slabdata      0      0      0
btrfs_delayed_ref_head      0      0    160   25    1 : tunables    0    0    0 : slabdata      0      0      0
btrfs_delayed_node      0      0    304   26    2 : tunables    0    0    0 : slabdata      0      0      0
btrfs_ordered_extent      0      0    424   38    4 : tunables    0    0    0 : slabdata      0      0      0
bio-2                 27    125    320   25    2 : tunables    0    0    0 : slabdata      5      5      0
btrfs_extent_buffer     36    145    280   29    2 : tunables    0    0    0 : slabdata      5      5      0
btrfs_extent_state     53    255     80   51    1 : tunables    0    0    0 : slabdata      5      5      0
btrfs_delalloc_work      0      0    152   26    1 : tunables    0    0    0 : slabdata      0      0      0
btrfs_free_space       2     64     64   64    1 : tunables    0    0    0 : slabdata      1      1      0
btrfs_path            42    140    144   28    1 : tunables    0    0    0 : slabdata      5      5      0
btrfs_transaction      0      0    288   28    2 : tunables    0    0    0 : slabdata      0      0      0
btrfs_trans_handle     23     69    176   23    1 : tunables    0    0    0 : slabdata      3      3      0
btrfs_inode            2     66    968   33    8 : tunables    0    0    0 : slabdata      2      2      0
nvidia_stack_cache    134    134  12288    2    8 : tunables    0    0    0 : slabdata     67     67      0
kvm_async_pf           0      0    136   30    1 : tunables    0    0    0 : slabdata      0      0      0
kvm_vcpu               0      0  15312    2    8 : tunables    0    0    0 : slabdata      0      0      0
UDPLITEv6              0      0   1088   30    8 : tunables    0    0    0 : slabdata      0      0      0
UDPv6                240    240   1088   30    8 : tunables    0    0    0 : slabdata      8      8      0
tw_sock_TCPv6          0      0    192   21    1 : tunables    0    0    0 : slabdata      0      0      0
TCPv6                 21    119   1920   17    8 : tunables    0    0    0 : slabdata      7      7      0
kcopyd_job             0      0   3312    9    8 : tunables    0    0    0 : slabdata      0      0      0
cfq_io_cq            235    324    112   36    1 : tunables    0    0    0 : slabdata      9      9      0
bsg_cmd                0      0    312   26    2 : tunables    0    0    0 : slabdata      0      0      0
mqueue_inode_cache     36     36    896   36    8 : tunables    0    0    0 : slabdata      1      1      0
xfs_icr                0      0    144   28    1 : tunables    0    0    0 : slabdata      0      0      0
xfs_ili            27300  27300    152   26    1 : tunables    0    0    0 : slabdata   1050   1050      0
xfs_inode          85835  85850    960   34    8 : tunables    0    0    0 : slabdata   2525   2525      0
xfs_efd_item         340    340    400   20    2 : tunables    0    0    0 : slabdata     17     17      0
xfs_trans            930   1470    232   35    2 : tunables    0    0    0 : slabdata     42     42      0
xfs_da_state         170    204    480   34    4 : tunables    0    0    0 : slabdata      6      6      0
pid_namespace         65    112   2224   14    8 : tunables    0    0    0 : slabdata      8      8      0
posix_timers_cache      0      0    216   37    2 : tunables    0    0    0 : slabdata      0      0      0
ip4-frags              0      0    168   24    1 : tunables    0    0    0 : slabdata      0      0      0
UDP-Lite               0      0    896   36    8 : tunables    0    0    0 : slabdata      0      0      0
flow_cache           713    858    104   39    1 : tunables    0    0    0 : slabdata     22     22      0
xfrm_dst_cache        66    216    448   36    4 : tunables    0    0    0 : slabdata      6      6      0
UDP                  288    288    896   36    8 : tunables    0    0    0 : slabdata      8      8      0
tw_sock_TCP          168    168    192   21    1 : tunables    0    0    0 : slabdata      8      8      0
TCP                  155    216   1728   18    8 : tunables    0    0    0 : slabdata     12     12      0
blkdev_queue          29    102   1832   17    8 : tunables    0    0    0 : slabdata      6      6      0
blkdev_requests      342    396    368   22    2 : tunables    0    0    0 : slabdata     18     18      0
sock_inode_cache    1046   1300    640   25    4 : tunables    0    0    0 : slabdata     52     52      0
file_lock_cache      312    312    208   39    2 : tunables    0    0    0 : slabdata      8      8      0
file_lock_ctx        249    584     56   73    1 : tunables    0    0    0 : slabdata      8      8      0
net_namespace         11     32   3904    8    8 : tunables    0    0    0 : slabdata      4      4      0
shmem_inode_cache   1951   2288    624   26    4 : tunables    0    0    0 : slabdata     88     88      0
taskstats            139    192    328   24    2 : tunables    0    0    0 : slabdata      8      8      0
proc_inode_cache    2842   3024    592   27    4 : tunables    0    0    0 : slabdata    112    112      0
sigqueue             200    200    160   25    1 : tunables    0    0    0 : slabdata      8      8      0
bdev_cache            51    234    832   39    8 : tunables    0    0    0 : slabdata      6      6      0
kernfs_node_cache  27234  27234    120   34    1 : tunables    0    0    0 : slabdata    801    801      0
mnt_cache          14316  14406    384   21    2 : tunables    0    0    0 : slabdata    686    686      0
inode_cache         7019   7170    536   30    4 : tunables    0    0    0 : slabdata    239    239      0
dentry            498951 499359    192   21    1 : tunables    0    0    0 : slabdata  23779  23779      0
buffer_head       575585 665886    104   39    1 : tunables    0    0    0 : slabdata  17074  17074      0
vm_area_struct     36069  36982    184   22    1 : tunables    0    0    0 : slabdata   1681   1681      0
mm_struct           1085   1296    896   36    8 : tunables    0    0    0 : slabdata     36     36      0
files_cache          270    375    640   25    4 : tunables    0    0    0 : slabdata     15     15      0
signal_cache         713   1020   1088   30    8 : tunables    0    0    0 : slabdata     34     34      0
sighand_cache        516    585   2112   15    8 : tunables    0    0    0 : slabdata     39     39      0
task_xstate          774   1209    832   39    8 : tunables    0    0    0 : slabdata     31     31      0
task_struct          886   1050   2080   15    8 : tunables    0    0    0 : slabdata     70     70      0
Acpi-ParseExt       3678   3752     72   56    1 : tunables    0    0    0 : slabdata     67     67      0
Acpi-State           408    408     80   51    1 : tunables    0    0    0 : slabdata      8      8      0
Acpi-Namespace      4794   4794     40  102    1 : tunables    0    0    0 : slabdata     47     47      0
anon_vma           17191  17748     80   51    1 : tunables    0    0    0 : slabdata    348    348      0
numa_policy         1360   1360     24  170    1 : tunables    0    0    0 : slabdata      8      8      0
radix_tree_node   165502 165676    584   28    4 : tunables    0    0    0 : slabdata   5917   5917      0
ftrace_event_file    933   1058     88   46    1 : tunables    0    0    0 : slabdata     23     23      0
ftrace_event_field   2295   2295     48   85    1 : tunables    0    0    0 : slabdata     27     27      0
idr_layer_cache      510    510   2096   15    8 : tunables    0    0    0 : slabdata     34     34      0
dma-kmalloc-8192       0      0   8192    4    8 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-4096       0      0   4096    8    8 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-2048       0      0   2048   16    8 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-1024       0      0   1024   32    8 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-512       32    128    512   32    4 : tunables    0    0    0 : slabdata      4      4      0
dma-kmalloc-256        0      0    256   32    2 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-128        0      0    128   32    1 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-64         0      0     64   64    1 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-32         0      0     32  128    1 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-16         0      0     16  256    1 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-8          0      0      8  512    1 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-192        0      0    192   21    1 : tunables    0    0    0 : slabdata      0      0      0
dma-kmalloc-96         0      0     96   42    1 : tunables    0    0    0 : slabdata      0      0      0
kmalloc-8192         229    240   8192    4    8 : tunables    0    0    0 : slabdata     60     60      0
kmalloc-4096         468    616   4096    8    8 : tunables    0    0    0 : slabdata     77     77      0
kmalloc-2048         705    896   2048   16    8 : tunables    0    0    0 : slabdata     56     56      0
kmalloc-1024        1944   2368   1024   32    8 : tunables    0    0    0 : slabdata     74     74      0
kmalloc-512         2660   3136    512   32    4 : tunables    0    0    0 : slabdata     98     98      0
kmalloc-256         9005  10560    256   32    2 : tunables    0    0    0 : slabdata    330    330      0
kmalloc-192         7725   8841    192   21    1 : tunables    0    0    0 : slabdata    421    421      0
kmalloc-128        26491  26816    128   32    1 : tunables    0    0    0 : slabdata    838    838      0
kmalloc-96          7350   7350     96   42    1 : tunables    0    0    0 : slabdata    175    175      0
kmalloc-64         65055  66624     64   64    1 : tunables    0    0    0 : slabdata   1041   1041      0
kmalloc-32         37000  37248     32  128    1 : tunables    0    0    0 : slabdata    291    291      0
kmalloc-16         15872  15872     16  256    1 : tunables    0    0    0 : slabdata     62     62      0
kmalloc-8           7168   7168      8  512    1 : tunables    0    0    0 : slabdata     14     14      0
kmem_cache_node      157    448     64   64    1 : tunables    0    0    0 : slabdata      7      7      0
kmem_cache           125    256    256   32    2 : tunables    0    0    0 : slabdata      8      8      0
`

func TestSlabinfoGather(t *testing.T) {
	tf, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tf.Name())
	defer func(p string) { slabinfoPath = p }(slabinfoPath)
	slabinfoPath = tf.Name()

	_, err = tf.Write([]byte(slabinfoSample))
	require.NoError(t, err)

	si := Slabinfo{}

	var acc testutil.Accumulator
	require.NoError(t, si.Gather(&acc))

	// just spot check a few of them
	tags := map[string]string{
		"name": "isofs_inode_cache",
	}
	fields := map[string]interface{}{
		"active_objs":  uint64(0),
		"num_objs":     uint64(140),
		"objsize":      uint64(584),
		"objperslab":   uint64(28),
		"pagesperslab": uint64(4),
		"limit":        uint64(0),
		"batchcount":   uint64(0),
		"sharedfactor": uint64(0),
		"active_slabs": uint64(5),
		"num_slabs":    uint64(5),
		"sharedavail":  uint64(0),
	}
	acc.AssertContainsTaggedFields(t, "slabinfo", fields, tags)

	tags = map[string]string{
		"name": "dentry",
	}
	fields = map[string]interface{}{
		"active_objs":  uint64(498951),
		"num_objs":     uint64(499359),
		"objsize":      uint64(192),
		"objperslab":   uint64(21),
		"pagesperslab": uint64(1),
		"limit":        uint64(0),
		"batchcount":   uint64(0),
		"sharedfactor": uint64(0),
		"active_slabs": uint64(23779),
		"num_slabs":    uint64(23779),
		"sharedavail":  uint64(0),
	}
	acc.AssertContainsTaggedFields(t, "slabinfo", fields, tags)
}
