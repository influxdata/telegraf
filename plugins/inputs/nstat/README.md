# Kernel Network Statistics Input Plugin

This plugin collects network metrics from `/proc/net/netstat`, `/proc/net/snmp`
and `/proc/net/snmp6` files

⭐ Telegraf v0.13.1
🏷️ network, system
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Collect kernel snmp counters and network interface statistics
[[inputs.nstat]]
  ## file paths for proc files. If empty default paths will be used:
  ##    /proc/net/netstat, /proc/net/snmp, /proc/net/snmp6
  ## These can also be overridden with env variables, see README.
  proc_net_netstat = "/proc/net/netstat"
  proc_net_snmp = "/proc/net/snmp"
  proc_net_snmp6 = "/proc/net/snmp6"
  ## dump metrics with 0 values too
  dump_zeros       = true
```

The plugin firstly tries to read file paths from config values if it is empty,
then it reads from env variables.

* `PROC_NET_NETSTAT`
* `PROC_NET_SNMP`
* `PROC_NET_SNMP6`

If these variables are also not set,
then it tries to read the proc root from env - `PROC_ROOT`,
and sets `/proc` as a root path if `PROC_ROOT` is also empty.

Then appends default file paths:

* `/net/netstat`
* `/net/snmp`
* `/net/snmp6`

So if nothing is given, no paths in config and in env vars, the plugin takes the
default paths.

* `/proc/net/netstat`
* `/proc/net/snmp`
* `/proc/net/snmp6`

In case that `proc_net_snmp6` path doesn't exist (e.g. IPv6 is not enabled) no
error would be raised.

## Metrics

* nstat
  * Icmp6InCsumErrors
  * Icmp6InDestUnreachs
  * Icmp6InEchoReplies
  * Icmp6InEchos
  * Icmp6InErrors
  * Icmp6InGroupMembQueries
  * Icmp6InGroupMembReductions
  * Icmp6InGroupMembResponses
  * Icmp6InMLDv2Reports
  * Icmp6InMsgs
  * Icmp6InNeighborAdvertisements
  * Icmp6InNeighborSolicits
  * Icmp6InParmProblems
  * Icmp6InPktTooBigs
  * Icmp6InRedirects
  * Icmp6InRouterAdvertisements
  * Icmp6InRouterSolicits
  * Icmp6InTimeExcds
  * Icmp6OutDestUnreachs
  * Icmp6OutEchoReplies
  * Icmp6OutEchos
  * Icmp6OutErrors
  * Icmp6OutGroupMembQueries
  * Icmp6OutGroupMembReductions
  * Icmp6OutGroupMembResponses
  * Icmp6OutMLDv2Reports
  * Icmp6OutMsgs
  * Icmp6OutNeighborAdvertisements
  * Icmp6OutNeighborSolicits
  * Icmp6OutParmProblems
  * Icmp6OutPktTooBigs
  * Icmp6OutRedirects
  * Icmp6OutRouterAdvertisements
  * Icmp6OutRouterSolicits
  * Icmp6OutTimeExcds
  * Icmp6OutType133
  * Icmp6OutType135
  * Icmp6OutType143
  * IcmpInAddrMaskReps
  * IcmpInAddrMasks
  * IcmpInCsumErrors
  * IcmpInDestUnreachs
  * IcmpInEchoReps
  * IcmpInEchos
  * IcmpInErrors
  * IcmpInMsgs
  * IcmpInParmProbs
  * IcmpInRedirects
  * IcmpInSrcQuenchs
  * IcmpInTimeExcds
  * IcmpInTimestampReps
  * IcmpInTimestamps
  * IcmpMsgInType3
  * IcmpMsgOutType3
  * IcmpOutAddrMaskReps
  * IcmpOutAddrMasks
  * IcmpOutDestUnreachs
  * IcmpOutEchoReps
  * IcmpOutEchos
  * IcmpOutErrors
  * IcmpOutMsgs
  * IcmpOutParmProbs
  * IcmpOutRedirects
  * IcmpOutSrcQuenchs
  * IcmpOutTimeExcds
  * IcmpOutTimestampReps
  * IcmpOutTimestamps
  * Ip6FragCreates
  * Ip6FragFails
  * Ip6FragOKs
  * Ip6InAddrErrors
  * Ip6InBcastOctets
  * Ip6InCEPkts
  * Ip6InDelivers
  * Ip6InDiscards
  * Ip6InECT0Pkts
  * Ip6InECT1Pkts
  * Ip6InHdrErrors
  * Ip6InMcastOctets
  * Ip6InMcastPkts
  * Ip6InNoECTPkts
  * Ip6InNoRoutes
  * Ip6InOctets
  * Ip6InReceives
  * Ip6InTooBigErrors
  * Ip6InTruncatedPkts
  * Ip6InUnknownProtos
  * Ip6OutBcastOctets
  * Ip6OutDiscards
  * Ip6OutForwDatagrams
  * Ip6OutMcastOctets
  * Ip6OutMcastPkts
  * Ip6OutNoRoutes
  * Ip6OutOctets
  * Ip6OutRequests
  * Ip6ReasmFails
  * Ip6ReasmOKs
  * Ip6ReasmReqds
  * Ip6ReasmTimeout
  * IpDefaultTTL
  * IpExtInBcastOctets
  * IpExtInBcastPkts
  * IpExtInCEPkts
  * IpExtInCsumErrors
  * IpExtInECT0Pkts
  * IpExtInECT1Pkts
  * IpExtInMcastOctets
  * IpExtInMcastPkts
  * IpExtInNoECTPkts
  * IpExtInNoRoutes
  * IpExtInOctets
  * IpExtInTruncatedPkts
  * IpExtOutBcastOctets
  * IpExtOutBcastPkts
  * IpExtOutMcastOctets
  * IpExtOutMcastPkts
  * IpExtOutOctets
  * IpForwDatagrams
  * IpForwarding
  * IpFragCreates
  * IpFragFails
  * IpFragOKs
  * IpInAddrErrors
  * IpInDelivers
  * IpInDiscards
  * IpInHdrErrors
  * IpInReceives
  * IpInUnknownProtos
  * IpOutDiscards
  * IpOutNoRoutes
  * IpOutRequests
  * IpReasmFails
  * IpReasmOKs
  * IpReasmReqds
  * IpReasmTimeout
  * TcpActiveOpens
  * TcpAttemptFails
  * TcpCurrEstab
  * TcpEstabResets
  * TcpExtArpFilter
  * TcpExtBusyPollRxPackets
  * TcpExtDelayedACKLocked
  * TcpExtDelayedACKLost
  * TcpExtDelayedACKs
  * TcpExtEmbryonicRsts
  * TcpExtIPReversePathFilter
  * TcpExtListenDrops
  * TcpExtListenOverflows
  * TcpExtLockDroppedIcmps
  * TcpExtOfoPruned
  * TcpExtOutOfWindowIcmps
  * TcpExtPAWSActive
  * TcpExtPAWSEstab
  * TcpExtPAWSPassive
  * TcpExtPruneCalled
  * TcpExtRcvPruned
  * TcpExtSyncookiesFailed
  * TcpExtSyncookiesRecv
  * TcpExtSyncookiesSent
  * TcpExtTCPACKSkippedChallenge
  * TcpExtTCPACKSkippedFinWait2
  * TcpExtTCPACKSkippedPAWS
  * TcpExtTCPACKSkippedSeq
  * TcpExtTCPACKSkippedSynRecv
  * TcpExtTCPACKSkippedTimeWait
  * TcpExtTCPAbortFailed
  * TcpExtTCPAbortOnClose
  * TcpExtTCPAbortOnData
  * TcpExtTCPAbortOnLinger
  * TcpExtTCPAbortOnMemory
  * TcpExtTCPAbortOnTimeout
  * TcpExtTCPAutoCorking
  * TcpExtTCPBacklogDrop
  * TcpExtTCPChallengeACK
  * TcpExtTCPDSACKIgnoredNoUndo
  * TcpExtTCPDSACKIgnoredOld
  * TcpExtTCPDSACKOfoRecv
  * TcpExtTCPDSACKOfoSent
  * TcpExtTCPDSACKOldSent
  * TcpExtTCPDSACKRecv
  * TcpExtTCPDSACKUndo
  * TcpExtTCPDeferAcceptDrop
  * TcpExtTCPDirectCopyFromBacklog
  * TcpExtTCPDirectCopyFromPrequeue
  * TcpExtTCPFACKReorder
  * TcpExtTCPFastOpenActive
  * TcpExtTCPFastOpenActiveFail
  * TcpExtTCPFastOpenCookieReqd
  * TcpExtTCPFastOpenListenOverflow
  * TcpExtTCPFastOpenPassive
  * TcpExtTCPFastOpenPassiveFail
  * TcpExtTCPFastRetrans
  * TcpExtTCPForwardRetrans
  * TcpExtTCPFromZeroWindowAdv
  * TcpExtTCPFullUndo
  * TcpExtTCPHPAcks
  * TcpExtTCPHPHits
  * TcpExtTCPHPHitsToUser
  * TcpExtTCPHystartDelayCwnd
  * TcpExtTCPHystartDelayDetect
  * TcpExtTCPHystartTrainCwnd
  * TcpExtTCPHystartTrainDetect
  * TcpExtTCPKeepAlive
  * TcpExtTCPLossFailures
  * TcpExtTCPLossProbeRecovery
  * TcpExtTCPLossProbes
  * TcpExtTCPLossUndo
  * TcpExtTCPLostRetransmit
  * TcpExtTCPMD5NotFound
  * TcpExtTCPMD5Unexpected
  * TcpExtTCPMTUPFail
  * TcpExtTCPMTUPSuccess
  * TcpExtTCPMemoryPressures
  * TcpExtTCPMinTTLDrop
  * TcpExtTCPOFODrop
  * TcpExtTCPOFOMerge
  * TcpExtTCPOFOQueue
  * TcpExtTCPOrigDataSent
  * TcpExtTCPPartialUndo
  * TcpExtTCPPrequeueDropped
  * TcpExtTCPPrequeued
  * TcpExtTCPPureAcks
  * TcpExtTCPRcvCoalesce
  * TcpExtTCPRcvCollapsed
  * TcpExtTCPRenoFailures
  * TcpExtTCPRenoRecovery
  * TcpExtTCPRenoRecoveryFail
  * TcpExtTCPRenoReorder
  * TcpExtTCPReqQFullDoCookies
  * TcpExtTCPReqQFullDrop
  * TcpExtTCPRetransFail
  * TcpExtTCPSACKDiscard
  * TcpExtTCPSACKReneging
  * TcpExtTCPSACKReorder
  * TcpExtTCPSYNChallenge
  * TcpExtTCPSackFailures
  * TcpExtTCPSackMerged
  * TcpExtTCPSackRecovery
  * TcpExtTCPSackRecoveryFail
  * TcpExtTCPSackShiftFallback
  * TcpExtTCPSackShifted
  * TcpExtTCPSchedulerFailed
  * TcpExtTCPSlowStartRetrans
  * TcpExtTCPSpuriousRTOs
  * TcpExtTCPSpuriousRtxHostQueues
  * TcpExtTCPSynRetrans
  * TcpExtTCPTSReorder
  * TcpExtTCPTimeWaitOverflow
  * TcpExtTCPTimeouts
  * TcpExtTCPToZeroWindowAdv
  * TcpExtTCPWantZeroWindowAdv
  * TcpExtTCPWinProbe
  * TcpExtTW
  * TcpExtTWKilled
  * TcpExtTWRecycled
  * TcpInCsumErrors
  * TcpInErrs
  * TcpInSegs
  * TcpMaxConn
  * TcpOutRsts
  * TcpOutSegs
  * TcpPassiveOpens
  * TcpRetransSegs
  * TcpRtoAlgorithm
  * TcpRtoMax
  * TcpRtoMin
  * Udp6IgnoredMulti
  * Udp6InCsumErrors
  * Udp6InDatagrams
  * Udp6InErrors
  * Udp6NoPorts
  * Udp6OutDatagrams
  * Udp6RcvbufErrors
  * Udp6SndbufErrors
  * UdpIgnoredMulti
  * UdpInCsumErrors
  * UdpInDatagrams
  * UdpInErrors
  * UdpLite6InCsumErrors
  * UdpLite6InDatagrams
  * UdpLite6InErrors
  * UdpLite6NoPorts
  * UdpLite6OutDatagrams
  * UdpLite6RcvbufErrors
  * UdpLite6SndbufErrors
  * UdpLiteIgnoredMulti
  * UdpLiteInCsumErrors
  * UdpLiteInDatagrams
  * UdpLiteInErrors
  * UdpLiteNoPorts
  * UdpLiteOutDatagrams
  * UdpLiteRcvbufErrors
  * UdpLiteSndbufErrors
  * UdpNoPorts
  * UdpOutDatagrams
  * UdpRcvbufErrors
  * UdpSndbufErrors

### Tags

* All measurements have the following tags
  * host (host of the system)
  * name (the type of the metric: snmp, snmp6 or netstat)

## Example Output

```text
> nstat,host=Hugin,name=netstat IpExtInBcastOctets=2142i,IpExtInBcastPkts=1i,IpExtInCEPkts=0i,IpExtInCsumErrors=0i,IpExtInECT0Pkts=0i,IpExtInECT1Pkts=0i,IpExtInMcastOctets=2636i,IpExtInMcastPkts=14i,IpExtInNoECTPkts=234065i,IpExtInNoRoutes=2i,IpExtInOctets=135040263i,IpExtInTruncatedPkts=0i,IpExtOutBcastOctets=2162i,IpExtOutBcastPkts=2i,IpExtOutMcastOctets=3196i,IpExtOutMcastPkts=28i,IpExtOutOctets=45962238i,IpExtReasmOverlaps=0i,MPTcpExtAddAddr=0i,MPTcpExtAddAddrDrop=0i,MPTcpExtAddAddrTx=0i,MPTcpExtAddAddrTxDrop=0i,MPTcpExtBlackhole=0i,MPTcpExtDSSCorruptionFallback=0i,MPTcpExtDSSCorruptionReset=0i,MPTcpExtDSSNoMatchTCP=0i,MPTcpExtDSSNotMatching=0i,MPTcpExtDataCsumErr=0i,MPTcpExtDuplicateData=0i,MPTcpExtEchoAdd=0i,MPTcpExtEchoAddTx=0i,MPTcpExtEchoAddTxDrop=0i,MPTcpExtInfiniteMapRx=0i,MPTcpExtInfiniteMapTx=0i,MPTcpExtMPCapableACKRX=0i,MPTcpExtMPCapableEndpAttempt=0i,MPTcpExtMPCapableFallbackACK=0i,MPTcpExtMPCapableFallbackSYNACK=0i,MPTcpExtMPCapableSYNACKRX=0i,MPTcpExtMPCapableSYNRX=0i,MPTcpExtMPCapableSYNTX=0i,MPTcpExtMPCapableSYNTXDisabled=0i,MPTcpExtMPCapableSYNTXDrop=0i,MPTcpExtMPCurrEstab=0i,MPTcpExtMPFailRx=0i,MPTcpExtMPFailTx=0i,MPTcpExtMPFallbackTokenInit=0i,MPTcpExtMPFastcloseRx=0i,MPTcpExtMPFastcloseTx=0i,MPTcpExtMPJoinAckHMacFailure=0i,MPTcpExtMPJoinAckRx=0i,MPTcpExtMPJoinNoTokenFound=0i,MPTcpExtMPJoinPortAckRx=0i,MPTcpExtMPJoinPortSynAckRx=0i,MPTcpExtMPJoinPortSynRx=0i,MPTcpExtMPJoinSynAckBackupRx=0i,MPTcpExtMPJoinSynAckHMacFailure=0i,MPTcpExtMPJoinSynAckRx=0i,MPTcpExtMPJoinSynBackupRx=0i,MPTcpExtMPJoinSynRx=0i,MPTcpExtMPJoinSynTx=0i,MPTcpExtMPJoinSynTxBindErr=0i,MPTcpExtMPJoinSynTxConnectErr=0i,MPTcpExtMPJoinSynTxCreatSkErr=0i,MPTcpExtMPPrioRx=0i,MPTcpExtMPPrioTx=0i,MPTcpExtMPRstRx=0i,MPTcpExtMPRstTx=0i,MPTcpExtMPTCPRetrans=0i,MPTcpExtMismatchPortAckRx=0i,MPTcpExtMismatchPortSynRx=0i,MPTcpExtNoDSSInWindow=0i,MPTcpExtOFOMerge=0i,MPTcpExtOFOQueue=0i,MPTcpExtOFOQueueTail=0i,MPTcpExtPortAdd=0i,MPTcpExtRcvPruned=0i,MPTcpExtRcvWndConflict=0i,MPTcpExtRcvWndConflictUpdate=0i,MPTcpExtRcvWndShared=0i,MPTcpExtRmAddr=0i,MPTcpExtRmAddrDrop=0i,MPTcpExtRmAddrTx=0i,MPTcpExtRmAddrTxDrop=0i,MPTcpExtRmSubflow=0i,MPTcpExtSndWndShared=0i,MPTcpExtSubflowRecover=0i,MPTcpExtSubflowStale=0i,TcpExtArpFilter=0i,TcpExtBusyPollRxPackets=0i,TcpExtDelayedACKLocked=1i,TcpExtDelayedACKLost=123i,TcpExtDelayedACKs=1313i,TcpExtEmbryonicRsts=0i,TcpExtIPReversePathFilter=0i,TcpExtListenDrops=0i,TcpExtListenOverflows=0i,TcpExtLockDroppedIcmps=0i,TcpExtOfoPruned=0i,TcpExtOutOfWindowIcmps=0i,TcpExtPAWSActive=0i,TcpExtPAWSEstab=0i,TcpExtPFMemallocDrop=0i,TcpExtPruneCalled=0i,TcpExtRcvPruned=0i,TcpExtSyncookiesFailed=0i,TcpExtSyncookiesRecv=0i,TcpExtSyncookiesSent=0i,TcpExtTCPACKSkippedChallenge=0i,TcpExtTCPACKSkippedFinWait2=0i,TcpExtTCPACKSkippedPAWS=0i,TcpExtTCPACKSkippedSeq=1i,TcpExtTCPACKSkippedSynRecv=0i,TcpExtTCPACKSkippedTimeWait=0i,TcpExtTCPAOBad=0i,TcpExtTCPAODroppedIcmps=0i,TcpExtTCPAOGood=0i,TcpExtTCPAOKeyNotFound=0i,TcpExtTCPAORequired=0i,TcpExtTCPAbortFailed=0i,TcpExtTCPAbortOnClose=132i,TcpExtTCPAbortOnData=457i,TcpExtTCPAbortOnLinger=0i,TcpExtTCPAbortOnMemory=0i,TcpExtTCPAbortOnTimeout=0i,TcpExtTCPAckCompressed=15i,TcpExtTCPAutoCorking=1471i,TcpExtTCPBacklogCoalesce=113i,TcpExtTCPBacklogDrop=0i,TcpExtTCPChallengeACK=0i,TcpExtTCPDSACKIgnoredDubious=0i,TcpExtTCPDSACKIgnoredNoUndo=65i,TcpExtTCPDSACKIgnoredOld=0i,TcpExtTCPDSACKOfoRecv=0i,TcpExtTCPDSACKOfoSent=0i,TcpExtTCPDSACKOldSent=123i,TcpExtTCPDSACKRecv=78i,TcpExtTCPDSACKRecvSegs=78i,TcpExtTCPDSACKUndo=0i,TcpExtTCPDeferAcceptDrop=0i,TcpExtTCPDelivered=95905i,TcpExtTCPDeliveredCE=0i,TcpExtTCPFastOpenActive=0i,TcpExtTCPFastOpenActiveFail=0i,TcpExtTCPFastOpenBlackhole=0i,TcpExtTCPFastOpenCookieReqd=0i,TcpExtTCPFastOpenListenOverflow=0i,TcpExtTCPFastOpenPassive=0i,TcpExtTCPFastOpenPassiveAltKey=0i,TcpExtTCPFastOpenPassiveFail=0i,TcpExtTCPFastRetrans=3i,TcpExtTCPFromZeroWindowAdv=1i,TcpExtTCPFullUndo=2i,TcpExtTCPHPAcks=40380i,TcpExtTCPHPHits=46243i,TcpExtTCPHystartDelayCwnd=81i,TcpExtTCPHystartDelayDetect=3i,TcpExtTCPHystartTrainCwnd=0i,TcpExtTCPHystartTrainDetect=0i,TcpExtTCPKeepAlive=2816i,TcpExtTCPLossFailures=0i,TcpExtTCPLossProbeRecovery=0i,TcpExtTCPLossProbes=85i,TcpExtTCPLossUndo=0i,TcpExtTCPLostRetransmit=0i,TcpExtTCPMD5Failure=0i,TcpExtTCPMD5NotFound=0i,TcpExtTCPMD5Unexpected=0i,TcpExtTCPMTUPFail=0i,TcpExtTCPMTUPSuccess=0i,TcpExtTCPMemoryPressures=0i,TcpExtTCPMemoryPressuresChrono=0i,TcpExtTCPMigrateReqFailure=0i,TcpExtTCPMigrateReqSuccess=0i,TcpExtTCPMinTTLDrop=0i,TcpExtTCPOFODrop=0i,TcpExtTCPOFOMerge=0i,TcpExtTCPOFOQueue=509i,TcpExtTCPOrigDataSent=89707i,TcpExtTCPPLBRehash=0i,TcpExtTCPPartialUndo=1i,TcpExtTCPPureAcks=35929i,TcpExtTCPRcvCoalesce=9070i,TcpExtTCPRcvCollapsed=0i,TcpExtTCPRcvQDrop=0i,TcpExtTCPRenoFailures=0i,TcpExtTCPRenoRecovery=0i,TcpExtTCPRenoRecoveryFail=0i,TcpExtTCPRenoReorder=0i,TcpExtTCPReqQFullDoCookies=0i,TcpExtTCPReqQFullDrop=0i,TcpExtTCPRetransFail=0i,TcpExtTCPSACKDiscard=0i,TcpExtTCPSACKReneging=0i,TcpExtTCPSACKReorder=79i,TcpExtTCPSYNChallenge=0i,TcpExtTCPSackFailures=0i,TcpExtTCPSackMerged=3i,TcpExtTCPSackRecovery=3i,TcpExtTCPSackRecoveryFail=0i,TcpExtTCPSackShiftFallback=116i,TcpExtTCPSackShifted=2i,TcpExtTCPSlowStartRetrans=0i,TcpExtTCPSpuriousRTOs=0i,TcpExtTCPSpuriousRtxHostQueues=0i,TcpExtTCPSynRetrans=1i,TcpExtTCPTSReorder=1i,TcpExtTCPTimeWaitOverflow=0i,TcpExtTCPTimeouts=1i,TcpExtTCPToZeroWindowAdv=1i,TcpExtTCPWantZeroWindowAdv=4i,TcpExtTCPWinProbe=0i,TcpExtTCPWqueueTooBig=0i,TcpExtTCPZeroWindowDrop=0i,TcpExtTW=7592i,TcpExtTWKilled=0i,TcpExtTWRecycled=0i,TcpExtTcpDuplicateDataRehash=0i,TcpExtTcpTimeoutRehash=1i 1742837823000000000
2025-03-24T17:37:03Z D! [agent] Stopping service inputs
> nstat,host=Hugin,name=snmp IcmpInAddrMaskReps=0i,IcmpInAddrMasks=0i,IcmpInCsumErrors=0i,IcmpInDestUnreachs=1i,IcmpInEchoReps=0i,IcmpInEchos=0i,IcmpInErrors=0i,IcmpInMsgs=1i,IcmpInParmProbs=0i,IcmpInRedirects=0i,IcmpInSrcQuenchs=0i,IcmpInTimeExcds=0i,IcmpInTimestampReps=0i,IcmpInTimestamps=0i,IcmpMsgInType3=1i,IcmpMsgOutType3=1i,IcmpOutAddrMaskReps=0i,IcmpOutAddrMasks=0i,IcmpOutDestUnreachs=1i,IcmpOutEchoReps=0i,IcmpOutEchos=0i,IcmpOutErrors=0i,IcmpOutMsgs=1i,IcmpOutParmProbs=0i,IcmpOutRateLimitGlobal=0i,IcmpOutRateLimitHost=0i,IcmpOutRedirects=0i,IcmpOutSrcQuenchs=0i,IcmpOutTimeExcds=0i,IcmpOutTimestampReps=0i,IcmpOutTimestamps=0i,IpDefaultTTL=64i,IpForwDatagrams=12i,IpForwarding=1i,IpFragCreates=2i,IpFragFails=0i,IpFragOKs=1i,IpInAddrErrors=0i,IpInDelivers=227088i,IpInDiscards=0i,IpInHdrErrors=0i,IpInReceives=227102i,IpInUnknownProtos=0i,IpOutDiscards=0i,IpOutNoRoutes=44i,IpOutRequests=194901i,IpOutTransmits=194914i,IpReasmFails=0i,IpReasmOKs=0i,IpReasmReqds=0i,IpReasmTimeout=0i,TcpActiveOpens=12453i,TcpAttemptFails=4005i,TcpCurrEstab=9i,TcpEstabResets=2312i,TcpInCsumErrors=0i,TcpInErrs=0i,TcpInSegs=206125i,TcpMaxConn=-1i,TcpOutRsts=6928i,TcpOutSegs=189170i,TcpPassiveOpens=7628i,TcpRetransSegs=86i,TcpRtoAlgorithm=1i,TcpRtoMax=120000i,TcpRtoMin=200i,UdpIgnoredMulti=0i,UdpInCsumErrors=0i,UdpInDatagrams=27391i,UdpInErrors=0i,UdpLiteIgnoredMulti=0i,UdpLiteInCsumErrors=0i,UdpLiteInDatagrams=0i,UdpLiteInErrors=0i,UdpLiteMemErrors=0i,UdpLiteNoPorts=0i,UdpLiteOutDatagrams=0i,UdpLiteRcvbufErrors=0i,UdpLiteSndbufErrors=0i,UdpMemErrors=0i,UdpNoPorts=1i,UdpOutDatagrams=14308i,UdpRcvbufErrors=0i,UdpSndbufErrors=0i 1742837823000000000
2025-03-24T17:37:03Z D! [agent] Input channel closed
2025-03-24T17:37:03Z D! [agent] Stopped Successfully
> nstat,host=Hugin,name=snmp6 Icmp6InCsumErrors=0i,Icmp6InDestUnreachs=0i,Icmp6InEchoReplies=0i,Icmp6InEchos=0i,Icmp6InErrors=0i,Icmp6InGroupMembQueries=0i,Icmp6InGroupMembReductions=0i,Icmp6InGroupMembResponses=0i,Icmp6InMLDv2Reports=0i,Icmp6InMsgs=0i,Icmp6InNeighborAdvertisements=0i,Icmp6InNeighborSolicits=0i,Icmp6InParmProblems=0i,Icmp6InPktTooBigs=0i,Icmp6InRedirects=0i,Icmp6InRouterAdvertisements=0i,Icmp6InRouterSolicits=0i,Icmp6InTimeExcds=0i,Icmp6OutDestUnreachs=0i,Icmp6OutEchoReplies=0i,Icmp6OutEchos=0i,Icmp6OutErrors=0i,Icmp6OutGroupMembQueries=0i,Icmp6OutGroupMembReductions=0i,Icmp6OutGroupMembResponses=0i,Icmp6OutMLDv2Reports=271i,Icmp6OutMsgs=430i,Icmp6OutNeighborAdvertisements=0i,Icmp6OutNeighborSolicits=62i,Icmp6OutParmProblems=0i,Icmp6OutPktTooBigs=0i,Icmp6OutRateLimitHost=0i,Icmp6OutRedirects=0i,Icmp6OutRouterAdvertisements=0i,Icmp6OutRouterSolicits=97i,Icmp6OutTimeExcds=0i,Icmp6OutType133=97i,Icmp6OutType135=62i,Icmp6OutType143=271i,Ip6FragCreates=0i,Ip6FragFails=0i,Ip6FragOKs=0i,Ip6InAddrErrors=0i,Ip6InBcastOctets=0i,Ip6InCEPkts=0i,Ip6InDelivers=6433i,Ip6InDiscards=0i,Ip6InECT0Pkts=0i,Ip6InECT1Pkts=0i,Ip6InHdrErrors=0i,Ip6InMcastOctets=2652i,Ip6InMcastPkts=11i,Ip6InNoECTPkts=6433i,Ip6InNoRoutes=0i,Ip6InOctets=763395i,Ip6InReceives=6433i,Ip6InTooBigErrors=0i,Ip6InTruncatedPkts=0i,Ip6InUnknownProtos=0i,Ip6OutBcastOctets=0i,Ip6OutDiscards=12i,Ip6OutForwDatagrams=0i,Ip6OutMcastOctets=35016i,Ip6OutMcastPkts=453i,Ip6OutNoRoutes=4652i,Ip6OutOctets=795759i,Ip6OutRequests=6875i,Ip6OutTransmits=6875i,Ip6ReasmFails=0i,Ip6ReasmOKs=0i,Ip6ReasmReqds=0i,Ip6ReasmTimeout=0i,Udp6IgnoredMulti=0i,Udp6InCsumErrors=0i,Udp6InDatagrams=45i,Udp6InErrors=0i,Udp6MemErrors=0i,Udp6NoPorts=0i,Udp6OutDatagrams=24i,Udp6RcvbufErrors=0i,Udp6SndbufErrors=0i,UdpLite6InCsumErrors=0i,UdpLite6InDatagrams=0i,UdpLite6InErrors=0i,UdpLite6MemErrors=0i,UdpLite6NoPorts=0i,UdpLite6OutDatagrams=0i,UdpLite6RcvbufErrors=0i,UdpLite6SndbufErrors=0i 1742837823000000000
```
