#!/usr/bin/env bash

set -eu

# this scripts is used when migrating v2 to v3.
# usage: cd ${GOPATH}/src/github.com/shirou/gopsutil && bash tools/v3migration/v3migration.sh



DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)"
ROOT=$(cd "${DIR}"/../.. && pwd)


## 1. refresh
cd "${ROOT}"

/bin/rm -rf v3

## 2. copy directories
# docker is removed, #464 will be fixed
mkdir -p v3
cp -rp cpu disk docker host internal load mem net process winservices v3
cp Makefile v3

# build migartion tool
go build -o v3/v3migration "${DIR}"/v3migration.go


V3DIR=$(cd "${ROOT}"/v3 && pwd)
cd "${V3DIR}"

## 3. mod
go mod init

###  change import path
find . -name "*.go" -print0 | xargs -0 -I@ sed -i 's|"github.com/shirou/gopsutil/|"github.com/shirou/gopsutil/v3/|g' @

############ Issues

# #429 process.NetIOCounters is pointless on Linux
./v3migration "$(pwd)" 429
sed -i '/NetIOCounters/d' process/process.go
sed -i "/github.com\/shirou\/gopsutil\/v3\/net/d" process/process_bsd.go


# #464 CgroupMem : fix typo and wrong file names
sed -i 's|memoryLimitInBbytes|memoryLimitInBytes|g' docker/docker.go
sed -i 's|memoryLimitInBbytes|memory.limit_in_bytes|g' docker/docker_linux.go
sed -i 's|memoryFailcnt|memory.failcnt|g' docker/docker_linux.go


# fix #346
sed -i 's/Soft     int32/Soft     uint64/' process/process.go
sed -i 's/Hard     int32/Hard     uint64/' process/process.go
sed -i 's| //TODO too small. needs to be uint64||' process/process.go
sed -i 's|limitToInt(val string) (int32, error)|limitToUint(val string) (uint64, error)|' process/process_*.go
sed -i 's|limitToInt|limitToUint|' process/process_*.go
sed -i 's|return int32(res), nil|return uint64(res), nil|' process/process_*.go
sed -i 's|math.MaxInt32|math.MaxUint64|' process/process_*.go

