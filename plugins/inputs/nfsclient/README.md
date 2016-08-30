# Telegraf plugin: NFSCLIENT

#### Plugin arguments:
- **fullstat** bool: Collect per-operation statistics

#### Description

The NFSCLIENT plugin collects data from /proc/self/mountstats, by default it will only include a quite limited set of IO metrics. 
If fullstat is set, it will collect a lot of per-operation statistics.

#### References
[nfsiostat](http://git.linux-nfs.org/?p=steved/nfs-utils.git;a=summary)

[What is in /proc/self/mountstats for NFS mounts: an introduction](https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex)
