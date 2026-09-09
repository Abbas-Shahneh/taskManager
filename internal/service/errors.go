package service

import "errors"

var (
	ErrInvalidTaskID   = errors.New("invalid task id")
	ErrInvalidPage     = errors.New("invalid page")
	ErrInvalidPageSize = errors.New("invalid page size")
)
