package sflow_a10

import (
	"encoding/hex"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSFlow_A10(t *testing.T) {
	sflow := &SFlow_A10{
		ServiceAddress:   "udp://127.0.0.1:0",
		Log:              testutil.Logger{},
		IgnoreZeroValues: true,
	}

	data, err := ioutil.ReadFile("sflow_3_2_t2.xml")
	require.NoError(t, err)

	err = sflow.initInternal([]byte(data))
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = sflow.Start(&acc)
	require.NoError(t, err)
	defer sflow.Stop()

	client, err := net.Dial(sflow.Address().Network(), sflow.Address().String())
	require.NoError(t, err)

	// send the metrics multiple times to make sure it gets parsed correctly at least once
	for i := 0; i < 5; i++ {
		// 271 - hex 10f
		packetBytes, err := hex.DecodeString("00000005000000010a41101f0000000000008a6766e234bc00000001000000020000005a00000000000210c70000000109f8a10F2468000000020004000000000A0001031EC0A805060F")
		require.NoError(t, err)
		client.Write(packetBytes)

		// 260 - hex 104
		packetBytes, err = hex.DecodeString("00000005000000010a41101f0000000000008a6766e234bc00000001000000020000005a00000000000210c70000000109f8a10400000046050c00567a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		client.Write(packetBytes)

		// 219 - hex 0db
		packetBytes, err = hex.DecodeString("00000005000000010a41101f0000000000008a9066e235c800000002000000020000026c00000000000210c70000000109f8a0db000002580000004a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000007400000000000210c70000000109f8a0d000000060000000080000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		client.Write(packetBytes)

		// 207 - hex 0cf
		packetBytes, err = hex.DecodeString("00000005000000010a01000600000000002c6d611a33ba2f00000001000000020000027c000000000000a27e0000000109f8a0cf0000026800a6004c00000000000000000292a4dc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000059b9909a0000002ce72d028c00000000000000000000000000000000000000000000000000000000022e30fc00000000efc72290000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001392de4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011c8a000000000000948d000000000001b11500000000026350560000000106f457cc000000000000000000000000000000000000000002c7c2860000000013722fbc00000000000000000000000000000000000000000000004100000000000000400000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000012550000000000003795000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		client.Write(packetBytes)

		// 293 - hex 125
		packetBytes, err = hex.DecodeString("00000005000000010a1d000800000000000007f70003ae8800000001000000020000002c00000008020001030000000109f8a12500000018000000000000001a00000000000000000000000000000000")
		require.NoError(t, err)
		client.Write(packetBytes)

		// 294 - hex 126
		packetBytes, err = hex.DecodeString("00000005000000010a1d000800000000000008ac0003d59800000001000000020000006c00000009040000010000000109f8a1260000005800000001000000060000000ba43b7400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000038000000000000000000000000000000150000000000000000")
		require.NoError(t, err)
		client.Write(packetBytes)
	}
	acc.Wait(3)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"sflow_a10",
			map[string]string{
				"agent_address": "10.29.0.8",
				"ifindex":       "1",
			},
			map[string]interface{}{
				"ifintcppkts":  uint64(56),
				"ifouttcppkts": uint64(21),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"sflow_a10",
			map[string]string{
				"agent_address": "10.29.0.8",
			},
			map[string]interface{}{
				"bw_limit_ignored": uint64(26),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"sflow_a10",
			map[string]string{
				"agent_address":  "10.65.16.31",
				"ip_address":     "10.0.1.3_192.168.5.6",
				"port_number":    "86",
				"port_range_end": "0",
				"port_type":      "SIP_UDP",
				"table_type":     "Zone",
			},
			map[string]interface{}{
				"protocol": uint64(8),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func BenchmarkSFlow(b *testing.B) {
	sflow := &SFlow_A10{
		ServiceAddress: "udp://127.0.0.1:0",
		Log:            testutil.Logger{},
	}
	err := sflow.initInternal([]byte(testXMLStringSFlow))
	require.NoError(b, err)

	var acc testutil.Accumulator
	err = sflow.Start(&acc)
	require.NoError(b, err)
	defer sflow.Stop()

	client, err := net.Dial(sflow.Address().Network(), sflow.Address().String())
	require.NoError(b, err)

	// 271 - hex 10f
	packetBytes260, err := hex.DecodeString("00000005000000010a41101f0000000000008a6766e234bc00000001000000020000005a00000000000210c70000000109f8a10F2468000000020004000000000A0001031EC0A805060F")
	require.NoError(b, err)

	// 260 - hex 104
	packetBytes271, err := hex.DecodeString("00000005000000010a41101f0000000000008a6766e234bc00000001000000020000005a00000000000210c70000000109f8a10400000046050c00567a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(b, err)

	packetBytesCounter, err := hex.DecodeString("00000005000000010a41101f0000000000008a9066e235c800000002000000020000026c00000000000210c70000000109f8a0db000002580000004a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000007400000000000210c70000000109f8a0d000000060000000080000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		client.Write(packetBytes260)
		client.Write(packetBytes271)
		client.Write(packetBytesCounter)
		acc.Wait(1)
	}
}

const testXMLStringSFlow = `
<?xml version="1.0"?>
<ctr:allctrblocks xmlns:ctr="-">
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>208</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_IP_PORT_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>20</ctr:ctrBlkSz>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PROTOCOL</ctr:enumName>
			<ctr:fieldName>Protocol</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_STATE</ctr:enumName>
			<ctr:fieldName>State</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>2</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_EXCEED_BYTE</ctr:enumName>
			<ctr:fieldName>Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>3</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_LOCKUP_PERIOD</ctr:enumName>
			<ctr:fieldName>LockU Time</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>4</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CURR_CONN</ctr:enumName>
			<ctr:fieldName>Curr Conn</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>5</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CONN_LIMIT</ctr:enumName>
			<ctr:fieldName>Conn Limit</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>6</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CURR_CONN_RATE</ctr:enumName>
			<ctr:fieldName>Curr Conn Rate</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>7</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CONN_RATE_LIMIT</ctr:enumName>
			<ctr:fieldName>Conn Rate Limit</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>8</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CURR_PKT_RATE</ctr:enumName>
			<ctr:fieldName>Curr Pkt Rate</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>9</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PKT_RATE_LIMIT</ctr:enumName>
			<ctr:fieldName>Pkt Rate Limit</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>10</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CURR_SYN_COOKIE</ctr:enumName>
			<ctr:fieldName>Curr Syn Cookie</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>11</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_SYN_COOKIE_THR</ctr:enumName>
			<ctr:fieldName>Syn Cookie Thr</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>12</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_BL_DROP_CT</ctr:enumName>
			<ctr:fieldName>Bl Drop Ct</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>13</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CONN_RATE_EXCEED_CT</ctr:enumName>
			<ctr:fieldName>Conn Rate Exceed Ct</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>14</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_PKT_RATE_EXCEED_CT</ctr:enumName>
			<ctr:fieldName>Pkt Rate Exceed Ct</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>15</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CONN_LIMIT_EXCEED_CT</ctr:enumName>
			<ctr:fieldName>Conn Limit Exceed Ct</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>16</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CURR_KBIT_RATE</ctr:enumName>
			<ctr:fieldName>Curr Kbit Rate</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>17</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_KBIT_RATE_LIMIT</ctr:enumName>
			<ctr:fieldName>Kbit Rate Limit</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>18</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_CURR_FRAG_RATE</ctr:enumName>
			<ctr:fieldName>Curr Frag Rate</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>19</ctr:offset>
			<ctr:dtype>u32</ctr:dtype>
			<ctr:enumName>DDOS_PORT_IP_T2_FRAG_RATE_LIMIT</ctr:enumName>
			<ctr:fieldName>Frag Rate Limit</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
	<ctr:counterBlock>
		<ctr:mapVersion>v2</ctr:mapVersion>
		<ctr:tag>219</ctr:tag>
		<ctr:ctrBlkSzMacroName>SFLOW_DDOS_SIP_IP_COUNTERS_V2_TOTAL_NUM</ctr:ctrBlkSzMacroName>
		<ctr:ctrBlkType>Fixed</ctr:ctrBlkType>
		<ctr:ctrBlkSz>74</ctr:ctrBlkSz>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Counter Offset</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Total Counter Num</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Reserved1</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:offsetHeader>
			<ctr:dtype>u16</ctr:dtype>
			<ctr:fieldName>Reserved2</ctr:fieldName>
		</ctr:offsetHeader>
		<ctr:counter>
			<ctr:offset>0</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_POLICY_DROP</ctr:enumName>
			<ctr:fieldName>Policy Drop</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>1</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_POLICY_VIOLATION</ctr:enumName>
			<ctr:fieldName>Policy Violation</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>2</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_IDLE_TIMEOUT</ctr:enumName>
			<ctr:fieldName>Proxy Idle Timeout</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>3</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_OFO_TIMEOUT</ctr:enumName>
			<ctr:fieldName>Proxy Out of Order Timeout</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>4</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_SEQ_CHECK_OFO</ctr:enumName>
			<ctr:fieldName>Proxy TCP Sequence Check Out of Order</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>5</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_PKTS_OFO_TOTAL</ctr:enumName>
			<ctr:fieldName>Total Out of Order Packets</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>6</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_OFO_QUEUE_SZ_EXCEED</ctr:enumName>
			<ctr:fieldName>Out of Order Queue Size Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>7</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_SEQ_CHECK_RETRANS_FIN</ctr:enumName>
			<ctr:fieldName>Proxy TCP Sequence Check Retransmitted FIN</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>8</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_SEQ_CHECK_RETRANS_RST</ctr:enumName>
			<ctr:fieldName>Proxy TCP Sequence Check Retransmitted RST</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>9</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_SEQ_CHECK_RETRANS_PUSH</ctr:enumName>
			<ctr:fieldName>Proxy TCP Sequence Check Retransmitted PSH</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>10</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_SEQ_CHECK_RETRANS_OTHER</ctr:enumName>
			<ctr:fieldName>Proxy TCP Sequence Check Retransmitted Other</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>11</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_PKTS_RETRANS_TOTAL</ctr:enumName>
			<ctr:fieldName>Total Retransmitted Packets</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>12</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_CLIENT_RST</ctr:enumName>
			<ctr:fieldName>Client TCP RST Received</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>13</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_ERR_CONDITION</ctr:enumName>
			<ctr:fieldName>Error Condition</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>14</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_ACK</ctr:enumName>
			<ctr:fieldName>Request Method ACK</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>15</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_BYE</ctr:enumName>
			<ctr:fieldName>Request Method BYE</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>16</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_CANCEL</ctr:enumName>
			<ctr:fieldName>Request Method CANCEL</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>17</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_INVITE</ctr:enumName>
			<ctr:fieldName>Request Method INVITE</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>18</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_INFO</ctr:enumName>
			<ctr:fieldName>Request Method INFO</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>19</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_MESSAGE</ctr:enumName>
			<ctr:fieldName>Request Method MESSAGE</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>20</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_NOTIFY</ctr:enumName>
			<ctr:fieldName>Request Method NOTIFY</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>21</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_OPTIONS</ctr:enumName>
			<ctr:fieldName>Request Method OPTIONS</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>22</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_PRACK</ctr:enumName>
			<ctr:fieldName>Request Method PRACK</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>23</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_PUBLISH</ctr:enumName>
			<ctr:fieldName>Request Method PUBLISH</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>24</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_REGISTER</ctr:enumName>
			<ctr:fieldName>Request Method REGISTER</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>25</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_REFER</ctr:enumName>
			<ctr:fieldName>Request Method REFER</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>26</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_SUBSCRIBE</ctr:enumName>
			<ctr:fieldName>Request Method SUBSCRIBE</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>27</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_UPDATE</ctr:enumName>
			<ctr:fieldName>Request Method UPDATE</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>28</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_METHOD_UNKNOWN</ctr:enumName>
			<ctr:fieldName>Unknown Request Method</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>29</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_UNKNOWN_VERSION</ctr:enumName>
			<ctr:fieldName>Unknown Request Version</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>30</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_KEEP_ALIVE_MSG</ctr:enumName>
			<ctr:fieldName>KeepAlive Message</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>31</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_RATE1_LIMIT_EXCEED</ctr:enumName>
			<ctr:fieldName>Dst Request Rate 1 Limit Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>32</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_RATE2_LIMIT_EXCEED</ctr:enumName>
			<ctr:fieldName>Dst Request Rate 2 Limit Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>33</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_SRC_RATE1_LIMIT_EXCEED</ctr:enumName>
			<ctr:fieldName>Src Request Rate 1 Limit Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>34</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_REQUEST_SRC_RATE2_LIMIT_EXCEED</ctr:enumName>
			<ctr:fieldName>Src Request Rate 2 Limit Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>35</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_1XX</ctr:enumName>
			<ctr:fieldName>Response Status Code 1xx</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>36</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_2XX</ctr:enumName>
			<ctr:fieldName>Response Status Code 2xx</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>37</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_3XX</ctr:enumName>
			<ctr:fieldName>Response Status Code 3xx</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>38</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_4XX</ctr:enumName>
			<ctr:fieldName>Response Status Code 4xx</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>39</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_5XX</ctr:enumName>
			<ctr:fieldName>Response Status Code 5xx</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>40</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_6XX</ctr:enumName>
			<ctr:fieldName>Response Status Code 6xx</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>41</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_UNKNOWN</ctr:enumName>
			<ctr:fieldName>Unknown Response Status Code</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>42</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_RESPONSE_UNKNOWN_VERSION</ctr:enumName>
			<ctr:fieldName>Unknown Response Version</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>43</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_READ_START_LINE_ERR</ctr:enumName>
			<ctr:fieldName>Start Line Read Erro</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>44</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_INVALID_START_LINE_ERR</ctr:enumName>
			<ctr:fieldName>Invalid Start Line</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>45</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_PARSE_START_LINE_ERR</ctr:enumName>
			<ctr:fieldName>Start Line Parse Error</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>46</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_LINE_TOO_LONG</ctr:enumName>
			<ctr:fieldName>Line Too Long</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>47</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_LINE_MEM_ALLOCATED</ctr:enumName>
			<ctr:fieldName>Line Memory Allocated</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>48</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_LINE_MEM_FREED</ctr:enumName>
			<ctr:fieldName>Line Memory Freed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>49</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_MAX_URI_LEN_EXCEED</ctr:enumName>
			<ctr:fieldName>Max URI Length Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>50</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_TOO_MANY_HEADER</ctr:enumName>
			<ctr:fieldName>Max Header Count Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>51</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_INVALID_HEADER</ctr:enumName>
			<ctr:fieldName>Invalid Header</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>52</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_NAME_TOO_LONG</ctr:enumName>
			<ctr:fieldName>Max Header Name Length Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>53</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_PARSE_HEADER_FAIL_ERR</ctr:enumName>
			<ctr:fieldName>Header Parse Fail</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>54</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_MAX_HEADER_VALUE_LEN_EXCEED</ctr:enumName>
			<ctr:fieldName>Max Header Value Length Excee</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>55</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_MAX_CALL_ID_LEN_EXCEED</ctr:enumName>
			<ctr:fieldName>Max Call ID Length Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>56</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>57</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_NOT_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter Not Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>58</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_NONE_MATCH</ctr:enumName>
			<ctr:fieldName>None Header Filter Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>59</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_ACTION_DROP</ctr:enumName>
			<ctr:fieldName>Header Filter Action Drop</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>60</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_ACTION_BLACKLIST</ctr:enumName>
			<ctr:fieldName>Header Filter Action Blacklist</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>61</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_ACTION_WHITELIST</ctr:enumName>
			<ctr:fieldName>Header Filter Action Whitelist</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>62</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_ACTION_DEFAULT_PASS</ctr:enumName>
			<ctr:fieldName>Header Filter Action Default Pass</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>63</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_FILTER1_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter 1 Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>64</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_FILTER2_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter 2 Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>65</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_FILTER3_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter 3 Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>66</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_FILTER4_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter 4 Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>67</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_HEADER_FILTER_FILTER5_MATCH</ctr:enumName>
			<ctr:fieldName>Header Filter 5 Match</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>68</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_MAX_SDP_LEN_EXCEED</ctr:enumName>
			<ctr:fieldName>Max SDP Length Exceed</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>69</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_BODY_TOO_BIG</ctr:enumName>
			<ctr:fieldName>Body Too Big</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>70</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_GET_CONTENT_FAIL_ERR</ctr:enumName>
			<ctr:fieldName>Get Content Fail</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>71</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_CONCATENATE_MSG</ctr:enumName>
			<ctr:fieldName>Concatenate Message</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>72</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_MEM_ALLOC_FAIL_ERR</ctr:enumName>
			<ctr:fieldName>Memory Allocate Fail</ctr:fieldName>
		</ctr:counter>
		<ctr:counter>
			<ctr:offset>73</ctr:offset>
			<ctr:dtype>u64</ctr:dtype>
			<ctr:enumName>DDOS_SIP_T2_MALFORM_REQUEST</ctr:enumName>
			<ctr:fieldName>Malformed Request</ctr:fieldName>
		</ctr:counter>
	</ctr:counterBlock>
</ctr:allctrblocks>`
