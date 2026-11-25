package identity

import (
	"time"

	"github.com/google/uuid"
)

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
