package gnmi

import gnmiLib "github.com/openconfig/gnmi/proto/gnmi"

type Worker struct {
	address  string
	tagStore *tagNode
}

type tagNode struct {
	elem     *gnmiLib.PathElem
	tagName  string
	value    *gnmiLib.TypedValue
	tagStore map[string][]*tagNode
}

type tagResults struct {
	names  []string
	values []*gnmiLib.TypedValue
}

func (w *Worker) storeTags(update *gnmiLib.Update, sub TagSubscription) {
	updateKeys := pathKeys(update.Path)
	var foundKey bool
	for _, requiredKey := range sub.Elements {
		foundKey = false
		for _, elem := range updateKeys {
			if elem.Name == requiredKey {
				foundKey = true
			}
		}
		if !foundKey {
			return
		}
	}
	// All required keys present for this TagSubscription
	w.tagStore.insert(updateKeys, sub.Name, update.Val)
}

func (node *tagNode) insert(keys []*gnmiLib.PathElem, name string, value *gnmiLib.TypedValue) {
	if len(keys) == 0 {
		node.value = value
		node.tagName = name
		return
	}
	var found *tagNode
	key := keys[0]
	keyName := key.Name
	if node.tagStore == nil {
		node.tagStore = make(map[string][]*tagNode)
	}
	if _, ok := node.tagStore[keyName]; !ok {
		node.tagStore[keyName] = make([]*tagNode, 0)
	}
	for _, node := range node.tagStore[keyName] {
		if compareKeys(node.elem.Key, key.Key) {
			found = node
			break
		}
	}
	if found == nil {
		found = &tagNode{elem: keys[0]}
		node.tagStore[keyName] = append(node.tagStore[keyName], found)
	}
	found.insert(keys[1:], name, value)
}

func (node *tagNode) retrieve(keys []*gnmiLib.PathElem, tagResults *tagResults) {
	if node.value != nil {
		tagResults.names = append(tagResults.names, node.tagName)
		tagResults.values = append(tagResults.values, node.value)
	}
	for _, key := range keys {
		if elems, ok := node.tagStore[key.Name]; ok {
			for _, node := range elems {
				if compareKeys(node.elem.Key, key.Key) {
					node.retrieve(keys, tagResults)
				}
			}
		}
	}
}

func (w *Worker) checkTags(fullPath *gnmiLib.Path) map[string]interface{} {
	results := &tagResults{}
	w.tagStore.retrieve(pathKeys(fullPath), results)
	tags := make(map[string]interface{})
	for idx := range results.names {
		vals, _ := gnmiToFields(results.names[idx], results.values[idx])
		for k, v := range vals {
			tags[k] = v
		}
	}
	return tags
}
