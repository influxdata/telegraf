package knx_listener

import (
	"strconv"
	"strings"
	"time"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

// Helper to abstract away router or tunnel interfaces
func sendRegularly(di *KNXDummyInterface, period time.Duration, data []knx.GroupEvent) {
	idx := 0
	for range time.Tick(period) {
		di.Send(data[idx])
		idx = (idx + 1) % len(data)
	}
}

func generateEvent(a string, d dpt.DatapointValue) knx.GroupEvent {
	parts := strings.Split(a, "/")
	addr := make([]uint8, 3)
	for i, p := range parts {
		x, err := strconv.Atoi(p)
		if err != nil {
			return knx.GroupEvent{}
		}
		addr[i] = uint8(x)
	}

	return knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(addr[0], addr[1], addr[2]),
		Data:        d.Pack(),
	}
}

func generateData() []knx.GroupEvent {
	data := make([]knx.GroupEvent, 0)

	// DPT 1.xxx
	d_1001 := dpt.DPT_1001(true)
	data = append(data, generateEvent("1/0/1", &d_1001))
	d_1002 := dpt.DPT_1002(false)
	data = append(data, generateEvent("1/0/2", &d_1002))
	d_1003 := dpt.DPT_1003(true)
	data = append(data, generateEvent("1/0/3", &d_1003))
	d_1009 := dpt.DPT_1009(false)
	data = append(data, generateEvent("1/0/9", &d_1009))
	d_1010 := dpt.DPT_1010(true)
	data = append(data, generateEvent("1/1/0", &d_1010))

	// DPT 5.xxx
	d_5001 := dpt.DPT_5001(33.33)
	data = append(data, generateEvent("5/0/1", &d_5001))
	d_5003 := dpt.DPT_5003(120.1)
	data = append(data, generateEvent("5/0/3", &d_5003))
	d_5004 := dpt.DPT_5004(25)
	data = append(data, generateEvent("5/0/4", &d_5004))

	// DPT 9.xxx
	d_9001 := dpt.DPT_9001(18.56)
	data = append(data, generateEvent("9/0/1", &d_9001))
	d_9004 := dpt.DPT_9004(243.9)
	data = append(data, generateEvent("9/0/4", &d_9004))
	d_9005 := dpt.DPT_9005(12.01)
	data = append(data, generateEvent("9/0/5", &d_9005))
	d_9007 := dpt.DPT_9007(59.32)
	data = append(data, generateEvent("9/0/7", &d_9007))

	// DPT 12.xxx
	d_12001 := dpt.DPT_12001(1234567)
	data = append(data, generateEvent("12/0/1", &d_12001))

	// DPT 13.xxx
	d_13001 := dpt.DPT_13001(13001)
	data = append(data, generateEvent("13/0/1", &d_13001))
	d_13002 := dpt.DPT_13002(13002)
	data = append(data, generateEvent("13/0/2", &d_13002))
	d_13010 := dpt.DPT_13010(130010)
	data = append(data, generateEvent("13/1/0", &d_13010))
	d_13011 := dpt.DPT_13011(130011)
	data = append(data, generateEvent("13/1/1", &d_13011))
	d_13012 := dpt.DPT_13012(130012)
	data = append(data, generateEvent("13/1/2", &d_13012))
	d_13013 := dpt.DPT_13013(130013)
	data = append(data, generateEvent("13/1/3", &d_13013))
	d_13014 := dpt.DPT_13014(130014)
	data = append(data, generateEvent("13/1/4", &d_13014))
	d_13015 := dpt.DPT_13015(130015)
	data = append(data, generateEvent("13/1/5", &d_13015))

	return data
}
