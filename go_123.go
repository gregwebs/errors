//go:build go1.23

package errors

import "iter"

// UnwrapGroups will unwrap errors visiting each one.
// Any ErrorGroup is expanded and traversed
// This is a depth-first traversal, doing the unwrap first and the expansion second.
// This can be used for functionality similar to errors.As but it also expands error groups.
func UnwrapGroups(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		_ = WalkDeep(err, func(e error) bool {
			return !yield(e)
		})
	}
}
