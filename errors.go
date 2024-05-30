package ipam

import "errors"

var (
	// ErrNotFound is returned if prefix or cidr was not found
	ErrNotFound = errors.New("NotFound")
	// ErrNoIPAvailable is returned if no IP is available anymore
	ErrNoIPAvailable = errors.New("NoIPAvailableError")
	// ErrAlreadyAllocated is returned if the requested address is not available
	ErrAlreadyAllocated = errors.New("AlreadyAllocatedError")
	// ErrNamespaceDoesNotExist is returned when an operation is perfomed in a namespace that does not exist.
	ErrNamespaceDoesNotExist = errors.New("NamespaceDoesNotExist")
	// ErrNamespaceExist is returned when an operation is perfomed in a namespace that exist.
	ErrNamespaceExist = errors.New("NamespaceDoesExist")
	// ErrNameTooLong is returned when a name exceeds the databases max identifier length
	ErrNameTooLong = errors.New("NameTooLong")
	//ErrNamespaceInconsistent is returned when namespace provided inconsistent with namespace in context
	ErrNamespaceInconsistent = errors.New("prefix namespace inconsistent with context")

	//ErrDbNil is returned when db is nil
	ErrDbNil = errors.New("db is nil")
)
