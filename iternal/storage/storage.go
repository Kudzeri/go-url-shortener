package storage

import "errors"

var (
	ErrNotFound      = errors.New("URL not found")
	ErrAlreadyExists = errors.New("URL already exists")
)
