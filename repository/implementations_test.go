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
	expectedQuery := "SELECT um.id, um.name, um.password_hash, ul.successful_login FROM user_master um " +
		"LEFT JOIN user_login ul ON um.id = ul.user_id WHERE um.phone_number = (.+)"

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

		mock.ExpectQuery(expectedQuery).
			WithArgs(input.PhoneNumber).WillReturnRows(sqlmock.NewRows([]string{"id", "name",
			"password", "successful_login"}).AddRow(expectedOutput.Id, expectedOutput.Name,
			expectedOutput.Password, expectedOutput.NumOfSuccessfulLogin))

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

		mock.ExpectQuery(expectedQuery).
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

		mock.ExpectQuery(expectedQuery).
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

func TestRepository_UpsertUserLogin(t *testing.T) {
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
	expectedQuery := "INSERT INTO user_login (.+)"

	t.Run("positive", func(t *testing.T) {
		var (
			input = UpsertUserLoginInput{
				UserId:               uuid.New(),
				NumOfSuccessfulLogin: 1,
			}
		)

		mock.ExpectExec(expectedQuery).
			WithArgs(input.UserId, input.NumOfSuccessfulLogin).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpsertUserLogin(ctx, input)
		assert.Nil(t, err)
	})

	t.Run("exec context returns error", func(t *testing.T) {
		var (
			input = UpsertUserLoginInput{
				UserId:               uuid.New(),
				NumOfSuccessfulLogin: 1,
			}
		)

		mock.ExpectExec(expectedQuery).
			WithArgs(input.UserId, input.NumOfSuccessfulLogin).
			WillReturnError(errors.New("error"))

		err := repo.UpsertUserLogin(ctx, input)
		assert.EqualError(t, err, "error")
	})
}

func TestRepository_GetUserById(t *testing.T) {
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
	expectedQuery := "SELECT id, name, phone_number FROM user_master WHERE id = (.+)"

	t.Run("positive", func(t *testing.T) {
		var (
			id    = uuid.New()
			input = GetUserByIdInput{
				Id: id.String(),
			}

			expectedOutput = GetUserByIdOutput{
				Id:          id,
				Name:        "Sakino Yui",
				PhoneNumber: "+6287341234234",
			}
		)

		mock.ExpectQuery(expectedQuery).
			WithArgs(input.Id).WillReturnRows(sqlmock.NewRows([]string{"id", "name",
			"phone_number"}).AddRow(expectedOutput.Id, expectedOutput.Name, expectedOutput.PhoneNumber))

		output, err := repo.GetUserById(ctx, input)
		assert.Equal(t, expectedOutput, output)
		assert.Nil(t, err)
	})

	t.Run("query row context returns no rows", func(t *testing.T) {
		var (
			id    = uuid.New()
			input = GetUserByIdInput{
				Id: id.String(),
			}
		)

		mock.ExpectQuery(expectedQuery).
			WithArgs(input.Id).WillReturnError(sql.ErrNoRows)

		output, err := repo.GetUserById(ctx, input)
		assert.EqualError(t, common.ErrUserNotFound, err.Error())
		assert.Empty(t, output)
	})

	t.Run("query row context returns other error", func(t *testing.T) {
		var (
			id    = uuid.New()
			input = GetUserByIdInput{
				Id: id.String(),
			}
		)

		mock.ExpectQuery(expectedQuery).
			WithArgs(input.Id).WillReturnError(errors.New("error"))

		output, err := repo.GetUserById(ctx, input)
		assert.EqualError(t, err, "error")
		assert.Empty(t, output)
	})
}

func TestRepository_UpdateUser(t *testing.T) {
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
	expectedQuery := "UPDATE user_master (.+)"

	t.Run("positive", func(t *testing.T) {
		var (
			input = UpdateUserInput{
				Id:          uuid.New().String(),
				PhoneNumber: "+628787878",
				Name:        "Ruru's Mirapas",
			}
		)

		mock.ExpectExec(expectedQuery).
			WithArgs(input.Id, input.PhoneNumber, input.Name).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateUser(ctx, input)
		assert.Nil(t, err)
	})

	t.Run("exec context returns error", func(t *testing.T) {
		var (
			input = UpdateUserInput{
				Id:          uuid.New().String(),
				PhoneNumber: "+628787878",
				Name:        "Ruru's Mirapas",
			}
		)

		mock.ExpectExec(expectedQuery).
			WithArgs(input.Id, input.PhoneNumber, input.Name).
			WillReturnError(errors.New("error"))

		err := repo.UpdateUser(ctx, input)
		assert.EqualError(t, err, "error")
	})
}
