//
// Created by Yaz Saito on 06/10/12.
//

// A red-black tree with an API similar to C++ STL's.
//
// The implementation is inspired (read: stolen) from:
// http://en.literateprograms.org/Red-black_tree_(C)#chunk use:private function prototypes.
//
package rbtree

//
// Public definitions
//

// Item is the object stored in each tree node.
type Item interface{}

// CompareFunc returns 0 if a==b, <0 if a<b, >0 if a>b.
type CompareFunc func(a, b Item) int

type Tree struct {
	// Root of the tree
	root             *node

	// The minimum and maximum nodes under the root.
	minNode, maxNode *node

	// Number of nodes under root, including the root
	count            int
	compare          CompareFunc
}

// Create a new empty tree.
func NewTree(compare CompareFunc) *Tree {
	return &Tree{compare: compare}
}

// Return the number of elements in the tree.
func (root *Tree) Len() int {
	return root.count
}

// A convenience function for finding an element equal to key. Return
// nil if not found.
func (root *Tree) Get(key Item) Item {
	n, exact := root.findGE(key)
	if exact {
		return n.item
	}
	return nil
}

// Create an iterator that points to the minimum item in the tree
// If the tree is empty, return Limit()
func (root *Tree) Min() Iterator {
	return Iterator{root, root.minNode}
}

// Create an iterator that points at the maximum item in the tree
//
// If the tree is empty, return NegativeLimit()
func (root *Tree) Max() Iterator {
	if root.maxNode == nil {
		// TODO: there are a few checks of this form.
		// Perhaps set maxNode=negativeLimit when the tree is empty
		return Iterator{root, negativeLimitNode}
	}
	return  Iterator{root, root.maxNode}
}

// Create an iterator that points beyond the maximum item in the tree
func (root *Tree) Limit() Iterator {
	return Iterator{root, nil}
}

// Create an iterator that points before the minimum item in the tree
func (root *Tree) NegativeLimit() Iterator {
	return  Iterator{root, negativeLimitNode}
}

// Find the smallest element N such that N >= key, and return the
// iterator pointing to the element. If no such element is found,
// return root.Limit().
func (root *Tree) FindGE(key Item) Iterator {
	n, _ := root.findGE(key)
	return Iterator{root, n}
}

// Find the largest element N such that N <= key, and return the
// iterator pointing to the element. If no such element is found,
// return iter.NegativeLimit().
func (root *Tree) FindLE(key Item) Iterator {
	n, exact := root.findGE(key)
	if exact {
		return Iterator{root, n}
	}
	if n != nil {
		return Iterator{root, n.doPrev()}
	}
	if root.maxNode == nil {
		return Iterator{root, negativeLimitNode}
	}
	return Iterator{root, root.maxNode}
}

// Insert an item. If the item is already in the tree, do nothing and
// return false. Else return true.
func (root *Tree) Insert(item Item) bool {
	// TODO: delay creating n until it is found to be inserted
	n := root.doInsert(item)
	if n == nil {
		return false
	}

	n.color = red

	for true {
		// Case 1: N is at the root
		if n.parent == nil {
			n.color = black
			break
		}

		// Case 2: The parent is black, so the tree already
		// satisfies the RB properties
		if n.parent.color == black {
			break
		}

		// Case 3: parent and uncle are both red.
		// Then paint both black and make grandparent red.
		grandparent := n.parent.parent
		var uncle *node
		if n.parent.isLeftChild() {
			uncle = grandparent.right
		} else {
			uncle = grandparent.left
		}
		if uncle != nil && uncle.color == red {
			n.parent.color = black
			uncle.color = black
			grandparent.color = red
			n = grandparent
			continue
		}

		// Case 4: parent is red, uncle is black (1)
		if n.isRightChild() && n.parent.isLeftChild() {
			root.rotateLeft(n.parent)
			n = n.left
			continue
		}
		if n.isLeftChild() && n.parent.isRightChild() {
			root.rotateRight(n.parent)
			n = n.right
			continue
		}

		// Case 5: parent is read, uncle is black (2)
		n.parent.color = black
		grandparent.color = red
		if n.isLeftChild() {
			root.rotateRight(grandparent)
		} else {
			root.rotateLeft(grandparent)
		}
		break
	}
	return true
}

