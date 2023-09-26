package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dityuiri/UserServiceTest/common"
)

func TestRepository_GetUserByPhoneNumber(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	ctx := context.Background()
	repo := &Repository{Db: db}

	t.Run("positive", func(t *testing.T) {
		var (
			input = GetUserByPhoneNumberInput{
				PhoneNumber: "+628159972915",
			}

			expectedOutput = GetUserByPhoneNumberOutput{
				Id:   uuid.New(),
				Name: "Sakino Yui",
			}
		)

		mock.ExpectQuery("SELECT id, name FROM user_master WHERE phone_number = (.+)").
			WithArgs(input.PhoneNumber).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(expectedOutput.Id, expectedOutput.Name))

		output, err := repo.GetUserByPhoneNumber(ctx, input)
		assert.Equal(t, expectedOutput, output)
		assert.Nil(t, err)
	})

	t.Run("query row context returns no rows", func(t *testing.T) {
		var (
			input = GetUserByPhoneNumberInput{
				PhoneNumber: "+628159972915",
			}
		)

		mock.ExpectQuery("SELECT id, name FROM user_master WHERE phone_number = (.+)").
			WithArgs(input.PhoneNumber).WillReturnError(sql.ErrNoRows)

		output, err := repo.GetUserByPhoneNumber(ctx, input)
		assert.EqualError(t, common.ErrUserNotFound, err.Error())
		assert.Empty(t, output)
	})

	t.Run("query row context returns other error", func(t *testing.T) {
		var (
			input = GetUserByPhoneNumberInput{
				PhoneNumber: "+628159972915",
			}
		)

		mock.ExpectQuery("SELECT id, name FROM user_master WHERE phone_number = (.+)").
			WithArgs(input.PhoneNumber).WillReturnError(errors.New("error"))

		output, err := repo.GetUserByPhoneNumber(ctx, input)
		assert.EqualError(t, err, "error")
		assert.Empty(t, output)
	})
}

func TestRepository_InsertUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	ctx := context.Background()
	repo := &Repository{Db: db}

	t.Run("positive", func(t *testing.T) {
		var (
			input = InsertUserInput{
				Id:          uuid.New(),
				Name:        "Sakino Yui",
				PhoneNumber: "+6285320993",
				Password:    "polarBearYui!",
			}
		)

		mock.ExpectExec("INSERT INTO user_master (.+)").
			WithArgs(input.Id, input.PhoneNumber, input.Name, input.Password).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.InsertUser(ctx, input)
		assert.Nil(t, err)
	})

	t.Run("exec context returns error", func(t *testing.T) {
		var (
			input = InsertUserInput{
				Id:          uuid.New(),
				Name:        "Sakino Yui",
				PhoneNumber: "+6285320993",
				Password:    "polarBearYui!",
			}
		)

		mock.ExpectExec("INSERT INTO user_master (.+)").
			WithArgs(input.Id, input.PhoneNumber, input.Name, input.Password).
			WillReturnError(errors.New("error"))

		err := repo.InsertUser(ctx, input)
		assert.EqualError(t, err, "error")
	})
}
