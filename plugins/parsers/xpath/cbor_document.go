package xpath

import (
	"reflect"
	"strconv"
	"strings"

	path "github.com/antchfx/xpath"
	"github.com/srebhan/cborquery"
)

type cborDocument struct{}

func (d *cborDocument) Parse(buf []byte) (dataNode, error) {
	return cborquery.Parse(strings.NewReader(string(buf)))
}

func (d *cborDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
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

func (d *cborDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return cborquery.CreateXPathNavigator(node.(*cborquery.Node))
}

func (d *cborDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	names := make([]string, 0)

	// If these panic it's a programming error as we changed the document type while processing
	nativeNode := node.(*cborquery.Node)
	nativeRelativeTo := relativeTo.(*cborquery.Node)

	// Climb up the tree and collect the node names
	n := nativeNode.Parent
	for n != nil && n != nativeRelativeTo {
		kind := reflect.Invalid
		if n.Parent != nil && n.Parent.Value() != nil {
			kind = reflect.TypeOf(n.Parent.Value()).Kind()
		}

		switch kind {
		case reflect.Slice, reflect.Array:
			// Determine the index for array elements
			names = append(names, d.index(n))
		default:
			// Use the name if not an array
			names = append(names, n.Name)
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

func (d *cborDocument) OutputXML(node dataNode) string {
	native := node.(*cborquery.Node)
	return native.OutputXML()
}

func (d *cborDocument) index(node *cborquery.Node) string {
	idx := 0

	for n := node; n.PrevSibling != nil; n = n.PrevSibling {
		idx++
	}

	return strconv.Itoa(idx)
}
