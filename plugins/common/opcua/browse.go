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

// BrowsedNode is a single node discovered from the address space. Path is the
// slash-joined browse path from the browse root to this node, exclusive of the
// root itself, suitable for matching against a filter.Filter compiled with "/"
// as the separator.
type BrowsedNode struct {
	NodeID      *ua.NodeID
	BrowseName  string
	DisplayName string
	NodeClass   ua.NodeClass
	Path        string
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
	Client    *opcua.Client
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
	visited := map[string]bool{rootID.String(): true}

	type queueItem struct {
		nodeID *ua.NodeID
		path   string
		depth  int
	}
	queue := []queueItem{{nodeID: rootID}}

	for len(queue) > 0 {
		// Consume up to batchSize items from the queue, skipping items past
		// MaxDepth in place, and build the per-item BrowseDescription in the
		// same pass.
		batch := make([]queueItem, 0, batchSize)
		descs := make([]*ua.BrowseDescription, 0, batchSize)
		consumed := 0
		for ; consumed < len(queue) && len(batch) < batchSize; consumed++ {
			item := queue[consumed]
			if b.MaxDepth > 0 && item.depth >= b.MaxDepth {
				continue
			}
			batch = append(batch, item)
			descs = append(descs, &ua.BrowseDescription{
				NodeID:          item.nodeID,
				BrowseDirection: ua.BrowseDirectionForward,
				ReferenceTypeID: ua.NewNumericNodeID(0, id.HierarchicalReferences),
				IncludeSubtypes: true,
				NodeClassMask:   uint32(ua.NodeClassAll),
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			})
		}
		queue = queue[consumed:]
		if len(batch) == 0 {
			continue
		}

		// Batched Browse: one RPC carries up to batchSize BrowseDescriptions.
		// gopcua's Node.References sends a single BrowseDescription per call, so we
		// use the raw Browse/BrowseNext API to keep large address-space walks to
		// O(N/batchSize) round trips instead of O(N).
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
			// Drain continuation points: BrowseNext returns the *next* page of
			// references for this BrowseDescription, not a fresh first page,
			// so we append onto refs rather than replacing it.
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
				// One continuation point in, one result out.
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
				if visited[key] {
					continue
				}
				visited[key] = true

				childPath := ref.BrowseName.Name
				if batch[i].path != "" {
					childPath = batch[i].path + "/" + ref.BrowseName.Name
				}

				childID := ua.NewNodeIDFromExpandedNodeID(ref.NodeID)
				nodes = append(nodes, &BrowsedNode{
					NodeID:      childID,
					BrowseName:  ref.BrowseName.Name,
					DisplayName: ref.DisplayName.Text,
					NodeClass:   ref.NodeClass,
					Path:        childPath,
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
