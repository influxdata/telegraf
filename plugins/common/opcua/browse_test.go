package opcua

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/server"
	"github.com/gopcua/opcua/server/attrs"
	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestBrowseSingleLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that spins up an OPC UA server")
	}
	ts := startBrowseTestServer(t)
	root := ts.addFolder(t, nil, "root")
	ts.addFolder(t, root, "Plant1")
	ts.addVariable(t, root, "ServerTime")
	ts.start(t)

	client := ts.connect(t)
	nodes, err := newBrowser(client).Browse(t.Context(), root.id)
	require.NoError(t, err)
	require.Len(t, nodes, 2)

	plant1 := findByName(nodes, "Plant1")
	require.NotNil(t, plant1)
	require.Equal(t, "Plant1", plant1.Path)
	require.Equal(t, ua.NodeClassObject, plant1.NodeClass)

	serverTime := findByName(nodes, "ServerTime")
	require.NotNil(t, serverTime)
	require.Equal(t, "ServerTime", serverTime.Path)
	require.Equal(t, ua.NodeClassVariable, serverTime.NodeClass)
}

func TestBrowseDescendsOnlyContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that spins up an OPC UA server")
	}
	ts := startBrowseTestServer(t)
	root := ts.addFolder(t, nil, "root")
	plant := ts.addFolder(t, root, "Plant1")
	ts.addVariable(t, plant, "MV01")
	ts.addVariable(t, root, "Sensor")
	ts.start(t)

	client := ts.connect(t)
	nodes, err := newBrowser(client).Browse(t.Context(), root.id)
	require.NoError(t, err)

	require.ElementsMatch(t, []string{"Plant1", "MV01", "Sensor"}, collectBrowseNames(nodes))
	mv01 := findByName(nodes, "MV01")
	require.NotNil(t, mv01)
	require.Equal(t, "Plant1/MV01", mv01.Path)
}

func TestBrowseMaxDepth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that spins up an OPC UA server")
	}
	ts := startBrowseTestServer(t)
	root := ts.addFolder(t, nil, "root")
	l1 := ts.addFolder(t, root, "L1")
	l2 := ts.addFolder(t, l1, "L2")
	ts.addFolder(t, l2, "L3")
	ts.start(t)

	browser := newBrowser(ts.connect(t))
	browser.MaxDepth = 2

	nodes, err := browser.Browse(t.Context(), root.id)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"L1", "L2"}, collectBrowseNames(nodes))
}

func TestBrowseMaxNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that spins up an OPC UA server")
	}
	ts := startBrowseTestServer(t)
	root := ts.addFolder(t, nil, "root")
	for _, name := range []string{"A", "B", "C", "D"} {
		ts.addFolder(t, root, name)
	}
	ts.start(t)

	browser := newBrowser(ts.connect(t))
	browser.MaxNodes = 2

	nodes, err := browser.Browse(t.Context(), root.id)
	require.NoError(t, err)
	require.Len(t, nodes, 2)
}

func TestBrowsePathPreserved(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that spins up an OPC UA server")
	}
	ts := startBrowseTestServer(t)
	root := ts.addFolder(t, nil, "root")
	objects := ts.addFolder(t, root, "Objects")
	plant1 := ts.addFolder(t, objects, "Plant1")
	device1 := ts.addFolder(t, plant1, "Device1")
	ts.addVariable(t, device1, "MV01")
	ts.start(t)

	nodes, err := newBrowser(ts.connect(t)).Browse(t.Context(), root.id)
	require.NoError(t, err)

	mv01 := findByName(nodes, "MV01")
	require.NotNil(t, mv01)
	require.Equal(t, "Objects/Plant1/Device1/MV01", mv01.Path)
}

func newBrowser(c *opcua.Client) *AddressSpaceBrowser {
	return &AddressSpaceBrowser{Client: c, Log: testutil.Logger{}, BatchSize: 50}
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

type browseTestServer struct {
	srv *server.Server
	ns  *server.NodeNameSpace
	url string
}

type browseTestNode struct {
	id   *ua.NodeID
	node *server.Node
}

func startBrowseTestServer(t *testing.T) *browseTestServer {
	t.Helper()

	// Bind a free port up front; close the listener so the server can
	// claim it. The brief gap is acceptable for serial unit tests.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())

	srv := server.New(
		server.EnableSecurity("None", ua.MessageSecurityModeNone),
		server.EnableAuthMode(ua.UserTokenTypeAnonymous),
		server.EndPoint("127.0.0.1", port),
	)

	ns := server.NewNodeNameSpace(srv, "telegraf-test")
	srv.AddNamespace(ns)

	// Hook the test namespace's Objects folder under the standard one so
	// the standard Objects(i=85) browse path remains intact.
	rootNS, err := srv.Namespace(0)
	require.NoError(t, err)
	rootNS.Objects().AddRef(ns.Objects(), id.HasComponent, true)

	return &browseTestServer{
		srv: srv,
		ns:  ns,
		url: fmt.Sprintf("opc.tcp://127.0.0.1:%d", port),
	}
}

func (ts *browseTestServer) addFolder(t *testing.T, parent *browseTestNode, name string) *browseTestNode {
	t.Helper()
	// Build the folder/object node by hand: server.NewFolderNode in
	// gopcua v0.8.0 panics building the Description value (it passes the
	// NodeClass enum to a variant constructor that does not support it).
	nodeID := ua.NewStringNodeID(ts.ns.ID(), name)
	folder := server.NewNode(
		nodeID,
		map[ua.AttributeID]*ua.DataValue{
			ua.AttributeIDNodeClass:   server.DataValueFromValue(uint32(ua.NodeClassObject)),
			ua.AttributeIDBrowseName:  server.DataValueFromValue(attrs.BrowseName(name)),
			ua.AttributeIDDisplayName: server.DataValueFromValue(attrs.DisplayName(name, name)),
		},
		nil,
		nil,
	)
	ts.ns.AddNode(folder)
	if parent == nil {
		ts.ns.Objects().AddRef(folder, id.HasComponent, true)
	} else {
		parent.node.AddRef(folder, id.HasComponent, true)
	}
	return &browseTestNode{id: nodeID, node: folder}
}

func (ts *browseTestServer) addVariable(t *testing.T, parent *browseTestNode, name string) {
	t.Helper()
	nodeID := ua.NewStringNodeID(ts.ns.ID(), name)
	v := server.NewVariableNode(nodeID, name, int32(0))
	ts.ns.AddNode(v)
	parent.node.AddRef(v, id.HasComponent, true)
}

func (ts *browseTestServer) start(t *testing.T) {
	t.Helper()
	require.NoError(t, ts.srv.Start(context.Background()))
	t.Cleanup(func() {
		_ = ts.srv.Close()
	})
	// Wait for the listener to accept connections before returning.
	require.Eventually(t, func() bool {
		c, err := net.DialTimeout("tcp", ts.url[len("opc.tcp://"):], 100*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	}, 5*time.Second, 50*time.Millisecond, "OPC UA server did not start listening")
}

func (ts *browseTestServer) connect(t *testing.T) *opcua.Client {
	t.Helper()
	c, err := opcua.NewClient(ts.url, opcua.SecurityMode(ua.MessageSecurityModeNone))
	require.NoError(t, err)
	require.NoError(t, c.Connect(context.Background()))
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})
	return c
}
