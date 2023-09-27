package handler

import "golang.org/x/crypto/bcrypt"

// Function aliasing for helping cover unit test
var (
	GenerateFromPassword   = bcrypt.GenerateFromPassword
	CompareHashAndPassword = bcrypt.CompareHashAndPassword
)
