package xpath

import (
	"bytes"
	"encoding/json"
	"fmt"

	path "github.com/antchfx/xpath"
	"github.com/fxamacker/cbor/v2"
	"github.com/srebhan/cborquery"
	"github.com/tinylib/msgp/msgp"
)

type msgpackDocument jsonDocument

func (*msgpackDocument) Parse(buf []byte) (dataNode, error) {
	var jsonBuf bytes.Buffer

	// Unmarshal the message-pack binary message to JSON
	if _, err := msgp.UnmarshalAsJSON(&jsonBuf, buf); err != nil {
		return nil, fmt.Errorf("unmarshalling to json failed: %w", err)
	}

	// Parse JSON to interface{}
	var data interface{}
	if err := json.Unmarshal(jsonBuf.Bytes(), &data); err != nil {
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

func (d *msgpackDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	return (*jsonDocument)(d).QueryAll(node, expr)
}

func (d *msgpackDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
	return (*jsonDocument)(d).CreateXPathNavigator(node)
}

func (d *msgpackDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	return (*jsonDocument)(d).GetNodePath(node, relativeTo, sep)
}
func (d *msgpackDocument) GetNodeName(node dataNode, sep string, withParent bool) string {
	return (*jsonDocument)(d).GetNodeName(node, sep, withParent)
}

func (d *msgpackDocument) OutputXML(node dataNode) string {
	return (*jsonDocument)(d).OutputXML(node)
}
