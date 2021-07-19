package xpath

import (
	"strings"

	"github.com/antchfx/xmlquery"
	path "github.com/antchfx/xpath"
)

type xmlDocument struct{}

func (d *xmlDocument) Parse(buf []byte) (dataNode, error) {
	return xmlquery.Parse(strings.NewReader(string(buf)))
}

func (d *xmlDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
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

func (d *xmlDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return xmlquery.CreateXPathNavigator(node.(*xmlquery.Node))
}

func (d *xmlDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
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
	nodepath := ""
	for _, name := range names {
		nodepath = name + sep + nodepath
	}

	return nodepath[:len(nodepath)-1]
}

func (d *xmlDocument) OutputXML(node dataNode) string {
	native := node.(*xmlquery.Node)
	return native.OutputXML(false)
}
