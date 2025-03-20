package treemap

// Refer: https://github.com/emirpasic/gods/blob/master/maps/treemap/serialization.go

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