# fix #545
# variable names
sed -i 's|WritebackTmp|WriteBackTmp|g' mem/*.go
sed -i 's|Writeback|WriteBack|g' mem/*.go
sed -i 's|SReclaimable|Sreclaimable|g' mem/*.go
sed -i 's|SUnreclaim|Sunreclaim|g' mem/*.go
sed -i 's|VMallocTotal|VmallocTotal|g' mem/*.go
sed -i 's|VMallocUsed|VmallocUsed|g' mem/*.go
sed -i 's|VMallocChunk|VmallocChunk|g' mem/*.go

# json field name
sed -i 's|hostid|hostId|g' host/host.go
sed -i 's|hostid|hostId|g' host/host_test.go
sed -i 's|sensorTemperature|temperature|g' host/host.go
sed -i 's|sensorTemperature|temperature|g' host/host_test.go

sed -i 's|writeback|writeBack|g' mem/*.go
sed -i 's|writeBacktmp|writeBackTmp|g' mem/*.go
sed -i 's|pagetables|pageTables|g' mem/*.go
sed -i 's|swapcached|swapCached|g' mem/*.go
sed -i 's|commitlimit|commitLimit|g' mem/*.go
sed -i 's|committedas|committedAS|g' mem/*.go
sed -i 's|hightotal|highTotal|g' mem/*.go
sed -i 's|highfree|highFree|g' mem/*.go
sed -i 's|lowtotal|lowTotal|g' mem/*.go
sed -i 's|lowfree|lowFree|g' mem/*.go
sed -i 's|swaptotal|swapTotal|g' mem/*.go
sed -i 's|swapfree|swapFree|g' mem/*.go
sed -i 's|vmalloctotal|vmallocTotal|g' mem/*.go
sed -i 's|vmallocused|vmallocUsed|g' mem/*.go
sed -i 's|vmallocchunk|vmallocChunk|g' mem/*.go
sed -i 's|hugepagestotal|hugePagesTotal|g' mem/*.go
sed -i 's|hugepagesfree|hugePagesFree|g' mem/*.go
sed -i 's|hugepagesize|hugePageSize|g' mem/*.go
sed -i 's|pgin|pgIn|g' mem/*.go
sed -i 's|pgout|pgOut|g' mem/*.go
sed -i 's|pgfault|pgFault|g' mem/*.go
sed -i 's|pgmajfault|pgMajFault|g' mem/*.go

sed -i 's|hardwareaddr|hardwareAddr|g' net/*.go
sed -i 's|conntrackCount|connTrackCount|g' net/*.go
sed -i 's|conntrackMax|connTrackMax|g' net/*.go
sed -i 's|delete_list|deleteList|g' net/*.go
sed -i 's|insert_failed|insertFailed|g' net/*.go
sed -i 's|early_drop|earlyDrop|g' net/*.go
sed -i 's|expect_create|expectCreate|g' net/*.go
sed -i 's|expect_delete|expectDelete|g' net/*.go
sed -i 's|search_restart|searchRestart|g' net/*.go
sed -i 's|icmp_error|icmpError|g' net/*.go
sed -i 's|expect_new|expectNew|g' net/*.go



# fix no more public API/types/constants defined only for some platforms

sed -i 's|CTLKern|ctlKern|g' cpu/*.go
sed -i 's|CPNice|cpNice|g' cpu/*.go
sed -i 's|CPSys|cpSys|g' cpu/*.go
sed -i 's|CPIntr|cpIntr|g' cpu/*.go
sed -i 's|CPIdle|cpIdle|g' cpu/*.go
sed -i 's|CPUStates|cpUStates|g' cpu/*.go
sed -i 's|CTLKern|ctlKern|g' cpu/cpu_openbsd.go
sed -i 's|CTLHw|ctlHw|g' cpu/cpu_openbsd.go
sed -i 's|SMT|sMT|g' cpu/cpu_openbsd.go
sed -i 's|KernCptime|kernCptime|g' cpu/cpu_openbsd.go
sed -i 's|KernCptime2|kernCptime2|g' cpu/cpu_openbsd.go
sed -i 's|Win32_Processor|win32Processor|g' cpu/cpu_windows.go

sed -i 's|DEVSTAT_NO_DATA|devstat_NO_DATA|g' disk/*.go
sed -i 's|DEVSTAT_READ|devstat_READ|g' disk/*.go
sed -i 's|DEVSTAT_WRITE|devstat_WRITE|g' disk/*.go
sed -i 's|DEVSTAT_FREE|devstat_FREE|g' disk/*.go
sed -i 's|Devstat|devstat|g' disk/*.go
sed -i 's|Bintime|bintime|g' disk/*.go
sed -i 's|SectorSize|sectorSize|g' disk/disk_linux.go
sed -i 's|FileFileCompression|fileFileCompression|g' disk/disk_windows.go
sed -i 's|FileReadOnlyVolume|fileReadOnlyVolume|g' disk/disk_windows.go

sed -i 's|USER_PROCESS|user_PROCESS|g' host/host_*.go
sed -i 's|LSB|lsbStruct|g' host/host_linux*

sed -i 's| BcacheStats | bcacheStats |g' mem/*.go

sed -i 's|TCPStatuses|tcpStatuses|g' net/*.go
sed -i 's|CT_ENTRIES|ctENTRIES|g' net/net_linux.go
sed -i 's|CT_SEARCHED|ctSEARCHED|g' net/net_linux.go
sed -i 's|CT_FOUND|ctFOUND|g' net/net_linux.go
sed -i 's|CT_NEW|ctNEW|g' net/net_linux.go
sed -i 's|CT_INVALID|ctINVALID|g' net/net_linux.go
sed -i 's|CT_IGNORE|ctIGNORE|g' net/net_linux.go
sed -i 's|CT_DELETE|ctDELETE|g' net/net_linux.go
sed -i 's|CT_DELETE_LIST|ctDELETE_LIST|g' net/net_linux.go
sed -i 's|CT_INSERT|ctINSERT|g' net/net_linux.go
sed -i 's|CT_INSERT_FAILED|ctINSERT_FAILED|g' net/net_linux.go
sed -i 's|CT_DROP|ctDROP|g' net/net_linux.go
sed -i 's|CT_EARLY_DROP|ctEARLY_DROP|g' net/net_linux.go
sed -i 's|CT_ICMP_ERROR|ctICMP_ERROR|g' net/net_linux.go
sed -i 's|CT_EXPECT_NEW|ctEXPECT_NEW|g' net/net_linux.go
sed -i 's|CT_EXPECT_CREATE|ctEXPECT_CREATE|g' net/net_linux.go
sed -i 's|CT_EXPECT_DELETE|ctEXPECT_DELETE|g' net/net_linux.go
sed -i 's|CT_SEARCH_RESTART|ctSEARCH_RESTART|g' net/net_linux.go

sed -i 's|PageSize|pageSize|g' process/process_*.go
sed -i 's|PrioProcess|prioProcess|g' process/process_*.go
sed -i 's|ClockTicks|clockTicks|g' process/process_*.go


./v3migration "$(pwd)" issueRemoveUnusedValue


############ SHOULD BE FIXED BY HAND
