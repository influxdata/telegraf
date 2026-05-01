package opcua

import (
	"context"
	"errors"
	"testing"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestBrowseEmptyTree(t *testing.T) {
	fake := newFakeBrowseClient()
	rootID := ua.NewNumericNodeID(0, 85)

	tree, err := newBrowser(fake).Browse(t.Context(), rootID)
	require.NoError(t, err)
	require.NotNil(t, tree.Root)
	require.Equal(t, rootID, tree.Root.NodeID)
	require.Empty(t, tree.AllNodes)
	require.Empty(t, tree.Root.Children)
	require.Equal(t, 1, fake.browseCalls)
}

func TestBrowseSingleLevel(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=Plant1", "Plant1", ua.NodeClassObject),
		makeRef(t, "ns=2;s=ServerTime", "ServerTime", ua.NodeClassVariable),
	}

	tree, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)
	require.Len(t, tree.AllNodes, 2)
	require.Len(t, tree.Root.Children, 2)

	require.Equal(t, "Plant1", tree.AllNodes[0].BrowseName)
	require.Equal(t, []string{"Plant1"}, tree.AllNodes[0].PathSegments)
	require.Equal(t, ua.NodeClassObject, tree.AllNodes[0].NodeClass)

	require.Equal(t, "ServerTime", tree.AllNodes[1].BrowseName)
	require.Equal(t, []string{"ServerTime"}, tree.AllNodes[1].PathSegments)
	require.Equal(t, ua.NodeClassVariable, tree.AllNodes[1].NodeClass)
}

func TestBrowseDescendsOnlyContainers(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=Plant1", "Plant1", ua.NodeClassObject),
		makeRef(t, "ns=2;s=Sensor", "Sensor", ua.NodeClassVariable),
	}
	fake.refs["ns=2;s=Plant1"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=MV01", "MV01", ua.NodeClassVariable),
	}
	// Children under a Variable would be a bug if descended.
	fake.refs["ns=2;s=Sensor"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=ShouldNotAppear", "ShouldNotAppear", ua.NodeClassVariable),
	}

	tree, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)

	names := collectBrowseNames(tree.AllNodes)
	require.ElementsMatch(t, []string{"Plant1", "Sensor", "MV01"}, names)

	mv01 := findByName(tree.AllNodes, "MV01")
	require.NotNil(t, mv01)
	require.Equal(t, []string{"Plant1", "MV01"}, mv01.PathSegments)
}

func TestBrowseCycleDetection(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=A", "A", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=A"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=B", "B", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=B"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=A", "A", ua.NodeClassObject),
	}

	tree, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)
	require.Len(t, tree.AllNodes, 2, "cycle must not produce duplicates")
}

func TestBrowseMaxDepth(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=L1", "L1", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=L1"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=L2", "L2", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=L2"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=L3", "L3", ua.NodeClassObject),
	}

	browser := newBrowser(fake)
	browser.MaxDepth = 2

	tree, err := browser.Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)

	names := collectBrowseNames(tree.AllNodes)
	require.ElementsMatch(t, []string{"L1", "L2"}, names)
}

func TestBrowseMaxNodes(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=A", "A", ua.NodeClassObject),
		makeRef(t, "ns=2;s=B", "B", ua.NodeClassObject),
		makeRef(t, "ns=2;s=C", "C", ua.NodeClassObject),
		makeRef(t, "ns=2;s=D", "D", ua.NodeClassObject),
	}

	browser := newBrowser(fake)
	browser.MaxNodes = 2

	tree, err := browser.Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)
	require.Len(t, tree.AllNodes, 2)
}

func TestBrowsePerResultBadStatusSkipped(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=Good", "Good", ua.NodeClassObject),
		makeRef(t, "ns=2;s=Forbidden", "Forbidden", ua.NodeClassObject),
	}
	fake.statuses["ns=2;s=Forbidden"] = ua.StatusBadUserAccessDenied
	fake.refs["ns=2;s=Good"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=GoodChild", "GoodChild", ua.NodeClassVariable),
	}

	tree, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)

	names := collectBrowseNames(tree.AllNodes)
	require.ElementsMatch(t, []string{"Good", "Forbidden", "GoodChild"}, names)
}

func TestBrowseContinuationPoints(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.chunkSize = 2
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=A", "A", ua.NodeClassVariable),
		makeRef(t, "ns=2;s=B", "B", ua.NodeClassVariable),
		makeRef(t, "ns=2;s=C", "C", ua.NodeClassVariable),
		makeRef(t, "ns=2;s=D", "D", ua.NodeClassVariable),
		makeRef(t, "ns=2;s=E", "E", ua.NodeClassVariable),
	}

	tree, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)

	names := collectBrowseNames(tree.AllNodes)
	require.ElementsMatch(t, []string{"A", "B", "C", "D", "E"}, names)
	require.GreaterOrEqual(t, fake.nextCalls, 1, "BrowseNext must be invoked when chunked")
}

func TestBrowseRPCErrorPropagates(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.browseErr = errors.New("network down")

	_, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.ErrorContains(t, err, "browse request failed")
}

func TestBrowseNextRPCErrorPropagates(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.chunkSize = 1
	fake.browseNextErr = errors.New("next failed")
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=A", "A", ua.NodeClassVariable),
		makeRef(t, "ns=2;s=B", "B", ua.NodeClassVariable),
	}

	_, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.ErrorContains(t, err, "browse-next request failed")
}

