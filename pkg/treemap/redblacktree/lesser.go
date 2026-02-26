//go:build !go1.26

package redblacktree

type Lesser[T any] interface {
	comparable
	Less(other T) bool
}
