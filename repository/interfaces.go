package repository

import "context"

type RepositoryInterface interface {
	GetUserByPhoneNumber(ctx context.Context, input GetUserByPhoneNumberInput) (output GetUserByPhoneNumberOutput, err error)
	InsertUser(ctx context.Context, input InsertUserInput) (err error)
	UpdateUserLogin(ctx context.Context, input UpdateUserLoginInput) (err error)
}
