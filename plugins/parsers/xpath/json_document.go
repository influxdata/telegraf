package xpath

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	path "github.com/antchfx/xpath"
	"github.com/fxamacker/cbor/v2"
	"github.com/srebhan/cborquery"
)

type jsonDocument struct{}

func (*jsonDocument) Parse(buf []byte) (dataNode, error) {
	// First parse JSON to an interface{}
	var data interface{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to CBOR to leverage cborquery's correct array handling
	cborData, err := cbor.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to CBOR: %w", err)
	}

	// Parse with cborquery which handles arrays correctly
	return cborquery.Parse(bytes.NewReader(cborData))
}

func (*jsonDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	// If this panics it's a programming error as we changed the document type while processing
	native, err := cborquery.QueryAll(node.(*cborquery.Node), expr)
	if err != nil {
		return nil, err
	}

	nodes := make([]dataNode, 0, len(native))
	for _, n := range native {
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (*jsonDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return cborquery.CreateXPathNavigator(node.(*cborquery.Node))
}

func (d *jsonDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	names := make([]string, 0)

	// If these panic it's a programming error as we changed the document type while processing
	nativeNode := node.(*cborquery.Node)
	nativeRelativeTo := relativeTo.(*cborquery.Node)

	// Climb up the tree and collect the node names
	n := nativeNode.Parent
	for n != nil && n != nativeRelativeTo {
		nodeName := d.GetNodeName(n, sep, false)
		names = append(names, nodeName)
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

func (d *jsonDocument) GetNodeName(node dataNode, sep string, withParent bool) string {
	// If this panics it's a programming error as we changed the document type while processing
	nativeNode := node.(*cborquery.Node)

	name := nativeNode.Name

	// In cborquery, array elements appear as siblings with the same name.
	// Check if this node is part of an array by looking for siblings with the same name.
	if nativeNode.Parent != nil && name != "" {
		idx, count := d.siblingIndex(nativeNode)
		if count > 1 {
			// This is an array element, append the index
			return name + sep + strconv.Itoa(idx)
		}
	}

	return name
}

func (*jsonDocument) OutputXML(node dataNode) string {
	native := node.(*cborquery.Node)
	return native.OutputXML()
}

func (*jsonDocument) siblingIndex(node *cborquery.Node) (idx, count int) {
	if node.Parent == nil {
		return 0, 1
	}

	// Count siblings with the same name and find our index among them
	for sibling := node.Parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Name == node.Name {
			if sibling == node {
				idx = count
			}
			count++
		}
	}
	return idx, count
}
