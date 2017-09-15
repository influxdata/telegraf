package oomkiller

import (
	"testing"
)

func TestIsOomkillerEvent(t *testing.T) {
	var notOomkillerEvent = []string{
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825217]  [<ffffffff811926c2>] oom_kill_process+0x202/0x3c0",
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825292] [ pid ]   uid  tgid total_vm      rss nr_ptes nr_pmds swapents oom_score_adj name",
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825382] [ 4602]     0  4602  1274525   467547    2494       8   806022             0 oom-killer-sim",
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825383] Out of memory: Kill process 4602 (oom-killer-sim) score 818 or sacrifice child",
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.826227] Killed process 4602 (oom-killer-sim) total-vm:5098100kB, anon-rss:1869024kB, file-rss:1164kB",
	}
	for _, log := range notOomkillerEvent {
		if IsOomkillerEvent(log) {
			t.Fatalf("%s should not be oom killer event", log)
		}
	}
	var IsOomkillerEvents = []string{
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825192] ospfd invoked oom-killer: gfp_mask=0x24201ca, order=0, oom_score_adj=0",
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825192] dockerd invoked oom-killer: gfp_mask=0x24201ca, order=0, oom_score_adj=0",
		"Sep 14 12:13:49 fib-r10-u23 kernel: [98991.825192] oomk-sim invoked oom-killer: gfp_mask=0x24201ca, order=0, oom_score_adj=0",
	}
	for _, log := range IsOomkillerEvents {
		if !IsOomkillerEvent(log) {
			t.Fatalf("%s should be a oom killer event", log)
		}
	}
}


