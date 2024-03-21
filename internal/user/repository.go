package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return UserRepo{db: db}
}

func (r UserRepo) CreateUser(ctx context.Context, user User) error {
	query := `
		INSERT INTO users
			(id, email, phone, name, password)
		VALUES
			(:id, :email, :phone, :name, :password)
	`

	updatedQuery, args, err := sqlx.Named(query, user)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return err
	}

	return nil
}

func (r UserRepo) GetUserByCredential(ctx context.Context, credentialType, credentialValue string) (User, error) {
	var result User

	rawQuery := `
		SELECT
			id,
			email,
			phone,
			name,
			password
		FROM
			users
		WHERE
			%s = $1
		LIMIT 1
	`

	var query string
	if credentialType == "email" {
		query = fmt.Sprintf(rawQuery, "email")
	} else {
		query = fmt.Sprintf(rawQuery, "phone")
	}

	err := r.db.GetContext(ctx, &result, query, credentialValue)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r UserRepo) GetUserByID(ctx context.Context, id string) (User, error) {
	var result User

	query := `
		SELECT
			id,
			email,
			phone,
			name,
			password,
			image_url
		FROM
			users
		WHERE
			id = $1
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &result, query, id)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r UserRepo) UpdateUser(ctx context.Context, tx *sql.Tx, user User) error {
	query := `
		UPDATE
			users
		SET
			email = :email,
			phone = :phone,
			image_url = :image_url,
			name = :name
		WHERE
			id = :id
	`

	updatedQuery, args, err := sqlx.Named(query, user)
	if err != nil {
		return err
	}

	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	} else {
		_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	}
	if err != nil {
		return err
	}

	return nil
}