// Delete an item with the given key. Return true iff the item was
// found.
func (root *Tree) DeleteWithKey(key Item) bool {
	n, exact := root.findGE(key)
	if exact {
		root.doDelete(n)
		return true
	}
	return false
}

// Delete the current item.
//
// REQUIRES: !iter.Limit() && !iter.NegativeLimit()
func (root *Tree) DeleteWithIterator(iter Iterator) {
	doAssert(!iter.Limit() && !iter.NegativeLimit())
	root.doDelete(iter.node)
}

// Iterator allows scanning tree elements in sort order.
//
// Iterator invalidation rule is the same as C++ std::map<>'s. That
// is, if you delete the element that an iterator points to, the
// iterator becomes invalid. For other operation types, the iterator
// remains valid.
type Iterator struct {
	root *Tree
	node *node
}

func (iter Iterator) Equal(iter2 Iterator) bool {
	return iter.node == iter2.node
}

// Check if the iterator points beyond the max element in the tree
func (iter Iterator) Limit() bool {
	return iter.node == nil
}

// Check if the iterator points to the minimum element in the tree
func (iter Iterator) Min() bool {
	return iter.node == iter.root.minNode
}

// Check if the iterator points to the maximum element in the tree
func (iter Iterator) Max() bool {
	return iter.node == iter.root.maxNode
}

// Check if the iterator points before the minumum element in the tree
func (iter Iterator) NegativeLimit() bool {
	return iter.node == negativeLimitNode
}

// Return the current element.
//
// REQUIRES: !iter.Limit() && !iter.NegativeLimit()
func (iter Iterator) Item() interface{} {
	return iter.node.item
}

// Create a new iterator that points to the successor of the current element.
//
// REQUIRES: !iter.Limit()
func (iter Iterator) Next() Iterator {
	doAssert(!iter.Limit())
	if iter.NegativeLimit() {
		return Iterator{iter.root, iter.root.minNode}
	}
	return Iterator{iter.root, iter.node.doNext()}
}

// Create a new iterator that points to the predecessor of the current
// node.
//
// REQUIRES: !iter.NegativeLimit()
func (iter Iterator) Prev() Iterator {
	doAssert(!iter.NegativeLimit())
	if !iter.Limit() {
		return Iterator{iter.root, iter.node.doPrev()}
	}
	if iter.root.maxNode == nil {
		return Iterator{iter.root, negativeLimitNode}
	}
	return Iterator{iter.root, iter.root.maxNode}
}

func doAssert(b bool) {
	if !b {
		panic("rbtree internal assertion failed")
	}
}

const red = iota
const black = 1 + iota

type node struct {
	item                Item
	parent, left, right *node
	color               int // black or red
}

var negativeLimitNode *node

//
// Internal node attribute accessors
//
func getColor(n *node) int {
	if n == nil {
		return black
	}
	return n.color
}

func (n *node) isLeftChild() bool {
	return n == n.parent.left
}

func (n *node) isRightChild() bool {
	return n == n.parent.right
}

func (n *node) sibling() *node {
	doAssert(n.parent != nil)
	if n.isLeftChild() {
		return n.parent.right
	}
	return n.parent.left
}

// Return the minimum node that's larger than N. Return nil if no such
// node is found.
func (n *node) doNext() *node {
	if n.right != nil {
		m := n.right
		for m.left != nil {
			m = m.left
		}
		return m
	}

	for n != nil {
		p := n.parent
		if p == nil {
			return nil
		}
		if n.isLeftChild() {
			return p
		}
		n = p
	}
	return nil
}

// Return the maximum node that's smaller than N. Return nil if no
// such node is found.
func (n *node) doPrev() *node {
	if n.left != nil {
		return maxPredecessor(n)
	}

	for n != nil {
		p := n.parent
		if p == nil {
			break
		}
		if n.isRightChild() {
			return p
		}
		n = p
	}
	return negativeLimitNode
}

