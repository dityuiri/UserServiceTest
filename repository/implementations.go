package repository

import (
	"context"
	"database/sql"

	"github.com/dityuiri/UserServiceTest/common"
)

func (r *Repository) GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT name FROM test WHERE id = $1", input.Id).Scan(&output.Name)
	if err != nil {
		return
	}
	return
}

func (r *Repository) GetUserByPhoneNumber(ctx context.Context, input GetUserByPhoneNumberInput) (output GetUserByPhoneNumberOutput, err error) {
	var query = `SELECT id, name FROM user WHERE phone_number = $1`

	err = r.Db.QueryRowContext(ctx, query, input.PhoneNumber).Scan(&output.Id, &output.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return output, common.ErrUserNotFound
		}

		return
	}

	return
}

func (r *Repository) InsertUser(ctx context.Context, input InsertUserInput) (err error) {
	var query = `
		INSERT INTO user 
		    (id, phone_number, name, password_hash, salt)
		VALUES
			($1, $2, $3, $4, $5)
	`

	_, err = r.Db.ExecContext(ctx, query, input.Id, input.PhoneNumber, input.Name, input.Password, input.Salt)
	return
}
