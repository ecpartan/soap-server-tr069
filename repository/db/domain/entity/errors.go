package entity

import "errors"

// ErrNotFound not found
var ErrNotFound = errors.New("Not found")

// ErrInvalidEntity invalid entity
var ErrInvalidEntity = errors.New("Invalid entity")

// ErrCannotBeDeleted cannot be deleted
var ErrCannotBeDeleted = errors.New("Cannot Be Deleted")

// ErrCannotBeDeleted cannot be deleted
var ErrCRULR = errors.New("Connection request URL is invalid")
