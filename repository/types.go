// This file contains types that are used in the repository layer.
package repository

import "github.com/google/uuid"

type InsertUserInput struct {
	Id          uuid.UUID
	PhoneNumber string
	Name        string
	Password    string
	Salt        string
}

type GetUserByPhoneNumberInput struct {
	PhoneNumber string
}

type GetUserByPhoneNumberOutput struct {
	Id   uuid.UUID
	Name string
}
