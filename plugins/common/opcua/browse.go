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

// BrowsedNode is a single node in the discovered address-space tree.
type BrowsedNode struct {
	NodeID       *ua.NodeID
	BrowseName   string
	DisplayName  string
	NodeClass    ua.NodeClass
	PathSegments []string
	Children     []*BrowsedNode
}

// BrowseTree is the address space discovered from a single browse run.
// AllNodes is a flat index over the discovered descendants (the root itself
// is excluded; only nodes returned by the server appear here).
type BrowseTree struct {
	Root     *BrowsedNode
	AllNodes []*BrowsedNode
}

// AddressSpaceBrowser walks an OPC UA server's address space using the
// Browse service. Traversal is breadth-first with cycle detection;
// hierarchical references are followed forward, and only Object and
// ObjectType nodes are descended into. Variable and other terminal classes
// are recorded but not expanded.
//
// MaxDepth caps tree depth (0 = unlimited). MaxNodes caps total discovered
// nodes (0 = unlimited); when reached, browsing stops and the partial tree
// is returned. BatchSize controls how many nodes are browsed per request
// (0 falls back to defaultBrowseBatchSize).
type AddressSpaceBrowser struct {
	Client    browseClient
	Log       telegraf.Logger
	MaxDepth  int
	MaxNodes  int
	BatchSize int
}

// Browse walks the address space starting from rootID and returns the
// discovered tree. The root node is included as a placeholder; only its
// descendants are populated from the server.
func (b *AddressSpaceBrowser) Browse(ctx context.Context, rootID *ua.NodeID) (*BrowseTree, error) {
	batchSize := b.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBrowseBatchSize
	}

	root := &BrowsedNode{NodeID: rootID}
	tree := &BrowseTree{Root: root}
	visited := map[string]struct{}{rootID.String(): {}}

	type queueItem struct {
		node  *BrowsedNode
		depth int
	}
	queue := []queueItem{{node: root, depth: 0}}

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
				NodeID:          item.node.NodeID,
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
				b.Log.Debugf("Browse failed for %s: %v", batch[i].node.NodeID, result.StatusCode)
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
					b.Log.Debugf("Browse-next failed for %s: %v", batch[i].node.NodeID, nextResult.StatusCode)
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

				path := make([]string, len(batch[i].node.PathSegments)+1)
				copy(path, batch[i].node.PathSegments)
				path[len(path)-1] = ref.BrowseName.Name

				child := &BrowsedNode{
					NodeID:       ua.NewNodeIDFromExpandedNodeID(ref.NodeID),
					BrowseName:   ref.BrowseName.Name,
					DisplayName:  ref.DisplayName.Text,
					NodeClass:    ref.NodeClass,
					PathSegments: path,
				}
				batch[i].node.Children = append(batch[i].node.Children, child)
				tree.AllNodes = append(tree.AllNodes, child)

				if b.MaxNodes > 0 && len(tree.AllNodes) >= b.MaxNodes {
					b.Log.Warnf("Reached max_nodes limit (%d), stopping browse", b.MaxNodes)
					return tree, nil
				}

				if ref.NodeClass == ua.NodeClassObject || ref.NodeClass == ua.NodeClassObjectType {
					queue = append(queue, queueItem{node: child, depth: batch[i].depth + 1})
				}
			}
		}
	}

	return tree, nil
}
