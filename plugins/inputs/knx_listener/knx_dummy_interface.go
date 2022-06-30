package knx_listener

import (
	"github.com/vapourismo/knx-go/knx"
)

type KNXDummyInterface struct {
	inbound chan knx.GroupEvent
}

func NewDummyInterface() (di KNXDummyInterface, err error) {
	di, err = KNXDummyInterface{}, nil
	di.inbound = make(chan knx.GroupEvent)

	return di, err
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