// Return the predecessor of "n".
func maxPredecessor(n *node) *node {
	doAssert(n.left != nil)
	m := n.left
	for m.right != nil {
		m = m.right
	}
	return m
}

//
// Tree methods
//

//
// Private methods
//

func (root *Tree) recomputeMinNode() {
	root.minNode = root.root
	if root.minNode != nil {
		for root.minNode.left != nil {
			root.minNode = root.minNode.left
		}
	}
}

func (root *Tree) recomputeMaxNode() {
	root.maxNode = root.root
	if root.maxNode != nil {
		for root.maxNode.right != nil {
			root.maxNode = root.maxNode.right
		}
	}
}

func (root *Tree) maybeSetMinNode(n *node) {
	if root.minNode == nil {
		root.minNode = n
		root.maxNode = n
	} else if root.compare(n.item, root.minNode.item) < 0 {
		root.minNode = n
	}
}

func (root *Tree) maybeSetMaxNode(n *node) {
	if root.maxNode == nil {
		root.minNode = n
		root.maxNode = n
	} else if root.compare(n.item, root.maxNode.item) > 0 {
		root.maxNode = n
	}
}

// Try inserting "item" into the tree. Return nil if the item is
// already in the tree. Otherwise return a new (leaf) node.
func (root *Tree) doInsert(item Item) *node {
	if root.root == nil {
		n := &node{item: item}
		root.root = n
		root.minNode = n
		root.maxNode = n
		root.count++
		return n
	}
	parent := root.root
	for true {
		comp := root.compare(item, parent.item)
		if comp == 0 {
			return nil
		} else if comp < 0 {
			if parent.left == nil {
				n := &node{item: item, parent: parent}
				parent.left = n
				root.count++
				root.maybeSetMinNode(n)
				return n
			} else {
				parent = parent.left
			}
		} else {
			if parent.right == nil {
				n := &node{item: item, parent: parent}
				parent.right = n
				root.count++
				root.maybeSetMaxNode(n)
				return n
			} else {
				parent = parent.right
			}
		}
	}
	panic("should not reach here")
}

// Find a node whose item >= key. The 2nd return value is true iff the
// node.item==key. Returns (nil, false) if all nodes in the tree are <
// key.
func (root *Tree) findGE(key Item) (*node, bool) {
	n := root.root
	for true {
		if n == nil {
			return nil, false
		}
		comp := root.compare(key, n.item)
		if comp == 0 {
			return n, true
		} else if comp < 0 {
			if n.left != nil {
				n = n.left
			} else {
				return n, false
			}
		} else {
			if n.right != nil {
				n = n.right
			} else {
				succ := n.doNext()
				if succ == nil {
					return nil, false
				} else {
					comp = root.compare(key, succ.item)
					return succ, (comp == 0)
				}
			}
		}
	}
	panic("should not reach here")
}


// Delete N from the tree.
func (root *Tree) doDelete(n *node) {
	if n.left != nil && n.right != nil {
		pred := maxPredecessor(n)
		root.swapNodes(n, pred)
	}

	doAssert(n.left == nil || n.right == nil)
	child := n.right
	if child == nil {
		child = n.left
	}
	if n.color == black {
		n.color = getColor(child)
		root.deleteCase1(n)
	}
	root.replaceNode(n, child)
	if n.parent == nil && child != nil {
		child.color = black
	}
	root.count--
	if root.count == 0 {
		root.minNode = nil
		root.maxNode = nil
	} else {
		if root.minNode == n {
			root.recomputeMinNode()
		}
		if root.maxNode == n {
			root.recomputeMaxNode()
		}
	}
}

