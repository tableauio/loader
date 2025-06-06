package redblacktree

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

// Refer: https://github.com/emirpasic/gods/blob/master/trees/redblacktree/redblacktree.go

type color bool

const (
	black, red color = true, false
)

// Tree holds elements of the red-black tree
type Tree[K constraints.Ordered, V any] struct {
	Root *Node[K, V]
	size int
}

// Node is a single element within the tree
type Node[K constraints.Ordered, V any] struct {
	Key    K
	Value  V
	color  color
	Left   *Node[K, V]
	Right  *Node[K, V]
	Parent *Node[K, V]
}

func New[K constraints.Ordered, V any]() *Tree[K, V] {
	return &Tree[K, V]{}
}

// Put inserts node into the tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Put(key K, value V) {
	var insertedNode *Node[K, V]
	if tree.Root == nil {
		tree.Root = &Node[K, V]{Key: key, Value: value, color: red}
		insertedNode = tree.Root
	} else {
		node := tree.Root
		loop := true
		for loop {
			switch {
			case key == node.Key:
				node.Key = key
				node.Value = value
				return
			case key < node.Key:
				if node.Left == nil {
					node.Left = &Node[K, V]{Key: key, Value: value, color: red}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case key > node.Key:
				if node.Right == nil {
					node.Right = &Node[K, V]{Key: key, Value: value, color: red}
					insertedNode = node.Right
					loop = false
				} else {
					node = node.Right
				}
			}
		}
		insertedNode.Parent = node
	}
	tree.insertCase1(insertedNode)
	tree.size++
}

// Get searches the node in the tree by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Get(key K) (value V, found bool) {
	node := tree.lookup(key)
	if node != nil {
		return node.Value, true
	}
	return value, false
}

// GetNode searches the node in the tree by key and returns its node or nil if key is not found in tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) GetNode(key K) *Node[K, V] {
	return tree.lookup(key)
}

// Remove remove the node from the tree by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Remove(key K) {
	var child *Node[K, V]
	node := tree.lookup(key)
	if node == nil {
		return
	}
	if node.Left != nil && node.Right != nil {
		pred := node.Left.maximumNode()
		node.Key = pred.Key
		node.Value = pred.Value
		node = pred
	}
	if node.Left == nil || node.Right == nil {
		if node.Right == nil {
			child = node.Left
		} else {
			child = node.Right
		}
		if node.color == black {
			node.color = child.nodeColor()
			tree.deleteCase1(node)
		}
		tree.replaceNode(node, child)
		if node.Parent == nil && child != nil {
			child.color = black
		}
	}
	tree.size--
}

// Empty returns true if tree does not contain any nodes
func (tree *Tree[K, V]) Empty() bool {
	return tree.size == 0
}

// Size returns number of nodes in the tree.
func (tree *Tree[K, V]) Size() int {
	return tree.size
}

// Size returns the number of elements stored in the subtree.
// Computed dynamically on each call, i.e. the subtree is traversed to count the number of the nodes.
func (node *Node[K, V]) Size() int {
	if node == nil {
		return 0
	}
	size := 1
	if node.Left != nil {
		size += node.Left.Size()
	}
	if node.Right != nil {
		size += node.Right.Size()
	}
	return size
}

// Keys returns all keys in-order
func (tree *Tree[K, V]) Keys() []K {
	keys := make([]K, tree.size)
	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

// Values returns all values in-order based on the key.
func (tree *Tree[K, V]) Values() []V {
	values := make([]V, tree.size)
	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values
}

// Left returns the left-most (min) node or nil if tree is empty.
func (tree *Tree[K, V]) Left() *Node[K, V] {
	var parent *Node[K, V]
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Left
	}
	return parent
}

// Right returns the right-most (max) node or nil if tree is empty.
func (tree *Tree[K, V]) Right() *Node[K, V] {
	var parent *Node[K, V]
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Right
	}
	return parent
}

