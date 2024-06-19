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

// WalkDeep does a depth-first traversal of all errors.
// Any ErrorGroup is traversed (after going deep).
// The visitor function can return true to end the traversal early
// In that case, WalkDeep will return true, otherwise false.
func WalkDeep(err error, visitor func(err error) bool) bool {
	// Go deep
	unErr := err
	for unErr != nil {
		if done := visitor(unErr); done {
			return true
		}
		unErr = Unwrap(unErr)
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
