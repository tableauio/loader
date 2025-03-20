package redblacktree

import "encoding/json"

// Refer: https://github.com/emirpasic/gods/blob/master/trees/redblacktree/serialization.go

// ToJSON outputs the JSON representation of the tree.
func (tree *Tree[K, V]) ToJSON() ([]byte, error) {
	elements := make(map[K]V)
	it := tree.Iterator()
	for it.Next() {
		elements[it.Key()] = it.Value()
	}
	return json.Marshal(&elements)
}

// FromJSON populates the tree from the input JSON representation.
func (tree *Tree[K, V]) FromJSON(data []byte) error {
	elements := make(map[K]V)
	err := json.Unmarshal(data, &elements)
	if err == nil {
		tree.Clear()
		for key, value := range elements {
			tree.Put(key, value)
		}
	}
	return err
}

// UnmarshalJSON @implements json.Unmarshaler
func (tree *Tree[K, V]) UnmarshalJSON(bytes []byte) error {
	return tree.FromJSON(bytes)
}

// MarshalJSON @implements json.Marshaler
func (tree *Tree[K, V]) MarshalJSON() ([]byte, error) {
	return tree.ToJSON()
}
