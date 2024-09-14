package errors

// ErrorGroup is an interface for multiple errors that are not a chain.
// This happens for example when executing multiple operations in parallel.
// Also can now define Unwraps() []error
type ErrorGroup interface {
	Errors() []error
}

type unwraps interface {
	Unwrap() []error
}

// If the error is not nil and an ErrorGroup or satisfies Unwraps() []error, return its list of errors
// otherwise return nil
func Errors(err error) []error {
	if err == nil {
		return nil
	}
	if group, ok := err.(ErrorGroup); ok {
		return group.Errors()
	} else if group, ok := err.(unwraps); ok {
		return group.Unwrap()
	}
	return nil
}

// Deprecated: WalkDeep was created before iterators.
// UnwrapGroups is now preferred for those using Go version >= 1.23.
// Note that WalkDeep uses the opposite convention for boolean return values compared to golang iterators.
// WalkDeep does a depth-first traversal of all errors.
// Any ErrorGroup is traversed (after first unwrapping deeply).
// The visitor function can return true to end the traversal early
// If iteration is ended early, WalkDeep will return true, otherwise false.
func WalkDeep(err error, visitor func(err error) bool) bool {
	if err == nil {
		return false
	}
	if done := visitor(err); done {
		return true
	}
	if done := WalkDeep(Unwrap(err), visitor); done {
		return true
	}

	// Go wide
	if errors := Errors(err); len(errors) > 0 {
		for _, err := range errors {
			if early := WalkDeep(err, visitor); early {
				return true
			}
		}
	}

	return false
}

func walkDeepStack(err error, visitor func(error, int) bool, stack int) bool {
	if err == nil {
		return false
	}
	if done := visitor(err, stack); done {
		return true
	}
	if done := walkDeepStack(Unwrap(err), visitor, stack+1); done {
		return true
	}

	// Go wide
	if errors := Errors(err); len(errors) > 0 {
		for _, err := range errors {
			if early := walkDeepStack(err, visitor, stack+1); early {
				return true
			}
		}
	}

	return false
}
