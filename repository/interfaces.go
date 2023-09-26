package repository

import "context"

type RepositoryInterface interface {
	GetUserByPhoneNumber(ctx context.Context, input GetUserByPhoneNumberInput) (output GetUserByPhoneNumberOutput, err error)
	InsertUser(ctx context.Context, input InsertUserInput) (err error)
}
