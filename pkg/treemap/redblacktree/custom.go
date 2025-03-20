package redblacktree

// IsBegin returns true if the iterator is in initial state (one-before-first)
func (iterator *Iterator[K, V]) IsBegin() bool {
	return iterator.position == begin
}

// IsEnd returns true if the iterator is past the last element (one-past-the-end).
func (iterator *Iterator[K, V]) IsEnd() bool {
	return iterator.position == end
}
