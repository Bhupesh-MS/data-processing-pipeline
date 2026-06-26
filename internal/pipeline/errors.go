package pipeline

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrProcessingFailed = errors.New("processing failed")
)
