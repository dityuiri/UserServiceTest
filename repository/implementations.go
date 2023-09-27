package repository

import (
	"context"
	"database/sql"

	"github.com/dityuiri/UserServiceTest/common"
)

func (r *Repository) GetUserByPhoneNumber(ctx context.Context, input GetUserByPhoneNumberInput) (output GetUserByPhoneNumberOutput, err error) {
	var query = `
	SELECT um.id, um.name, um.password, ul.successful_login 
	FROM user_master um 
	INNER JOIN user_login ul ON um.id = ul.user_id
	WHERE um.phone_number = $1`

	err = r.Db.QueryRowContext(ctx, query, input.PhoneNumber).Scan(&output.Id, &output.Name, &output.Password, &output.NumOfSuccessfulLogin)
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
		INSERT INTO user_master
		    (id, phone_number, name, password_hash)
		VALUES
			($1, $2, $3, $4)
	`

	_, err = r.Db.ExecContext(ctx, query, input.Id, input.PhoneNumber, input.Name, input.Password)
	return
}

func (r *Repository) UpdateUserLogin(ctx context.Context, input UpdateUserLoginInput) (err error) {
	var query = `
		UPDATE user_login
		SET
			successful_login = $2, last_login_at = NOW()
		WHERE
			user_id = $1
	`

	_, err = r.Db.ExecContext(ctx, query, input.UserId, input.NumOfSuccessfulLogin)
	return
}
