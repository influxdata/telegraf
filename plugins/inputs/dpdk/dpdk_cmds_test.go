//go:build linux

package dpdk

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func Test_LinkStatusCommand(t *testing.T) {
	t.Run("when 'status' field is DOWN then return 'link_status'=0", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q:{%q: "DOWN"}}`, ethdevLinkStatusCommand, linkStatusStringFieldName)
		simulateResponse(mockConn, response, nil)
		dpdkConn := dpdk.connectors[0]
		dpdkConn.processCommand(mockAcc, testutil.Logger{}, fmt.Sprintf("%s,1", ethdevLinkStatusCommand), nil)

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": ethdevLinkStatusCommand,
					"params":  "1",
				},
				map[string]interface{}{
					linkStatusStringFieldName:  "DOWN",
					linkStatusIntegerFieldName: int64(0),
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})

	t.Run("when 'status' field is UP then return 'link_status'=1", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q:{%q: "UP"}}`, ethdevLinkStatusCommand, linkStatusStringFieldName)
		simulateResponse(mockConn, response, nil)
		dpdkConn := dpdk.connectors[0]
		dpdkConn.processCommand(mockAcc, testutil.Logger{}, fmt.Sprintf("%s,1", ethdevLinkStatusCommand), nil)

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": ethdevLinkStatusCommand,
					"params":  "1",
				},
				map[string]interface{}{
					linkStatusStringFieldName:  "UP",
					linkStatusIntegerFieldName: int64(1),
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})

	t.Run("when link status output doesn't have any fields then don't return 'link_status' field", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q:{}}`, ethdevLinkStatusCommand)
		simulateResponse(mockConn, response, nil)
		dpdkConn := dpdk.connectors[0]
		dpdkConn.processCommand(mockAcc, testutil.Logger{}, fmt.Sprintf("%s,1", ethdevLinkStatusCommand), nil)

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, nil, actual, testutil.IgnoreTime())
	})

	t.Run("when link status output doesn't have status field then don't return 'link_status' field", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q:{"tag1": 1}}`, ethdevLinkStatusCommand)
		simulateResponse(mockConn, response, nil)
		dpdkConn := dpdk.connectors[0]
		dpdkConn.processCommand(mockAcc, testutil.Logger{}, fmt.Sprintf("%s,1", ethdevLinkStatusCommand), nil)
		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": ethdevLinkStatusCommand,
					"params":  "1",
				},
				map[string]interface{}{
					"tag1": float64(1),
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})

	t.Run("when link status output is invalid then don't return 'link_status' field", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q:{%q: "BOB"}}`, ethdevLinkStatusCommand, linkStatusStringFieldName)
		simulateResponse(mockConn, response, nil)
		dpdkConn := dpdk.connectors[0]
		dpdkConn.processCommand(mockAcc, testutil.Logger{}, fmt.Sprintf("%s,1", ethdevLinkStatusCommand), nil)

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": ethdevLinkStatusCommand,
					"params":  "1",
				},
				map[string]interface{}{
					linkStatusStringFieldName: "BOB",
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})
}
