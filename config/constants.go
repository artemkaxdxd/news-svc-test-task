package config

import "errors"

var ( // Errors
	ErrEmptyTitle   = errors.New("post title cannot be empty")
	ErrEmptyContent = errors.New("post content cannot be empty")
	ErrPostNotFound = errors.New("post not found")
)
