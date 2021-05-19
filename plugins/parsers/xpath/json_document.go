package xpath

import (
	"strings"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xpath"
)

type jsonDocument struct{}

func (xh *jsonDocument) Parse(buf []byte) (dataNode, error) {
	return jsonquery.Parse(strings.NewReader(string(buf)))
}

func (xh *jsonDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	// If this panics it's a programming error as we changed the document type while processing
	native, err := jsonquery.QueryAll(node.(*jsonquery.Node), expr)
	if err != nil {
		return nil, err
	}

	nodes := make([]dataNode, len(native))
	for i, n := range native {
		nodes[i] = n
	}
	return nodes, nil
}

func (xh *jsonDocument) CreateXPathNavigator(node dataNode) xpath.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return jsonquery.CreateXPathNavigator(node.(*jsonquery.Node))
}

func (xh *jsonDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	names := make([]string, 0)

	// If these panic it's a programming error as we changed the document type while processing
	nativeNode := node.(*jsonquery.Node)
	nativeRelativeTo := relativeTo.(*jsonquery.Node)

	// Climb up the tree and collect the node names
	n := nativeNode.Parent
	for n != nil && n != nativeRelativeTo {
		names = append(names, n.Data)
		n = n.Parent
	}

	if len(names) < 1 {
		return ""
	}

	// Construct the nodes
	path := ""
	for _, name := range names {
		path = name + sep + path
	}

	return path[:len(path)-1]
}

func (xh *jsonDocument) OutputXML(node dataNode) string {
	native := node.(*jsonquery.Node)
	return native.OutputXML()
}