// Move n to the pred's place, and vice versa
//
// TODO: this code is overly convoluted
func (root *Tree) swapNodes(n, pred *node) {
	doAssert(pred != n)
	isLeft := pred.isLeftChild()
	tmp := *pred
	root.replaceNode(n, pred)
	pred.color = n.color

	if tmp.parent == n {
		// swap the positions of n and pred
		if isLeft {
			pred.left = n
			pred.right = n.right
			if pred.right != nil {
				pred.right.parent = pred
			}
		} else {
			pred.left = n.left
			if pred.left != nil {
				pred.left.parent = pred
			}
			pred.right = n
		}
		n.item = tmp.item
		n.parent = pred

		n.left = tmp.left
		if n.left != nil {
			n.left.parent = n
		}
		n.right = tmp.right
		if n.right != nil {
			n.right.parent = n
		}
	} else {
		pred.left = n.left
		if pred.left != nil {
			pred.left.parent = pred
		}
		pred.right = n.right
		if pred.right != nil {
			pred.right.parent = pred
		}
		if isLeft {
			tmp.parent.left = n
		} else {
			tmp.parent.right = n
		}
		n.item = tmp.item
		n.parent = tmp.parent
		n.left = tmp.left
		if n.left != nil {
			n.left.parent = n
		}
		n.right = tmp.right
		if n.right != nil {
			n.right.parent = n
		}
	}
	n.color = tmp.color
}

func (root *Tree) deleteCase1(n *node) {
	for true {
		if n.parent != nil {
			if getColor(n.sibling()) == red {
				n.parent.color = red
				n.sibling().color = black
				if n == n.parent.left {
					root.rotateLeft(n.parent)
				} else {
					root.rotateRight(n.parent)
				}
			}
			if getColor(n.parent) == black &&
				getColor(n.sibling()) == black &&
				getColor(n.sibling().left) == black &&
				getColor(n.sibling().right) == black {
				n.sibling().color = red
				n = n.parent
				continue
			} else {
				// case 4
				if getColor(n.parent) == red &&
					getColor(n.sibling()) == black &&
					getColor(n.sibling().left) == black &&
					getColor(n.sibling().right) == black {
					n.sibling().color = red
					n.parent.color = black
				} else {
					root.deleteCase5(n)
				}
			}
		}
		break
	}
}

func (root *Tree) deleteCase5(n *node) {
	if n == n.parent.left &&
		getColor(n.sibling()) == black &&
		getColor(n.sibling().left) == red &&
		getColor(n.sibling().right) == black {
		n.sibling().color = red
		n.sibling().left.color = black
		root.rotateRight(n.sibling())
	} else if n == n.parent.right &&
		getColor(n.sibling()) == black &&
		getColor(n.sibling().right) == red &&
		getColor(n.sibling().left) == black {
		n.sibling().color = red
		n.sibling().right.color = black
		root.rotateLeft(n.sibling())
	}

	// case 6
	n.sibling().color = getColor(n.parent)
	n.parent.color = black
	if n == n.parent.left {
		doAssert(getColor(n.sibling().right) == red)
		n.sibling().right.color = black
		root.rotateLeft(n.parent)
	} else {
		doAssert(getColor(n.sibling().left) == red)
		n.sibling().left.color = black
		root.rotateRight(n.parent)
	}
}

func (root *Tree) replaceNode(oldn, newn *node) {
	if oldn.parent == nil {
		root.root = newn
	} else {
		if oldn == oldn.parent.left {
			oldn.parent.left = newn
		} else {
			oldn.parent.right = newn
		}
	}
	if newn != nil {
		newn.parent = oldn.parent
	}
}

/*
    X		     Y
  A   Y	    =>     X   C
     B C 	  A B
*/
func (root *Tree) rotateLeft(x *node) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		root.root = y
	} else {
		if x.isLeftChild() {
			x.parent.left = y
		} else {
			x.parent.right = y
		}
	}
	y.left = x
	x.parent = y
}

/*
     Y           X
   X   C  =>   A   Y
  A B             B C
*/
func (root *Tree) rotateRight(y *node) {
	x := y.left

	// Move "B"
	y.left = x.right
	if x.right != nil {
		x.right.parent = y
	}

	x.parent = y.parent
	if y.parent == nil {
		root.root = x
	} else {
		if y.isLeftChild() {
			y.parent.left = x
		} else {
			y.parent.right = x
		}
	}
	x.right = y
	y.parent = x
}

func init() {
	negativeLimitNode = &node{}
}
