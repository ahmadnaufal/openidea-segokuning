package user

type RegisterUserRequest struct {
	CredentialType  string `json:"username" validate:"required,min=5,max=15"`
	CredentialValue string `json:"credential_value" validate:"required,min=5,max=15"`
	Name            string `json:"name" validate:"required,min=5,max=50"`
	Password        string `json:"password" validate:"required,min=5,max=15"`
}

type AuthenticateRequest struct {
	CredentialType  string `json:"username" validate:"required,min=5,max=15"`
	CredentialValue string `json:"credential_value" validate:"required,min=5,max=15"`
	Password        string `json:"password" validate:"required,min=5,max=15"`
}

type LinkCredentialRequest struct {
	Email string `json:"email" validate:"required_if=CredentialType email"`
	Phone string `json:"phone" validate:"required_if=CredentialType phone"`

	CredentialType string
}

type User struct {
	ID       string `db:"id"`
	Email    string `db:"email"`
	Phone    string `db:"phone"`
	Name     string `db:"name"`
	Password string `db:"password"`
	ImageURL string `db:"image_url"`
}
