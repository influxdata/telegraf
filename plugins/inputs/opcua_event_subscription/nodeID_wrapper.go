package opcua_event_subscription

import (
	"fmt"
	"github.com/gopcua/opcua/ua"
)

type NodeIDWrapper struct {
	ID *ua.NodeID
}

func (n *NodeIDWrapper) UnmarshalText(text []byte) error {
	nodeID, err := ua.ParseNodeID(string(text))
	if err != nil {
		return fmt.Errorf("failed to parse NodeID from text: %w", err)
	}
	n.ID = nodeID
	return nil
}
