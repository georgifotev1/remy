package main

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/georgifotev1/remy/internal/data"
	"github.com/georgifotev1/remy/internal/password"
	"github.com/georgifotev1/remy/internal/validator"
)

var slugRX = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
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

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(validator.NotBlank(input.User.Email), "user.email", "must be provided")
	v.Check(validator.MaxChars(input.User.Email, 254), "user.email", "must not be more than 254 characters")
	v.Check(validator.Matches(input.User.Email, validator.EmailRX), "user.email", "must be a valid email address")
	v.Check(validator.MinChars(input.User.Password, 8), "user.password", "must be at least 8 characters")
	v.Check(validator.NotBlank(input.User.Name), "user.name", "must be provided")
	v.Check(validator.MaxChars(input.User.Name, 100), "user.name", "must not be more than 100 characters")

	v.Check(validator.NotBlank(input.Tenant.Name), "tenant.name", "must be provided")
	v.Check(validator.MaxChars(input.Tenant.Name, 100), "tenant.name", "must not be more than 100 characters")
	v.Check(validator.NotBlank(input.Tenant.Slug), "tenant.slug", "must be provided")
	v.Check(validator.MinChars(input.Tenant.Slug, 3), "tenant.slug", "must be at least 3 characters")
	v.Check(validator.MaxChars(input.Tenant.Slug, 40), "tenant.slug", "must not be more than 40 characters")
	v.Check(validator.Matches(input.Tenant.Slug, slugRX), "tenant.slug", "may only contain lowercase letters, numbers and single hyphens")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	hash, err := password.Hash(input.User.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:         input.User.Name,
		Email:        input.User.Email,
		PasswordHash: hash,
	}

	tenant := &data.Tenant{
		Name: input.Tenant.Name,
		Slug: input.Tenant.Slug,
	}

	session, err := data.NewSession(sessionTTL)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Signup(r.Context(), user, tenant, session)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("user.email", "is already taken")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicateSlug):
			v.AddError("tenant.slug", "is already taken")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.setSessionCookie(w, session)

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user, "tenant": tenant}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
