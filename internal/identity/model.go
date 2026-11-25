package identity

import (
	"regexp"
	"time"

	"github.com/georgifotev1/remy/internal/platform/apperr"
	"github.com/georgifotev1/remy/internal/platform/validator"
	"github.com/google/uuid"
)

var slugRX = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type User struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	Created_At time.Time `json:"created_at"`
	Updated_AT time.Time `json:"updated_at"`
}

type Tenant struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	Created_At time.Time `json:"created_at"`
	Updated_AT time.Time `json:"updated_at"`
}

type SignupRequest struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"user"`
	Tenant struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"tenant"`
}

func (r *SignupRequest) Validate() error {
	v := validator.New()

	// user.*
	v.Check(validator.NotBlank(r.User.Email), "user.email", "must be provided")
	v.Check(validator.MaxChars(r.User.Email, 254), "user.email", "must not be more than 254 characters")
	v.Check(validator.Matches(r.User.Email, validator.EmailRX), "user.email", "must be a valid email address")
	v.Check(validator.MinChars(r.User.Password, 8), "user.password", "must be at least 8 characters")
	v.Check(validator.NotBlank(r.User.Name), "user.name", "must be provided")
	v.Check(validator.MaxChars(r.User.Name, 100), "user.name", "must not be more than 100 characters")

	// tenant.*
	v.Check(validator.NotBlank(r.Tenant.Name), "tenant.name", "must be provided")
	v.Check(validator.MaxChars(r.Tenant.Name, 100), "tenant.name", "must not be more than 100 characters")
	v.Check(validator.NotBlank(r.Tenant.Slug), "tenant.slug", "must be provided")
	v.Check(validator.MinChars(r.Tenant.Slug, 3), "tenant.slug", "must be at least 3 characters")
	v.Check(validator.MaxChars(r.Tenant.Slug, 40), "tenant.slug", "must not be more than 40 characters")
	v.Check(validator.Matches(r.Tenant.Slug, slugRX), "tenant.slug", "may only contain lowercase letters, numbers and single hyphens")

	if !v.Valid() {
		return &apperr.ValidationError{Fields: v.Errors}
	}
	return nil
}
