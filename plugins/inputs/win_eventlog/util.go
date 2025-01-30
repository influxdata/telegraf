//go:build windows

package win_eventlog

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/sys/windows"
)

// decodeUTF16 to UTF8 bytes
func decodeUTF16(b []byte) ([]byte, error) {
	if len(b)%2 != 0 {
		return nil, errors.New("must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.Bytes(), nil
}

// getFromSnapProcess finds information about process by the given pid
// Returns process name
func getFromSnapProcess(pid uint32) (string, error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(snap)
	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))
	if err := windows.Process32First(snap, &pe32); err != nil {
		return "", err
	}
	for {
		if pe32.ProcessID == pid {
			szexe := windows.UTF16ToString(pe32.ExeFile[:])
			return szexe, nil
		}
		if err = windows.Process32Next(snap, &pe32); err != nil {
			break
		}
	}
	return "", fmt.Errorf("couldn't find pid: %d", pid)
}

type xmlnode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Text    string     `xml:",chardata"`
	Nodes   []xmlnode  `xml:",any"`
}

// eventField for unique rendering
type eventField struct {
	Name  string
	Value string
}

// UnmarshalXML redefined for xml elements walk
func (n *xmlnode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node xmlnode

	return d.DecodeElement((*node)(n), &start)
}

// unrollXMLFields extracts fields from xml data
func unrollXMLFields(data []byte, fieldsUsage map[string]int, separator string) ([]eventField, map[string]int) {
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	var fields []eventField
	for {
		var node xmlnode
		err := dec.Decode(&node)
		if err != nil {
			break
		}

		var parents []string
		walkXML([]xmlnode{node}, parents, separator, func(node xmlnode, parents []string, separator string) bool {
			innerText := strings.TrimSpace(node.Text)
			if len(innerText) > 0 {
				valueName := strings.Join(parents, separator)
				fieldsUsage[valueName]++
				field := eventField{Name: valueName, Value: innerText}
				fields = append(fields, field)
			}
			return true
		})
	}
	return fields, fieldsUsage
}

func walkXML(nodes []xmlnode, parents []string, separator string, f func(xmlnode, []string, string) bool) {
	for _, node := range nodes {
		parentName := node.XMLName.Local
		for _, attr := range node.Attrs {
			attrName := strings.ToLower(attr.Name.Local)
			if attrName == "name" {
				// Add Name attribute to parent name
				parentName = strings.Join([]string{parentName, attr.Value}, separator)
			}
		}
		nodeParents := append(parents, parentName)
		if f(node, nodeParents, separator) {
			walkXML(node.Nodes, nodeParents, separator, f)
		}
	}
}

// uniqueFieldNames forms unique field names by adding _<num> if there are several of them
func uniqueFieldNames(fields []eventField, fieldsUsage map[string]int, separator string) []eventField {
	var fieldsCounter = make(map[string]int, len(fields))
	fieldsUnique := make([]eventField, 0, len(fields))
	for _, field := range fields {
		fieldName := field.Name
		if fieldsUsage[field.Name] > 1 {
			fieldsCounter[field.Name]++
			fieldName = fmt.Sprint(field.Name, separator, fieldsCounter[field.Name])
		}
		fieldsUnique = append(fieldsUnique, eventField{
			Name:  fieldName,
			Value: field.Value,
		})
	}
	return fieldsUnique
}