// Floor Finds floor node of the input key, return the floor node or nil if no floor is found.
// Second return parameter is true if floor was found, otherwise false.
//
// Floor node is defined as the largest node that is smaller than or equal to the given node.
// A floor node may not be found, either because the tree is empty, or because
// all nodes in the tree are larger than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Floor(key K) (floor *Node[K, V], found bool) {
	found = false
	node := tree.Root
	for node != nil {
		switch {
		case key == node.Key:
			return node, true
		case key < node.Key:
			node = node.Left
		case key > node.Key:
			floor, found = node, true
			node = node.Right
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

// Ceiling finds ceiling node of the input key, return the ceiling node or nil if no ceiling is found.
// Second return parameter is true if ceiling was found, otherwise false.
//
// Ceiling node is defined as the smallest node that is larger than or equal to the given node.
// A ceiling node may not be found, either because the tree is empty, or because
// all nodes in the tree are smaller than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Ceiling(key K) (ceiling *Node[K, V], found bool) {
	found = false
	node := tree.Root
	for node != nil {
		switch {
		case key == node.Key:
			return node, true
		case key < node.Key:
			ceiling, found = node, true
			node = node.Left
		case key > node.Key:
			node = node.Right
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

// Clear removes all nodes from the tree.
func (tree *Tree[K, V]) Clear() {
	tree.Root = nil
	tree.size = 0
}

// String returns a string representation of container
func (tree *Tree[K, V]) String() string {
	str := "RedBlackTree\n"
	if !tree.Empty() {
		tree.Root.output("", true, &str)
	}
	return str
}

func (node *Node[K, V]) String() string {
	return fmt.Sprintf("%v", node.Key)
}

func (node *Node[K, V]) output(prefix string, isTail bool, str *string) {
	if node.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		node.Right.output(newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		node.Left.output(newPrefix, true, str)
	}
}

func (tree *Tree[K, V]) lookup(key K) *Node[K, V] {
	node := tree.Root
	for node != nil {
		switch {
		case key == node.Key:
			return node
		case key < node.Key:
			node = node.Left
		case key > node.Key:
			node = node.Right
		}
	}
	return nil
}

func (node *Node[K, V]) grandparent() *Node[K, V] {
	if node != nil && node.Parent != nil {
		return node.Parent.Parent
	}
	return nil
}

func (node *Node[K, V]) uncle() *Node[K, V] {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

func (node *Node[K, V]) sibling() *Node[K, V] {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

func (tree *Tree[K, V]) rotateLeft(node *Node[K, V]) {
	right := node.Right
	tree.replaceNode(node, right)
	node.Right = right.Left
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node
	node.Parent = right
}

func (tree *Tree[K, V]) rotateRight(node *Node[K, V]) {
	left := node.Left
	tree.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (tree *Tree[K, V]) replaceNode(old *Node[K, V], new *Node[K, V]) {
	if old.Parent == nil {
		tree.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}

func (tree *Tree[K, V]) insertCase1(node *Node[K, V]) {
	if node.Parent == nil {
		node.color = black
	} else {
		tree.insertCase2(node)
	}
}

func (tree *Tree[K, V]) insertCase2(node *Node[K, V]) {
	if node.Parent.nodeColor() == black {
		return
	}
	tree.insertCase3(node)
}

func (tree *Tree[K, V]) insertCase3(node *Node[K, V]) {
	uncle := node.uncle()
	if uncle.nodeColor() == red {
		node.Parent.color = black
		uncle.color = black
		node.grandparent().color = red
		tree.insertCase1(node.grandparent())
	} else {
		tree.insertCase4(node)
	}
}

func (tree *Tree[K, V]) insertCase4(node *Node[K, V]) {
	grandparent := node.grandparent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		tree.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		tree.rotateRight(node.Parent)
		node = node.Right
	}
	tree.insertCase5(node)
}

func (tree *Tree[K, V]) insertCase5(node *Node[K, V]) {
	node.Parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.Parent.Left && node.Parent == grandparent.Left {
		tree.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		tree.rotateLeft(grandparent)
	}
}

func (node *Node[K, V]) maximumNode() *Node[K, V] {
	if node == nil {
		return nil
	}
	for node.Right != nil {
		node = node.Right
	}
	return node
}

func (tree *Tree[K, V]) deleteCase1(node *Node[K, V]) {
	if node.Parent == nil {
		return
	}
	tree.deleteCase2(node)
}

func (tree *Tree[K, V]) deleteCase2(node *Node[K, V]) {
	sibling := node.sibling()
	if sibling.nodeColor() == red {
		node.Parent.color = red
		sibling.color = black
		if node == node.Parent.Left {
			tree.rotateLeft(node.Parent)
		} else {
			tree.rotateRight(node.Parent)
		}
	}
	tree.deleteCase3(node)
}

func (tree *Tree[K, V]) deleteCase3(node *Node[K, V]) {
	sibling := node.sibling()
	if node.Parent.nodeColor() == black &&
		sibling.nodeColor() == black &&
		sibling.Left.nodeColor() == black &&
		sibling.Right.nodeColor() == black {
		sibling.color = red
		tree.deleteCase1(node.Parent)
	} else {
		tree.deleteCase4(node)
	}
}

func (tree *Tree[K, V]) deleteCase4(node *Node[K, V]) {
	sibling := node.sibling()
	if node.Parent.nodeColor() == red &&
		sibling.nodeColor() == black &&
		sibling.Left.nodeColor() == black &&
		sibling.Right.nodeColor() == black {
		sibling.color = red
		node.Parent.color = black
	} else {
		tree.deleteCase5(node)
	}
}

func (tree *Tree[K, V]) deleteCase5(node *Node[K, V]) {
	sibling := node.sibling()
	if node == node.Parent.Left &&
		sibling.nodeColor() == black &&
		sibling.Left.nodeColor() == red &&
		sibling.Right.nodeColor() == black {
		sibling.color = red
		sibling.Left.color = black
		tree.rotateRight(sibling)
	} else if node == node.Parent.Right &&
		sibling.nodeColor() == black &&
		sibling.Right.nodeColor() == red &&
		sibling.Left.nodeColor() == black {
		sibling.color = red
		sibling.Right.color = black
		tree.rotateLeft(sibling)
	}
	tree.deleteCase6(node)
}

func (tree *Tree[K, V]) deleteCase6(node *Node[K, V]) {
	sibling := node.sibling()
	sibling.color = node.Parent.nodeColor()
	node.Parent.color = black
	if node == node.Parent.Left && sibling.Right.nodeColor() == red {
		sibling.Right.color = black
		tree.rotateLeft(node.Parent)
	} else if sibling.Left.nodeColor() == red {
		sibling.Left.color = black
		tree.rotateRight(node.Parent)
	}
}

func (node *Node[K, V]) nodeColor() color {
	if node == nil {
		return black
	}
	return node.color
}
