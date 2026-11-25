package main

import (
	"context"
	"errors"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Error("internal error",
		"err", err,
		"method", r.Method,
		"uri", r.URL.RequestURI(),
		"request_id", contextGetRequestID(r),
	)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	if err := app.writeJSON(w, status, env, nil); err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		// Client went away; nobody is listening for a response.
		app.logger.Info("request cancelled",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)
		return
	case errors.Is(err, context.DeadlineExceeded):
		app.logger.Warn("request timed out",
			"err", err,
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)
		app.errorResponse(w, r, http.StatusServiceUnavailable, "the request timed out, please try again")
		return
	}

	app.logError(r, err)
	app.errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusUnauthorized, "invalid email or password")
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errs map[string]string) {
	if err := app.writeJSON(w, http.StatusUnprocessableEntity, envelope{"errors": errs}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
