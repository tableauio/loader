# treemap

This implements a generic-type-supported sorted map with rbtree.

Refer:
- https://github.com/emirpasic/gods#treemap
- https://github.com/emirpasic/gods#redblacktree

## Usage

```go
package main

import "github.com/tableauio/loader/pkg/treemap"

func main() {
	m := treemap.New[int, string]() // empty
	m.Put(1, "x")                   // 1->x
	m.Put(2, "b")                   // 1->x, 2->b (in order)
	m.Put(1, "a")                   // 1->a, 2->b (in order)
	_, _ = m.Get(2)                 // b, true
	_, _ = m.Get(3)                 // "", false
	_ = m.Values()                  // [a b] (in order)
	_ = m.Keys()                    // [1 2] (in order)
	m.Remove(1)                     // 2->b
	m.Clear()                       // empty
	m.Empty()                       // true
	m.Size()                        // 0

	// Other:
	m.Min() // Returns the minimum key and its value from map.
	m.Max() // Returns the maximum key and its value from map.
	m.Range(func(key int, value string) bool{
		if key == 2 {
			return false
		}
		return true
	})
}
```

## Iterate

```go
package main

import (
	"github.com/tableauio/loader/pkg/treemap"
)

func main() {
	m := treemap.New[int, string]()
	m.Put(2, "b")
	m.Put(1, "a")
	m.Put(3, "c")

	// iterate
	iter := m.Iterator()
	for iter.Begin(); iter.Next(); {
		_, _ = iter.Key(), iter.Value() // 1->a, 2->b, 3->c
	}

	// iterate in reverse order
	for iter.End(); iter.Prev(); {
		_, _ = iter.Key(), iter.Value() // 3->c, 2->b, 1->a
	}

	iter = m.LowerBound(0)
	if !iter.IsEnd() {
		_, _ = iter.Key(), iter.Value() // 1->a
	}
	iter = m.LowerBound(2)
	if !iter.IsEnd() {
		_, _ = iter.Key(), iter.Value() // 2->b
	}
	iter = m.LowerBound(4)
	if !iter.IsEnd() {
		_, _ = iter.Key(), iter.Value() // panic if code reaches here
	}
}
```