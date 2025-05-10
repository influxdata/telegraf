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

// Send simulates sending a GroupEvent over the KNX interface.
func (di *knxDummyInterface) Send(event knx.GroupEvent) {
	di.inbound <- event
}

// Inbound returns a read-only channel for receiving GroupEvents.
func (di *knxDummyInterface) Inbound() <-chan knx.GroupEvent {
	return di.inbound
}

// Close closes the inbound channel to simulate shutting down the interface.
func (di *knxDummyInterface) Close() {
	close(di.inbound)
}
