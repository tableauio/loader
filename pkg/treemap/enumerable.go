package treemap

// Refer: https://github.com/emirpasic/gods/blob/master/maps/treemap/enumerable.go

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
// the first (key,value) for which the function is true or nil,nil otherwise if no element
// matches the criteria.
func (m *TreeMap[K, V]) Find(f func(key K, value V) bool) (k K, v V) {
	iterator := m.Iterator()
	for iterator.Next() {
		if f(iterator.Key(), iterator.Value()) {
			return iterator.Key(), iterator.Value()
		}
	}
	return k, v
}
