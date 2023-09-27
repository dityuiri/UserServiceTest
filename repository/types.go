// This file contains types that are used in the repository layer.
package repository

import "github.com/google/uuid"

type InsertUserInput struct {
	Id          uuid.UUID
	PhoneNumber string
	Name        string
	Password    string //hashed
}

type GetUserByPhoneNumberInput struct {
	PhoneNumber string
}

type GetUserByPhoneNumberOutput struct {
	Id                   uuid.UUID
	Name                 string
	Password             string
	NumOfSuccessfulLogin int
}
