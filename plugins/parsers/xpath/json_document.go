package xpath

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/antchfx/jsonquery"
	path "github.com/antchfx/xpath"
)

type jsonDocument struct{}

func (d *jsonDocument) Parse(buf []byte) (dataNode, error) {
	return jsonquery.Parse(strings.NewReader(string(buf)))
}

func (d *jsonDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	// If this panics it's a programming error as we changed the document type while processing
	native, err := jsonquery.QueryAll(node.(*jsonquery.Node), expr)
	if err != nil {
		return nil, err
	}

	nodes := make([]dataNode, 0, len(native))
	for _, n := range native {
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (d *jsonDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return jsonquery.CreateXPathNavigator(node.(*jsonquery.Node))
}

func (d *jsonDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	names := make([]string, 0)

	// If these panic it's a programming error as we changed the document type while processing
	nativeNode := node.(*jsonquery.Node)
	nativeRelativeTo := relativeTo.(*jsonquery.Node)

	// Climb up the tree and collect the node names
	n := nativeNode.Parent
	for n != nil && n != nativeRelativeTo {
		switch reflect.TypeOf(n.Parent.Value()).Kind() {
		case reflect.Slice, reflect.Array:
			// Determine the index for array elements
			names = append(names, d.index(n))
		default:
			// Use the name if not an array
			names = append(names, n.Data)
		}
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

func (d *jsonDocument) OutputXML(node dataNode) string {
	native := node.(*jsonquery.Node)
	return native.OutputXML()
}

func (d *jsonDocument) index(node *jsonquery.Node) string {
	idx := 0

	for n := node; n.PrevSibling != nil; n = n.PrevSibling {
		idx++
	}

	return strconv.Itoa(idx)
}
