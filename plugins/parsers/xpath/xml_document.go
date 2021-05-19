package xpath

import (
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
)

type xmlDocument struct{}

func (xh *xmlDocument) Parse(buf []byte) (dataNode, error) {
	return xmlquery.Parse(strings.NewReader(string(buf)))
}

func (xh *xmlDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	// If this panics it's a programming error as we changed the document type while processing
	native, err := xmlquery.QueryAll(node.(*xmlquery.Node), expr)
	if err != nil {
		return nil, err
	}

	nodes := make([]dataNode, len(native))
	for i, n := range native {
		nodes[i] = n
	}
	return nodes, nil
}

func (xh *xmlDocument) CreateXPathNavigator(node dataNode) xpath.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return xmlquery.CreateXPathNavigator(node.(*xmlquery.Node))
}

func (xh *xmlDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	names := make([]string, 0)

	// If these panic it's a programming error as we changed the document type while processing
	nativeNode := node.(*xmlquery.Node)
	nativeRelativeTo := relativeTo.(*xmlquery.Node)

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

func (xh *xmlDocument) OutputXML(node dataNode) string {
	native := node.(*xmlquery.Node)
	return native.OutputXML(false)
}
