package user

import (
	"net/mail"
	"regexp"

	"github.com/ahmadnaufal/openidea-segokuning/internal/config"
)

func (r *RegisterUserRequest) Validate() error {
	var validationErrs config.ValidationErrors

	// validate credential type
	if r.CredentialType == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "credentialType",
			Message: "required",
		})
	} else if r.CredentialType != "email" && r.CredentialType != "phone" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "credentialType",
			Message: "unaccepted value",
		})
	}

	// validate credential value
	if r.CredentialValue == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "credentialValue",
			Message: "required",
		})
	} else {
		if r.CredentialType == "email" {
			email := r.CredentialValue

			// do email validation
			if _, err := mail.ParseAddress(email); err != nil {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "is not a valid email format",
				})
			}

			if len(r.CredentialValue) < 5 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is less than 5",
				})
			}

			if len(r.CredentialValue) > 30 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is more than 30",
				})
			}
		} else if r.CredentialType == "phone" {
			phone := r.CredentialValue

			// do email validation
			pat := regexp.MustCompile(`^\+(\d)+$`)
			if !pat.MatchString(phone) {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "is not a valid phone format",
				})
			}

			if len(r.CredentialValue) < 7 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is less than 7",
				})
			}

			if len(r.CredentialValue) > 13 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is more than 13",
				})
			}
		}
	}

	// validate name
	if r.Name == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "name",
			Message: "required",
		})
	} else {
		if len(r.Name) < 5 {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "name",
				Message: "length is less than 5",
			})
		}
		if len(r.Name) > 50 {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "name",
				Message: "length is more than 50",
			})
		}
	}

	// validate password
	if r.Password == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "password",
			Message: "required",
		})
	} else {
		if len(r.Password) < 5 {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "password",
				Message: "length is less than 5",
			})
		}
		if len(r.Password) > 15 {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "password",
				Message: "length is more than 15",
			})
		}
	}

	if validationErrs == nil {
		return nil
	} else {
		return &validationErrs
	}
}

func (r *AuthenticateRequest) Validate() error {
	var validationErrs config.ValidationErrors

	// validate credential type
	if r.CredentialType == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "credentialType",
			Message: "required",
		})
	} else if r.CredentialType != "email" && r.CredentialType != "phone" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "credentialType",
			Message: "unaccepted value",
		})
	}

	// validate credential value
	if r.CredentialValue == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "credentialValue",
			Message: "required",
		})
	} else {
		if r.CredentialType == "email" {
			email := r.CredentialValue

			// do email validation
			if _, err := mail.ParseAddress(email); err != nil {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "is not a valid email format",
				})
			}

			if len(r.CredentialValue) < 5 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is less than 5",
				})
			}

			if len(r.CredentialValue) > 30 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is more than 30",
				})
			}
		} else if r.CredentialType == "phone" {
			phone := r.CredentialValue

			// do email validation
			pat := regexp.MustCompile(`^\+(\d)+$`)
			if !pat.MatchString(phone) {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "is not a valid phone format",
				})
			}

			if len(r.CredentialValue) < 7 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is less than 7",
				})
			}

			if len(r.CredentialValue) > 13 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is more than 13",
				})
			}
		}
	}

	// validate password
	if r.Password == "" {
		validationErrs = append(validationErrs, config.ValidationError{
			Field:   "password",
			Message: "required",
		})
	} else {
		if len(r.Password) < 5 {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "password",
				Message: "length is less than 5",
			})
		}
		if len(r.Password) > 15 {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "password",
				Message: "length is more than 15",
			})
		}
	}

	if validationErrs == nil {
		return nil
	} else {
		return &validationErrs
	}
}

func (r *LinkCredentialRequest) Validate() error {
	var validationErrs config.ValidationErrors

	// validate credential type
	if r.CredentialType == "email" {
		if r.Email == "" {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "email",
				Message: "required",
			})
		} else {
			// do email validation
			if _, err := mail.ParseAddress(r.Email); err != nil {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "is not a valid email format",
				})
			}

			if len(r.Email) < 5 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is less than 5",
				})
			}

			if len(r.Email) > 30 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is more than 30",
				})
			}
		}
	} else if r.CredentialType == "phone" {
		if r.Phone == "" {
			validationErrs = append(validationErrs, config.ValidationError{
				Field:   "phone",
				Message: "required",
			})
		} else {
			// do phone validation
			pat := regexp.MustCompile(`^\+(\d)+$`)
			if !pat.MatchString(r.Phone) {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "is not a valid phone format",
				})
			}

			if len(r.Phone) < 7 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is less than 7",
				})
			}

			if len(r.Phone) > 13 {
				validationErrs = append(validationErrs, config.ValidationError{
					Field:   "credentialValue",
					Message: "length is more than 13",
				})
			}
		}
	}

	if validationErrs == nil {
		return nil
	} else {
		return &validationErrs
	}
}
