//go:build go1.26

package redblacktree

type Lesser[T Lesser[T]] interface {
	comparable
	Less(other T) bool
}
