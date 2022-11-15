package knx_listener

import (
	"github.com/vapourismo/knx-go/knx"
)

type KNXDummyInterface struct {
	inbound chan knx.GroupEvent
}

func NewDummyInterface() KNXDummyInterface {
	di := KNXDummyInterface{}
	di.inbound = make(chan knx.GroupEvent)

	return di
}

func (di *KNXDummyInterface) Send(event knx.GroupEvent) {
	di.inbound <- event
}

func (di *KNXDummyInterface) Inbound() <-chan knx.GroupEvent {
	return di.inbound
}

func (di *KNXDummyInterface) Close() {
	close(di.inbound)
}
