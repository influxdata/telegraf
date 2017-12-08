package test

import (
	"fmt"
	"io"
	"net"
	"text/tabwriter"

	"github.com/kentik/libkflow/api"
	"github.com/kentik/libkflow/chf"
)

func Print(out io.Writer, i int, flow chf.CHF, dev *api.Device) {
	w := tabwriter.NewWriter(out, 0, 4, 1, ' ', 0)

	fmt.Fprintf(w, "FLOW #%02d\n", i)
	fmt.Fprintf(w, "  timestampNano:\t%v\n", flow.TimestampNano())
	fmt.Fprintf(w, "  dstAs:\t%v\n", flow.DstAs())
	fmt.Fprintf(w, "  dstGeo:\t%v\n", flow.DstGeo())
	fmt.Fprintf(w, "  dstMac:\t%v\n", flow.DstMac())
	fmt.Fprintf(w, "  headerLen:\t%v\n", flow.HeaderLen())
	fmt.Fprintf(w, "  inBytes:\t%v\n", flow.InBytes())
	fmt.Fprintf(w, "  inPkts:\t%v\n", flow.InPkts())
	fmt.Fprintf(w, "  inputPort:\t%v\n", flow.InputPort())
	fmt.Fprintf(w, "  ipSize:\t%v\n", flow.IpSize())
	fmt.Fprintf(w, "  ipv4DstAddr:\t%v\n", ip(flow.Ipv4DstAddr()))
	fmt.Fprintf(w, "  ipv4SrcAddr:\t%v\n", ip(flow.Ipv4SrcAddr()))
	fmt.Fprintf(w, "  l4DstPort:\t%v\n", flow.L4DstPort())
	fmt.Fprintf(w, "  l4SrcPort:\t%v\n", flow.L4SrcPort())
	fmt.Fprintf(w, "  outputPort:\t%v\n", flow.OutputPort())
	fmt.Fprintf(w, "  protocol:\t%v\n", flow.Protocol())
	fmt.Fprintf(w, "  sampledPacketSize:\t%v\n", flow.SampledPacketSize())
	fmt.Fprintf(w, "  srcAs:\t%v\n", flow.SrcAs())
	fmt.Fprintf(w, "  srcGeo:\t%v\n", flow.SrcGeo())
	fmt.Fprintf(w, "  srcMac:\t%v\n", flow.SrcMac())
	fmt.Fprintf(w, "  tcpFlags:\t%v\n", flow.TcpFlags())
	fmt.Fprintf(w, "  tos:\t%v\n", flow.Tos())
	fmt.Fprintf(w, "  vlanIn:\t%v\n", flow.VlanIn())
	fmt.Fprintf(w, "  vlanOut:\t%v\n", flow.VlanOut())
	fmt.Fprintf(w, "  ipv4NextHop:\t%v\n", ip(flow.Ipv4NextHop()))
	fmt.Fprintf(w, "  mplsType:\t%v\n", flow.MplsType())
	fmt.Fprintf(w, "  outBytes:\t%v\n", flow.OutBytes())
	fmt.Fprintf(w, "  outPkts:\t%v\n", flow.OutPkts())
	fmt.Fprintf(w, "  tcpRetransmit:\t%v\n", flow.TcpRetransmit())
	fmt.Fprintf(w, "  srcFlowTags:\t%#v\n", str(flow.SrcFlowTags()))
	fmt.Fprintf(w, "  dstFlowTags:\t%#v\n", str(flow.DstFlowTags()))
	fmt.Fprintf(w, "  sampleRate:\t%v\n", flow.SampleRate())
	fmt.Fprintf(w, "  deviceId:\t%v\n", flow.DeviceId())
	fmt.Fprintf(w, "  flowTags:\t%#v\n", str(flow.FlowTags()))
	fmt.Fprintf(w, "  timestamp:\t%v\n", flow.Timestamp())
	fmt.Fprintf(w, "  dstBgpAsPath:\t%#v\n", str(flow.DstBgpAsPath()))
	fmt.Fprintf(w, "  dstBgpCommunity:\t%#v\n", str(flow.DstBgpCommunity()))
	fmt.Fprintf(w, "  srcBgpAsPath:\t%#v\n", str(flow.SrcBgpAsPath()))
	fmt.Fprintf(w, "  srcBgpCommunity:\t%#v\n", str(flow.SrcBgpCommunity()))
	fmt.Fprintf(w, "  srcNextHopAs:\t%v\n", flow.SrcNextHopAs())
	fmt.Fprintf(w, "  dstNextHopAs:\t%v\n", flow.DstNextHopAs())
	fmt.Fprintf(w, "  srcGeoRegion:\t%v\n", flow.SrcGeoRegion())
	fmt.Fprintf(w, "  dstGeoRegion:\t%v\n", flow.DstGeoRegion())
	fmt.Fprintf(w, "  srcGeoCity:\t%v\n", flow.SrcGeoCity())
	fmt.Fprintf(w, "  dstGeoCity:\t%v\n", flow.DstGeoCity())
	fmt.Fprintf(w, "  big:\t%v\n", flow.Big())
	fmt.Fprintf(w, "  sampleAdj:\t%v\n", flow.SampleAdj())
	fmt.Fprintf(w, "  ipv4DstNextHop:\t%v\n", ip(flow.Ipv4DstNextHop()))
	fmt.Fprintf(w, "  ipv4SrcNextHop:\t%v\n", ip(flow.Ipv4SrcNextHop()))
	fmt.Fprintf(w, "  srcRoutePrefix:\t%v\n", flow.SrcRoutePrefix())
	fmt.Fprintf(w, "  dstRoutePrefix:\t%v\n", flow.DstRoutePrefix())
	fmt.Fprintf(w, "  srcRouteLength:\t%v\n", flow.SrcRouteLength())
	fmt.Fprintf(w, "  dstRouteLength:\t%v\n", flow.DstRouteLength())
	fmt.Fprintf(w, "  srcSecondAsn:\t%v\n", flow.SrcSecondAsn())
	fmt.Fprintf(w, "  dstSecondAsn:\t%v\n", flow.DstSecondAsn())
	fmt.Fprintf(w, "  srcThirdAsn:\t%v\n", flow.SrcThirdAsn())
	fmt.Fprintf(w, "  dstThirdAsn:\t%v\n", flow.DstThirdAsn())
	fmt.Fprintf(w, "  ipv6DstAddr:\t%v\n", ip(flow.Ipv6DstAddr()))
	fmt.Fprintf(w, "  ipv6SrcAddr:\t%v\n", ip(flow.Ipv6SrcAddr()))
	fmt.Fprintf(w, "  srcEthMac:\t%v\n", mac(flow.SrcEthMac()))
	fmt.Fprintf(w, "  dstEthMac:\t%v\n", mac(flow.DstEthMac()))

	customs, _ := flow.Custom()
	fmt.Fprintf(w, "  CUSTOM FIELDS (%d)\t\n", customs.Len())

	for i := 0; i < customs.Len(); i++ {
		c := customs.At(i)
		v := c.Value()

		var name = "INVALID"
		for _, d := range dev.Customs {
			if d.ID == uint64(c.Id()) {
				name = d.Name
				break
			}
		}

		var value interface{}
		switch v.Which() {
		case chf.Custom_value_Which_strVal:
			value, _ = v.StrVal()
		case chf.Custom_value_Which_uint32Val:
			value = v.Uint32Val()
		case chf.Custom_value_Which_float32Val:
			value = v.Float32Val()
		}

		fmt.Fprintf(w, "    %s:\t%v\n", name, value)
	}

	w.Flush()
}

func ip(v interface{}, _ ...error) net.IP {
	switch v := v.(type) {
	case uint32:
		return net.IPv4(byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	case []byte:
		return net.IP(v)
	default:
		return (net.IP)(nil)
	}
}

func mac(v uint64) net.HardwareAddr {
	return net.HardwareAddr([]byte{
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
}

func str(v interface{}, _ error) interface{} {
	return v
}
