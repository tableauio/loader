package treemap

import (
	"cmp"
	"fmt"
	"strings"

	rbt "github.com/tableauio/loader/pkg/treemap/redblacktree"
)

type TreeMap[K, V any] struct {
	tree *rbt.Tree[K, V]
}

type Ordered[T any] interface {
	Less(other T) bool
}

func New[K cmp.Ordered, V any]() *TreeMap[K, V] {
	return &TreeMap[K, V]{tree: rbt.New[K, V]()}
}

func New2[K Ordered[K], V any]() *TreeMap[K, V] {
	return &TreeMap[K, V]{tree: rbt.New2[K, V]()}
}

func new3[K, V any](less func(K, K) bool) *TreeMap[K, V] {
	return &TreeMap[K, V]{tree: rbt.New3[K, V](less)}
}

// Put inserts key-value pair into the map.
func (m *TreeMap[K, V]) Put(key K, value V) {
	m.tree.Put(key, value)
}

// Get searches the element in the map by key and returns its value or empty value if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
func (m *TreeMap[K, V]) Get(key K) (value V, found bool) {
	return m.tree.Get(key)
}

// Remove removes the element from the map by key.
func (m *TreeMap[K, V]) Remove(key K) {
	m.tree.Remove(key)
}

// Empty returns true if map does not contain any elements.
func (m *TreeMap[K, V]) Empty() bool {
	return m.tree.Empty()
}

// Size returns number of elements in the map.
func (m *TreeMap[K, V]) Size() int {
	return m.tree.Size()
}

// Keys returns all keys in-order.
func (m *TreeMap[K, V]) Keys() []K {
	return m.tree.Keys()
}

// Values returns all values in-order based on the key.
func (m *TreeMap[K, V]) Values() []V {
	return m.tree.Values()
}

// Clear removes all elements from the map.
func (m *TreeMap[K, V]) Clear() {
	m.tree.Clear()
}

// Min returns the minimum key and its value from the tree map.
// If the map is empty, the third return parameter will be false.
func (m *TreeMap[K, V]) Min() (key K, value V, ok bool) {
	if node := m.tree.Left(); node != nil {
		return node.Key, node.Value, true
	}
	return key, value, false
}

// Max returns the maximum key and its value from the tree map.
// If the map is empty, the third return parameter will be false.
func (m *TreeMap[K, V]) Max() (key K, value V, ok bool) {
	if node := m.tree.Right(); node != nil {
		return node.Key, node.Value, true
	}
	return key, value, false
}

// Floor finds the floor key-value pair for the input key.
// In case that no floor is found, then both returned values will be corresponding type's empty value.
//
// Floor key is defined as the greatest key that is less than or equal to the given key.
// A floor key may not be found, either because the map is empty, or because
// all keys in the map are greater than the given key.
func (m *TreeMap[K, V]) Floor(key K) (foundKey K, foundValue V, ok bool) {
	node, found := m.tree.Floor(key)
	if found {
		return node.Key, node.Value, true
	}
	return foundKey, foundValue, false
}

// Ceiling finds the ceiling key-value pair for the input key.
// In case that no ceiling is found, then both returned values will be corresponding type's empty value.
//
// Ceiling key is defined as the least key that is greater than or equal to the given key.
// A ceiling key may not be found, either because the map is empty, or because
// all keys in the map are less than the given key.
func (m *TreeMap[K, V]) Ceiling(key K) (foundKey K, foundValue V, ok bool) {
	node, found := m.tree.Ceiling(key)
	if found {
		return node.Key, node.Value, true
	}
	return foundKey, foundValue, false
}

// String returns a string representation of container
func (m *TreeMap[K, V]) String() string {
	str := "TreeMap\nmap["
	it := m.Iterator()
	for it.Next() {
		str += fmt.Sprintf("%v:%v ", it.Key(), it.Value())
	}
	return strings.TrimRight(str, " ") + "]"
}
