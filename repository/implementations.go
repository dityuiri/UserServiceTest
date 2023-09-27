package repository

import (
	"context"
	"database/sql"

	"github.com/dityuiri/UserServiceTest/common"
)

func (r *Repository) GetUserByPhoneNumber(ctx context.Context, input GetUserByPhoneNumberInput) (output GetUserByPhoneNumberOutput, err error) {
	var query = `
	SELECT um.id, um.name, um.password_hash, ul.successful_login 
	FROM user_master um 
	LEFT JOIN user_login ul ON um.id = ul.user_id
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

func (r *Repository) UpsertUserLogin(ctx context.Context, input UpsertUserLoginInput) (err error) {
	var query = `
		INSERT INTO user_login (user_id, successful_login, last_login_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id)
		DO UPDATE
		SET successful_login = EXCLUDED.successful_login, last_login_at = EXCLUDED.last_login_at
	`

	_, err = r.Db.ExecContext(ctx, query, input.UserId, input.NumOfSuccessfulLogin)
	return
}