func TestBrowsePathSegmentsPreserved(t *testing.T) {
	fake := newFakeBrowseClient()
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=Objects", "Objects", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=Objects"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=Plant1", "Plant1", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=Plant1"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=Device1", "Device1", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=Device1"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=MV01", "MV01", ua.NodeClassVariable),
	}

	tree, err := newBrowser(fake).Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)

	mv01 := findByName(tree.AllNodes, "MV01")
	require.NotNil(t, mv01)
	require.Equal(t, []string{"Objects", "Plant1", "Device1", "MV01"}, mv01.PathSegments)
}

func TestBrowseBatching(t *testing.T) {
	fake := newFakeBrowseClient()
	// Three siblings under root, each with a child. With batch size 2,
	// the second-level expansion should issue a single batched browse.
	fake.refs["i=85"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=A", "A", ua.NodeClassObject),
		makeRef(t, "ns=2;s=B", "B", ua.NodeClassObject),
	}
	fake.refs["ns=2;s=A"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=AC", "AC", ua.NodeClassVariable),
	}
	fake.refs["ns=2;s=B"] = []*ua.ReferenceDescription{
		makeRef(t, "ns=2;s=BC", "BC", ua.NodeClassVariable),
	}

	browser := newBrowser(fake)
	browser.BatchSize = 5

	tree, err := browser.Browse(t.Context(), ua.NewNumericNodeID(0, 85))
	require.NoError(t, err)
	require.Len(t, tree.AllNodes, 4)
	require.Equal(t, 2, fake.browseCalls, "root browse + one batched browse for the two children")
}

type fakeBrowseClient struct {
	refs          map[string][]*ua.ReferenceDescription
	statuses      map[string]ua.StatusCode
	browseErr     error
	browseNextErr error
	chunkSize     int
	continuations map[string][]*ua.ReferenceDescription
	browseCalls   int
	nextCalls     int
}

func newFakeBrowseClient() *fakeBrowseClient {
	return &fakeBrowseClient{
		refs:          make(map[string][]*ua.ReferenceDescription),
		statuses:      make(map[string]ua.StatusCode),
		continuations: make(map[string][]*ua.ReferenceDescription),
	}
}

func (f *fakeBrowseClient) Browse(_ context.Context, req *ua.BrowseRequest) (*ua.BrowseResponse, error) {
	f.browseCalls++
	if f.browseErr != nil {
		return nil, f.browseErr
	}
	resp := &ua.BrowseResponse{Results: make([]*ua.BrowseResult, len(req.NodesToBrowse))}
	for i, desc := range req.NodesToBrowse {
		key := desc.NodeID.String()
		result := &ua.BrowseResult{}
		if status, ok := f.statuses[key]; ok {
			result.StatusCode = status
			resp.Results[i] = result
			continue
		}
		refs := f.refs[key]
		if f.chunkSize > 0 && len(refs) > f.chunkSize {
			cp := []byte("cp-" + key)
			f.continuations[string(cp)] = refs[f.chunkSize:]
			result.References = refs[:f.chunkSize]
			result.ContinuationPoint = cp
		} else {
			result.References = refs
		}
		resp.Results[i] = result
	}
	return resp, nil
}

func (f *fakeBrowseClient) BrowseNext(_ context.Context, req *ua.BrowseNextRequest) (*ua.BrowseNextResponse, error) {
	f.nextCalls++
	if f.browseNextErr != nil {
		return nil, f.browseNextErr
	}
	resp := &ua.BrowseNextResponse{Results: make([]*ua.BrowseResult, len(req.ContinuationPoints))}
	for i, cp := range req.ContinuationPoints {
		remaining, ok := f.continuations[string(cp)]
		if !ok {
			resp.Results[i] = &ua.BrowseResult{}
			continue
		}
		delete(f.continuations, string(cp))
		result := &ua.BrowseResult{}
		if f.chunkSize > 0 && len(remaining) > f.chunkSize {
			nextCp := []byte(string(cp) + "+")
			f.continuations[string(nextCp)] = remaining[f.chunkSize:]
			result.References = remaining[:f.chunkSize]
			result.ContinuationPoint = nextCp
		} else {
			result.References = remaining
		}
		resp.Results[i] = result
	}
	return resp, nil
}

func makeRef(t *testing.T, nodeID, browseName string, class ua.NodeClass) *ua.ReferenceDescription {
	t.Helper()
	nid, err := ua.ParseNodeID(nodeID)
	require.NoError(t, err)
	return &ua.ReferenceDescription{
		NodeID:      &ua.ExpandedNodeID{NodeID: nid},
		BrowseName:  &ua.QualifiedName{Name: browseName},
		DisplayName: &ua.LocalizedText{Text: browseName},
		NodeClass:   class,
	}
}

func newBrowser(client browseClient) *AddressSpaceBrowser {
	return &AddressSpaceBrowser{Client: client, Log: testutil.Logger{}}
}

func collectBrowseNames(nodes []*BrowsedNode) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.BrowseName)
	}
	return names
}

func findByName(nodes []*BrowsedNode, name string) *BrowsedNode {
	for _, n := range nodes {
		if n.BrowseName == name {
			return n
		}
	}
	return nil
}
