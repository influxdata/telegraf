package knx_listener

import (
	"github.com/vapourismo/knx-go/knx"
)

type knxDummyInterface struct {
	inbound chan knx.GroupEvent
}

func newDummyInterface() knxDummyInterface {
	di := knxDummyInterface{}
	di.inbound = make(chan knx.GroupEvent)

	return di
}

func (di *knxDummyInterface) Send(event knx.GroupEvent) {
	di.inbound <- event
}

func (di *knxDummyInterface) Inbound() <-chan knx.GroupEvent {
	return di.inbound
}

func (di *knxDummyInterface) Close() {
	close(di.inbound)
}
