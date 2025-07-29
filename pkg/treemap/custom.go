package treemap

// IsBegin returns true if the iterator is in initial state (one-before-first)
func (iterator *TreeMapIterator[K, V]) IsBegin() bool {
	return iterator.iterator.IsBegin()
}

// IsEnd returns true if the iterator is past the last element (one-past-the-end).
func (iterator *TreeMapIterator[K, V]) IsEnd() bool {
	return iterator.iterator.IsEnd()
}

// FindIter returns an iterator pointing to the element with specified key.
// If no such element is found, a past-the-end iterator is returned.
// See: https://en.cppreference.com/w/cpp/container/map/find
func (m *TreeMap[K, V]) FindIter(key K) *TreeMapIterator[K, V] {
	iter := m.tree.Iterator()
	iter.End()
	node := m.tree.GetNode(key)
	if node != nil {
		iter = m.tree.IteratorAt(node)
	}
	return &TreeMapIterator[K, V]{iter}
}

// UpperBound returns an iterator pointing to the first element that is greater than key.
// If no such element is found, a past-the-end iterator is returned.
// See: https://en.cppreference.com/w/cpp/container/map/upper_bound
func (m *TreeMap[K, V]) UpperBound(key K) *TreeMapIterator[K, V] {
	iter := m.tree.Iterator()
	node, found := m.tree.Floor(key)
	if found {
		iter = m.tree.IteratorAt(node)
	}
	iter.Next()
	return &TreeMapIterator[K, V]{iter}
}

// LowerBound returns an iterator pointing to the first element that is not less than key.
// If no such element is found, a past-the-end iterator is returned.
// See: https://en.cppreference.com/w/cpp/container/map/lower_bound
func (m *TreeMap[K, V]) LowerBound(key K) *TreeMapIterator[K, V] {
	iter := m.tree.Iterator()
	iter.End()
	node, found := m.tree.Ceiling(key)
	if found {
		iter = m.tree.IteratorAt(node)
	}
	return &TreeMapIterator[K, V]{iter}
}

// FloorOrMin finds the floor key-value pair for the input key.
// If no floor is found, returns the key-value pair for the least key.
// In case that map is empty, then both returned values will be corresponding type's empty value.
//
// Floor key is defined as the greatest key that is less than or equal to the given key.
// A floor key may not be found, either because the map is empty, or because
// all keys in the map are greater than the given key.
func (m *TreeMap[K, V]) FloorOrMin(key K) (foundKey K, foundValue V, ok bool) {
	node, found := m.tree.Floor(key)
	if found {
		return node.Key, node.Value, true
	}
	return m.Min()
}

// Ceiling finds the ceiling key-value pair for the input key.
// If no ceiling is found, returns the key-value pair for the greatest key.
// In case that map is empty, then both returned values will be corresponding type's empty value.
//
// Ceiling key is defined as the least key that is greater than or equal to the given key.
// A ceiling key may not be found, either because the map is empty, or because
// all keys in the map are less than the given key.
func (m *TreeMap[K, V]) CeilingOrMax(key K) (foundKey K, foundValue V, ok bool) {
	node, found := m.tree.Ceiling(key)
	if found {
		return node.Key, node.Value, true
	}
	return m.Max()
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *TreeMap[K, V]) Range(f func(key K, value V) bool) {
	iterator := m.Iterator()
	for iterator.Next() {
		if !f(iterator.Key(), iterator.Value()) {
			break
		}
	}
}
