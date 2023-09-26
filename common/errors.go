// Package common errors: Error variables that can be shared across layer
package common

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
)
