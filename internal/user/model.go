package user

import (
	"database/sql"
	"time"
)

type RegisterUserRequest struct {
	CredentialType  string `json:"credentialType" validate:"required,oneof=email phone"`
	CredentialValue string `json:"credentialValue" validate:"required,min=5,max=15"`
	Name            string `json:"name" validate:"required,min=5,max=50"`
	Password        string `json:"password" validate:"required,min=5,max=15"`
}

type AuthenticateRequest struct {
	CredentialType  string `json:"credentialType" validate:"required,oneof=email phone"`
	CredentialValue string `json:"credentialValue" validate:"required,min=5,max=15"`
	Password        string `json:"password" validate:"required,min=5,max=15"`
}

type LinkCredentialRequest struct {
	Email string `json:"email" validate:"required_if=CredentialType email,email,min=7,max=50"`
	Phone string `json:"phone" validate:"required_if=CredentialType phone,startswith=+,min=7,max=13"`

	CredentialType string
	UserID         string
}

type UpdateUserRequest struct {
	ImageURL string `json:"imageUrl" validate:"required,url"`
	Name     string `json:"name" validate:"required,min=5,max=50"`

	UserID string
}

type User struct {
	ID        string         `db:"id"`
	Email     sql.NullString `db:"email"`
	Phone     sql.NullString `db:"phone"`
	Name      string         `db:"name"`
	Password  string         `db:"password"`
	ImageURL  sql.NullString `db:"image_url"`
	CreatedAt time.Time      `db:"created_at"`
}
