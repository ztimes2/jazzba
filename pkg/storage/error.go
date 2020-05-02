package storage

// NotFoundError indicates an error case during attempts to fetch a resource
// which does not exist in a data store.
type NotFoundError struct{}

// NewNotFoundError initializes a new instance of NotFoundError.
func NewNotFoundError() *NotFoundError {
	return &NotFoundError{}
}

// Error implements error interface's Error function.
func (e NotFoundError) Error() string {
	return "resource is not found"
}

// DuplicateError indicates an error case during attempts to create or update a
// resource using data which breaks uniqueness of other existing resource in a
// data store.
// (e.g. unique constraint violation)
type DuplicateError struct{}

// NewDuplicateError initializes a new instance of DuplicateError.
func NewDuplicateError() *DuplicateError {
	return &DuplicateError{}
}

// Error implements error interface's Error function.
func (e DuplicateError) Error() string {
	return "resource is duplicated"
}

// ReferenceError indicates an error case during attempts to reference a resource
// associated with some other resource in a data store using an invalid identifier.
// (e.g. foreign key violation)
type ReferenceError struct{}

// NewReferenceError initializes a new instance of ReferenceError.
func NewReferenceError() *ReferenceError {
	return &ReferenceError{}
}

// Error implements error interface's Error function.
func (e ReferenceError) Error() string {
	return "resource is referenced using invalid identifier"
}
