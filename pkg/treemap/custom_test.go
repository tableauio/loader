package treemap

import "testing"

func TestMapRange(t *testing.T) {
	m := New[int, string]()

	m.Put(5, "e")
	m.Put(6, "f")
	m.Put(7, "g")
	m.Put(3, "c")
	m.Put(4, "d")
	m.Put(1, "x")
	m.Put(2, "b")

	expectedKey := 1
	m.Range(func(key int, value string) bool {
		if key != expectedKey {
			t.Errorf("[Range] expected %d, got %d", expectedKey, key)
			return false
		}
		expectedKey++
		return true
	})

	expectedKey = 7
	m.ReverseRange(func(key int, value string) bool {
		if key != expectedKey {
			t.Errorf("[ReverseRange] expected %d, got %d", expectedKey, key)
			return false
		}
		expectedKey--
		return true
	})
}
