package errors

// ErrorGroup is an interface for multiple errors that are not a chain.
// This happens for example when executing multiple operations in parallel.
type ErrorGroup interface {
	Errors() []error
}

// If the error is an ErrorGroup, return its list of errors
// If it is not an error group, the list will contain just itself.
// If nil return nil
func Errors(err error) []error {
	if err == nil {
		return nil
	}
	if group, ok := err.(ErrorGroup); ok {
		errors := group.Errors()
		result := make([]error, len(errors))
		copy(result, errors)
		return result
	} else {
		return []error{err}
	}
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
	if group, ok := err.(ErrorGroup); ok {
		for _, err := range group.Errors() {
			if early := WalkDeep(err, visitor); early {
				return true
			}
		}
	}

	return false
}
