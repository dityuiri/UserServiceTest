package repository

import "context"

type RepositoryInterface interface {
	GetUserByPhoneNumber(ctx context.Context, input GetUserByPhoneNumberInput) (output GetUserByPhoneNumberOutput, err error)
	GetUserById(ctx context.Context, input GetUserByIdInput) (output GetUserByIdOutput, err error)
	InsertUser(ctx context.Context, input InsertUserInput) (err error)
	UpsertUserLogin(ctx context.Context, input UpsertUserLoginInput) (err error)
}
