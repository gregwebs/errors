//go:build go1.23

package errwrap

import (
	"iter"
)

// UnwrapGroups will unwrap errors visiting each one.
// An error that defines Unwrap() []error is expanded and traversed
// This is a depth-first traversal, doing the unwrap first and the expansion second.
// This can be used for functionality similar to errors.As but it also expands error groups.
func UnwrapGroups(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		_ = WalkDeep(err, func(e error) bool {
			return !yield(e)
		})
	}
}

// UnwrapGroupsStack is similar to UnwrapGroups.
// It adds a second parameter the level of depth in the error tree.
func UnwrapGroupsLevel(err error) iter.Seq2[int, error] {
	return func(yield func(int, error) bool) {
		_ = WalkDeepLevel(err, func(e error, level int) bool {
			return !yield(level, e)
		})
	}
}
