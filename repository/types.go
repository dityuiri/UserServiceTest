// This file contains types that are used in the repository layer.
package repository

import (
	"database/sql"
	"github.com/google/uuid"
)

type InsertUserInput struct {
	Id          uuid.UUID
	PhoneNumber string
	Name        string
	Password    string //hashed
}

type GetUserByPhoneNumberInput struct {
	PhoneNumber string
}

type GetUserByIdInput struct {
	Id string
}

type GetUserByIdOutput struct {
	Id          uuid.UUID
	Name        string
	PhoneNumber string
}

type GetUserByPhoneNumberOutput struct {
	Id                   uuid.UUID
	Name                 string
	Password             string
	NumOfSuccessfulLogin sql.NullInt32
}

type UpsertUserLoginInput struct {
	UserId               uuid.UUID
	NumOfSuccessfulLogin int32
}

type UpdateUserInput struct {
	Id          string
	PhoneNumber string
	Name        string
}
