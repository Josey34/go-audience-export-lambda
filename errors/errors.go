package errors

import "errors"

var ErrUnauthorized = errors.New("unauthorized")
var ErrInvalidInput = errors.New("invalid input")
var ErrJobNotFound = errors.New("job not found")
