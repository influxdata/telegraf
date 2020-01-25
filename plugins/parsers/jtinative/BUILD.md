### Build info

```
cd to root telegraf directory
export GOPATH=${PWD}
GIT_TAG="v1.1.0"
go get -d -u github.com/golang/protobuf/protoc-gen-go
git -C "$(go env GOPATH)"/src/github.com/golang/protobuf checkout $GIT_TAG
go install github.com/golang/protobuf/protoc-gen-go
PATH=${PWD}/bin:$PATH

add the following line to telemetry_top proto file
option go_package = "github.com/influxdata/telegraf/plugins/parsers/jtinative/telemetry_top"

add the following line to pbj proto file
option go_package = "github.com/influxdata/telegraf/plugins/parsers/jtinative/pbj"

protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/agent/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/agentd/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/ancpd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/authd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-smgd_ancp_stats_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-smgd_pppoe_stats_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-smgd_rsmon_debug_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-smgd_rsmon_stats_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-smgd_smd_queue_stats_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-smgd_sub_mgmt_network_stats_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/bbe-statsd-telemetry_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/chassisd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/cmerror/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/cmerror_data/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/cpu_memory_utilization/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/dcd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/eventd/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/fabric/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/firewall/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/inline_jflow/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/ipsec_telemetry/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/jdhcpd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/jkhmd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/jl2tpd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/jpppd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/junos-xmlproxyd_junos-rsvp-interface/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/junos-xmlproxyd_junos-rtg-task-memory/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/kernel-ifstate-render/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/kmd_render/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/l2ald_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/l2ald_oc_intf/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/l2cpd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/lacpd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/logical_port/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/lsp_stats/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/mib2d_arp_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/mib2d_nd6_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/mib2d_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/npu_memory_utilization/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/npu_utilization/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/optics/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/packet_stats/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pbj/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_ifl_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_mpls_sr_egress_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_mpls_sr_ingress_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_mpls_sr_sid_egress_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_mpls_sr_sid_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_mpls_sr_te_bsid_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_mpls_sr_te_ip_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_npu_resource/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfe_port_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/pfed_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/port/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/port_exp/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rmopd_render/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_bgp_rib_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_ipv6_ra_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_isis_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_loc_rib_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_ni_bgp_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_rsvp_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/rpd_te_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/session_telemetry/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/smid_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/sr_stats_per_if_egress/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/sr_stats_per_if_ingress/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/sr_stats_per_sid/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/svcset_telemetry/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/telemetry_top/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/vrrpd_oc/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ plugins/parsers/jtinative/xmlproxyd_show_local_interface_OC/*.proto
protoc -I. --go_out=:. --proto_path=${PWD}/plugins/parsers/jtinative/telemetry_top/ --proto_path=${PWD}/plugins/parsers/jtinative/pbj/ plugins/parsers/jtinative/qmon/*.proto

```
