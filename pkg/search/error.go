package search

// NotFoundError indicates an error case during attempts to modify a resource
// which does not exist in a search engine.
type NotFoundError struct{}

// NewNotFoundError initializes a new instance of NotFoundError.
func NewNotFoundError() *NotFoundError {
	return &NotFoundError{}
}

// Error implements error interface's Error function.
func (e NotFoundError) Error() string {
	return "resource is not found"
}

// DuplicateError indicates an error case during attempts to create a resource
// which already exists in a search engine.
type DuplicateError struct{}

// NewDuplicateError initializes a new instance of DuplicateError.
func NewDuplicateError() *DuplicateError {
	return &DuplicateError{}
}

// Error implements error interface's Error function.
func (e DuplicateError) Error() string {
	return "resource is duplicated"
}
