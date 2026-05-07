package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
)

const defaultBrowseBatchSize = 50

// browseClient is the subset of *opcua.Client used by AddressSpaceBrowser.
// Defined as an interface so tests can supply a fake without a live server.
type browseClient interface {
	Browse(ctx context.Context, req *ua.BrowseRequest) (*ua.BrowseResponse, error)
	BrowseNext(ctx context.Context, req *ua.BrowseNextRequest) (*ua.BrowseNextResponse, error)
}

// Compile-time check that *opcua.Client satisfies browseClient.
var _ browseClient = (*opcua.Client)(nil)

// BrowsedNode is a single node discovered from the address space. PathSegments
// holds the browse-path segments from the browse root to this node, exclusive
// of the root itself.
type BrowsedNode struct {
	NodeID       *ua.NodeID
	BrowseName   string
	DisplayName  string
	NodeClass    ua.NodeClass
	PathSegments []string
}

// AddressSpaceBrowser walks an OPC UA server's address space using the
// Browse service. Traversal is breadth-first with cycle detection;
// hierarchical references are followed forward, and only Object and
// ObjectType nodes are descended into. Variable and other terminal classes
// are recorded but not expanded.
//
// MaxDepth caps the number of levels descended below the root (0 = unlimited).
// MaxNodes caps total discovered nodes (0 = unlimited); when reached,
// browsing stops and the partial result is returned. BatchSize controls how
// many nodes are browsed per request (0 falls back to defaultBrowseBatchSize).
type AddressSpaceBrowser struct {
	Client    browseClient
	Log       telegraf.Logger
	MaxDepth  int
	MaxNodes  int
	BatchSize int
}

// Browse walks the address space starting from rootID and returns the
// discovered descendants. The root itself is not included in the result.
func (b *AddressSpaceBrowser) Browse(ctx context.Context, rootID *ua.NodeID) ([]*BrowsedNode, error) {
	batchSize := b.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBrowseBatchSize
	}

	var nodes []*BrowsedNode
	visited := map[string]struct{}{rootID.String(): {}}

	type queueItem struct {
		nodeID *ua.NodeID
		path   []string
		depth  int
	}
	queue := []queueItem{{nodeID: rootID}}

	for len(queue) > 0 {
		var batch []queueItem
		for len(queue) > 0 && len(batch) < batchSize {
			item := queue[0]
			queue = queue[1:]
			if b.MaxDepth > 0 && item.depth >= b.MaxDepth {
				continue
			}
			batch = append(batch, item)
		}
		if len(batch) == 0 {
			continue
		}

		descs := make([]*ua.BrowseDescription, len(batch))
		for i, item := range batch {
			descs[i] = &ua.BrowseDescription{
				NodeID:          item.nodeID,
				BrowseDirection: ua.BrowseDirectionForward,
				ReferenceTypeID: ua.NewNumericNodeID(0, id.HierarchicalReferences),
				IncludeSubtypes: true,
				NodeClassMask:   uint32(ua.NodeClassAll),
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			}
		}

		resp, err := b.Client.Browse(ctx, &ua.BrowseRequest{NodesToBrowse: descs})
		if err != nil {
			return nil, fmt.Errorf("browse request failed: %w", err)
		}

		for i, result := range resp.Results {
			if result.StatusCode != ua.StatusOK {
				b.Log.Debugf("Browse failed for %s: %v", batch[i].nodeID, result.StatusCode)
				continue
			}

			refs := result.References
			cont := result.ContinuationPoint
			for len(cont) > 0 {
				next, err := b.Client.BrowseNext(ctx, &ua.BrowseNextRequest{
					ContinuationPoints: [][]byte{cont},
				})
				if err != nil {
					return nil, fmt.Errorf("browse-next request failed: %w", err)
				}
				if len(next.Results) == 0 {
					break
				}
				nextResult := next.Results[0]
				if nextResult.StatusCode != ua.StatusOK {
					b.Log.Debugf("Browse-next failed for %s: %v", batch[i].nodeID, nextResult.StatusCode)
					break
				}
				refs = append(refs, nextResult.References...)
				cont = nextResult.ContinuationPoint
			}

			for _, ref := range refs {
				key := ref.NodeID.String()
				if _, seen := visited[key]; seen {
					continue
				}
				visited[key] = struct{}{}

				childPath := make([]string, len(batch[i].path)+1)
				copy(childPath, batch[i].path)
				childPath[len(childPath)-1] = ref.BrowseName.Name

				childID := ua.NewNodeIDFromExpandedNodeID(ref.NodeID)
				nodes = append(nodes, &BrowsedNode{
					NodeID:       childID,
					BrowseName:   ref.BrowseName.Name,
					DisplayName:  ref.DisplayName.Text,
					NodeClass:    ref.NodeClass,
					PathSegments: childPath,
				})

				if b.MaxNodes > 0 && len(nodes) >= b.MaxNodes {
					b.Log.Warnf("Reached max_nodes limit (%d), stopping browse", b.MaxNodes)
					return nodes, nil
				}

				if ref.NodeClass == ua.NodeClassObject || ref.NodeClass == ua.NodeClassObjectType {
					queue = append(queue, queueItem{
						nodeID: childID,
						path:   childPath,
						depth:  batch[i].depth + 1,
					})
				}
			}
		}
	}

	return nodes, nil
}
