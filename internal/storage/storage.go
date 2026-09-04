package storage

import "errors"

var (
	ErrUrlNotFound = errors.New("URL not found")
	ErrUrlExists   = errors.New("URL already exists")
)
