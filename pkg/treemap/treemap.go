package treemap

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

type TreeMap[K constraints.Ordered, V any] struct {
	tree *Tree[K, V]
}

func New[K constraints.Ordered, V any]() *TreeMap[K, V] {
	return &TreeMap[K, V]{tree: &Tree[K, V]{}}
}

// Put inserts key-value pair into the map.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (m *TreeMap[K, V]) Put(key K, value V) {
	m.tree.Put(key, value)
}

// Get searches the element in the map by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (m *TreeMap[K, V]) Get(key K) (value V, found bool) {
	return m.tree.Get(key)
}

// Remove removes the element from the map by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (m *TreeMap[K, V]) Remove(key K) {
	m.tree.Remove(key)
}

// Empty returns true if map does not contain any elements
func (m *TreeMap[K, V]) Empty() bool {
	return m.tree.Empty()
}

// Size returns number of elements in the map.
func (m *TreeMap[K, V]) Size() int {
	return m.tree.Size()
}

// Keys returns all keys in-order
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

// UpperBound returns an iterator pointing to the first element that is greater than key.
// If no such element is found, a past-the-end iterator is returned.
// See: https://en.cppreference.com/w/cpp/container/map/upper_bound
func (m *TreeMap[K, V]) UpperBound(key K) TreeMapIterator[K, V] {
	iter := m.tree.Iterator()
	iter.Begin()
	node, found := m.tree.Floor(key)
	if found {
		iter = m.tree.IteratorAt(node)
	}
	iter.Next()
	return TreeMapIterator[K, V]{iter}
}

// LowerBound returns an iterator pointing to the first element that is not less than key.
// If no such element is found, a past-the-end iterator is returned.
// See: https://en.cppreference.com/w/cpp/container/map/lower_bound
func (m *TreeMap[K, V]) LowerBound(key K) TreeMapIterator[K, V] {
	iter := m.tree.Iterator()
	iter.End()
	node, found := m.tree.Ceiling(key)
	if found {
		iter = m.tree.IteratorAt(node)
	}
	return TreeMapIterator[K, V]{iter}
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

// Iterator returns a stateful iterator whose elements are key/value pairs.
func (m *TreeMap[K, V]) Iterator() TreeMapIterator[K, V] {
	return TreeMapIterator[K, V]{iterator: m.tree.Iterator()}
}

// Each calls the given function once for each element, passing that element's key and value.
func (m *TreeMap[K, V]) Each(f func(key K, value V)) {
	iterator := m.Iterator()
	for iterator.Next() {
		f(iterator.Key(), iterator.Value())
	}
}

// Map invokes the given function once for each element and returns a container
// containing the values returned by the given function as key/value pairs.
func (m *TreeMap[K, V]) Map(f func(key1 K, value1 V) (K, V)) *TreeMap[K, V] {
	newMap := New[K, V]()
	iterator := m.Iterator()
	for iterator.Next() {
		key2, value2 := f(iterator.Key(), iterator.Value())
		newMap.Put(key2, value2)
	}
	return newMap
}

// Select returns a new container containing all elements for which the given function returns a true value.
func (m *TreeMap[K, V]) Select(f func(key K, value V) bool) *TreeMap[K, V] {
	newMap := New[K, V]()
	iterator := m.Iterator()
	for iterator.Next() {
		if f(iterator.Key(), iterator.Value()) {
			newMap.Put(iterator.Key(), iterator.Value())
		}
	}
	return newMap
}

// Any passes each element of the container to the given function and
// returns true if the function ever returns true for any element.
func (m *TreeMap[K, V]) Any(f func(key K, value V) bool) bool {
	iterator := m.Iterator()
	for iterator.Next() {
		if f(iterator.Key(), iterator.Value()) {
			return true
		}
	}
	return false
}

// All passes each element of the container to the given function and
// returns true if the function returns true for all elements.
func (m *TreeMap[K, V]) All(f func(key K, value V) bool) bool {
	iterator := m.Iterator()
	for iterator.Next() {
		if !f(iterator.Key(), iterator.Value()) {
			return false
		}
	}
	return true
}

// Find passes each element of the container to the given function and returns
// iterator to the first element for which the function is true.
// If no element matches the criteria, this returns the iterator past the last element (one-past-the-end).
func (m *TreeMap[K, V]) Find(f func(key K, value V) bool) TreeMapIterator[K, V] {
	iterator := m.Iterator()
	for iterator.Next() {
		if f(iterator.Key(), iterator.Value()) {
			return iterator
		}
	}
	iter := m.tree.Iterator()
	iter.End()
	return TreeMapIterator[K, V]{iter}
}

// ToJSON outputs the JSON representation of the map.
func (m *TreeMap[K, V]) ToJSON() ([]byte, error) {
	return m.tree.ToJSON()
}

// FromJSON populates the map from the input JSON representation.
func (m *TreeMap[K, V]) FromJSON(data []byte) error {
	return m.tree.FromJSON(data)
}

// UnmarshalJSON @implements json.Unmarshaler
func (m *TreeMap[K, V]) UnmarshalJSON(bytes []byte) error {
	return m.FromJSON(bytes)
}

// MarshalJSON @implements json.Marshaler
func (m *TreeMap[K, V]) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

type TreeMapIterator[K constraints.Ordered, V any] struct {
	iterator Iterator[K, V]
}

// Next moves the iterator to the next element and returns true if there was a next element in the container.
// If Next() returns true, then next element's key and value can be retrieved by Key() and Value().
// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
// Modifies the state of the iterator.
func (iterator *TreeMapIterator[K, V]) Next() bool {
	return iterator.iterator.Next()
}

// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
// If Prev() returns true, then previous element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *TreeMapIterator[K, V]) Prev() bool {
	return iterator.iterator.Prev()
}

// IsBegin returns true if the iterator is in initial state (one-before-first)
func (iterator *TreeMapIterator[K, V]) IsBegin() bool {
	return iterator.iterator.position == begin
}

// IsEnd returns true if the iterator is past the last element (one-past-the-end).
func (iterator *TreeMapIterator[K, V]) IsEnd() bool {
	return iterator.iterator.position == end
}

// Value returns the current element's value.
// Does not modify the state of the iterator.
func (iterator *TreeMapIterator[K, V]) Value() V {
	return iterator.iterator.Value()
}

// Key returns the current element's key.
// Does not modify the state of the iterator.
func (iterator *TreeMapIterator[K, V]) Key() K {
	return iterator.iterator.Key()
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (iterator *TreeMapIterator[K, V]) Begin() {
	iterator.iterator.Begin()
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (iterator *TreeMapIterator[K, V]) End() {
	iterator.iterator.End()
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator
func (iterator *TreeMapIterator[K, V]) First() bool {
	return iterator.iterator.First()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *TreeMapIterator[K, V]) Last() bool {
	return iterator.iterator.Last()
}

// NextTo moves the iterator to the next element from current position that satisfies the condition given by the
// passed function, and returns true if there was a next element in the container.
// If NextTo() returns true, then next element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *TreeMapIterator[K, V]) NextTo(f func(key K, value V) bool) bool {
	for iterator.Next() {
		key, value := iterator.Key(), iterator.Value()
		if f(key, value) {
			return true
		}
	}
	return false
}

// PrevTo moves the iterator to the previous element from current position that satisfies the condition given by the
// passed function, and returns true if there was a next element in the container.
// If PrevTo() returns true, then next element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *TreeMapIterator[K, V]) PrevTo(f func(key K, value V) bool) bool {
	for iterator.Prev() {
		key, value := iterator.Key(), iterator.Value()
		if f(key, value) {
			return true
		}
	}
	return false
}
