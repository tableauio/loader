package treemap

import (
	rbt "github.com/tableauio/loader/pkg/treemap/redblacktree"
	"golang.org/x/exp/constraints"
)

// Refer: https://github.com/emirpasic/gods/blob/master/maps/treemap/iterator.go

type TreeMapIterator[K constraints.Ordered, V any] struct {
	iterator *rbt.Iterator[K, V]
}

// Iterator returns a stateful iterator whose elements are key/value pairs.
func (m *TreeMap[K, V]) Iterator() *TreeMapIterator[K, V] {
	return &TreeMapIterator[K, V]{iterator: m.tree.Iterator()}
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
