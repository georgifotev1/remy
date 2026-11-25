package main

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/georgifotev1/remy/internal/data"
	"github.com/georgifotev1/remy/internal/password"
	"github.com/georgifotev1/remy/internal/validator"
)

const (
	sessionCookieName = "session"
	sessionTTL        = 30 * 24 * time.Hour
)

var dummyPasswordHash = sync.OnceValue(func() string {
	hash, err := password.Hash("dummy-password-for-timing")
	if err != nil {
		panic(err)
	}
	return hash
})

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(validator.NotBlank(input.Email), "email", "must be provided")
	v.Check(validator.NotBlank(input.Password), "password", "must be provided")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(r.Context(), input.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			password.Verify(input.Password, dummyPasswordHash())
			app.invalidCredentialsResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	ok, err := password.Verify(input.Password, user.PasswordHash)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !ok {
		app.invalidCredentialsResponse(w, r)
		return
	}

	session, err := data.NewSession(sessionTTL)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	session.UserID = user.ID

	if err := app.models.Sessions.Insert(r.Context(), session); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.setSessionCookie(w, session)

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		if err := app.models.Sessions.DeleteByToken(r.Context(), cookie.Value); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	app.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) setSessionCookie(w http.ResponseWriter, s *data.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    s.Plaintext,
		Path:     "/",
		Expires:  s.Expiry,
		MaxAge:   int(time.Until(s.Expiry).Seconds()),
		HttpOnly: true,
		Secure:   app.config.isProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

func (app *application) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.config.isProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}
